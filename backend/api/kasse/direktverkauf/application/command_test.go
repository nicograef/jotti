package application

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/nicograef/jotti/backend/api/kasse/enrichment"
	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/druckstation"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/produkt"
	"github.com/nicograef/jotti/backend/domain/steuer"
	"github.com/nicograef/jotti/backend/repository/kassenjournal_repo"
	"github.com/nicograef/jotti/backend/repository/repotest"
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
	druckauftraege []druckstation.NeuerDruckauftrag
	writeErr       error
	streamEvents   []event.Event
	readErr        error
}

type writtenEvent struct {
	event           event.Event
	streamType      kasse.StreamType
	kassensitzungNr int
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

func (s *spyEventRepo) WriteEventsAtomic(_ context.Context, kassensitzungNr int, events []kassenjournal_repo.NeuesEvent) error {
	if s.writeErr != nil {
		return s.writeErr
	}
	for _, neu := range events {
		id, err := s.WriteEvent(context.Background(), neu.Event, neu.StreamType, kassensitzungNr)
		if err != nil {
			return err
		}
		if neu.BuildAuftraege != nil {
			stored := neu.Event
			stored.ID = id
			s.druckauftraege = append(s.druckauftraege, neu.BuildAuftraege(stored)...)
		}
	}
	return nil
}

// The lookups find nothing: the idempotency path is not exercised in these unit tests.
func (s *spyEventRepo) ReadDirektverkaufEvents(_ context.Context, _ int) ([]event.Event, error) {
	return nil, nil
}

func (s *spyEventRepo) ReadPfandrueckgabeEvents(_ context.Context, _, _ int) ([]event.Event, error) {
	return nil, nil
}

func (s *spyEventRepo) GetDirektverkauf(_ context.Context, _ string) (event.Event, bool, error) {
	return event.Event{}, false, nil
}

func (s *spyEventRepo) GetPfandrueckgabe(_ context.Context, _ string) (event.Event, bool, error) {
	return event.Event{}, false, nil
}

func (s *spyEventRepo) GetPfandrueckgabeStornierung(_ context.Context, _ string) (event.Event, bool, error) {
	return event.Event{}, false, nil
}

type mockDruckstationRepo struct {
	konfig        map[string]druckstation.Druckstation
	zusatzdrucker map[int]druckstation.Zusatzdrucker // keyed by VarianteID
	err           error
}

func (m *mockDruckstationRepo) GetKonfigurierteDruckstationen(_ context.Context) (map[string]druckstation.Druckstation, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.konfig, nil
}

// GetZusatzdruckerJeVariante answers only the requested Varianten, like the repository.
func (m *mockDruckstationRepo) GetZusatzdruckerJeVariante(_ context.Context, varianteIDs []int) (map[int]druckstation.Zusatzdrucker, error) {
	if m.err != nil {
		return nil, m.err
	}
	result := map[int]druckstation.Zusatzdrucker{}
	for _, id := range varianteIDs {
		if zd, ok := m.zusatzdrucker[id]; ok {
			result[id] = zd
		}
	}
	return result, nil
}

func newProductMock() produktRepo {
	productMock := repotest.NewProduktRepo([]produkt.Produkt{testProduct}, nil)
	productMock.AddVariante(testProduct.ID, testVariant)
	return productMock
}

func newCommand(eventRepo eventRepo, ks *kasse.Kassensitzung) Command {
	return Command{
		EventRepo:           eventRepo,
		ProduktRepo:         newProductMock(),
		KassensitzungenRepo: repotest.NewKassensitzungenRepo(ks, nil),
		DruckstationRepo:    &mockDruckstationRepo{},
	}
}

func newCommandWithDruckstationen(eventRepo eventRepo, ks *kasse.Kassensitzung, stationen map[string]druckstation.Druckstation) Command {
	return Command{
		EventRepo:           eventRepo,
		ProduktRepo:         newProductMock(),
		KassensitzungenRepo: repotest.NewKassensitzungenRepo(ks, nil),
		DruckstationRepo:    &mockDruckstationRepo{konfig: stationen},
	}
}

var testInputs = []enrichment.PositionInput{
	{ProduktID: testProduct.ID, VarianteID: testVariant.ID, Menge: 2},
}

func TestDirektverkaufTaetigen_KasseNichtGeoeffnet(t *testing.T) {
	command := newCommand(&spyEventRepo{}, nil)

	err := command.DirektverkaufTaetigen(context.Background(), 1, "Test User", "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", testInputs, "", nil)
	if !errors.Is(err, ErrKasseNichtGeoeffnet) {
		t.Errorf("expected ErrKasseNichtGeoeffnet, got %v", err)
	}
}

func TestDirektverkaufTaetigen_WritesSingleEvent(t *testing.T) {
	spy := &spyEventRepo{}
	command := newCommand(spy, testOpenKS)

	err := command.DirektverkaufTaetigen(context.Background(), 1, "Test User", "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", testInputs, "Direktverkauf", nil)
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

	if want := "kassensitzung-1/direktverkauf-aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"; got.event.Subject != want {
		t.Errorf("expected subject %q, got %q", want, got.event.Subject)
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

// The sale freezes each position's Pfand flag from its product, for the DSFinV-K GV_TYP.
func TestDirektverkaufTaetigen_FriertPfandKennzeichenEin(t *testing.T) {
	pfand := produkt.Produkt{ID: 9, Name: "Pfand", Kategorie: produkt.SonstigesKategorie, Steuersatz: steuer.RegelSteuersatz, Pfand: true, Status: produkt.ActiveStatus}
	produkte := repotest.NewProduktRepo([]produkt.Produkt{testProduct, pfand}, nil)
	produkte.AddVariante(testProduct.ID, testVariant)
	produkte.AddVariante(pfand.ID, produkt.Variante{ID: 19, Name: "Glas", PreisCents: 200, Status: produkt.ActiveStatus})
	spy := &spyEventRepo{}
	command := newCommand(spy, testOpenKS)
	command.ProduktRepo = produkte

	err := command.DirektverkaufTaetigen(context.Background(), 1, "Test User", "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", []enrichment.PositionInput{
		{ProduktID: testProduct.ID, VarianteID: testVariant.ID, Menge: 1},
		{ProduktID: pfand.ID, VarianteID: 19, Menge: 1},
	}, "", nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(spy.written) != 1 || spy.written[0].event.Type != "direktverkauf-getaetigt:v2" {
		t.Fatalf("expected one direktverkauf-getaetigt:v2, got %+v", spy.written)
	}
	var data kasse.DirektverkaufGetaetigtV2Data
	if err := json.Unmarshal(spy.written[0].event.Data, &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	got := map[int]bool{}
	for _, p := range data.Positionen {
		got[p.VarianteID] = p.Pfand
	}
	if len(got) != 2 || got[testVariant.ID] || !got[19] {
		t.Errorf("pfand per variante = %v, want map[1:false 19:true]", got)
	}
}

func TestDirektverkaufTaetigen_SchreibtPositionskommentar(t *testing.T) {
	produkte := repotest.NewProduktRepo([]produkt.Produkt{testProduct}, nil)
	produkte.AddVariante(testProduct.ID, testVariant)
	zweiteVariante := testVariant
	zweiteVariante.ID = 2
	zweiteVariante.Name = "Cola 0,3l"
	produkte.AddVariante(testProduct.ID, zweiteVariante)
	spy := &spyEventRepo{}
	command := newCommand(spy, testOpenKS)
	command.ProduktRepo = produkte

	err := command.DirektverkaufTaetigen(context.Background(), 1, "Test User", "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", []enrichment.PositionInput{
		{ProduktID: testProduct.ID, VarianteID: testVariant.ID, Menge: 2, Kommentar: "ohne Eis"},
		{ProduktID: testProduct.ID, VarianteID: zweiteVariante.ID, Menge: 1},
	}, "", nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(spy.written) != 1 || spy.written[0].event.Type != "direktverkauf-getaetigt:v2" {
		t.Fatalf("expected one direktverkauf-getaetigt:v2, got %+v", spy.written)
	}
	var data struct {
		Positionen []map[string]json.RawMessage `json:"positionen"`
	}
	if err := json.Unmarshal(spy.written[0].event.Data, &data); err != nil {
		t.Fatalf("unmarshal event data: %v", err)
	}
	if got := string(data.Positionen[0]["kommentar"]); got != `"ohne Eis"` {
		t.Errorf("positionen[0].kommentar = %s, want \"ohne Eis\"", got)
	}
	if _, ok := data.Positionen[1]["kommentar"]; ok {
		t.Errorf("positionen[1] carries a kommentar key, want it absent: %v", data.Positionen[1])
	}
}

// Ein deaktivierte (inactive) Variante taucht im Verkaufs-Menü nicht auf, kann
// aber per direktem POST referenziert werden — der Command lehnt sie ab.
func TestDirektverkaufTaetigen_InactiveVariante(t *testing.T) {
	inactiveVariant := produkt.Variante{
		ID:         2,
		Name:       "Cola 0,5l",
		PreisCents: 350,
		Status:     produkt.InactiveStatus,
	}
	productMock := repotest.NewProduktRepo([]produkt.Produkt{testProduct}, nil)
	productMock.AddVariante(testProduct.ID, inactiveVariant)
	spy := &spyEventRepo{}
	command := Command{
		EventRepo:           spy,
		ProduktRepo:         productMock,
		KassensitzungenRepo: repotest.NewKassensitzungenRepo(testOpenKS, nil),
		DruckstationRepo:    &mockDruckstationRepo{},
	}

	inputs := []enrichment.PositionInput{
		{ProduktID: testProduct.ID, VarianteID: inactiveVariant.ID, Menge: 1},
	}

	err := command.DirektverkaufTaetigen(context.Background(), 1, "Test User", "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", inputs, "", nil)
	if !errors.Is(err, enrichment.ErrVarianteNichtAktiv) {
		t.Errorf("expected ErrVarianteNichtAktiv, got %v", err)
	}
	if len(spy.written) != 0 {
		t.Errorf("expected no event written, got %d", len(spy.written))
	}
}

func TestDirektverkaufTaetigen_ProduktNotFound(t *testing.T) {
	spy := &spyEventRepo{}
	command := Command{
		EventRepo:           spy,
		ProduktRepo:         repotest.NewProduktRepo([]produkt.Produkt{}, nil),
		KassensitzungenRepo: repotest.NewKassensitzungenRepo(testOpenKS, nil),
		DruckstationRepo:    &mockDruckstationRepo{},
	}

	err := command.DirektverkaufTaetigen(context.Background(), 1, "Test User", "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", testInputs, "", nil)
	if !errors.Is(err, enrichment.ErrProduktNotFound) {
		t.Errorf("expected ErrProduktNotFound, got %v", err)
	}
	if len(spy.written) != 0 {
		t.Errorf("expected no event written, got %d", len(spy.written))
	}
}

func TestDirektverkaufTaetigen_Conflict(t *testing.T) {
	spy := &spyEventRepo{writeErr: db.ErrAlreadyExists}
	command := newCommand(spy, testOpenKS)

	err := command.DirektverkaufTaetigen(context.Background(), 1, "Test User", "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", testInputs, "", nil)
	if !errors.Is(err, ErrConflict) {
		t.Errorf("expected ErrConflict, got %v", err)
	}
}

func TestDirektverkaufTaetigen_DeadlockMapsToConflict(t *testing.T) {
	spy := &spyEventRepo{writeErr: db.ErrConflict}
	command := newCommand(spy, testOpenKS)

	err := command.DirektverkaufTaetigen(context.Background(), 1, "Test User", "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", testInputs, "", nil)
	if !errors.Is(err, ErrConflict) {
		t.Errorf("expected ErrConflict, got %v", err)
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

	err := command.DirektverkaufTaetigen(context.Background(), 1, "Test User", "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", testInputs, "Direktverkauf", nil)
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

func TestDirektverkaufTaetigen_DrucktAmZusatzdrucker(t *testing.T) {
	spy := &spyEventRepo{}
	command := newCommand(spy, testOpenKS)
	command.DruckstationRepo = &mockDruckstationRepo{
		konfig: map[string]druckstation.Druckstation{
			string(testProduct.Kategorie): {DruckerIP: "192.168.1.50", Bonmodus: "pro_bestellung"},
		},
		zusatzdrucker: map[int]druckstation.Zusatzdrucker{
			testVariant.ID: {ID: 1, Name: "Bar", DruckerIP: "192.168.1.61", Bonmodus: "pro_bestellung"},
		},
	}

	if err := command.DirektverkaufTaetigen(context.Background(), 1, "Test User", "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaab", testInputs, "", nil); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(spy.druckauftraege) != 1 {
		t.Fatalf("expected exactly 1 druckauftrag, got %d", len(spy.druckauftraege))
	}
	if spy.druckauftraege[0].ZielIP != "192.168.1.61" {
		t.Errorf("expected ZielIP of the Zusatzdrucker 192.168.1.61, got %s", spy.druckauftraege[0].ZielIP)
	}
}

// Bonmodus pro_stueck der Abholbon-Station: die zwei Einheiten der Testposition
// ergeben zwei Druckaufträge.
func TestDirektverkaufTaetigen_AbholbonProStueckQueuesEinenAuftragJeEinheit(t *testing.T) {
	spy := &spyEventRepo{}
	command := newCommandWithDruckstationen(
		spy,
		testOpenKS,
		map[string]druckstation.Druckstation{
			"abholbon": {DruckerIP: "192.168.1.77", Bonmodus: "pro_stueck"},
		},
	)

	err := command.DirektverkaufTaetigen(context.Background(), 1, "Test User", "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", testInputs, "Direktverkauf", nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(spy.druckauftraege) != 2 {
		t.Fatalf("expected 2 druckauftraege (one per unit), got %d", len(spy.druckauftraege))
	}
	if spy.druckauftraege[0].Payload != spy.druckauftraege[1].Payload {
		t.Error("expected identical payloads for the two units of one position")
	}
}

// getaetigtEvent builds a real direktverkauf-getaetigt:v2 event with a single position and returns
// the event, its verkaufId, and the server-generated positionId for use in storno tests.
func getaetigtEvent(t *testing.T, einzelpreis, menge int) (event.Event, string, string) {
	t.Helper()
	verkaufID := uuid.New().String()
	subject := kasse.DirektverkaufSubject(testKassensitzungNr, verkaufID)
	evt, err := kasse.NewDirektverkaufGetaetigtEvent(subject, verkaufID, 1, "User", []kasse.Position{
		{VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", EinzelpreisCents: einzelpreis, Menge: menge},
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

	err := command.DirektverkaufStornieren(context.Background(), 2, "Leitung", uuid.NewString(), uuid.New().String(), []kasse.PositionRef{{PositionID: uuid.New().String(), Menge: 1}}, "Rückgabe")
	if !errors.Is(err, ErrKasseNichtGeoeffnet) {
		t.Errorf("expected ErrKasseNichtGeoeffnet, got %v", err)
	}
}

func TestDirektverkaufStornieren_VerkaufNichtGefunden(t *testing.T) {
	spy := &spyEventRepo{streamEvents: nil}
	command := newCommand(spy, testOpenKS)

	err := command.DirektverkaufStornieren(context.Background(), 2, "Leitung", uuid.NewString(), uuid.New().String(), []kasse.PositionRef{{PositionID: uuid.New().String(), Menge: 1}}, "Rückgabe")
	if !errors.Is(err, ErrVerkaufNichtGefunden) {
		t.Errorf("expected ErrVerkaufNichtGefunden, got %v", err)
	}
	if len(spy.written) != 0 {
		t.Errorf("expected no event written, got %d", len(spy.written))
	}
}

func TestDirektverkaufStornieren_UeberVerfuegbareMenge(t *testing.T) {
	getaetigt, verkaufID, positionID := getaetigtEvent(t, 500, 2)
	spy := &spyEventRepo{streamEvents: []event.Event{getaetigt}}
	command := newCommand(spy, testOpenKS)

	err := command.DirektverkaufStornieren(context.Background(), 2, "Leitung", uuid.NewString(), verkaufID, []kasse.PositionRef{{PositionID: positionID, Menge: 3}}, "Zu viel")
	if !errors.Is(err, ErrPositionNichtStornierbar) {
		t.Errorf("expected ErrPositionNichtStornierbar, got %v", err)
	}
	if len(spy.written) != 0 {
		t.Errorf("expected no event written, got %d", len(spy.written))
	}
}

func TestDirektverkaufStornieren_DuplikatPositionRefs(t *testing.T) {
	// Duplikate sind per se ungültig — auch wenn die Summe der Mengen (1+1) die
	// verfügbare Menge (3) nicht übersteigt.
	for _, menge := range []int{2, 1} {
		getaetigt, verkaufID, positionID := getaetigtEvent(t, 500, 3)
		spy := &spyEventRepo{streamEvents: []event.Event{getaetigt}}
		command := newCommand(spy, testOpenKS)

		refs := []kasse.PositionRef{
			{PositionID: positionID, Menge: menge},
			{PositionID: positionID, Menge: menge},
		}
		err := command.DirektverkaufStornieren(context.Background(), 2, "Leitung", uuid.NewString(), verkaufID, refs, "Duplikat")
		if !errors.Is(err, ErrPositionNichtStornierbar) {
			t.Errorf("menge %d: expected ErrPositionNichtStornierbar, got %v", menge, err)
		}
		if len(spy.written) != 0 {
			t.Errorf("menge %d: expected no event written, got %d", menge, len(spy.written))
		}
	}
}

func TestDirektverkaufStornieren_WritesStornoEventWithNextVersion(t *testing.T) {
	getaetigt, verkaufID, positionID := getaetigtEvent(t, 500, 2)
	spy := &spyEventRepo{streamEvents: []event.Event{getaetigt}}
	command := newCommand(spy, testOpenKS)

	err := command.DirektverkaufStornieren(context.Background(), 2, "Leitung", uuid.NewString(), verkaufID, []kasse.PositionRef{{PositionID: positionID, Menge: 1}}, "Rückgabe")
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
		t.Errorf("expected version 2 (last version+1), got %d", got.event.Version)
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

func TestDirektverkaufStornieren_WriteErrors(t *testing.T) {
	cases := []struct {
		name     string
		writeErr error
		want     error
	}{
		{"unique violation", db.ErrAlreadyExists, ErrConflict},
		{"deadlock", db.ErrConflict, ErrConflict},
		{"kassensitzung closed meanwhile", kassenjournal_repo.ErrKassensitzungNichtOffen, ErrKasseNichtGeoeffnet},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			getaetigt, verkaufID, positionID := getaetigtEvent(t, 500, 2)
			spy := &spyEventRepo{streamEvents: []event.Event{getaetigt}, writeErr: tc.writeErr}
			command := newCommand(spy, testOpenKS)

			err := command.DirektverkaufStornieren(context.Background(), 2, "Leitung", uuid.NewString(), verkaufID, []kasse.PositionRef{{PositionID: positionID, Menge: 1}}, "Rückgabe")
			if !errors.Is(err, tc.want) {
				t.Errorf("expected %v, got %v", tc.want, err)
			}
		})
	}
}
