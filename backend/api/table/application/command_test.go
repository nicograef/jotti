//go:build unit

package application

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	bondruckApp "github.com/nicograef/jotti/backend/api/bondruck/application"
	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/product"
	"github.com/nicograef/jotti/backend/domain/settings"
	"github.com/nicograef/jotti/backend/domain/steuer"
	"github.com/nicograef/jotti/backend/domain/table"
	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/nicograef/jotti/backend/repository/druckauftrag_repo"
	"github.com/nicograef/jotti/backend/repository/kassenjournal_repo"
	"github.com/nicograef/jotti/backend/repository/kassensitzungen_repo"
	"github.com/nicograef/jotti/backend/repository/product_repo"
	"github.com/nicograef/jotti/backend/repository/table_repo"
	"github.com/nicograef/jotti/backend/repository/tse_repo"
)

const testKassensitzungNr = 1

var testOpenKS = &kasse.Kassensitzung{
	ZNr:    testKassensitzungNr,
	Status: kasse.KassensitzungOffen,
}

func newTestCommand(tables []table.Tisch, products []product.Produkt) Command {
	eventMock := kassenjournal_repo.NewMock(nil, nil)
	return Command{
		TableRepo:           table_repo.NewMock(tables, nil),
		EventRepo:           eventMock,
		ProductRepo:         product_repo.NewMock(products, db.ErrNotFound),
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
	}
}

func newTestCommandWithEventMock(tables []table.Tisch, products []product.Produkt, eventMock *kassenjournal_repo.MockRepo) Command {
	return Command{
		TableRepo:           table_repo.NewMock(tables, nil),
		EventRepo:           eventMock,
		ProductRepo:         product_repo.NewMock(products, db.ErrNotFound),
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
	}
}

var testProduct = product.Produkt{
	ID:         1,
	Name:       "Cola",
	Kategorie:  product.GetraenkKategorie,
	Steuersatz: steuer.RegelSteuersatz,
	Status:     product.ActiveStatus,
	Varianten:  []product.Variante{},
	CreatedAt:  time.Now().UTC(),
	UpdatedAt:  time.Now().UTC(),
}

var testVariant = product.Variante{
	ID:         1,
	Name:       "Cola 0,5l",
	PreisCents: 350,
	Status:     product.ActiveStatus,
	CreatedAt:  time.Now().UTC(),
	UpdatedAt:  time.Now().UTC(),
}

var testActiveTisch = table.Tisch{
	ID:        1,
	Name:      "Tisch 1",
	Status:    table.ActiveStatus,
	CreatedAt: time.Now().UTC(),
	UpdatedAt: time.Now().UTC(),
}

var testInactiveTisch = table.Tisch{
	ID:        2,
	Name:      "Tisch 2",
	Status:    table.InactiveStatus,
	CreatedAt: time.Now().UTC(),
	UpdatedAt: time.Now().UTC(),
}

type mockDruckstationRepo struct {
	konfig map[string]bondruckApp.Druckstation
	err    error
}

func (m *mockDruckstationRepo) GetKonfigurierteDruckstationen(_ context.Context) (map[string]bondruckApp.Druckstation, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.konfig, nil
}

// kassenbelegStationen ist die konfigurierte Kassenbeleg-Druckstation für die
// KassenbelegDrucken-Tests (Ziel-IP des Kassenbeleg-Druckers).
var kassenbelegStationen = map[string]bondruckApp.Druckstation{
	"kassenbeleg": {IP: "192.168.1.80"},
}

type mockDruckauftragRepo struct {
	enqueued []druckauftrag_repo.NeuerDruckauftrag
	err      error
}

func (m *mockDruckauftragRepo) EnqueueDruckauftraege(_ context.Context, auftraege []druckauftrag_repo.NeuerDruckauftrag) error {
	if m.err != nil {
		return m.err
	}
	m.enqueued = append(m.enqueued, auftraege...)
	return nil
}

type mockSettingsRepo struct {
	betreiber      settings.Betreiber
	kassenident    settings.Kassenidentitaet
	tse            settings.TSEKonfiguration
	betreiberErr   error
	kassenidentErr error
	tseErr         error
}

type umbuchungPositionData struct {
	PositionID   string `json:"positionId"`
	VarianteID   int    `json:"varianteId"`
	ProduktName  string `json:"produktName"`
	VarianteName string `json:"varianteName"`
	Kategorie    string `json:"kategorie"`
	Einzelpreis  int    `json:"einzelpreis"`
	Menge        int    `json:"menge"`
}

type bestellungUmgebuchtData struct {
	UmbuchungID  string                  `json:"umbuchungId"`
	QuellTischID int                     `json:"quellTischId"`
	ZielTischID  int                     `json:"zielTischId"`
	Positionen   []umbuchungPositionData `json:"positionen"`
	GesamtCents  int                     `json:"gesamtCents"`
	Kommentar    string                  `json:"kommentar"`
}

type umbuchungTableRepoMock struct {
	tables map[int]table.Tisch
}

func (m *umbuchungTableRepoMock) GetTable(_ context.Context, id int) (table.Tisch, error) {
	tisch, ok := m.tables[id]
	if !ok {
		return table.Tisch{}, db.ErrNotFound
	}
	return tisch, nil
}

func (m *umbuchungTableRepoMock) CreateTable(_ context.Context, _ table.Tisch) (int, error) {
	return 0, nil
}

func (m *umbuchungTableRepoMock) UpdateTable(_ context.Context, _ table.Tisch) error {
	return nil
}

func (m *umbuchungTableRepoMock) GetAllTables(_ context.Context) ([]table.Tisch, error) {
	return nil, nil
}

func (m *umbuchungTableRepoMock) GetActiveTables(_ context.Context, _ int) ([]table.AktiverTisch, error) {
	return nil, nil
}

func (m *umbuchungTableRepoMock) GetActiveTablesWithFavorites(_ context.Context, _, _ int) ([]table.AktiverTischMitFavorit, error) {
	return nil, nil
}

func (m *mockSettingsRepo) GetBetreiber(_ context.Context) (settings.Betreiber, error) {
	if m.betreiberErr != nil {
		return settings.Betreiber{}, m.betreiberErr
	}
	return m.betreiber, nil
}

func (m *mockSettingsRepo) GetKassenidentitaet(_ context.Context) (settings.Kassenidentitaet, error) {
	if m.kassenidentErr != nil {
		return settings.Kassenidentitaet{}, m.kassenidentErr
	}
	return m.kassenident, nil
}

func (m *mockSettingsRepo) GetTSEKonfiguration(_ context.Context) (settings.TSEKonfiguration, error) {
	if m.tseErr != nil {
		return settings.TSEKonfiguration{}, m.tseErr
	}
	return m.tse, nil
}

// mockTSEAuftragRepo liefert den Signaturauftrags-Stand je Event-ID; Events
// ohne Eintrag gelten als nicht signaturpflichtig (db.ErrNotFound).
type mockTSEAuftragRepo struct {
	staende map[int]tse_repo.SignaturauftragStand
}

func (m *mockTSEAuftragRepo) GetSignaturauftragZuEvent(_ context.Context, eventID int) (tse_repo.SignaturauftragStand, error) {
	if stand, ok := m.staende[eventID]; ok {
		return stand, nil
	}
	return tse_repo.SignaturauftragStand{}, db.ErrNotFound
}

func TestTischErstellen(t *testing.T) {
	ctx := context.Background()
	command := newTestCommand(nil, nil)

	tischId, err := command.TischErstellen(ctx, "Tisch 1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tischId != 1 {
		t.Errorf("expected tisch ID 1, got %d", tischId)
	}

	tisch, err := command.TableRepo.GetTable(ctx, tischId)
	if err != nil {
		t.Fatalf("expected no error retrieving tisch, got %v", err)
	}
	if tisch.Name != "Tisch 1" {
		t.Errorf("expected tisch name 'Tisch 1', got %s", tisch.Name)
	}
}

func TestTischErstellen_Error(t *testing.T) {
	repo := table_repo.NewMock([]table.Tisch{}, db.ErrAlreadyExists)
	command := Command{TableRepo: repo}

	_, err := command.TischErstellen(context.Background(), "Tisch 1")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestTischAktualisieren(t *testing.T) {
	repo := table_repo.NewMock([]table.Tisch{{ID: 1, Name: "Old Name", Status: table.ActiveStatus, UpdatedAt: time.Now().UTC()}}, nil)
	command := Command{TableRepo: repo}

	err := command.TischAktualisieren(context.Background(), 1, "New Name")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	tisch, err := command.TableRepo.GetTable(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error retrieving tisch, got %v", err)
	}
	if tisch.Name != "New Name" {
		t.Errorf("expected tisch name to be 'New Name', got %s", tisch.Name)
	}
}

func TestTischAktualisieren_NotFound(t *testing.T) {
	repo := table_repo.NewMock([]table.Tisch{}, db.ErrNotFound)
	command := Command{TableRepo: repo}

	err := command.TischAktualisieren(context.Background(), 999, "New Name")
	if err != ErrTischNotFound {
		t.Fatalf("expected ErrTischNotFound, got %v", err)
	}
}

func TestTischAktivieren(t *testing.T) {
	repo := table_repo.NewMock([]table.Tisch{{ID: 1, Name: "Tisch 1", Status: table.InactiveStatus, UpdatedAt: time.Now().UTC()}}, nil)
	command := Command{TableRepo: repo}

	err := command.TischAktivieren(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	tbl, err := repo.GetTable(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error retrieving tisch, got %v", err)
	}
	if tbl.Status != table.ActiveStatus {
		t.Errorf("expected tisch status to be Active, got %v", tbl.Status)
	}
}

func TestTischAktivieren_NotFound(t *testing.T) {
	repo := table_repo.NewMock([]table.Tisch{}, db.ErrNotFound)
	command := Command{TableRepo: repo}

	err := command.TischAktivieren(context.Background(), 999)
	if err != ErrTischNotFound {
		t.Fatalf("expected ErrTischNotFound, got %v", err)
	}
}

func TestTischDeaktivieren(t *testing.T) {
	repo := table_repo.NewMock([]table.Tisch{{ID: 1, Name: "Tisch 1", Status: table.ActiveStatus, UpdatedAt: time.Now().UTC()}}, nil)
	command := Command{TableRepo: repo}

	err := command.TischDeaktivieren(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	tbl, err := repo.GetTable(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error retrieving tisch, got %v", err)
	}
	if tbl.Status != table.InactiveStatus {
		t.Errorf("expected tisch status to be Inactive, got %v", tbl.Status)
	}
}

func TestTischDeaktivieren_NotFound(t *testing.T) {
	repo := table_repo.NewMock([]table.Tisch{}, db.ErrNotFound)
	command := Command{TableRepo: repo}

	err := command.TischDeaktivieren(context.Background(), 999)
	if err != ErrTischNotFound {
		t.Fatalf("expected ErrTischNotFound, got %v", err)
	}
}

func TestBestellungAufnehmen_KasseNichtGeoeffnet(t *testing.T) {
	ctx := context.Background()
	productMock := product_repo.NewMock([]product.Produkt{testProduct}, nil)
	productMock.AddVariant(testProduct.ID, testVariant)
	eventMock := kassenjournal_repo.NewMock(nil, nil)
	// no open KS set
	command := Command{
		TableRepo:           table_repo.NewMock([]table.Tisch{testActiveTisch}, nil),
		EventRepo:           eventMock,
		ProductRepo:         productMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(nil, nil), // no open KS
	}

	inputs := []BestellPositionInput{
		{ProduktID: testProduct.ID, VarianteID: testVariant.ID, Menge: 1},
	}

	err := command.BestellungAufnehmen(ctx, 1, "Test User", 1, inputs, "")
	if err != ErrKasseNichtGeoeffnet {
		t.Fatalf("expected ErrKasseNichtGeoeffnet, got %v", err)
	}
}

func TestBestellungAufnehmen_WithOCC(t *testing.T) {
	ctx := context.Background()
	productMock := product_repo.NewMock([]product.Produkt{testProduct}, nil)
	productMock.AddVariant(testProduct.ID, testVariant)
	command := newTestCommand([]table.Tisch{testActiveTisch}, []product.Produkt{testProduct})
	command.ProductRepo = productMock

	inputs := []BestellPositionInput{
		{ProduktID: testProduct.ID, VarianteID: testVariant.ID, Menge: 2},
	}

	err := command.BestellungAufnehmen(ctx, 1, "Test User", 1, inputs, "Testkommentar")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestBestellungAufnehmen_EnqueueArbeitsbonDruckauftraege(t *testing.T) {
	ctx := context.Background()
	productMock := product_repo.NewMock([]product.Produkt{testProduct}, nil)
	productMock.AddVariant(testProduct.ID, testVariant)

	eventMock := kassenjournal_repo.NewMock(nil, nil)
	stationMock := &mockDruckstationRepo{konfig: map[string]bondruckApp.Druckstation{
		"getraenk": {IP: "192.168.1.50", Bonmodus: "pro_position"},
	}}

	command := newTestCommandWithEventMock([]table.Tisch{testActiveTisch}, []product.Produkt{testProduct}, eventMock)
	command.ProductRepo = productMock
	command.DruckstationRepo = stationMock

	inputs := []BestellPositionInput{{ProduktID: testProduct.ID, VarianteID: testVariant.ID, Menge: 2}}

	err := command.BestellungAufnehmen(ctx, 1, "Test User", 1, inputs, "")
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
	productMock := product_repo.NewMock([]product.Produkt{testProduct}, nil)
	productMock.AddVariant(testProduct.ID, testVariant)
	eventMock := kassenjournal_repo.NewMockWithWriteErr(nil, db.ErrAlreadyExists)
	command := newTestCommandWithEventMock([]table.Tisch{testActiveTisch}, []product.Produkt{testProduct}, eventMock)
	command.ProductRepo = productMock

	inputs := []BestellPositionInput{
		{ProduktID: testProduct.ID, VarianteID: testVariant.ID, Menge: 1},
	}

	err := command.BestellungAufnehmen(ctx, 1, "Test User", 1, inputs, "")
	if err != ErrConflict {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

func TestBestellungAufnehmen_DeadlockMapsToConflict(t *testing.T) {
	ctx := context.Background()
	productMock := product_repo.NewMock([]product.Produkt{testProduct}, nil)
	productMock.AddVariant(testProduct.ID, testVariant)
	eventMock := kassenjournal_repo.NewMockWithWriteErr(nil, db.ErrConflict)
	command := newTestCommandWithEventMock([]table.Tisch{testActiveTisch}, []product.Produkt{testProduct}, eventMock)
	command.ProductRepo = productMock

	inputs := []BestellPositionInput{
		{ProduktID: testProduct.ID, VarianteID: testVariant.ID, Menge: 1},
	}

	err := command.BestellungAufnehmen(ctx, 1, "Test User", 1, inputs, "")
	if err != ErrConflict {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

// --- Invariant Tests ---

func TestBestellungAufnehmen_InactiveTisch(t *testing.T) {
	ctx := context.Background()
	productMock := product_repo.NewMock([]product.Produkt{testProduct}, nil)
	productMock.AddVariant(testProduct.ID, testVariant)
	command := newTestCommand([]table.Tisch{testInactiveTisch}, []product.Produkt{testProduct})
	command.ProductRepo = productMock

	inputs := []BestellPositionInput{
		{ProduktID: testProduct.ID, VarianteID: testVariant.ID, Menge: 1},
	}

	err := command.BestellungAufnehmen(ctx, 1, "Test User", testInactiveTisch.ID, inputs, "")
	if err != ErrTischNotActive {
		t.Fatalf("expected ErrTischNotActive, got %v", err)
	}
}

func TestZahlungKassieren_NonOrderedPosition(t *testing.T) {
	ctx := context.Background()
	// No order events exist — paying a non-existent position should fail
	command := newTestCommand([]table.Tisch{testActiveTisch}, nil)

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
		SaldoCents:            0,
		UnbezahltePositionen:  []kasse.Position{},
		AusstehendePositionen: []kasse.Position{},
		GesamtZahlungenCents:  350,
	})
	command := Command{
		TableRepo:           table_repo.NewMock([]table.Tisch{testActiveTisch}, nil),
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
		[]kasse.Position{{PositionID: "22222222-2222-4222-8222-222222222222", VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", Einzelpreis: 350, Menge: 1}},
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
			PositionID:   "22222222-2222-4222-8222-222222222222",
			VarianteID:   1,
			ProduktName:  "Cola",
			VarianteName: "0,5l",
			Kategorie:    "getraenk",
			Steuersatz:   "regel",
			Einzelpreis:  350,
			Menge:        1,
		}},
		LastEventVersion: 1,
	})

	command := Command{
		TableRepo:           table_repo.NewMock([]table.Tisch{testActiveTisch}, nil),
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
			PositionID:   "22222222-2222-4222-8222-222222222222",
			VarianteID:   1,
			ProduktName:  "Cola",
			VarianteName: "0,5l",
			Kategorie:    "getraenk",
			Steuersatz:   "regel",
			Einzelpreis:  350,
			Menge:        1,
		}},
		LastEventVersion: 3,
	})

	command := Command{
		TableRepo:           table_repo.NewMock([]table.Tisch{testActiveTisch}, nil),
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

func TestAusgabeBestaetigen_NonOrderedPosition(t *testing.T) {
	ctx := context.Background()
	// No order events exist — issuing a non-existent position should fail
	command := newTestCommand([]table.Tisch{testActiveTisch}, nil)

	fakeRefs := []kasse.PositionRef{
		{PositionID: "00000000-0000-0000-0000-000000000001", Menge: 1},
	}

	err := command.AusgabeBestaetigen(ctx, 1, "Test User", testActiveTisch.ID, fakeRefs, "")
	if err != ErrPositionNichtAusgebbar {
		t.Fatalf("expected ErrPositionNichtAusgebbar, got %v", err)
	}
}

func TestStornierungErteilen_AlreadyPaidPosition_Succeeds(t *testing.T) {
	ctx := context.Background()
	subject := kasse.TischSessionSubject(testKassensitzungNr, testActiveTisch.ID)

	orderEvent, _ := kasse.NewBestellungAufgenommenEvent(subject, 1, "Test User",
		[]kasse.Position{
			{VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", Einzelpreis: 350, Menge: 1},
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
			{PositionID: posID, VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", Einzelpreis: 350, Menge: 1},
		}, 350, "")

	eventMock := kassenjournal_repo.NewMock(nil, nil)
	eventMock.SetTischSession(subject, kasse.TischSession{
		SaldoCents:            0,
		UnbezahltePositionen:  []kasse.Position{},
		AusstehendePositionen: []kasse.Position{},
		GesamtZahlungenCents:  350,
	})
	orderEvent.Subject = subject
	paymentEvent.Subject = subject
	eventMock.AddEvent(orderEvent)
	eventMock.AddEvent(paymentEvent)

	command := Command{
		TableRepo:           table_repo.NewMock([]table.Tisch{testActiveTisch}, nil),
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

	orderEvent, _ := kasse.NewBestellungAufgenommenEvent(subject, 1, "Test User",
		[]kasse.Position{
			{VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", Einzelpreis: 350, Menge: 1},
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
			{PositionID: posID, VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", Einzelpreis: 350, Menge: 1},
		}, 350, "Test")

	eventMock := kassenjournal_repo.NewMock(nil, nil)
	eventMock.SetTischSession(subject, kasse.TischSession{
		SaldoCents:            0,
		UnbezahltePositionen:  []kasse.Position{},
		AusstehendePositionen: []kasse.Position{},
		GesamtZahlungenCents:  0,
	})
	orderEvent.Subject = subject
	cancelEvent.Subject = subject
	eventMock.AddEvent(orderEvent)
	eventMock.AddEvent(cancelEvent)

	command := Command{
		TableRepo:           table_repo.NewMock([]table.Tisch{testActiveTisch}, nil),
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
			{PositionID: "pos-1", VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", Einzelpreis: 350, Menge: 1},
		},
		AusstehendePositionen: []kasse.Position{
			{PositionID: "pos-1", VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", Einzelpreis: 350, Menge: 1},
		},
	})
	command := Command{
		TableRepo:           table_repo.NewMock([]table.Tisch{testActiveTisch}, nil),
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
// unbezahlt und ausstehend, für die Duplikat-Ablehnungstests.
func duplikatTestSession() kasse.TischSession {
	pos := kasse.Position{PositionID: "pos-1", VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", Einzelpreis: 350, Menge: 3}
	return kasse.TischSession{
		SaldoCents:            1050,
		UnbezahltePositionen:  []kasse.Position{pos},
		AusstehendePositionen: []kasse.Position{pos},
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
		TableRepo:           table_repo.NewMock([]table.Tisch{testActiveTisch}, nil),
		EventRepo:           eventMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
	}

	err := command.ZahlungKassieren(ctx, 1, "Test User", testActiveTisch.ID, duplikatRefs, "")
	if err != ErrPositionNichtBezahlbar {
		t.Fatalf("expected ErrPositionNichtBezahlbar, got %v", err)
	}
}

func TestAusgabeBestaetigen_DuplikatPositionRefs(t *testing.T) {
	ctx := context.Background()
	eventMock := kassenjournal_repo.NewMock(nil, nil)
	eventMock.SetTischSession(kasse.TischSessionSubject(testKassensitzungNr, testActiveTisch.ID), duplikatTestSession())
	command := Command{
		TableRepo:           table_repo.NewMock([]table.Tisch{testActiveTisch}, nil),
		EventRepo:           eventMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
	}

	err := command.AusgabeBestaetigen(ctx, 1, "Test User", testActiveTisch.ID, duplikatRefs, "")
	if err != ErrPositionNichtAusgebbar {
		t.Fatalf("expected ErrPositionNichtAusgebbar, got %v", err)
	}
}

func TestStornierungErteilen_DuplikatPositionRefs(t *testing.T) {
	ctx := context.Background()
	subject := kasse.TischSessionSubject(testKassensitzungNr, testActiveTisch.ID)

	orderEvent, _ := kasse.NewBestellungAufgenommenEvent(subject, 1, "Test User",
		[]kasse.Position{
			{VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", Einzelpreis: 350, Menge: 3},
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
		TableRepo:           table_repo.NewMock([]table.Tisch{testActiveTisch}, nil),
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
	quellTisch := table.Tisch{ID: 1, Name: "Tisch Quelle", Status: table.ActiveStatus}
	zielTisch := table.Tisch{ID: 2, Name: "Tisch Ziel", Status: table.ActiveStatus}

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
				Einzelpreis: 350,
				Menge:       2,
			},
		},
	})

	command := Command{
		TableRepo:           table_repo.NewMock([]table.Tisch{quellTisch, zielTisch}, nil),
		EventRepo:           eventMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
	}

	err := command.BestellungUmbuchen(ctx, 1, "Test User", quellTisch.ID, zielTisch.ID, []kasse.PositionRef{{PositionID: quellPositionID, Menge: 1}})
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
	if quellData.Kommentar != "Umbuchung auf Tisch Tisch Ziel" {
		t.Fatalf("unexpected source comment: %q", quellData.Kommentar)
	}
	if len(quellData.Positionen) != 1 {
		t.Fatalf("expected 1 source position, got %d", len(quellData.Positionen))
	}
	if quellData.Positionen[0].PositionID != quellPositionID {
		t.Fatalf("expected source position ID %q, got %q", quellPositionID, quellData.Positionen[0].PositionID)
	}
	if quellData.Positionen[0].Einzelpreis != 350 {
		t.Fatalf("expected source einzelpreis 350, got %d", quellData.Positionen[0].Einzelpreis)
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
	if zielData.Kommentar != "Umbuchung von Tisch Tisch Quelle" {
		t.Fatalf("unexpected target comment: %q", zielData.Kommentar)
	}
	if len(zielData.Positionen) != 1 {
		t.Fatalf("expected 1 target position, got %d", len(zielData.Positionen))
	}
	if zielData.Positionen[0].PositionID == quellPositionID {
		t.Fatalf("expected target position ID to be regenerated, but remained %q", zielData.Positionen[0].PositionID)
	}
	if zielData.Positionen[0].Einzelpreis != 350 {
		t.Fatalf("expected target einzelpreis 350, got %d", zielData.Positionen[0].Einzelpreis)
	}
}

func TestBestellungUmbuchen_KommentarWirdGekuerzt(t *testing.T) {
	ctx := context.Background()
	quellTisch := table.Tisch{ID: 1, Name: strings.Repeat("Q", 100), Status: table.ActiveStatus}
	zielTisch := table.Tisch{ID: 2, Name: strings.Repeat("Z", 100), Status: table.ActiveStatus}

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
			Einzelpreis: 350,
			Menge:       1,
		}},
	})

	command := Command{
		TableRepo:           table_repo.NewMock([]table.Tisch{quellTisch, zielTisch}, nil),
		EventRepo:           eventMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
	}

	err := command.BestellungUmbuchen(ctx, 1, "Test User", quellTisch.ID, zielTisch.ID, []kasse.PositionRef{{PositionID: quellPositionID, Menge: 1}})
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
	if !strings.HasPrefix(quellData.Kommentar, "Umbuchung auf Tisch ") {
		t.Fatalf("expected source comment prefix, got %q", quellData.Kommentar)
	}
	if !strings.HasPrefix(zielData.Kommentar, "Umbuchung von Tisch ") {
		t.Fatalf("expected target comment prefix, got %q", zielData.Kommentar)
	}
}

func TestBestellungUmbuchen_PositionNichtUmbuchbar(t *testing.T) {
	ctx := context.Background()
	quellTisch := table.Tisch{ID: 1, Name: "Tisch Quelle", Status: table.ActiveStatus}
	zielTisch := table.Tisch{ID: 2, Name: "Tisch Ziel", Status: table.ActiveStatus}

	eventMock := kassenjournal_repo.NewMock(nil, nil)
	quellSubject := kasse.TischSessionSubject(testKassensitzungNr, quellTisch.ID)
	eventMock.SetTischSession(quellSubject, kasse.TischSession{
		UnbezahltePositionen: []kasse.Position{{
			PositionID:   uuid.New().String(),
			VarianteID:   1,
			ProduktName:  "Cola",
			VarianteName: "0,5l",
			Kategorie:    "getraenk", Steuersatz: "regel",
			Einzelpreis: 350,
			Menge:       1,
		}},
	})

	command := Command{
		TableRepo:           table_repo.NewMock([]table.Tisch{quellTisch, zielTisch}, nil),
		EventRepo:           eventMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
	}

	err := command.BestellungUmbuchen(ctx, 1, "Test User", quellTisch.ID, zielTisch.ID, []kasse.PositionRef{{PositionID: uuid.New().String(), Menge: 1}})
	if err != ErrPositionNichtUmbuchbar {
		t.Fatalf("expected ErrPositionNichtUmbuchbar, got %v", err)
	}
}

func TestBestellungUmbuchen_GleicherTisch(t *testing.T) {
	err := Command{}.BestellungUmbuchen(context.Background(), 1, "Test User", 3, 3, []kasse.PositionRef{{PositionID: uuid.New().String(), Menge: 1}})
	if err != ErrUmbuchungGleicherTisch {
		t.Fatalf("expected ErrUmbuchungGleicherTisch, got %v", err)
	}
}

func TestBestellungUmbuchen_ZielTischNotActive(t *testing.T) {
	ctx := context.Background()
	quellTisch := table.Tisch{ID: 1, Name: "Tisch Quelle", Status: table.ActiveStatus}
	zielTisch := table.Tisch{ID: 2, Name: "Tisch Ziel", Status: table.InactiveStatus}

	command := Command{
		TableRepo:           table_repo.NewMock([]table.Tisch{quellTisch, zielTisch}, nil),
		EventRepo:           kassenjournal_repo.NewMock(nil, nil),
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
	}

	err := command.BestellungUmbuchen(ctx, 1, "Test User", quellTisch.ID, zielTisch.ID, []kasse.PositionRef{{PositionID: uuid.New().String(), Menge: 1}})
	if err != ErrTischNotActive {
		t.Fatalf("expected ErrTischNotActive, got %v", err)
	}
}

func TestBestellungUmbuchen_ZielTischNotFound(t *testing.T) {
	ctx := context.Background()
	quellTisch := table.Tisch{ID: 1, Name: "Tisch Quelle", Status: table.ActiveStatus}

	command := Command{
		TableRepo: &umbuchungTableRepoMock{tables: map[int]table.Tisch{
			quellTisch.ID: quellTisch,
		}},
		EventRepo:           kassenjournal_repo.NewMock(nil, nil),
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
	}

	err := command.BestellungUmbuchen(ctx, 1, "Test User", quellTisch.ID, 99, []kasse.PositionRef{{PositionID: uuid.New().String(), Menge: 1}})
	if err != ErrTischNotFound {
		t.Fatalf("expected ErrTischNotFound, got %v", err)
	}
}

func TestBestellungUmbuchen_KasseNichtGeoeffnet(t *testing.T) {
	ctx := context.Background()
	command := Command{
		TableRepo:           table_repo.NewMock([]table.Tisch{testActiveTisch}, nil),
		EventRepo:           kassenjournal_repo.NewMock(nil, nil),
		KassensitzungenRepo: kassensitzungen_repo.NewMock(nil, nil),
	}

	err := command.BestellungUmbuchen(ctx, 1, "Test User", 1, 2, []kasse.PositionRef{{PositionID: uuid.New().String(), Menge: 1}})
	if err != ErrKasseNichtGeoeffnet {
		t.Fatalf("expected ErrKasseNichtGeoeffnet, got %v", err)
	}
}

func TestBestellungUmbuchen_Conflict(t *testing.T) {
	ctx := context.Background()
	quellTisch := table.Tisch{ID: 1, Name: "Tisch Quelle", Status: table.ActiveStatus}
	zielTisch := table.Tisch{ID: 2, Name: "Tisch Ziel", Status: table.ActiveStatus}

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
			Einzelpreis: 350,
			Menge:       1,
		}},
	})

	command := Command{
		TableRepo:           table_repo.NewMock([]table.Tisch{quellTisch, zielTisch}, nil),
		EventRepo:           eventMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
	}

	err := command.BestellungUmbuchen(ctx, 1, "Test User", quellTisch.ID, zielTisch.ID, []kasse.PositionRef{{PositionID: quellPositionID, Menge: 1}})
	if err != ErrConflict {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

func TestKassenbelegDrucken_SuccessAndReprint(t *testing.T) {
	ctx := context.Background()
	subject := kasse.TischSessionSubject(testKassensitzungNr, testActiveTisch.ID)

	zahlungEvent, err := kasse.NewZahlungKassiertEvent(subject, 1, "Test User", []kasse.Position{
		{
			PositionID:   "11111111-1111-1111-1111-111111111111",
			VarianteID:   1,
			ProduktName:  "Cola",
			VarianteName: "0,5l",
			Kategorie:    "getraenk", Steuersatz: "regel",
			Einzelpreis: 350,
			Menge:       2,
		},
	}, 700, "")
	if err != nil {
		t.Fatalf("expected no event error, got %v", err)
	}

	var eventData struct {
		ZahlungID string `json:"zahlungId"`
	}
	if err := json.Unmarshal(zahlungEvent.Data, &eventData); err != nil {
		t.Fatalf("expected no unmarshal error, got %v", err)
	}

	eventMock := kassenjournal_repo.NewMock(nil, nil)
	eventMock.AddEvent(zahlungEvent)

	auftragMock := &mockDruckauftragRepo{}
	settingsMock := &mockSettingsRepo{
		betreiber: settings.Betreiber{
			Vereinsname: "SV Musterstadt",
			Strasse:     "Musterstrasse 1",
			Plz:         "12345",
			Ort:         "Musterstadt",
			UpdatedAt:   time.Now(),
		},
		kassenident: settings.Kassenidentitaet{
			Seriennummer: uuid.MustParse("2e00c5d4-7adb-4f63-84d6-a34235f2b0f4"),
			AngelegtAm:   time.Now(),
		},
	}

	command := Command{
		EventRepo:           eventMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
		DruckstationRepo:    &mockDruckstationRepo{konfig: kassenbelegStationen},
		SettingsRepo:        settingsMock,
		DruckauftragRepo:    auftragMock,
		TSERepo:             &mockTSEAuftragRepo{},
	}

	status, err := command.KassenbelegDrucken(ctx, testActiveTisch.ID, eventData.ZahlungID, "", "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if status != BelegStatusEingereiht {
		t.Fatalf("expected status eingereiht, got %q", status)
	}
	status, err = command.KassenbelegDrucken(ctx, testActiveTisch.ID, eventData.ZahlungID, "", "")
	if err != nil {
		t.Fatalf("expected no reprint error, got %v", err)
	}
	if status != BelegStatusEingereiht {
		t.Fatalf("expected reprint status eingereiht, got %q", status)
	}

	if len(auftragMock.enqueued) != 2 {
		t.Fatalf("expected 2 enqueued auftraege, got %d", len(auftragMock.enqueued))
	}
	if auftragMock.enqueued[0].BonArt != "kassenbeleg" {
		t.Fatalf("expected bon_art kassenbeleg, got %s", auftragMock.enqueued[0].BonArt)
	}
	if auftragMock.enqueued[0].ZielIP != "192.168.1.80" {
		t.Fatalf("expected ziel_ip 192.168.1.80, got %s", auftragMock.enqueued[0].ZielIP)
	}
	if !strings.HasPrefix(auftragMock.enqueued[0].Referenz, "zahlung-kassiert:") {
		t.Fatalf("expected zahlung-kassiert referenz, got %s", auftragMock.enqueued[0].Referenz)
	}
}

func TestKassenbelegDrucken_ContainsSteuerkennzeichenUndSteuermatrix(t *testing.T) {
	ctx := context.Background()
	subject := kasse.TischSessionSubject(testKassensitzungNr, testActiveTisch.ID)

	zahlungEvent, err := kasse.NewZahlungKassiertEvent(subject, 1, "Test User", []kasse.Position{
		{
			PositionID:   "11111111-1111-1111-1111-111111111111",
			VarianteID:   1,
			ProduktName:  "Cola",
			VarianteName: "0,5l",
			Kategorie:    "getraenk", Steuersatz: "regel",
			Einzelpreis: 350,
			Menge:       2,
		},
		{
			PositionID:   "22222222-2222-2222-2222-222222222222",
			VarianteID:   2,
			ProduktName:  "Brezel",
			VarianteName: "normal",
			Kategorie:    "essen", Steuersatz: "ermaessigt",
			Einzelpreis: 300,
			Menge:       1,
		},
	}, 1000, "")
	if err != nil {
		t.Fatalf("expected no event error, got %v", err)
	}

	var eventData struct {
		ZahlungID string `json:"zahlungId"`
	}
	if err := json.Unmarshal(zahlungEvent.Data, &eventData); err != nil {
		t.Fatalf("expected no unmarshal error, got %v", err)
	}

	eventMock := kassenjournal_repo.NewMock(nil, nil)
	eventMock.AddEvent(zahlungEvent)

	auftragMock := &mockDruckauftragRepo{}
	settingsMock := &mockSettingsRepo{
		betreiber: settings.Betreiber{
			Vereinsname: "SV Musterstadt",
			Strasse:     "Musterstrasse 1",
			Plz:         "12345",
			Ort:         "Musterstadt",
			UpdatedAt:   time.Now(),
		},
		kassenident: settings.Kassenidentitaet{
			Seriennummer: uuid.MustParse("2e00c5d4-7adb-4f63-84d6-a34235f2b0f4"),
			AngelegtAm:   time.Now(),
		},
	}

	command := Command{
		EventRepo:           eventMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
		DruckstationRepo:    &mockDruckstationRepo{konfig: kassenbelegStationen},
		SettingsRepo:        settingsMock,
		DruckauftragRepo:    auftragMock,
		TSERepo:             &mockTSEAuftragRepo{},
	}

	if _, err := command.KassenbelegDrucken(ctx, testActiveTisch.ID, eventData.ZahlungID, "", ""); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(auftragMock.enqueued) != 1 {
		t.Fatalf("expected exactly 1 enqueued auftrag, got %d", len(auftragMock.enqueued))
	}

	payload, err := base64.StdEncoding.DecodeString(auftragMock.enqueued[0].Payload)
	if err != nil {
		t.Fatalf("expected base64 payload, got decode error: %v", err)
	}

	got := string(payload)
	checks := []string{
		"GESAMT: 10,00 EUR",
		"= 7,00 EUR (A)",
		"= 3,00 EUR (B)",
		"Steueraufteilung:",
		"A: Netto 5,88 EUR, Steuer 1,12 EUR, Brutto 7,00 EUR",
		"B: Netto 2,80 EUR, Steuer 0,20 EUR, Brutto 3,00 EUR",
	}

	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Fatalf("kassenbeleg payload enthaelt %q nicht; got:\n%q", check, got)
		}
	}
}

func TestKassenbelegDrucken_MitSignaturAmAuftrag_ContainsTSEBlock(t *testing.T) {
	ctx := context.Background()
	subject := kasse.TischSessionSubject(testKassensitzungNr, testActiveTisch.ID)

	zahlungEvent, err := kasse.NewZahlungKassiertEvent(subject, 1, "Test User", []kasse.Position{
		{
			PositionID:   "11111111-1111-1111-1111-111111111111",
			VarianteID:   1,
			ProduktName:  "Cola",
			VarianteName: "0,5l",
			Kategorie:    "getraenk", Steuersatz: "regel",
			Einzelpreis: 350,
			Menge:       1,
		},
	}, 350, "")
	if err != nil {
		t.Fatalf("expected no event error, got %v", err)
	}

	var eventData struct {
		ZahlungID string `json:"zahlungId"`
	}
	if err := json.Unmarshal(zahlungEvent.Data, &eventData); err != nil {
		t.Fatalf("expected no unmarshal error, got %v", err)
	}

	eventMock := kassenjournal_repo.NewMock(nil, nil)
	eventMock.AddEvent(zahlungEvent) // Event-ID 1

	// Die Signatur liegt am quittierten Auftrag — die einzige Signaturquelle.
	tseRepo := &mockTSEAuftragRepo{staende: map[int]tse_repo.SignaturauftragStand{
		1: {Status: tse_repo.StatusErledigt, Signatur: &tse.Signatur{
			TransaktionNummer: 3001,
			SignaturZaehler:   77,
			TSESeriennummer:   "SW-TSE-SN-0042",
			LogTimeStart:      time.Date(2026, 6, 10, 18, 0, 1, 0, time.UTC),
			LogTimeEnd:        time.Date(2026, 6, 10, 18, 0, 3, 0, time.UTC),
			Signatur:          "SIG-XYZ",
			QRCodeData:        "V0;XYZ",
		}},
	}}

	auftragMock := &mockDruckauftragRepo{}
	settingsMock := &mockSettingsRepo{
		betreiber: settings.Betreiber{
			Vereinsname: "SV Musterstadt",
			Strasse:     "Musterstrasse 1",
			Plz:         "12345",
			Ort:         "Musterstadt",
			UpdatedAt:   time.Now(),
		},
		kassenident: settings.Kassenidentitaet{
			Seriennummer: uuid.MustParse("2e00c5d4-7adb-4f63-84d6-a34235f2b0f4"),
			AngelegtAm:   time.Now(),
		},
	}

	command := Command{
		EventRepo:           eventMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
		DruckstationRepo:    &mockDruckstationRepo{konfig: kassenbelegStationen},
		SettingsRepo:        settingsMock,
		DruckauftragRepo:    auftragMock,
		TSERepo:             tseRepo,
	}

	status, err := command.KassenbelegDrucken(ctx, testActiveTisch.ID, eventData.ZahlungID, "", "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if status != BelegStatusEingereiht {
		t.Fatalf("expected status eingereiht, got %q", status)
	}

	if len(auftragMock.enqueued) != 1 {
		t.Fatalf("expected exactly 1 enqueued auftrag, got %d", len(auftragMock.enqueued))
	}

	payload, err := base64.StdEncoding.DecodeString(auftragMock.enqueued[0].Payload)
	if err != nil {
		t.Fatalf("expected base64 payload, got decode error: %v", err)
	}

	got := string(payload)
	checks := []string{
		"TSE-Daten:",
		"TSE-Transaktion: 3001",
		"Signaturzaehler: 77",
		"TSE-Seriennummer: SW-TSE-SN-0042",
		"TSE-Start: 10.06.2026 18:00:01",
		"TSE-Ende: 10.06.2026 18:00:03",
		"Signatur: SIG-XYZ",
	}

	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Fatalf("kassenbeleg payload enthaelt %q nicht; got:\n%q", check, got)
		}
	}
}

func TestKassenbelegDrucken_Tischzahlung_WithErsteBestellungKlartext(t *testing.T) {
	ctx := context.Background()
	subject := kasse.TischSessionSubject(testKassensitzungNr, testActiveTisch.ID)

	zahlungEvent, err := kasse.NewZahlungKassiertEvent(subject, 1, "Test User", []kasse.Position{
		{
			PositionID:   "11111111-1111-4111-8111-111111111111",
			VarianteID:   1,
			ProduktName:  "Cola",
			VarianteName: "0,5l",
			Kategorie:    "getraenk", Steuersatz: "regel",
			Einzelpreis: 350,
			Menge:       1,
		},
	}, 350, "")
	if err != nil {
		t.Fatalf("expected no event error, got %v", err)
	}

	var eventData struct {
		ZahlungID string `json:"zahlungId"`
	}
	if err := json.Unmarshal(zahlungEvent.Data, &eventData); err != nil {
		t.Fatalf("expected no unmarshal error, got %v", err)
	}

	ersteBestellung := time.Date(2026, 5, 1, 18, 1, 0, 0, time.UTC)

	eventMock := kassenjournal_repo.NewMock(nil, nil)
	eventMock.AddEvent(zahlungEvent)
	eventMock.SetTischSession(subject, kasse.TischSession{
		Subject:                subject,
		TischID:                testActiveTisch.ID,
		KassensitzungNr:        testKassensitzungNr,
		ErsteBestellungLogTime: &ersteBestellung,
	})

	auftragMock := &mockDruckauftragRepo{}
	settingsMock := &mockSettingsRepo{
		betreiber: settings.Betreiber{
			Vereinsname: "SV Musterstadt",
			Strasse:     "Musterstrasse 1",
			Plz:         "12345",
			Ort:         "Musterstadt",
			UpdatedAt:   time.Now(),
		},
		kassenident: settings.Kassenidentitaet{
			Seriennummer: uuid.MustParse("2e00c5d4-7adb-4f63-84d6-a34235f2b0f4"),
			AngelegtAm:   time.Now(),
		},
	}

	command := Command{
		EventRepo:           eventMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
		DruckstationRepo:    &mockDruckstationRepo{konfig: kassenbelegStationen},
		SettingsRepo:        settingsMock,
		DruckauftragRepo:    auftragMock,
		TSERepo:             &mockTSEAuftragRepo{},
	}

	if _, err := command.KassenbelegDrucken(ctx, testActiveTisch.ID, eventData.ZahlungID, "", ""); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	payload, err := base64.StdEncoding.DecodeString(auftragMock.enqueued[0].Payload)
	if err != nil {
		t.Fatalf("expected base64 payload, got decode error: %v", err)
	}

	got := string(payload)
	if !strings.Contains(got, "Erste Bestellung: 01.05.2026 18:01:00") {
		t.Fatalf("expected first order klartext in table receipt, got:\n%q", got)
	}
}

// Der Beleg-Abruf antwortet sofort mit dem Signaturstatus: Solange der Auftrag
// nicht quittiert ist, entsteht kein Druckauftrag (ausstehend, die UI fasst
// nach); nach der Quittierung liefert derselbe Aufruf den Beleg mit dem
// TSE-Abschnitt aus den Signaturspalten des Auftrags.
func TestKassenbelegDrucken_AusstehendDannEingereiht(t *testing.T) {
	ctx := context.Background()
	subject := kasse.TischSessionSubject(testKassensitzungNr, testActiveTisch.ID)

	zahlungEvent, err := kasse.NewZahlungKassiertEvent(subject, 1, "Test User", []kasse.Position{{
		PositionID:   "11111111-1111-4111-8111-111111111111",
		VarianteID:   1,
		ProduktName:  "Cola",
		VarianteName: "0,5l",
		Kategorie:    "getraenk",
		Steuersatz:   "regel",
		Einzelpreis:  350,
		Menge:        1,
	}}, 350, "")
	if err != nil {
		t.Fatalf("expected no event error, got %v", err)
	}

	var eventData struct {
		ZahlungID string `json:"zahlungId"`
	}
	if err := json.Unmarshal(zahlungEvent.Data, &eventData); err != nil {
		t.Fatalf("expected no unmarshal error, got %v", err)
	}

	eventMock := kassenjournal_repo.NewMock(nil, nil)
	eventMock.AddEvent(zahlungEvent) // Event-ID 1

	// Auftrag existiert, aber der Worker hat noch nicht quittiert.
	tseRepo := &mockTSEAuftragRepo{staende: map[int]tse_repo.SignaturauftragStand{
		1: {Status: tse_repo.StatusOffen, ErstelltAm: time.Now().UTC()},
	}}

	auftragMock := &mockDruckauftragRepo{}
	command := Command{
		EventRepo:           eventMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
		DruckstationRepo:    &mockDruckstationRepo{konfig: kassenbelegStationen},
		SettingsRepo:        belegTestSettingsMock(),
		DruckauftragRepo:    auftragMock,
		TSERepo:             tseRepo,
	}

	status, err := command.KassenbelegDrucken(ctx, testActiveTisch.ID, eventData.ZahlungID, "", "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if status != BelegStatusAusstehend {
		t.Fatalf("expected status ausstehend, got %q", status)
	}
	if len(auftragMock.enqueued) != 0 {
		t.Fatalf("expected no druckauftrag while signature is pending, got %d", len(auftragMock.enqueued))
	}

	// Der Worker quittiert — der nächste Abruf liefert den Beleg mit Signatur.
	tseRepo.staende[1] = tse_repo.SignaturauftragStand{Status: tse_repo.StatusErledigt, Signatur: &tse.Signatur{
		TransaktionNummer: 3002,
		SignaturZaehler:   78,
		TSESeriennummer:   "SW-TSE-SN-0043",
		LogTimeStart:      time.Date(2026, 6, 10, 18, 10, 1, 0, time.UTC),
		LogTimeEnd:        time.Date(2026, 6, 10, 18, 10, 3, 0, time.UTC),
		Signatur:          "SIG-NACHGEHOLT",
		QRCodeData:        "V0;NACHGEHOLT",
	}}

	status, err = command.KassenbelegDrucken(ctx, testActiveTisch.ID, eventData.ZahlungID, "", "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if status != BelegStatusEingereiht {
		t.Fatalf("expected status eingereiht, got %q", status)
	}
	if len(auftragMock.enqueued) != 1 {
		t.Fatalf("expected exactly 1 enqueued auftrag, got %d", len(auftragMock.enqueued))
	}

	payload, err := base64.StdEncoding.DecodeString(auftragMock.enqueued[0].Payload)
	if err != nil {
		t.Fatalf("expected base64 payload, got decode error: %v", err)
	}

	got := string(payload)
	if !strings.Contains(got, "TSE-Daten:") {
		t.Fatalf("expected TSE block from auftrag signature, got:\n%q", got)
	}
	if !strings.Contains(got, "SIG-NACHGEHOLT") {
		t.Fatalf("expected auftrag signature in payload, got:\n%q", got)
	}
}

func TestKassenbelegDrucken_ZahlungNichtGefunden(t *testing.T) {
	ctx := context.Background()
	command := Command{
		EventRepo:           kassenjournal_repo.NewMock(nil, nil),
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
		SettingsRepo:        &mockSettingsRepo{},
		DruckstationRepo:    &mockDruckstationRepo{},
		DruckauftragRepo:    &mockDruckauftragRepo{},
		TSERepo:             &mockTSEAuftragRepo{},
	}

	_, err := command.KassenbelegDrucken(ctx, testActiveTisch.ID, "11111111-1111-1111-1111-111111111111", "", "")
	if err != ErrZahlungNichtGefunden {
		t.Fatalf("expected ErrZahlungNichtGefunden, got %v", err)
	}
}

func TestKassenbelegDrucken_KassenbelegDruckerNichtKonfiguriert(t *testing.T) {
	ctx := context.Background()
	subject := kasse.TischSessionSubject(testKassensitzungNr, testActiveTisch.ID)

	zahlungEvent, err := kasse.NewZahlungKassiertEvent(subject, 1, "Test User", []kasse.Position{
		{
			PositionID:   "11111111-1111-1111-1111-111111111111",
			VarianteID:   1,
			ProduktName:  "Cola",
			VarianteName: "0,5l",
			Kategorie:    "getraenk", Steuersatz: "regel",
			Einzelpreis: 350,
			Menge:       1,
		},
	}, 350, "")
	if err != nil {
		t.Fatalf("expected no event error, got %v", err)
	}

	var eventData struct {
		ZahlungID string `json:"zahlungId"`
	}
	if err := json.Unmarshal(zahlungEvent.Data, &eventData); err != nil {
		t.Fatalf("expected no unmarshal error, got %v", err)
	}

	eventMock := kassenjournal_repo.NewMock(nil, nil)
	eventMock.AddEvent(zahlungEvent)

	command := Command{
		EventRepo:           eventMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
		SettingsRepo:        &mockSettingsRepo{},
		DruckstationRepo:    &mockDruckstationRepo{},
		DruckauftragRepo:    &mockDruckauftragRepo{},
		TSERepo:             &mockTSEAuftragRepo{},
	}

	if _, err := command.KassenbelegDrucken(ctx, testActiveTisch.ID, eventData.ZahlungID, "", ""); err != ErrKassenbelegDruckerNichtKonfiguriert {
		t.Fatalf("expected ErrKassenbelegDruckerNichtKonfiguriert, got %v", err)
	}
}

func TestKassenbelegDrucken_Direktverkauf_ExactlyOneAuftrag(t *testing.T) {
	ctx := context.Background()
	verkaufID := uuid.New().String()
	subject := kasse.DirektverkaufSubject(testKassensitzungNr, verkaufID)

	verkaufEvent, err := kasse.NewDirektverkaufGetaetigtEvent(subject, verkaufID, 1, "Test User", []kasse.Position{
		{
			VarianteID:   1,
			ProduktName:  "Cola",
			VarianteName: "0,5l",
			Kategorie:    "getraenk", Steuersatz: "regel",
			Einzelpreis: 350,
			Menge:       2,
		},
	}, "")
	if err != nil {
		t.Fatalf("expected no event error, got %v", err)
	}

	eventMock := kassenjournal_repo.NewMock(nil, nil)
	eventMock.AddEvent(verkaufEvent)

	auftragMock := &mockDruckauftragRepo{}
	settingsMock := &mockSettingsRepo{
		betreiber: settings.Betreiber{
			Vereinsname: "SV Musterstadt",
			Strasse:     "Musterstrasse 1",
			Plz:         "12345",
			Ort:         "Musterstadt",
			UpdatedAt:   time.Now(),
		},
		kassenident: settings.Kassenidentitaet{
			Seriennummer: uuid.MustParse("2e00c5d4-7adb-4f63-84d6-a34235f2b0f4"),
			AngelegtAm:   time.Now(),
		},
	}

	command := Command{
		EventRepo:           eventMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
		DruckstationRepo:    &mockDruckstationRepo{konfig: kassenbelegStationen},
		SettingsRepo:        settingsMock,
		DruckauftragRepo:    auftragMock,
		TSERepo:             &mockTSEAuftragRepo{},
	}

	if _, err := command.KassenbelegDrucken(ctx, 0, "", verkaufID, ""); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(auftragMock.enqueued) != 1 {
		t.Fatalf("expected exactly 1 enqueued auftrag, got %d", len(auftragMock.enqueued))
	}
	if auftragMock.enqueued[0].BonArt != "kassenbeleg" {
		t.Fatalf("expected bon_art kassenbeleg, got %s", auftragMock.enqueued[0].BonArt)
	}
	if !strings.HasPrefix(auftragMock.enqueued[0].Referenz, "direktverkauf-getaetigt:") {
		t.Fatalf("expected direktverkauf-getaetigt referenz, got %s", auftragMock.enqueued[0].Referenz)
	}

	payload, err := base64.StdEncoding.DecodeString(auftragMock.enqueued[0].Payload)
	if err != nil {
		t.Fatalf("expected base64 payload, got decode error: %v", err)
	}
	if strings.Contains(string(payload), "Erste Bestellung:") {
		t.Fatalf("Direktverkauf-Beleg darf keinen Durchbedienen-Klarschriftzeitpunkt enthalten, got:\n%q", string(payload))
	}
}

func TestKassenbelegDrucken_Direktverkauf_NichtGefunden(t *testing.T) {
	ctx := context.Background()
	command := Command{
		EventRepo:           kassenjournal_repo.NewMock(nil, nil),
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
		SettingsRepo:        &mockSettingsRepo{},
		DruckstationRepo:    &mockDruckstationRepo{},
		DruckauftragRepo:    &mockDruckauftragRepo{},
		TSERepo:             &mockTSEAuftragRepo{},
	}

	_, err := command.KassenbelegDrucken(ctx, 0, "", uuid.New().String(), "")
	if err != ErrVerkaufNichtGefunden {
		t.Fatalf("expected ErrVerkaufNichtGefunden, got %v", err)
	}
}

func TestKassenbelegDrucken_Direktverkauf_KassenbelegDruckerNichtKonfiguriert(t *testing.T) {
	ctx := context.Background()
	verkaufID := uuid.New().String()
	subject := kasse.DirektverkaufSubject(testKassensitzungNr, verkaufID)

	verkaufEvent, err := kasse.NewDirektverkaufGetaetigtEvent(subject, verkaufID, 1, "Test User", []kasse.Position{
		{
			VarianteID:   1,
			ProduktName:  "Cola",
			VarianteName: "0,5l",
			Kategorie:    "getraenk", Steuersatz: "regel",
			Einzelpreis: 350,
			Menge:       1,
		},
	}, "")
	if err != nil {
		t.Fatalf("expected no event error, got %v", err)
	}

	eventMock := kassenjournal_repo.NewMock(nil, nil)
	eventMock.AddEvent(verkaufEvent)

	command := Command{
		EventRepo:           eventMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
		SettingsRepo:        &mockSettingsRepo{},
		DruckstationRepo:    &mockDruckstationRepo{},
		DruckauftragRepo:    &mockDruckauftragRepo{},
		TSERepo:             &mockTSEAuftragRepo{},
	}

	if _, err := command.KassenbelegDrucken(ctx, 0, "", verkaufID, ""); err != ErrKassenbelegDruckerNichtKonfiguriert {
		t.Fatalf("expected ErrKassenbelegDruckerNichtKonfiguriert, got %v", err)
	}
}

func belegTestSettingsMock() *mockSettingsRepo {
	return &mockSettingsRepo{
		betreiber: settings.Betreiber{
			Vereinsname: "SV Musterstadt",
			Strasse:     "Musterstrasse 1",
			Plz:         "12345",
			Ort:         "Musterstadt",
			UpdatedAt:   time.Now(),
		},
		kassenident: settings.Kassenidentitaet{
			Seriennummer: uuid.MustParse("2e00c5d4-7adb-4f63-84d6-a34235f2b0f4"),
			AngelegtAm:   time.Now(),
		},
	}
}

func TestKassenbelegDrucken_Direktverkauf_MitSignaturAmAuftrag(t *testing.T) {
	ctx := context.Background()
	verkaufID := uuid.New().String()
	subject := kasse.DirektverkaufSubject(testKassensitzungNr, verkaufID)

	verkaufEvent, err := kasse.NewDirektverkaufGetaetigtEvent(subject, verkaufID, 1, "Test User", []kasse.Position{{
		VarianteID:   1,
		ProduktName:  "Cola",
		VarianteName: "0,5l",
		Kategorie:    "getraenk", Steuersatz: "regel",
		Einzelpreis: 350,
		Menge:       2,
	}}, "")
	if err != nil {
		t.Fatalf("expected no event error, got %v", err)
	}

	eventMock := kassenjournal_repo.NewMock(nil, nil)
	eventMock.AddEvent(verkaufEvent) // Event-ID 1

	tseRepo := &mockTSEAuftragRepo{staende: map[int]tse_repo.SignaturauftragStand{
		1: {Status: tse_repo.StatusErledigt, Signatur: &tse.Signatur{
			TransaktionNummer: 4001,
			SignaturZaehler:   99,
			TSESeriennummer:   "SW-TSE-SN-0044",
			LogTimeStart:      time.Date(2026, 6, 10, 19, 0, 1, 0, time.UTC),
			LogTimeEnd:        time.Date(2026, 6, 10, 19, 0, 3, 0, time.UTC),
			Signatur:          "SIG-DIREKTVERKAUF",
			QRCodeData:        "V0;DIREKTVERKAUF",
		}},
	}}

	auftragMock := &mockDruckauftragRepo{}
	command := Command{
		EventRepo:           eventMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
		DruckstationRepo:    &mockDruckstationRepo{konfig: kassenbelegStationen},
		SettingsRepo:        belegTestSettingsMock(),
		DruckauftragRepo:    auftragMock,
		TSERepo:             tseRepo,
	}

	if _, err := command.KassenbelegDrucken(ctx, 0, "", verkaufID, ""); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	payload, err := base64.StdEncoding.DecodeString(auftragMock.enqueued[0].Payload)
	if err != nil {
		t.Fatalf("expected base64 payload, got decode error: %v", err)
	}

	got := string(payload)
	for _, check := range []string{"TSE-Daten:", "TSE-Transaktion: 4001", "Signaturzaehler: 99", "TSE-Seriennummer: SW-TSE-SN-0044", "SIG-DIREKTVERKAUF", "V0;DIREKTVERKAUF"} {
		if !strings.Contains(got, check) {
			t.Fatalf("expected %q in direktverkauf receipt, got:\n%q", check, got)
		}
	}
}

// Solange der Signaturauftrag des Direktverkaufs nicht quittiert ist, antwortet
// der Beleg-Abruf mit ausstehend und legt keinen Druckauftrag an.
func TestKassenbelegDrucken_Direktverkauf_SignaturAusstehend_KeinDruckauftrag(t *testing.T) {
	ctx := context.Background()
	verkaufID := uuid.New().String()
	subject := kasse.DirektverkaufSubject(testKassensitzungNr, verkaufID)

	verkaufEvent, err := kasse.NewDirektverkaufGetaetigtEvent(subject, verkaufID, 1, "Test User", []kasse.Position{{
		VarianteID:   1,
		ProduktName:  "Cola",
		VarianteName: "0,5l",
		Kategorie:    "getraenk", Steuersatz: "regel",
		Einzelpreis: 350,
		Menge:       1,
	}}, "")
	if err != nil {
		t.Fatalf("expected no event error, got %v", err)
	}

	eventMock := kassenjournal_repo.NewMock(nil, nil)
	eventMock.AddEvent(verkaufEvent) // Event-ID 1

	tseRepo := &mockTSEAuftragRepo{staende: map[int]tse_repo.SignaturauftragStand{
		1: {Status: tse_repo.StatusOffen, ErstelltAm: time.Now().UTC()},
	}}

	auftragMock := &mockDruckauftragRepo{}
	command := Command{
		EventRepo:           eventMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
		DruckstationRepo:    &mockDruckstationRepo{konfig: kassenbelegStationen},
		SettingsRepo:        belegTestSettingsMock(),
		DruckauftragRepo:    auftragMock,
		TSERepo:             tseRepo,
	}

	status, err := command.KassenbelegDrucken(ctx, 0, "", verkaufID, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if status != BelegStatusAusstehend {
		t.Fatalf("expected status ausstehend, got %q", status)
	}
	if len(auftragMock.enqueued) != 0 {
		t.Fatalf("expected no druckauftrag while signature is pending, got %d", len(auftragMock.enqueued))
	}
}

func TestKassenbelegDrucken_DirektverkaufStorno_DruckbarAlsStornobeleg(t *testing.T) {
	ctx := context.Background()
	verkaufID := uuid.New().String()
	subject := kasse.DirektverkaufSubject(testKassensitzungNr, verkaufID)

	verkaufEvent, err := kasse.NewDirektverkaufGetaetigtEvent(subject, verkaufID, 1, "Test User", []kasse.Position{{
		VarianteID:   1,
		ProduktName:  "Cola",
		VarianteName: "0,5l",
		Kategorie:    "getraenk", Steuersatz: "regel",
		Einzelpreis: 350,
		Menge:       2,
	}}, "")
	if err != nil {
		t.Fatalf("expected no event error, got %v", err)
	}

	stornoEvent, err := kasse.NewDirektverkaufStorniertEvent(subject, verkaufID, 2, "Leitung", []kasse.Position{{
		PositionID:   "11111111-1111-4111-8111-111111111111",
		VarianteID:   1,
		ProduktName:  "Cola",
		VarianteName: "0,5l",
		Kategorie:    "getraenk", Steuersatz: "regel",
		Einzelpreis: 350,
		Menge:       2,
	}}, 700, "Rückgabe")
	if err != nil {
		t.Fatalf("expected no storno event error, got %v", err)
	}

	var stornoData kasse.DirektverkaufStorniertV1Data
	if err := json.Unmarshal(stornoEvent.Data, &stornoData); err != nil {
		t.Fatalf("expected no unmarshal error, got %v", err)
	}

	eventMock := kassenjournal_repo.NewMock(nil, nil)
	eventMock.AddEvent(verkaufEvent) // Event-ID 1
	eventMock.AddEvent(stornoEvent)  // Event-ID 2

	// Die Signatur des Storno-Vorgangs liegt am quittierten Auftrag.
	tseRepo := &mockTSEAuftragRepo{staende: map[int]tse_repo.SignaturauftragStand{
		2: {Status: tse_repo.StatusErledigt, Signatur: &tse.Signatur{
			TransaktionNummer: 4003,
			SignaturZaehler:   101,
			TSESeriennummer:   "SW-TSE-SN-0044",
			LogTimeStart:      time.Date(2026, 6, 10, 19, 20, 1, 0, time.UTC),
			LogTimeEnd:        time.Date(2026, 6, 10, 19, 20, 3, 0, time.UTC),
			Signatur:          "SIG-STORNO",
			QRCodeData:        "V0;STORNO",
		}},
	}}

	auftragMock := &mockDruckauftragRepo{}
	command := Command{
		EventRepo:           eventMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
		DruckstationRepo:    &mockDruckstationRepo{konfig: kassenbelegStationen},
		SettingsRepo:        belegTestSettingsMock(),
		DruckauftragRepo:    auftragMock,
		TSERepo:             tseRepo,
	}

	if _, err := command.KassenbelegDrucken(ctx, 0, "", verkaufID, stornoData.StornierungID); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(auftragMock.enqueued) != 1 {
		t.Fatalf("expected exactly 1 enqueued auftrag, got %d", len(auftragMock.enqueued))
	}
	if !strings.HasPrefix(auftragMock.enqueued[0].Referenz, "direktverkauf-storniert:") {
		t.Fatalf("expected direktverkauf-storniert referenz, got %s", auftragMock.enqueued[0].Referenz)
	}

	payload, err := base64.StdEncoding.DecodeString(auftragMock.enqueued[0].Payload)
	if err != nil {
		t.Fatalf("expected base64 payload, got decode error: %v", err)
	}

	got := string(payload)
	for _, check := range []string{"STORNOBELEG", "Storno zu Bon-Nr: 1", "GESAMT: -7,00 EUR", "-3,50 x 2 = -7,00 EUR", "SIG-STORNO"} {
		if !strings.Contains(got, check) {
			t.Fatalf("expected %q in stornobeleg, got:\n%q", check, got)
		}
	}
}

func TestKassenbelegDrucken_TischStorno_DruckbarAlsStornobeleg(t *testing.T) {
	ctx := context.Background()
	subject := kasse.TischSessionSubject(testKassensitzungNr, testActiveTisch.ID)

	orderEvent, _ := kasse.NewBestellungAufgenommenEvent(subject, 1, "Test User", []kasse.Position{{
		VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", Einzelpreis: 350, Menge: 2,
	}}, "")
	var orderData kasse.BestellungAufgenommenV1Data
	if err := json.Unmarshal(orderEvent.Data, &orderData); err != nil {
		t.Fatalf("expected no unmarshal error, got %v", err)
	}
	posID := orderData.Positionen[0].PositionID

	paymentEvent, _ := kasse.NewZahlungKassiertEvent(subject, 1, "Test User", []kasse.Position{{
		PositionID: posID, VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", Einzelpreis: 350, Menge: 2,
	}}, 700, "")
	var paymentData kasse.ZahlungKassiertV1Data
	if err := json.Unmarshal(paymentEvent.Data, &paymentData); err != nil {
		t.Fatalf("expected no unmarshal error, got %v", err)
	}

	stornoEvent, err := kasse.NewStornierungErteiltEvent(subject, 2, "Leitung", paymentData.ZahlungID, []kasse.Position{{
		PositionID: posID, VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", Einzelpreis: 350, Menge: 2,
	}}, 700, "Rückgabe")
	if err != nil {
		t.Fatalf("expected no storno event error, got %v", err)
	}
	var stornoData kasse.StornierungErteiltV1Data
	if err := json.Unmarshal(stornoEvent.Data, &stornoData); err != nil {
		t.Fatalf("expected no unmarshal error, got %v", err)
	}

	eventMock := kassenjournal_repo.NewMock(nil, nil)
	eventMock.AddEvent(orderEvent)   // ID 1
	eventMock.AddEvent(paymentEvent) // ID 2
	eventMock.AddEvent(stornoEvent)  // ID 3

	// Die Signatur des Storno-Vorgangs liegt am quittierten Auftrag.
	tseRepo := &mockTSEAuftragRepo{staende: map[int]tse_repo.SignaturauftragStand{
		3: {Status: tse_repo.StatusErledigt, Signatur: &tse.Signatur{
			TransaktionNummer: 4003,
			SignaturZaehler:   101,
			TSESeriennummer:   "SW-TSE-SN-0044",
			LogTimeStart:      time.Date(2026, 6, 10, 19, 20, 1, 0, time.UTC),
			LogTimeEnd:        time.Date(2026, 6, 10, 19, 20, 3, 0, time.UTC),
			Signatur:          "SIG-STORNO",
			QRCodeData:        "V0;STORNO",
		}},
	}}

	auftragMock := &mockDruckauftragRepo{}
	command := Command{
		EventRepo:           eventMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
		DruckstationRepo:    &mockDruckstationRepo{konfig: kassenbelegStationen},
		SettingsRepo:        belegTestSettingsMock(),
		DruckauftragRepo:    auftragMock,
		TSERepo:             tseRepo,
	}

	if _, err := command.KassenbelegDrucken(ctx, testActiveTisch.ID, "", "", stornoData.StornierungID); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(auftragMock.enqueued) != 1 {
		t.Fatalf("expected exactly 1 enqueued auftrag, got %d", len(auftragMock.enqueued))
	}
	if !strings.HasPrefix(auftragMock.enqueued[0].Referenz, "stornierung-erteilt:") {
		t.Fatalf("expected stornierung-erteilt referenz, got %s", auftragMock.enqueued[0].Referenz)
	}

	payload, err := base64.StdEncoding.DecodeString(auftragMock.enqueued[0].Payload)
	if err != nil {
		t.Fatalf("expected base64 payload, got decode error: %v", err)
	}

	got := string(payload)
	// Referenz auf den Ursprungs-Zahlungsbeleg (Event-ID 2) und negativer Betrag.
	for _, check := range []string{"STORNOBELEG", "Storno zu Bon-Nr: 2", "GESAMT: -7,00 EUR", "SIG-STORNO"} {
		if !strings.Contains(got, check) {
			t.Fatalf("expected %q in stornobeleg, got:\n%q", check, got)
		}
	}
}

func TestStornierungErteilen_GemischterStorno_AtomischKorrekturUndWarenruecknahme(t *testing.T) {
	ctx := context.Background()
	subject := kasse.TischSessionSubject(testKassensitzungNr, testActiveTisch.ID)

	// Bestellt 3 Bier, davon 1 bezahlt → ein Storno von 3 spaltet in 2 Korrektur
	// (unbezahlt) und 1 Warenrücknahme (bezahlt).
	orderEvent, _ := kasse.NewBestellungAufgenommenEvent(subject, 1, "Test User", []kasse.Position{{
		VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", Einzelpreis: 350, Menge: 3,
	}}, "")
	var orderData kasse.BestellungAufgenommenV1Data
	if err := json.Unmarshal(orderEvent.Data, &orderData); err != nil {
		t.Fatalf("expected no unmarshal error, got %v", err)
	}
	posID := orderData.Positionen[0].PositionID

	paymentEvent, _ := kasse.NewZahlungKassiertEvent(subject, 1, "Test User", []kasse.Position{{
		PositionID: posID, VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", Einzelpreis: 350, Menge: 1,
	}}, 350, "")

	eventMock := kassenjournal_repo.NewMock(nil, nil)
	eventMock.SetTischSession(subject, kasse.TischSession{})
	eventMock.AddEvent(orderEvent)
	eventMock.AddEvent(paymentEvent)

	command := Command{
		TableRepo:           table_repo.NewMock([]table.Tisch{testActiveTisch}, nil),
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

func TestKassenbelegDrucken_DirektverkaufStorno_NichtGefunden(t *testing.T) {
	ctx := context.Background()
	verkaufID := uuid.New().String()
	subject := kasse.DirektverkaufSubject(testKassensitzungNr, verkaufID)

	verkaufEvent, err := kasse.NewDirektverkaufGetaetigtEvent(subject, verkaufID, 1, "Test User", []kasse.Position{{
		VarianteID:   1,
		ProduktName:  "Cola",
		VarianteName: "0,5l",
		Kategorie:    "getraenk", Steuersatz: "regel",
		Einzelpreis: 350,
		Menge:       1,
	}}, "")
	if err != nil {
		t.Fatalf("expected no event error, got %v", err)
	}

	eventMock := kassenjournal_repo.NewMock(nil, nil)
	eventMock.AddEvent(verkaufEvent)

	command := Command{
		EventRepo:           eventMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
		SettingsRepo:        belegTestSettingsMock(),
		DruckstationRepo:    &mockDruckstationRepo{konfig: kassenbelegStationen},
		DruckauftragRepo:    &mockDruckauftragRepo{},
		TSERepo:             &mockTSEAuftragRepo{},
	}

	_, err = command.KassenbelegDrucken(ctx, 0, "", verkaufID, uuid.New().String())
	if err != ErrStornierungNichtGefunden {
		t.Fatalf("expected ErrStornierungNichtGefunden, got %v", err)
	}
}
