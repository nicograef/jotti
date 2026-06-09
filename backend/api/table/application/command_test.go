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
	bondruckApp "github.com/nicograef/jotti/backend/api/bondruck/application"
	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/product"
	"github.com/nicograef/jotti/backend/domain/settings"
	"github.com/nicograef/jotti/backend/domain/table"
	"github.com/nicograef/jotti/backend/repository/kassenjournal_repo"
	"github.com/nicograef/jotti/backend/repository/kassensitzungen_repo"
	"github.com/nicograef/jotti/backend/repository/product_repo"
	"github.com/nicograef/jotti/backend/repository/table_repo"
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
	ID:        1,
	Name:      "Cola",
	Kategorie: product.GetraenkKategorie,
	Status:    product.ActiveStatus,
	Varianten: []product.Variante{},
	CreatedAt: time.Now().UTC(),
	UpdatedAt: time.Now().UTC(),
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

type mockDruckauftragRepo struct {
	enqueued []bondruckApp.Druckauftrag
	err      error
}

func (m *mockDruckauftragRepo) EnqueueDruckauftraege(_ context.Context, auftraege []bondruckApp.Druckauftrag) error {
	if m.err != nil {
		return m.err
	}
	m.enqueued = append(m.enqueued, auftraege...)
	return nil
}

type mockSettingsRepo struct {
	bondruck       settings.BondruckEinstellungen
	betreiber      settings.Betreiber
	kassenident    settings.Kassenidentitaet
	bondruckErr    error
	betreiberErr   error
	kassenidentErr error
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

type stornierungErteiltData struct {
	Positionen             []umbuchungPositionData `json:"positionen"`
	GesamtStornierungCents int                     `json:"gesamtStornierungCents"`
	Kommentar              string                  `json:"kommentar"`
}

type bestellungAufgenommenData struct {
	Positionen       []umbuchungPositionData `json:"positionen"`
	GesamtPreisCents int                     `json:"gesamtPreisCents"`
	Kommentar        string                  `json:"kommentar"`
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

func (m *mockSettingsRepo) GetBondruckEinstellungen(_ context.Context) (settings.BondruckEinstellungen, error) {
	if m.bondruckErr != nil {
		return settings.BondruckEinstellungen{}, m.bondruckErr
	}
	return m.bondruck, nil
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
			{VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Einzelpreis: 350, Menge: 1},
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
			{PositionID: posID, VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Einzelpreis: 350, Menge: 1},
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
			{VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Einzelpreis: 350, Menge: 1},
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

	cancelEvent, _ := kasse.NewStornierungErteiltEvent(subject, 1, "Test User",
		[]kasse.Position{
			{PositionID: posID, VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Einzelpreis: 350, Menge: 1},
		}, 350, "Test")

	eventMock := kassenjournal_repo.NewMock(nil, nil)
	eventMock.SetTischSession(subject, kasse.TischSession{
		SaldoCents:            0,
		UnbezahltePositionen:  []kasse.Position{},
		AusstehendePositionen: []kasse.Position{},
		GesamtZahlungenCents:  350,
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
			{PositionID: "pos-1", VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Einzelpreis: 350, Menge: 1},
		},
		AusstehendePositionen: []kasse.Position{
			{PositionID: "pos-1", VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Einzelpreis: 350, Menge: 1},
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
				Kategorie:    "getraenk",
				Einzelpreis:  350,
				Menge:        2,
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
	if quellEvents[0].Type != string(kasse.EventTypeStornierungErteiltV1) {
		t.Fatalf("expected source event type %s, got %s", kasse.EventTypeStornierungErteiltV1, quellEvents[0].Type)
	}

	zielEvents, err := eventMock.ReadEventsBySubject(ctx, zielSubject)
	if err != nil {
		t.Fatalf("expected no error reading target events, got %v", err)
	}
	if len(zielEvents) != 1 {
		t.Fatalf("expected 1 target event, got %d", len(zielEvents))
	}
	if zielEvents[0].Type != string(kasse.EventTypeBestellungAufgenommenV1) {
		t.Fatalf("expected target event type %s, got %s", kasse.EventTypeBestellungAufgenommenV1, zielEvents[0].Type)
	}

	var stornoData stornierungErteiltData
	if err := json.Unmarshal(quellEvents[0].Data, &stornoData); err != nil {
		t.Fatalf("expected no unmarshal error for storno data, got %v", err)
	}

	if stornoData.GesamtStornierungCents != 350 {
		t.Fatalf("expected source amount 350, got %d", stornoData.GesamtStornierungCents)
	}
	if stornoData.Kommentar != "Umbuchung auf Tisch Tisch Ziel" {
		t.Fatalf("unexpected source comment: %q", stornoData.Kommentar)
	}
	if len(stornoData.Positionen) != 1 {
		t.Fatalf("expected 1 source position, got %d", len(stornoData.Positionen))
	}
	if stornoData.Positionen[0].PositionID != quellPositionID {
		t.Fatalf("expected source position ID %q, got %q", quellPositionID, stornoData.Positionen[0].PositionID)
	}
	if stornoData.Positionen[0].Einzelpreis != 350 {
		t.Fatalf("expected source einzelpreis 350, got %d", stornoData.Positionen[0].Einzelpreis)
	}

	var bestellungData bestellungAufgenommenData
	if err := json.Unmarshal(zielEvents[0].Data, &bestellungData); err != nil {
		t.Fatalf("expected no unmarshal error for bestellung data, got %v", err)
	}

	if bestellungData.GesamtPreisCents != 350 {
		t.Fatalf("expected target amount 350, got %d", bestellungData.GesamtPreisCents)
	}
	if bestellungData.Kommentar != "Umbuchung von Tisch Tisch Quelle" {
		t.Fatalf("unexpected target comment: %q", bestellungData.Kommentar)
	}
	if len(bestellungData.Positionen) != 1 {
		t.Fatalf("expected 1 target position, got %d", len(bestellungData.Positionen))
	}
	if bestellungData.Positionen[0].PositionID == quellPositionID {
		t.Fatalf("expected target position ID to be regenerated, but remained %q", bestellungData.Positionen[0].PositionID)
	}
	if bestellungData.Positionen[0].Einzelpreis != 350 {
		t.Fatalf("expected target einzelpreis 350, got %d", bestellungData.Positionen[0].Einzelpreis)
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
			Kategorie:    "getraenk",
			Einzelpreis:  350,
			Menge:        1,
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

	var stornoData stornierungErteiltData
	if err := json.Unmarshal(quellEvents[0].Data, &stornoData); err != nil {
		t.Fatalf("expected no unmarshal error for storno data, got %v", err)
	}
	var bestellungData bestellungAufgenommenData
	if err := json.Unmarshal(zielEvents[0].Data, &bestellungData); err != nil {
		t.Fatalf("expected no unmarshal error for bestellung data, got %v", err)
	}

	if utf8.RuneCountInString(stornoData.Kommentar) > 100 {
		t.Fatalf("expected source comment length <= 100 runes, got %d", utf8.RuneCountInString(stornoData.Kommentar))
	}
	if utf8.RuneCountInString(bestellungData.Kommentar) > 100 {
		t.Fatalf("expected target comment length <= 100 runes, got %d", utf8.RuneCountInString(bestellungData.Kommentar))
	}
	if !strings.HasPrefix(stornoData.Kommentar, "Umbuchung auf Tisch ") {
		t.Fatalf("expected source comment prefix, got %q", stornoData.Kommentar)
	}
	if !strings.HasPrefix(bestellungData.Kommentar, "Umbuchung von Tisch ") {
		t.Fatalf("expected target comment prefix, got %q", bestellungData.Kommentar)
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
			Kategorie:    "getraenk",
			Einzelpreis:  350,
			Menge:        1,
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
			Kategorie:    "getraenk",
			Einzelpreis:  350,
			Menge:        1,
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
			Kategorie:    "getraenk",
			Einzelpreis:  350,
			Menge:        2,
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
		bondruck: settings.BondruckEinstellungen{KassenbelegDruckerIP: "192.168.1.80", UpdatedAt: time.Now()},
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
		SettingsRepo:        settingsMock,
		DruckauftragRepo:    auftragMock,
	}

	err = command.KassenbelegDrucken(ctx, testActiveTisch.ID, eventData.ZahlungID, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	err = command.KassenbelegDrucken(ctx, testActiveTisch.ID, eventData.ZahlungID, "")
	if err != nil {
		t.Fatalf("expected no reprint error, got %v", err)
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

func TestKassenbelegDrucken_ZahlungNichtGefunden(t *testing.T) {
	ctx := context.Background()
	command := Command{
		EventRepo:           kassenjournal_repo.NewMock(nil, nil),
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
		SettingsRepo:        &mockSettingsRepo{},
		DruckauftragRepo:    &mockDruckauftragRepo{},
	}

	err := command.KassenbelegDrucken(ctx, testActiveTisch.ID, "11111111-1111-1111-1111-111111111111", "")
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
			Kategorie:    "getraenk",
			Einzelpreis:  350,
			Menge:        1,
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
		SettingsRepo: &mockSettingsRepo{
			bondruck: settings.BondruckEinstellungen{KassenbelegDruckerIP: "", UpdatedAt: time.Now()},
		},
		DruckauftragRepo: &mockDruckauftragRepo{},
	}

	err = command.KassenbelegDrucken(ctx, testActiveTisch.ID, eventData.ZahlungID, "")
	if err != ErrKassenbelegDruckerNichtKonfiguriert {
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
			Kategorie:    "getraenk",
			Einzelpreis:  350,
			Menge:        2,
		},
	}, "")
	if err != nil {
		t.Fatalf("expected no event error, got %v", err)
	}

	eventMock := kassenjournal_repo.NewMock(nil, nil)
	eventMock.AddEvent(verkaufEvent)

	auftragMock := &mockDruckauftragRepo{}
	settingsMock := &mockSettingsRepo{
		bondruck: settings.BondruckEinstellungen{KassenbelegDruckerIP: "192.168.1.80", UpdatedAt: time.Now()},
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
		SettingsRepo:        settingsMock,
		DruckauftragRepo:    auftragMock,
	}

	err = command.KassenbelegDrucken(ctx, 0, "", verkaufID)
	if err != nil {
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
}

func TestKassenbelegDrucken_Direktverkauf_NichtGefunden(t *testing.T) {
	ctx := context.Background()
	command := Command{
		EventRepo:           kassenjournal_repo.NewMock(nil, nil),
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
		SettingsRepo:        &mockSettingsRepo{},
		DruckauftragRepo:    &mockDruckauftragRepo{},
	}

	err := command.KassenbelegDrucken(ctx, 0, "", uuid.New().String())
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
			Kategorie:    "getraenk",
			Einzelpreis:  350,
			Menge:        1,
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
		SettingsRepo: &mockSettingsRepo{
			bondruck: settings.BondruckEinstellungen{KassenbelegDruckerIP: "", UpdatedAt: time.Now()},
		},
		DruckauftragRepo: &mockDruckauftragRepo{},
	}

	err = command.KassenbelegDrucken(ctx, 0, "", verkaufID)
	if err != ErrKassenbelegDruckerNichtKonfiguriert {
		t.Fatalf("expected ErrKassenbelegDruckerNichtKonfiguriert, got %v", err)
	}
}
