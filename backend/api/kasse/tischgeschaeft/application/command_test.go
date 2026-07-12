//go:build unit

package application

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/druckstation"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/produkt"
	"github.com/nicograef/jotti/backend/domain/steuer"
	"github.com/nicograef/jotti/backend/domain/tisch"
	"github.com/nicograef/jotti/backend/repository/kassenjournal_repo"
	"github.com/nicograef/jotti/backend/repository/kassensitzungen_repo"
	"github.com/nicograef/jotti/backend/repository/produkt_repo"
	"github.com/nicograef/jotti/backend/repository/tisch_repo"
)

const testKassensitzungNr = 1

var testOpenKS = &kasse.Kassensitzung{
	ZNr:    testKassensitzungNr,
	Status: kasse.KassensitzungOffen,
}

func newTestCommand(tables []tisch.Tisch, products []produkt.Produkt) Command {
	eventMock := kassenjournal_repo.NewMock(nil, nil)
	return Command{
		TischRepo:           tisch_repo.NewMock(tables, nil),
		EventRepo:           eventMock,
		ProduktRepo:         produkt_repo.NewMock(products, db.ErrNotFound),
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
		DruckstationRepo:    &mockDruckstationRepo{},
	}
}

func newTestCommandWithEventMock(tables []tisch.Tisch, products []produkt.Produkt, eventMock *kassenjournal_repo.MockRepo) Command {
	return Command{
		TischRepo:           tisch_repo.NewMock(tables, nil),
		EventRepo:           eventMock,
		ProduktRepo:         produkt_repo.NewMock(products, db.ErrNotFound),
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
		DruckstationRepo:    &mockDruckstationRepo{},
	}
}

var testProduct = produkt.Produkt{
	ID:         1,
	Name:       "Cola",
	Kategorie:  produkt.GetraenkKategorie,
	Steuersatz: steuer.RegelSteuersatz,
	Status:     produkt.ActiveStatus,
	Varianten:  []produkt.Variante{},
	CreatedAt:  time.Now().UTC(),
	UpdatedAt:  time.Now().UTC(),
}

var testVariant = produkt.Variante{
	ID:         1,
	Name:       "Cola 0,5l",
	PreisCents: 350,
	Status:     produkt.ActiveStatus,
	CreatedAt:  time.Now().UTC(),
	UpdatedAt:  time.Now().UTC(),
}

var testActiveTisch = tisch.Tisch{
	ID:        1,
	Name:      "Tisch 1",
	Status:    tisch.ActiveStatus,
	CreatedAt: time.Now().UTC(),
	UpdatedAt: time.Now().UTC(),
}

var testInactiveTisch = tisch.Tisch{
	ID:        2,
	Name:      "Tisch 2",
	Status:    tisch.InactiveStatus,
	CreatedAt: time.Now().UTC(),
	UpdatedAt: time.Now().UTC(),
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

type umbuchungPositionData struct {
	PositionID       string `json:"positionId"`
	VarianteID       int    `json:"varianteId"`
	ProduktName      string `json:"produktName"`
	VarianteName     string `json:"varianteName"`
	Kategorie        string `json:"kategorie"`
	EinzelpreisCents int    `json:"einzelpreisCents"`
	Menge            int    `json:"menge"`
}

type bestellungUmgebuchtData struct {
	UmbuchungID       string                  `json:"umbuchungId"`
	QuellTischID      int                     `json:"quellTischId"`
	ZielTischID       int                     `json:"zielTischId"`
	Positionen        []umbuchungPositionData `json:"positionen"`
	GesamtCents       int                     `json:"gesamtCents"`
	Kommentar         string                  `json:"kommentar"`
	BenutzerKommentar string                  `json:"benutzerKommentar,omitempty"`
}

type umbuchungTableRepoMock struct {
	tables map[int]tisch.Tisch
}

func (m *umbuchungTableRepoMock) GetTable(_ context.Context, id int) (tisch.Tisch, error) {
	entry, ok := m.tables[id]
	if !ok {
		return tisch.Tisch{}, db.ErrNotFound
	}
	return entry, nil
}

func (m *umbuchungTableRepoMock) CreateTable(_ context.Context, _ tisch.Tisch) (int, error) {
	return 0, nil
}

func (m *umbuchungTableRepoMock) UpdateTable(_ context.Context, _ tisch.Tisch) error {
	return nil
}

func (m *umbuchungTableRepoMock) GetAllTables(_ context.Context) ([]tisch.Tisch, error) {
	return nil, nil
}

func (m *umbuchungTableRepoMock) GetActiveTables(_ context.Context, _ int) ([]tisch.AktiverTisch, error) {
	return nil, nil
}

func (m *umbuchungTableRepoMock) GetActiveTablesWithFavorites(_ context.Context, _, _ int) ([]tisch.AktiverTischMitFavorit, error) {
	return nil, nil
}

func TestBestellungAufnehmen_KasseNichtGeoeffnet(t *testing.T) {
	ctx := context.Background()
	productMock := produkt_repo.NewMock([]produkt.Produkt{testProduct}, nil)
	productMock.AddVariant(testProduct.ID, testVariant)
	eventMock := kassenjournal_repo.NewMock(nil, nil)
	// no open KS set
	command := Command{
		TischRepo:           tisch_repo.NewMock([]tisch.Tisch{testActiveTisch}, nil),
		EventRepo:           eventMock,
		ProduktRepo:         productMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(nil, nil), // no open KS
	}

	inputs := []BestellPositionInput{
		{ProduktID: testProduct.ID, VarianteID: testVariant.ID, Menge: 1},
	}

	err := command.BestellungAufnehmen(ctx, 1, "Test User", "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", 1, inputs, "")
	if err != ErrKasseNichtGeoeffnet {
		t.Fatalf("expected ErrKasseNichtGeoeffnet, got %v", err)
	}
}

func TestBestellungAufnehmen_WithOCC(t *testing.T) {
	ctx := context.Background()
	productMock := produkt_repo.NewMock([]produkt.Produkt{testProduct}, nil)
	productMock.AddVariant(testProduct.ID, testVariant)
	command := newTestCommand([]tisch.Tisch{testActiveTisch}, []produkt.Produkt{testProduct})
	command.ProduktRepo = productMock

	inputs := []BestellPositionInput{
		{ProduktID: testProduct.ID, VarianteID: testVariant.ID, Menge: 2},
	}

	err := command.BestellungAufnehmen(ctx, 1, "Test User", "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", 1, inputs, "Testkommentar")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestBestellungAufnehmen_EnqueueArbeitsbonDruckauftraege(t *testing.T) {
	ctx := context.Background()
	productMock := produkt_repo.NewMock([]produkt.Produkt{testProduct}, nil)
	productMock.AddVariant(testProduct.ID, testVariant)

	eventMock := kassenjournal_repo.NewMock(nil, nil)
	stationMock := &mockDruckstationRepo{konfig: map[string]druckstation.Druckstation{
		"getraenk": {DruckerIP: "192.168.1.50", Bonmodus: "pro_position"},
	}}

	command := newTestCommandWithEventMock([]tisch.Tisch{testActiveTisch}, []produkt.Produkt{testProduct}, eventMock)
	command.ProduktRepo = productMock
	command.DruckstationRepo = stationMock

	inputs := []BestellPositionInput{{ProduktID: testProduct.ID, VarianteID: testVariant.ID, Menge: 2}}

	err := command.BestellungAufnehmen(ctx, 1, "Test User", "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", 1, inputs, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	enqueued := eventMock.CapturedDruckauftraege()
	if len(enqueued) != 1 {
		t.Fatalf("expected 1 enqueued druckauftrag, got %d", len(enqueued))
	}
	if enqueued[0].BonArt != "arbeitsbon" {
		t.Fatalf("expected BonArt arbeitsbon, got %s", enqueued[0].BonArt)
	}
	if enqueued[0].ZielIP != "192.168.1.50" {
		t.Fatalf("expected ZielIP 192.168.1.50, got %s", enqueued[0].ZielIP)
	}
	if enqueued[0].Payload == "" {
		t.Fatal("expected non-empty payload")
	}
}

func TestBestellungAufnehmen_Conflict(t *testing.T) {
	ctx := context.Background()
	productMock := produkt_repo.NewMock([]produkt.Produkt{testProduct}, nil)
	productMock.AddVariant(testProduct.ID, testVariant)
	eventMock := kassenjournal_repo.NewMockWithWriteErr(nil, db.ErrAlreadyExists)
	command := newTestCommandWithEventMock([]tisch.Tisch{testActiveTisch}, []produkt.Produkt{testProduct}, eventMock)
	command.ProduktRepo = productMock

	inputs := []BestellPositionInput{
		{ProduktID: testProduct.ID, VarianteID: testVariant.ID, Menge: 1},
	}

	err := command.BestellungAufnehmen(ctx, 1, "Test User", "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", 1, inputs, "")
	if err != ErrConflict {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

func TestBestellungAufnehmen_DeadlockMapsToConflict(t *testing.T) {
	ctx := context.Background()
	productMock := produkt_repo.NewMock([]produkt.Produkt{testProduct}, nil)
	productMock.AddVariant(testProduct.ID, testVariant)
	eventMock := kassenjournal_repo.NewMockWithWriteErr(nil, db.ErrConflict)
	command := newTestCommandWithEventMock([]tisch.Tisch{testActiveTisch}, []produkt.Produkt{testProduct}, eventMock)
	command.ProduktRepo = productMock

	inputs := []BestellPositionInput{
		{ProduktID: testProduct.ID, VarianteID: testVariant.ID, Menge: 1},
	}

	err := command.BestellungAufnehmen(ctx, 1, "Test User", "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", 1, inputs, "")
	if err != ErrConflict {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

// --- Invariant Tests ---

func TestBestellungAufnehmen_InactiveTisch(t *testing.T) {
	ctx := context.Background()
	productMock := produkt_repo.NewMock([]produkt.Produkt{testProduct}, nil)
	productMock.AddVariant(testProduct.ID, testVariant)
	command := newTestCommand([]tisch.Tisch{testInactiveTisch}, []produkt.Produkt{testProduct})
	command.ProduktRepo = productMock

	inputs := []BestellPositionInput{
		{ProduktID: testProduct.ID, VarianteID: testVariant.ID, Menge: 1},
	}

	err := command.BestellungAufnehmen(ctx, 1, "Test User", "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", testInactiveTisch.ID, inputs, "")
	if err != ErrTischNotActive {
		t.Fatalf("expected ErrTischNotActive, got %v", err)
	}
}

func TestZahlungKassieren_NonOrderedPosition(t *testing.T) {
	ctx := context.Background()
	// No order events exist — paying a non-existent position should fail
	command := newTestCommand([]tisch.Tisch{testActiveTisch}, nil)

	fakeRefs := []kasse.PositionRef{
		{PositionID: "00000000-0000-0000-0000-000000000001", Menge: 1},
	}

	err := command.ZahlungKassieren(ctx, 1, "Test User", testActiveTisch.ID, fakeRefs, "")
	if err != ErrPositionNichtBezahlbar {
		t.Fatalf("expected ErrPositionNichtBezahlbar, got %v", err)
	}
}

func TestZahlungKassieren_DoublePayment(t *testing.T) {
	ctx := context.Background()
	// After order + payment, unbezahlt is empty
	eventMock := kassenjournal_repo.NewMock(nil, nil)
	subject := kasse.TischSessionSubject(testKassensitzungNr, testActiveTisch.ID)
	eventMock.SetTischSession(subject, kasse.TischSession{
		SaldoCents:           0,
		UnbezahltePositionen: []kasse.Position{},
		GesamtZahlungenCents: 350,
	})
	command := Command{
		TischRepo:           tisch_repo.NewMock([]tisch.Tisch{testActiveTisch}, nil),
		EventRepo:           eventMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
	}

	refs := []kasse.PositionRef{
		{PositionID: "00000000-0000-0000-0000-000000000001", Menge: 1},
	}

	// Try to pay again — should fail
	err := command.ZahlungKassieren(ctx, 1, "Test User", testActiveTisch.ID, refs, "")
	if err != ErrPositionNichtBezahlbar {
		t.Fatalf("expected ErrPositionNichtBezahlbar, got %v", err)
	}
}

// TestZahlungKassieren_KonfliktBeiParallelemCommit stellt das Read–Validate–Sign–Write-
// Race nach: Der Command validiert gegen die Projektion (LastEventVersion 1), aber im
// Event-Store liegt bereits ein parallel committetes Event mit Version 2 (z. B. eine
// zweite Zahlung während der TSE-Signierung). Der Write mit erwarteter Version 1 muss
// am UNIQUE(subject, version)-Constraint scheitern — der zweite Request bekommt 409
// statt eine Doppelzahlung durchzuschreiben.
func TestZahlungKassieren_KonfliktBeiParallelemCommit(t *testing.T) {
	ctx := context.Background()
	subject := kasse.TischSessionSubject(testKassensitzungNr, testActiveTisch.ID)

	konkurrierendesEvent, err := kasse.NewZahlungKassiertEvent(subject, 2, "Andere Servicekraft",
		[]kasse.Position{{PositionID: "22222222-2222-4222-8222-222222222222", VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", EinzelpreisCents: 350, Menge: 1}},
		350, "")
	if err != nil {
		t.Fatalf("failed to create concurrent event: %v", err)
	}
	konkurrierendesEvent.ID = 5
	konkurrierendesEvent.Version = 2

	eventMock := kassenjournal_repo.NewMock([]event.Event{konkurrierendesEvent}, nil)
	// Die Projektion ist der Stand VOR dem parallelen Commit: Version 1, Position offen.
	eventMock.SetTischSession(subject, kasse.TischSession{
		SaldoCents: 350,
		UnbezahltePositionen: []kasse.Position{{
			PositionID:       "22222222-2222-4222-8222-222222222222",
			VarianteID:       1,
			ProduktName:      "Cola",
			VarianteName:     "0,5l",
			Kategorie:        "getraenk",
			Steuersatz:       "regel",
			EinzelpreisCents: 350,
			Menge:            1,
		}},
		LastEventVersion: 1,
	})

	command := Command{
		TischRepo:           tisch_repo.NewMock([]tisch.Tisch{testActiveTisch}, nil),
		EventRepo:           eventMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
	}

	err = command.ZahlungKassieren(ctx, 1, "Test User", testActiveTisch.ID,
		[]kasse.PositionRef{{PositionID: "22222222-2222-4222-8222-222222222222", Menge: 1}}, "")
	if err != ErrConflict {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

// Ohne parallelen Commit schreibt die Zahlung mit der Version des gelesenen Zustands + 1.
func TestZahlungKassieren_VersionAusGelesenerProjektion(t *testing.T) {
	ctx := context.Background()
	subject := kasse.TischSessionSubject(testKassensitzungNr, testActiveTisch.ID)

	eventMock := kassenjournal_repo.NewMock(nil, nil)
	eventMock.SetTischSession(subject, kasse.TischSession{
		SaldoCents: 350,
		UnbezahltePositionen: []kasse.Position{{
			PositionID:       "22222222-2222-4222-8222-222222222222",
			VarianteID:       1,
			ProduktName:      "Cola",
			VarianteName:     "0,5l",
			Kategorie:        "getraenk",
			Steuersatz:       "regel",
			EinzelpreisCents: 350,
			Menge:            1,
		}},
		LastEventVersion: 3,
	})

	command := Command{
		TischRepo:           tisch_repo.NewMock([]tisch.Tisch{testActiveTisch}, nil),
		EventRepo:           eventMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
	}

	err := command.ZahlungKassieren(ctx, 1, "Test User", testActiveTisch.ID,
		[]kasse.PositionRef{{PositionID: "22222222-2222-4222-8222-222222222222", Menge: 1}}, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	events, err := eventMock.ReadEventsBySubject(ctx, subject)
	if err != nil {
		t.Fatalf("expected no read error, got %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected exactly one event, got %d", len(events))
	}
	if events[0].Version != 4 {
		t.Fatalf("expected version 4 (LastEventVersion 3 + 1), got %d", events[0].Version)
	}
}

func TestStornierungErteilen_AlreadyPaidPosition_Succeeds(t *testing.T) {
	ctx := context.Background()
	subject := kasse.TischSessionSubject(testKassensitzungNr, testActiveTisch.ID)

	orderEvent, _ := kasse.NewBestellungAufgenommenEvent(subject, 1, "Test User", "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
		[]kasse.Position{
			{VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", EinzelpreisCents: 350, Menge: 1},
		}, "")

	var orderData struct {
		Positionen []struct {
			PositionID string `json:"positionId"`
		} `json:"positionen"`
	}
	if err := json.Unmarshal(orderEvent.Data, &orderData); err != nil {
		t.Fatalf("expected no unmarshal error, got %v", err)
	}
	posID := orderData.Positionen[0].PositionID

	paymentEvent, _ := kasse.NewZahlungKassiertEvent(subject, 1, "Test User",
		[]kasse.Position{
			{PositionID: posID, VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", EinzelpreisCents: 350, Menge: 1},
		}, 350, "")

	eventMock := kassenjournal_repo.NewMock(nil, nil)
	eventMock.SetTischSession(subject, kasse.TischSession{
		SaldoCents:           0,
		UnbezahltePositionen: []kasse.Position{},
		GesamtZahlungenCents: 350,
	})
	orderEvent.Subject = subject
	paymentEvent.Subject = subject
	eventMock.AddEvent(orderEvent)
	eventMock.AddEvent(paymentEvent)

	command := Command{
		TischRepo:           tisch_repo.NewMock([]tisch.Tisch{testActiveTisch}, nil),
		EventRepo:           eventMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
	}

	refs := []kasse.PositionRef{{PositionID: posID, Menge: 1}}

	err := command.StornierungErteilen(ctx, 1, "Test User", testActiveTisch.ID, refs, "Reklamation")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestStornierungErteilen_AlreadyCancelledPosition_Fails(t *testing.T) {
	ctx := context.Background()
	subject := kasse.TischSessionSubject(testKassensitzungNr, testActiveTisch.ID)

	orderEvent, _ := kasse.NewBestellungAufgenommenEvent(subject, 1, "Test User", "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
		[]kasse.Position{
			{VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", EinzelpreisCents: 350, Menge: 1},
		}, "")

	var orderData struct {
		Positionen []struct {
			PositionID string `json:"positionId"`
		} `json:"positionen"`
	}
	if err := json.Unmarshal(orderEvent.Data, &orderData); err != nil {
		t.Fatalf("expected no unmarshal error, got %v", err)
	}
	posID := orderData.Positionen[0].PositionID

	// Die Position wurde bereits geldneutral korrigiert (unbezahlt) — ein erneuter
	// Storno darf nicht mehr greifen.
	cancelEvent, _ := kasse.NewBestellungKorrigiertEvent(subject, 1, "Test User",
		[]kasse.Position{
			{PositionID: posID, VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", EinzelpreisCents: 350, Menge: 1},
		}, 350, "Test")

	eventMock := kassenjournal_repo.NewMock(nil, nil)
	eventMock.SetTischSession(subject, kasse.TischSession{
		SaldoCents:           0,
		UnbezahltePositionen: []kasse.Position{},
		GesamtZahlungenCents: 0,
	})
	orderEvent.Subject = subject
	cancelEvent.Subject = subject
	eventMock.AddEvent(orderEvent)
	eventMock.AddEvent(cancelEvent)

	command := Command{
		TischRepo:           tisch_repo.NewMock([]tisch.Tisch{testActiveTisch}, nil),
		EventRepo:           eventMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
	}

	refs := []kasse.PositionRef{{PositionID: posID, Menge: 1}}

	err := command.StornierungErteilen(ctx, 1, "Test User", testActiveTisch.ID, refs, "")
	if err != ErrPositionNichtStornierbar {
		t.Fatalf("expected ErrPositionNichtStornierbar, got %v", err)
	}
}

func TestZahlungKassieren_ExceedsAvailableMenge(t *testing.T) {
	ctx := context.Background()
	// State has 1 position with Menge 1
	eventMock := kassenjournal_repo.NewMock(nil, nil)
	subject := kasse.TischSessionSubject(testKassensitzungNr, testActiveTisch.ID)
	eventMock.SetTischSession(subject, kasse.TischSession{
		SaldoCents: 350,
		UnbezahltePositionen: []kasse.Position{
			{PositionID: "pos-1", VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", EinzelpreisCents: 350, Menge: 1},
		},
	})
	command := Command{
		TischRepo:           tisch_repo.NewMock([]tisch.Tisch{testActiveTisch}, nil),
		EventRepo:           eventMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
	}

	// Try to pay for Menge 2 when only 1 was ordered
	refs := []kasse.PositionRef{
		{PositionID: "pos-1", Menge: 2},
	}

	err := command.ZahlungKassieren(ctx, 1, "Test User", testActiveTisch.ID, refs, "")
	if err != ErrPositionNichtBezahlbar {
		t.Fatalf("expected ErrPositionNichtBezahlbar, got %v", err)
	}
}

// duplikatTestSession liefert eine Tisch-Session mit einer Position (Menge 3),
// unbezahlt, für die Duplikat-Ablehnungstests.
func duplikatTestSession() kasse.TischSession {
	pos := kasse.Position{PositionID: "pos-1", VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", EinzelpreisCents: 350, Menge: 3}
	return kasse.TischSession{
		SaldoCents:           1050,
		UnbezahltePositionen: []kasse.Position{pos},
	}
}

// Duplikate sind per se ungültig — auch wenn die Summe der Mengen (1+1) die
// verfügbare Menge (3) nicht übersteigt.
var duplikatRefs = []kasse.PositionRef{
	{PositionID: "pos-1", Menge: 1},
	{PositionID: "pos-1", Menge: 1},
}

func TestZahlungKassieren_DuplikatPositionRefs(t *testing.T) {
	ctx := context.Background()
	eventMock := kassenjournal_repo.NewMock(nil, nil)
	eventMock.SetTischSession(kasse.TischSessionSubject(testKassensitzungNr, testActiveTisch.ID), duplikatTestSession())
	command := Command{
		TischRepo:           tisch_repo.NewMock([]tisch.Tisch{testActiveTisch}, nil),
		EventRepo:           eventMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
	}

	err := command.ZahlungKassieren(ctx, 1, "Test User", testActiveTisch.ID, duplikatRefs, "")
	if err != ErrPositionNichtBezahlbar {
		t.Fatalf("expected ErrPositionNichtBezahlbar, got %v", err)
	}
}

func TestStornierungErteilen_DuplikatPositionRefs(t *testing.T) {
	ctx := context.Background()
	subject := kasse.TischSessionSubject(testKassensitzungNr, testActiveTisch.ID)

	orderEvent, _ := kasse.NewBestellungAufgenommenEvent(subject, 1, "Test User", "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
		[]kasse.Position{
			{VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", EinzelpreisCents: 350, Menge: 3},
		}, "")

	var orderData struct {
		Positionen []struct {
			PositionID string `json:"positionId"`
		} `json:"positionen"`
	}
	if err := json.Unmarshal(orderEvent.Data, &orderData); err != nil {
		t.Fatalf("expected no unmarshal error, got %v", err)
	}
	posID := orderData.Positionen[0].PositionID

	eventMock := kassenjournal_repo.NewMock(nil, nil)
	eventMock.SetTischSession(subject, duplikatTestSession())
	orderEvent.Subject = subject
	eventMock.AddEvent(orderEvent)

	command := Command{
		TischRepo:           tisch_repo.NewMock([]tisch.Tisch{testActiveTisch}, nil),
		EventRepo:           eventMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
	}

	refs := []kasse.PositionRef{
		{PositionID: posID, Menge: 1},
		{PositionID: posID, Menge: 1},
	}

	err := command.StornierungErteilen(ctx, 1, "Test User", testActiveTisch.ID, refs, "Duplikat")
	if err != ErrPositionNichtStornierbar {
		t.Fatalf("expected ErrPositionNichtStornierbar, got %v", err)
	}
}

func TestBestellungUmbuchen_HappyPath(t *testing.T) {
	ctx := context.Background()
	quellTisch := tisch.Tisch{ID: 1, Name: "Tisch Quelle", Status: tisch.ActiveStatus}
	zielTisch := tisch.Tisch{ID: 2, Name: "Tisch Ziel", Status: tisch.ActiveStatus}

	eventMock := kassenjournal_repo.NewMock(nil, nil)
	quellSubject := kasse.TischSessionSubject(testKassensitzungNr, quellTisch.ID)
	zielSubject := kasse.TischSessionSubject(testKassensitzungNr, zielTisch.ID)
	quellPositionID := uuid.New().String()

	eventMock.SetTischSession(quellSubject, kasse.TischSession{
		SaldoCents: 700,
		UnbezahltePositionen: []kasse.Position{
			{
				PositionID:   quellPositionID,
				VarianteID:   1,
				ProduktName:  "Cola",
				VarianteName: "0,5l",
				Kategorie:    "getraenk", Steuersatz: "regel",
				EinzelpreisCents: 350,
				Menge:            2,
			},
		},
	})

	command := Command{
		TischRepo:           tisch_repo.NewMock([]tisch.Tisch{quellTisch, zielTisch}, nil),
		EventRepo:           eventMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
	}

	err := command.BestellungUmbuchen(ctx, 1, "Test User", quellTisch.ID, zielTisch.ID, []kasse.PositionRef{{PositionID: quellPositionID, Menge: 1}}, "Gast gewechselt")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	quellEvents, err := eventMock.ReadEventsBySubject(ctx, quellSubject)
	if err != nil {
		t.Fatalf("expected no error reading source events, got %v", err)
	}
	if len(quellEvents) != 1 {
		t.Fatalf("expected 1 source event, got %d", len(quellEvents))
	}
	if quellEvents[0].Type != string(kasse.EventTypeBestellungUmgebuchtV1) {
		t.Fatalf("expected source event type %s, got %s", kasse.EventTypeBestellungUmgebuchtV1, quellEvents[0].Type)
	}

	zielEvents, err := eventMock.ReadEventsBySubject(ctx, zielSubject)
	if err != nil {
		t.Fatalf("expected no error reading target events, got %v", err)
	}
	if len(zielEvents) != 1 {
		t.Fatalf("expected 1 target event, got %d", len(zielEvents))
	}
	if zielEvents[0].Type != string(kasse.EventTypeBestellungUmgebuchtV1) {
		t.Fatalf("expected target event type %s, got %s", kasse.EventTypeBestellungUmgebuchtV1, zielEvents[0].Type)
	}

	var quellData bestellungUmgebuchtData
	if err := json.Unmarshal(quellEvents[0].Data, &quellData); err != nil {
		t.Fatalf("expected no unmarshal error for source umbuchung data, got %v", err)
	}

	if quellData.GesamtCents != 350 {
		t.Fatalf("expected source amount 350, got %d", quellData.GesamtCents)
	}
	if quellData.QuellTischID != quellTisch.ID || quellData.ZielTischID != zielTisch.ID {
		t.Fatalf("unexpected source tisch refs: quell=%d ziel=%d", quellData.QuellTischID, quellData.ZielTischID)
	}
	if quellData.Kommentar != "Umbuchung auf Tisch Ziel" {
		t.Fatalf("unexpected source comment: %q", quellData.Kommentar)
	}
	if quellData.BenutzerKommentar != "Gast gewechselt" {
		t.Fatalf("unexpected source benutzerKommentar: %q", quellData.BenutzerKommentar)
	}
	if len(quellData.Positionen) != 1 {
		t.Fatalf("expected 1 source position, got %d", len(quellData.Positionen))
	}
	if quellData.Positionen[0].PositionID != quellPositionID {
		t.Fatalf("expected source position ID %q, got %q", quellPositionID, quellData.Positionen[0].PositionID)
	}
	if quellData.Positionen[0].EinzelpreisCents != 350 {
		t.Fatalf("expected source einzelpreis 350, got %d", quellData.Positionen[0].EinzelpreisCents)
	}

	var zielData bestellungUmgebuchtData
	if err := json.Unmarshal(zielEvents[0].Data, &zielData); err != nil {
		t.Fatalf("expected no unmarshal error for target umbuchung data, got %v", err)
	}

	if zielData.GesamtCents != 350 {
		t.Fatalf("expected target amount 350, got %d", zielData.GesamtCents)
	}
	// Beide Seiten teilen sich dieselbe UmbuchungID.
	if zielData.UmbuchungID != quellData.UmbuchungID {
		t.Fatalf("expected shared UmbuchungID, got quell=%q ziel=%q", quellData.UmbuchungID, zielData.UmbuchungID)
	}
	if zielData.Kommentar != "Umbuchung von Tisch Quelle" {
		t.Fatalf("unexpected target comment: %q", zielData.Kommentar)
	}
	if zielData.BenutzerKommentar != "Gast gewechselt" {
		t.Fatalf("unexpected target benutzerKommentar: %q", zielData.BenutzerKommentar)
	}
	if len(zielData.Positionen) != 1 {
		t.Fatalf("expected 1 target position, got %d", len(zielData.Positionen))
	}
	if zielData.Positionen[0].PositionID == quellPositionID {
		t.Fatalf("expected target position ID to be regenerated, but remained %q", zielData.Positionen[0].PositionID)
	}
	if zielData.Positionen[0].EinzelpreisCents != 350 {
		t.Fatalf("expected target einzelpreis 350, got %d", zielData.Positionen[0].EinzelpreisCents)
	}
}

func TestBestellungUmbuchen_KommentarWirdGekuerzt(t *testing.T) {
	ctx := context.Background()
	quellTisch := tisch.Tisch{ID: 1, Name: strings.Repeat("Q", 100), Status: tisch.ActiveStatus}
	zielTisch := tisch.Tisch{ID: 2, Name: strings.Repeat("Z", 100), Status: tisch.ActiveStatus}

	eventMock := kassenjournal_repo.NewMock(nil, nil)
	quellSubject := kasse.TischSessionSubject(testKassensitzungNr, quellTisch.ID)
	zielSubject := kasse.TischSessionSubject(testKassensitzungNr, zielTisch.ID)
	quellPositionID := uuid.New().String()

	eventMock.SetTischSession(quellSubject, kasse.TischSession{
		UnbezahltePositionen: []kasse.Position{{
			PositionID:   quellPositionID,
			VarianteID:   1,
			ProduktName:  "Cola",
			VarianteName: "0,5l",
			Kategorie:    "getraenk", Steuersatz: "regel",
			EinzelpreisCents: 350,
			Menge:            1,
		}},
	})

	command := Command{
		TischRepo:           tisch_repo.NewMock([]tisch.Tisch{quellTisch, zielTisch}, nil),
		EventRepo:           eventMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
	}

	err := command.BestellungUmbuchen(ctx, 1, "Test User", quellTisch.ID, zielTisch.ID, []kasse.PositionRef{{PositionID: quellPositionID, Menge: 1}}, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	quellEvents, _ := eventMock.ReadEventsBySubject(ctx, quellSubject)
	zielEvents, _ := eventMock.ReadEventsBySubject(ctx, zielSubject)

	var quellData bestellungUmgebuchtData
	if err := json.Unmarshal(quellEvents[0].Data, &quellData); err != nil {
		t.Fatalf("expected no unmarshal error for source umbuchung data, got %v", err)
	}
	var zielData bestellungUmgebuchtData
	if err := json.Unmarshal(zielEvents[0].Data, &zielData); err != nil {
		t.Fatalf("expected no unmarshal error for target umbuchung data, got %v", err)
	}

	if utf8.RuneCountInString(quellData.Kommentar) > 100 {
		t.Fatalf("expected source comment length <= 100 runes, got %d", utf8.RuneCountInString(quellData.Kommentar))
	}
	if utf8.RuneCountInString(zielData.Kommentar) > 100 {
		t.Fatalf("expected target comment length <= 100 runes, got %d", utf8.RuneCountInString(zielData.Kommentar))
	}
	if !strings.HasPrefix(quellData.Kommentar, "Umbuchung auf ") {
		t.Fatalf("expected source comment prefix, got %q", quellData.Kommentar)
	}
	if !strings.HasPrefix(zielData.Kommentar, "Umbuchung von ") {
		t.Fatalf("expected target comment prefix, got %q", zielData.Kommentar)
	}
}

func TestBestellungUmbuchen_PositionNichtUmbuchbar(t *testing.T) {
	ctx := context.Background()
	quellTisch := tisch.Tisch{ID: 1, Name: "Tisch Quelle", Status: tisch.ActiveStatus}
	zielTisch := tisch.Tisch{ID: 2, Name: "Tisch Ziel", Status: tisch.ActiveStatus}

	eventMock := kassenjournal_repo.NewMock(nil, nil)
	quellSubject := kasse.TischSessionSubject(testKassensitzungNr, quellTisch.ID)
	eventMock.SetTischSession(quellSubject, kasse.TischSession{
		UnbezahltePositionen: []kasse.Position{{
			PositionID:   uuid.New().String(),
			VarianteID:   1,
			ProduktName:  "Cola",
			VarianteName: "0,5l",
			Kategorie:    "getraenk", Steuersatz: "regel",
			EinzelpreisCents: 350,
			Menge:            1,
		}},
	})

	command := Command{
		TischRepo:           tisch_repo.NewMock([]tisch.Tisch{quellTisch, zielTisch}, nil),
		EventRepo:           eventMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
	}

	err := command.BestellungUmbuchen(ctx, 1, "Test User", quellTisch.ID, zielTisch.ID, []kasse.PositionRef{{PositionID: uuid.New().String(), Menge: 1}}, "")
	if err != ErrPositionNichtUmbuchbar {
		t.Fatalf("expected ErrPositionNichtUmbuchbar, got %v", err)
	}
}

func TestBestellungUmbuchen_GleicherTisch(t *testing.T) {
	err := Command{}.BestellungUmbuchen(context.Background(), 1, "Test User", 3, 3, []kasse.PositionRef{{PositionID: uuid.New().String(), Menge: 1}}, "")
	if err != ErrUmbuchungGleicherTisch {
		t.Fatalf("expected ErrUmbuchungGleicherTisch, got %v", err)
	}
}

func TestBestellungUmbuchen_ZielTischNotActive(t *testing.T) {
	ctx := context.Background()
	quellTisch := tisch.Tisch{ID: 1, Name: "Tisch Quelle", Status: tisch.ActiveStatus}
	zielTisch := tisch.Tisch{ID: 2, Name: "Tisch Ziel", Status: tisch.InactiveStatus}

	command := Command{
		TischRepo:           tisch_repo.NewMock([]tisch.Tisch{quellTisch, zielTisch}, nil),
		EventRepo:           kassenjournal_repo.NewMock(nil, nil),
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
	}

	err := command.BestellungUmbuchen(ctx, 1, "Test User", quellTisch.ID, zielTisch.ID, []kasse.PositionRef{{PositionID: uuid.New().String(), Menge: 1}}, "")
	if err != ErrTischNotActive {
		t.Fatalf("expected ErrTischNotActive, got %v", err)
	}
}

func TestBestellungUmbuchen_ZielTischNotFound(t *testing.T) {
	ctx := context.Background()
	quellTisch := tisch.Tisch{ID: 1, Name: "Tisch Quelle", Status: tisch.ActiveStatus}

	command := Command{
		TischRepo: &umbuchungTableRepoMock{tables: map[int]tisch.Tisch{
			quellTisch.ID: quellTisch,
		}},
		EventRepo:           kassenjournal_repo.NewMock(nil, nil),
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
	}

	err := command.BestellungUmbuchen(ctx, 1, "Test User", quellTisch.ID, 99, []kasse.PositionRef{{PositionID: uuid.New().String(), Menge: 1}}, "")
	if err != ErrTischNotFound {
		t.Fatalf("expected ErrTischNotFound, got %v", err)
	}
}

func TestBestellungUmbuchen_KasseNichtGeoeffnet(t *testing.T) {
	ctx := context.Background()
	command := Command{
		TischRepo:           tisch_repo.NewMock([]tisch.Tisch{testActiveTisch}, nil),
		EventRepo:           kassenjournal_repo.NewMock(nil, nil),
		KassensitzungenRepo: kassensitzungen_repo.NewMock(nil, nil),
	}

	err := command.BestellungUmbuchen(ctx, 1, "Test User", 1, 2, []kasse.PositionRef{{PositionID: uuid.New().String(), Menge: 1}}, "")
	if err != ErrKasseNichtGeoeffnet {
		t.Fatalf("expected ErrKasseNichtGeoeffnet, got %v", err)
	}
}

func TestBestellungUmbuchen_Conflict(t *testing.T) {
	ctx := context.Background()
	quellTisch := tisch.Tisch{ID: 1, Name: "Tisch Quelle", Status: tisch.ActiveStatus}
	zielTisch := tisch.Tisch{ID: 2, Name: "Tisch Ziel", Status: tisch.ActiveStatus}

	eventMock := kassenjournal_repo.NewMockWithWriteErr(nil, db.ErrAlreadyExists)
	quellSubject := kasse.TischSessionSubject(testKassensitzungNr, quellTisch.ID)
	quellPositionID := uuid.New().String()
	eventMock.SetTischSession(quellSubject, kasse.TischSession{
		UnbezahltePositionen: []kasse.Position{{
			PositionID:   quellPositionID,
			VarianteID:   1,
			ProduktName:  "Cola",
			VarianteName: "0,5l",
			Kategorie:    "getraenk", Steuersatz: "regel",
			EinzelpreisCents: 350,
			Menge:            1,
		}},
	})

	command := Command{
		TischRepo:           tisch_repo.NewMock([]tisch.Tisch{quellTisch, zielTisch}, nil),
		EventRepo:           eventMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
	}

	err := command.BestellungUmbuchen(ctx, 1, "Test User", quellTisch.ID, zielTisch.ID, []kasse.PositionRef{{PositionID: quellPositionID, Menge: 1}}, "")
	if err != ErrConflict {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

func TestStornierungErteilen_GemischterStorno_AtomischKorrekturUndWarenruecknahme(t *testing.T) {
	ctx := context.Background()
	subject := kasse.TischSessionSubject(testKassensitzungNr, testActiveTisch.ID)

	// Bestellt 3 Bier, davon 1 bezahlt → ein Storno von 3 spaltet in 2 Korrektur
	// (unbezahlt) und 1 Warenrücknahme (bezahlt).
	orderEvent, _ := kasse.NewBestellungAufgenommenEvent(subject, 1, "Test User", "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", []kasse.Position{{
		VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", EinzelpreisCents: 350, Menge: 3,
	}}, "")
	var orderData kasse.BestellungAufgenommenV1Data
	if err := json.Unmarshal(orderEvent.Data, &orderData); err != nil {
		t.Fatalf("expected no unmarshal error, got %v", err)
	}
	posID := orderData.Positionen[0].PositionID

	paymentEvent, _ := kasse.NewZahlungKassiertEvent(subject, 1, "Test User", []kasse.Position{{
		PositionID: posID, VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", EinzelpreisCents: 350, Menge: 1,
	}}, 350, "")

	eventMock := kassenjournal_repo.NewMock(nil, nil)
	eventMock.SetTischSession(subject, kasse.TischSession{})
	eventMock.AddEvent(orderEvent)
	eventMock.AddEvent(paymentEvent)

	command := Command{
		TischRepo:           tisch_repo.NewMock([]tisch.Tisch{testActiveTisch}, nil),
		EventRepo:           eventMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
	}

	err := command.StornierungErteilen(ctx, 2, "Leitung", testActiveTisch.ID, []kasse.PositionRef{{PositionID: posID, Menge: 3}}, "Reklamation")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	events, err := eventMock.ReadEventsBySubject(ctx, subject)
	if err != nil {
		t.Fatalf("expected no error reading events, got %v", err)
	}

	var korrektur, warenruecknahme int
	for _, evt := range events {
		switch evt.Type {
		case string(kasse.EventTypeBestellungKorrigiertV1):
			korrektur++
		case string(kasse.EventTypeStornierungErteiltV1):
			warenruecknahme++
		}
	}
	if korrektur != 1 {
		t.Fatalf("expected exactly 1 bestellung-korrigiert, got %d", korrektur)
	}
	if warenruecknahme != 1 {
		t.Fatalf("expected exactly 1 stornierung-erteilt, got %d", warenruecknahme)
	}
}
