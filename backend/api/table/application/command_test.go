//go:build unit

package application

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	bondruckApp "github.com/nicograef/jotti/backend/api/bondruck/application"
	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/product"
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

	auftragMock := &mockDruckauftragRepo{}
	stationMock := &mockDruckstationRepo{konfig: map[string]bondruckApp.Druckstation{
		"getraenk": {IP: "192.168.1.50", Bonmodus: "pro_position"},
	}}

	command := newTestCommand([]table.Tisch{testActiveTisch}, []product.Produkt{testProduct})
	command.ProductRepo = productMock
	command.DruckstationRepo = stationMock
	command.DruckauftragRepo = auftragMock

	inputs := []BestellPositionInput{{ProduktID: testProduct.ID, VarianteID: testVariant.ID, Menge: 2}}

	err := command.BestellungAufnehmen(ctx, 1, "Test User", 1, inputs, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(auftragMock.enqueued) != 1 {
		t.Fatalf("expected 1 enqueued druckauftrag, got %d", len(auftragMock.enqueued))
	}
	if auftragMock.enqueued[0].BonArt != "arbeitsbon" {
		t.Fatalf("expected BonArt arbeitsbon, got %s", auftragMock.enqueued[0].BonArt)
	}
	if auftragMock.enqueued[0].ZielIP != "192.168.1.50" {
		t.Fatalf("expected ZielIP 192.168.1.50, got %s", auftragMock.enqueued[0].ZielIP)
	}
	if auftragMock.enqueued[0].Payload == "" {
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
