//go:build unit

package application

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/druckstation"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/produkt"
	"github.com/nicograef/jotti/backend/domain/steuer"
	"github.com/nicograef/jotti/backend/repository/druckauftrag_repo"
	"github.com/nicograef/jotti/backend/repository/kassensitzungen_repo"
	"github.com/nicograef/jotti/backend/repository/produkt_repo"
)

const testKassensitzungNr = 1

var testOpenKS = &kasse.Kassensitzung{
	ZNr:    testKassensitzungNr,
	Status: kasse.KassensitzungOffen,
}

var testProduct = produkt.Produkt{
	ID:         1,
	Name:       "Cola",
	Kategorie:  produkt.GetraenkKategorie,
	Steuersatz: steuer.RegelSteuersatz,
	Status:     produkt.ActiveStatus,
}

var testVariant = produkt.Variante{
	ID:         1,
	Name:       "Cola 0,5l",
	PreisCents: 350,
	Status:     produkt.ActiveStatus,
}

// spyEventRepo records WriteEvent calls so tests can assert the exact write contract
// (single event, stream type, version, subject).
type spyEventRepo struct {
	written        []writtenEvent
	druckauftraege []druckauftrag_repo.NeuerDruckauftrag
	maxVersion     int
	writeErr       error
	streamEvents   []event.Event
	readErr        error
}

type writtenEvent struct {
	event           event.Event
	streamType      kasse.StreamType
	kassensitzungNr int
}

func (s *spyEventRepo) GetMaxVersion(_ context.Context, _ string) (int, error) {
	return s.maxVersion, nil
}

func (s *spyEventRepo) ReadEventsBySubject(_ context.Context, _ string) ([]event.Event, error) {
	return s.streamEvents, s.readErr
}

func (s *spyEventRepo) WriteEvent(_ context.Context, e event.Event, streamType kasse.StreamType, kassensitzungNr int) (int, error) {
	if s.writeErr != nil {
		return 0, s.writeErr
	}
	s.written = append(s.written, writtenEvent{event: e, streamType: streamType, kassensitzungNr: kassensitzungNr})
	return len(s.written), nil
}

func (s *spyEventRepo) WriteEventWithDruckauftraege(_ context.Context, e event.Event, streamType kasse.StreamType, kassensitzungNr int, buildAuftraege func(event.Event) []druckauftrag_repo.NeuerDruckauftrag) (int, error) {
	id, err := s.WriteEvent(context.Background(), e, streamType, kassensitzungNr)
	if err != nil {
		return 0, err
	}

	e.ID = id
	s.druckauftraege = append(s.druckauftraege, buildAuftraege(e)...)
	return id, nil
}

type mockDruckstationRepo struct {
	konfig map[string]druckstation.Druckstation
	err    error
}

func (m *mockDruckstationRepo) GetKonfigurierteDruckstationen(_ context.Context) (map[string]druckstation.Druckstation, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.konfig, nil
}

func newProductMock() productRepo {
	productMock := produkt_repo.NewMock([]produkt.Produkt{testProduct}, nil)
	productMock.AddVariant(testProduct.ID, testVariant)
	return productMock
}

func newCommand(eventRepo eventRepo, ks *kasse.Kassensitzung) Command {
	return Command{
		EventRepo:           eventRepo,
		ProductRepo:         newProductMock(),
		KassensitzungenRepo: kassensitzungen_repo.NewMock(ks, nil),
	}
}

func newCommandWithDruckstationen(eventRepo eventRepo, ks *kasse.Kassensitzung, stationen map[string]druckstation.Druckstation) Command {
	return Command{
		EventRepo:           eventRepo,
		ProductRepo:         newProductMock(),
		KassensitzungenRepo: kassensitzungen_repo.NewMock(ks, nil),
		DruckstationRepo:    &mockDruckstationRepo{konfig: stationen},
	}
}

var testInputs = []VerkaufPositionInput{
	{ProduktID: testProduct.ID, VarianteID: testVariant.ID, Menge: 2},
}

func TestDirektverkaufTaetigen_KasseNichtGeoeffnet(t *testing.T) {
	command := newCommand(&spyEventRepo{}, nil)

	err := command.DirektverkaufTaetigen(context.Background(), 1, "Test User", testInputs, "")
	if err != ErrKasseNichtGeoeffnet {
		t.Fatalf("expected ErrKasseNichtGeoeffnet, got %v", err)
	}
}

func TestDirektverkaufTaetigen_WritesSingleEvent(t *testing.T) {
	spy := &spyEventRepo{}
	command := newCommand(spy, testOpenKS)

	err := command.DirektverkaufTaetigen(context.Background(), 1, "Test User", testInputs, "Direktverkauf")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(spy.written) != 1 {
		t.Fatalf("expected exactly 1 written event, got %d", len(spy.written))
	}

	got := spy.written[0]
	if got.streamType != kasse.StreamTypeDirektverkauf {
		t.Errorf("expected stream type %s, got %s", kasse.StreamTypeDirektverkauf, got.streamType)
	}
	if got.event.Version != 1 {
		t.Errorf("expected version 1, got %d", got.event.Version)
	}
	if got.kassensitzungNr != testKassensitzungNr {
		t.Errorf("expected kassensitzungNr %d, got %d", testKassensitzungNr, got.kassensitzungNr)
	}

	verkaufID, err := kasse.ParseVerkaufIDFromSubject(got.event.Subject)
	if err != nil {
		t.Fatalf("expected a direktverkauf subject, got %q (%v)", got.event.Subject, err)
	}
	if verkaufID == "" {
		t.Error("expected a non-empty verkaufId in the subject")
	}

	var data struct {
		GesamtbetragCents int `json:"gesamtbetragCents"`
	}
	if err := json.Unmarshal(got.event.Data, &data); err != nil {
		t.Fatalf("failed to unmarshal event data: %v", err)
	}
	if data.GesamtbetragCents != testVariant.PreisCents*2 {
		t.Errorf("expected gesamtbetragCents %d, got %d", testVariant.PreisCents*2, data.GesamtbetragCents)
	}
}

func TestDirektverkaufTaetigen_ProduktNotFound(t *testing.T) {
	spy := &spyEventRepo{}
	command := Command{
		EventRepo:           spy,
		ProductRepo:         produkt_repo.NewMock([]produkt.Produkt{}, nil),
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
	}

	err := command.DirektverkaufTaetigen(context.Background(), 1, "Test User", testInputs, "")
	if err != ErrProduktNotFound {
		t.Fatalf("expected ErrProduktNotFound, got %v", err)
	}
	if len(spy.written) != 0 {
		t.Fatalf("expected no event written, got %d", len(spy.written))
	}
}

func TestDirektverkaufTaetigen_Conflict(t *testing.T) {
	spy := &spyEventRepo{writeErr: db.ErrAlreadyExists}
	command := newCommand(spy, testOpenKS)

	err := command.DirektverkaufTaetigen(context.Background(), 1, "Test User", testInputs, "")
	if err != ErrConflict {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

func TestDirektverkaufTaetigen_DeadlockMapsToConflict(t *testing.T) {
	spy := &spyEventRepo{writeErr: db.ErrConflict}
	command := newCommand(spy, testOpenKS)

	err := command.DirektverkaufTaetigen(context.Background(), 1, "Test User", testInputs, "")
	if err != ErrConflict {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

func TestDirektverkaufTaetigen_AbholbonModeQueuesExactlyOneAuftrag(t *testing.T) {
	spy := &spyEventRepo{}
	command := newCommandWithDruckstationen(
		spy,
		testOpenKS,
		map[string]druckstation.Druckstation{
			"abholbon": {DruckerIP: "192.168.1.77", Bonmodus: "pro_bestellung"},
		},
	)

	err := command.DirektverkaufTaetigen(context.Background(), 1, "Test User", testInputs, "Direktverkauf")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(spy.druckauftraege) != 1 {
		t.Fatalf("expected exactly 1 druckauftrag, got %d", len(spy.druckauftraege))
	}

	auftrag := spy.druckauftraege[0]
	if auftrag.ZielIP != "192.168.1.77" {
		t.Errorf("expected ZielIP 192.168.1.77, got %s", auftrag.ZielIP)
	}
	if auftrag.BonArt != "arbeitsbon" {
		t.Errorf("expected BonArt arbeitsbon, got %s", auftrag.BonArt)
	}
	if auftrag.Referenz != "direktverkauf-getaetigt:1" {
		t.Errorf("expected referenz direktverkauf-getaetigt:1, got %s", auftrag.Referenz)
	}
	if auftrag.Payload == "" {
		t.Error("expected non-empty payload")
	}
}

// getaetigtEvent builds a real direktverkauf-getaetigt:v1 event with a single position and returns
// the event, its verkaufId, and the server-generated positionId for use in storno tests.
func getaetigtEvent(t *testing.T, einzelpreis, menge int) (event.Event, string, string) {
	t.Helper()
	verkaufID := uuid.New().String()
	subject := kasse.DirektverkaufSubject(testKassensitzungNr, verkaufID)
	evt, err := kasse.NewDirektverkaufGetaetigtEvent(subject, verkaufID, 1, "User", []kasse.Position{
		{VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", Einzelpreis: einzelpreis, Menge: menge},
	}, "")
	if err != nil {
		t.Fatalf("failed to create getaetigt event: %v", err)
	}
	// Ein direktverkauf-getaetigt ist immer version = 1 seines Streams; der Storno-Pfad
	// leitet seine erwartete Version aus dem Replay ab (OCC gegen den gelesenen Zustand).
	evt.Version = 1

	var data struct {
		Positionen []struct {
			PositionID string `json:"positionId"`
		} `json:"positionen"`
	}
	if err := json.Unmarshal(evt.Data, &data); err != nil {
		t.Fatalf("failed to unmarshal getaetigt data: %v", err)
	}
	return evt, verkaufID, data.Positionen[0].PositionID
}

func TestDirektverkaufStornieren_KasseNichtGeoeffnet(t *testing.T) {
	command := newCommand(&spyEventRepo{}, nil)

	err := command.DirektverkaufStornieren(context.Background(), 2, "Leitung", uuid.New().String(), []kasse.PositionRef{{PositionID: uuid.New().String(), Menge: 1}}, "Rückgabe")
	if err != ErrKasseNichtGeoeffnet {
		t.Fatalf("expected ErrKasseNichtGeoeffnet, got %v", err)
	}
}

func TestDirektverkaufStornieren_VerkaufNichtGefunden(t *testing.T) {
	spy := &spyEventRepo{streamEvents: nil}
	command := newCommand(spy, testOpenKS)

	err := command.DirektverkaufStornieren(context.Background(), 2, "Leitung", uuid.New().String(), []kasse.PositionRef{{PositionID: uuid.New().String(), Menge: 1}}, "Rückgabe")
	if err != ErrVerkaufNichtGefunden {
		t.Fatalf("expected ErrVerkaufNichtGefunden, got %v", err)
	}
	if len(spy.written) != 0 {
		t.Fatalf("expected no event written, got %d", len(spy.written))
	}
}

func TestDirektverkaufStornieren_UeberVerfuegbareMenge(t *testing.T) {
	getaetigt, verkaufID, positionID := getaetigtEvent(t, 500, 2)
	spy := &spyEventRepo{maxVersion: 1, streamEvents: []event.Event{getaetigt}}
	command := newCommand(spy, testOpenKS)

	err := command.DirektverkaufStornieren(context.Background(), 2, "Leitung", verkaufID, []kasse.PositionRef{{PositionID: positionID, Menge: 3}}, "Zu viel")
	if err != ErrPositionNichtStornierbar {
		t.Fatalf("expected ErrPositionNichtStornierbar, got %v", err)
	}
	if len(spy.written) != 0 {
		t.Fatalf("expected no event written, got %d", len(spy.written))
	}
}

func TestDirektverkaufStornieren_DuplikatPositionRefs(t *testing.T) {
	// Duplikate sind per se ungültig — auch wenn die Summe der Mengen (1+1) die
	// verfügbare Menge (3) nicht übersteigt.
	for _, menge := range []int{2, 1} {
		getaetigt, verkaufID, positionID := getaetigtEvent(t, 500, 3)
		spy := &spyEventRepo{maxVersion: 1, streamEvents: []event.Event{getaetigt}}
		command := newCommand(spy, testOpenKS)

		refs := []kasse.PositionRef{
			{PositionID: positionID, Menge: menge},
			{PositionID: positionID, Menge: menge},
		}
		err := command.DirektverkaufStornieren(context.Background(), 2, "Leitung", verkaufID, refs, "Duplikat")
		if err != ErrPositionNichtStornierbar {
			t.Fatalf("menge %d: expected ErrPositionNichtStornierbar, got %v", menge, err)
		}
		if len(spy.written) != 0 {
			t.Fatalf("menge %d: expected no event written, got %d", menge, len(spy.written))
		}
	}
}

func TestDirektverkaufStornieren_WritesStornoEventWithNextVersion(t *testing.T) {
	getaetigt, verkaufID, positionID := getaetigtEvent(t, 500, 2)
	spy := &spyEventRepo{maxVersion: 1, streamEvents: []event.Event{getaetigt}}
	command := newCommand(spy, testOpenKS)

	err := command.DirektverkaufStornieren(context.Background(), 2, "Leitung", verkaufID, []kasse.PositionRef{{PositionID: positionID, Menge: 1}}, "Rückgabe")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(spy.written) != 1 {
		t.Fatalf("expected exactly 1 written event, got %d", len(spy.written))
	}

	got := spy.written[0]
	if got.streamType != kasse.StreamTypeDirektverkauf {
		t.Errorf("expected stream type %s, got %s", kasse.StreamTypeDirektverkauf, got.streamType)
	}
	if got.event.Version != 2 {
		t.Errorf("expected version 2 (maxVersion+1), got %d", got.event.Version)
	}
	if got.event.Type != string(kasse.EventTypeDirektverkaufStorniertV1) {
		t.Errorf("expected type %s, got %s", kasse.EventTypeDirektverkaufStorniertV1, got.event.Type)
	}

	var data struct {
		GesamtStornierungCents int `json:"gesamtStornierungCents"`
	}
	if err := json.Unmarshal(got.event.Data, &data); err != nil {
		t.Fatalf("failed to unmarshal storno event data: %v", err)
	}
	if data.GesamtStornierungCents != 500 {
		t.Errorf("expected gesamtStornierungCents 500, got %d", data.GesamtStornierungCents)
	}
}

func TestDirektverkaufStornieren_Conflict(t *testing.T) {
	getaetigt, verkaufID, positionID := getaetigtEvent(t, 500, 2)
	spy := &spyEventRepo{maxVersion: 1, streamEvents: []event.Event{getaetigt}, writeErr: db.ErrAlreadyExists}
	command := newCommand(spy, testOpenKS)

	err := command.DirektverkaufStornieren(context.Background(), 2, "Leitung", verkaufID, []kasse.PositionRef{{PositionID: positionID, Menge: 1}}, "Rückgabe")
	if err != ErrConflict {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}
