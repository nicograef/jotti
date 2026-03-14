//go:build unit

package application

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/product"
	"github.com/nicograef/jotti/backend/domain/table"
	"github.com/nicograef/jotti/backend/repository/event_repo"
	"github.com/nicograef/jotti/backend/repository/product_repo"
	"github.com/nicograef/jotti/backend/repository/table_repo"
)

func newTestCommand(tables []table.Tisch, products []product.Produkt) Command {
	return Command{
		TableRepo:   table_repo.NewMock(tables, nil),
		EventRepo:   event_repo.NewMock(nil, nil),
		ProductRepo: product_repo.NewMock(products, db.ErrNotFound),
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

func TestResolvePositions(t *testing.T) {
	available := []table.Position{
		{PositionID: "pos-1", VarianteID: 1, ProduktName: "Beer", VarianteName: "Pils 0.5l", Kategorie: "getraenk", Einzelpreis: 500, Menge: 3},
		{PositionID: "pos-2", VarianteID: 2, ProduktName: "Wurst", VarianteName: "Bratwurst", Kategorie: "essen", Einzelpreis: 400, Menge: 2},
	}

	refs := []table.PositionRef{
		{PositionID: "pos-1", Menge: 2},
		{PositionID: "pos-2", Menge: 1},
	}

	resolved, total := resolvePositions(available, refs)
	// 500*2 + 400*1 = 1400
	if total != 1400 {
		t.Fatalf("expected 1400, got %d", total)
	}
	if len(resolved) != 2 {
		t.Fatalf("expected 2 resolved positions, got %d", len(resolved))
	}
	if resolved[0].ProduktName != "Beer" || resolved[0].Menge != 2 {
		t.Fatalf("expected Beer with menge 2, got %s with menge %d", resolved[0].ProduktName, resolved[0].Menge)
	}
	if resolved[1].ProduktName != "Wurst" || resolved[1].Menge != 1 {
		t.Fatalf("expected Wurst with menge 1, got %s with menge %d", resolved[1].ProduktName, resolved[1].Menge)
	}
}

func TestResolvePositions_Empty(t *testing.T) {
	resolved, total := resolvePositions(nil, nil)
	if total != 0 {
		t.Fatalf("expected 0, got %d", total)
	}
	if len(resolved) != 0 {
		t.Fatalf("expected 0 resolved, got %d", len(resolved))
	}
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

func TestBestellungAufnehmen_WithOCC(t *testing.T) {
	ctx := context.Background()
	productMock := product_repo.NewMock([]product.Produkt{testProduct}, nil)
	productMock.AddVariant(testProduct.ID, testVariant)
	command := Command{
		TableRepo:   table_repo.NewMock([]table.Tisch{testActiveTisch}, nil),
		EventRepo:   event_repo.NewMock(nil, nil),
		ProductRepo: productMock,
	}

	inputs := []BestellPositionInput{
		{ProduktID: testProduct.ID, VarianteID: testVariant.ID, Menge: 2},
	}

	err := command.BestellungAufnehmen(ctx, 1, "Test User", 1, inputs, "Testkommentar")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestBestellungAufnehmen_Conflict(t *testing.T) {
	ctx := context.Background()
	productMock := product_repo.NewMock([]product.Produkt{testProduct}, nil)
	productMock.AddVariant(testProduct.ID, testVariant)
	eventMock := event_repo.NewMockWithWriteErr(nil, db.ErrAlreadyExists)
	command := Command{
		TableRepo:   table_repo.NewMock([]table.Tisch{testActiveTisch}, nil),
		EventRepo:   eventMock,
		ProductRepo: productMock,
	}

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
	command := Command{
		TableRepo:   table_repo.NewMock([]table.Tisch{testInactiveTisch}, nil),
		EventRepo:   event_repo.NewMock(nil, nil),
		ProductRepo: productMock,
	}

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
	command := Command{
		TableRepo: table_repo.NewMock([]table.Tisch{testActiveTisch}, nil),
		EventRepo: event_repo.NewMock(nil, nil),
	}

	fakeRefs := []table.PositionRef{
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
	eventMock := event_repo.NewMock(nil, nil)
	eventMock.SetTableState(testActiveTisch.ID, table.TischState{
		SaldoCents:            0,
		UnbezahltePositionen:  []table.Position{},
		AusstehendePositionen: []table.Position{},
		GesamtZahlungenCents:  350,
	})
	command := Command{
		TableRepo: table_repo.NewMock([]table.Tisch{testActiveTisch}, nil),
		EventRepo: eventMock,
	}

	refs := []table.PositionRef{
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
	command := Command{
		TableRepo: table_repo.NewMock([]table.Tisch{testActiveTisch}, nil),
		EventRepo: event_repo.NewMock(nil, nil),
	}

	fakeRefs := []table.PositionRef{
		{PositionID: "00000000-0000-0000-0000-000000000001", Menge: 1},
	}

	err := command.AusgabeBestaetigen(ctx, 1, "Test User", testActiveTisch.ID, fakeRefs, "")
	if err != ErrPositionNichtAusgebbar {
		t.Fatalf("expected ErrPositionNichtAusgebbar, got %v", err)
	}
}

func TestStornierungErteilen_AlreadyPaidPosition_Succeeds(t *testing.T) {
	ctx := context.Background()
	orderEvent, _ := table.NewBestellungAufgenommenEvent(1, "Test User", testActiveTisch.ID,
		[]table.Position{
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

	paymentEvent, _ := table.NewZahlungKassiertEvent(1, "Test User", testActiveTisch.ID,
		[]table.Position{
			{PositionID: posID, VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Einzelpreis: 350, Menge: 1},
		}, 350, "")

	eventMock := event_repo.NewMock(nil, nil)
	eventMock.SetTableState(testActiveTisch.ID, table.TischState{
		SaldoCents:            0,
		UnbezahltePositionen:  []table.Position{},
		AusstehendePositionen: []table.Position{},
		GesamtZahlungenCents:  350,
	})
	orderEvent.Subject = "tisch:1"
	paymentEvent.Subject = "tisch:1"
	eventMock.AddEvent(orderEvent)
	eventMock.AddEvent(paymentEvent)

	command := Command{
		TableRepo: table_repo.NewMock([]table.Tisch{testActiveTisch}, nil),
		EventRepo: eventMock,
	}

	refs := []table.PositionRef{{PositionID: posID, Menge: 1}}

	err := command.StornierungErteilen(ctx, 1, "Test User", testActiveTisch.ID, refs, "Reklamation")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestStornierungErteilen_AlreadyCancelledPosition_Fails(t *testing.T) {
	ctx := context.Background()
	orderEvent, _ := table.NewBestellungAufgenommenEvent(1, "Test User", testActiveTisch.ID,
		[]table.Position{
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

	cancelEvent, _ := table.NewStornierungErteiltEvent(1, "Test User", testActiveTisch.ID,
		[]table.Position{
			{PositionID: posID, VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Einzelpreis: 350, Menge: 1},
		}, 350, "")

	eventMock := event_repo.NewMock(nil, nil)
	eventMock.SetTableState(testActiveTisch.ID, table.TischState{
		SaldoCents:            0,
		UnbezahltePositionen:  []table.Position{},
		AusstehendePositionen: []table.Position{},
		GesamtZahlungenCents:  350,
	})
	orderEvent.Subject = "tisch:1"
	cancelEvent.Subject = "tisch:1"
	eventMock.AddEvent(orderEvent)
	eventMock.AddEvent(cancelEvent)

	command := Command{
		TableRepo: table_repo.NewMock([]table.Tisch{testActiveTisch}, nil),
		EventRepo: eventMock,
	}

	refs := []table.PositionRef{{PositionID: posID, Menge: 1}}

	err := command.StornierungErteilen(ctx, 1, "Test User", testActiveTisch.ID, refs, "")
	if err != ErrPositionNichtStornierbar {
		t.Fatalf("expected ErrPositionNichtStornierbar, got %v", err)
	}
}

func TestZahlungKassieren_ExceedsAvailableMenge(t *testing.T) {
	ctx := context.Background()
	// State has 1 position with Menge 1
	eventMock := event_repo.NewMock(nil, nil)
	eventMock.SetTableState(testActiveTisch.ID, table.TischState{
		SaldoCents: 350,
		UnbezahltePositionen: []table.Position{
			{PositionID: "pos-1", VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Einzelpreis: 350, Menge: 1},
		},
		AusstehendePositionen: []table.Position{
			{PositionID: "pos-1", VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Einzelpreis: 350, Menge: 1},
		},
	})
	command := Command{
		TableRepo: table_repo.NewMock([]table.Tisch{testActiveTisch}, nil),
		EventRepo: eventMock,
	}

	// Try to pay for Menge 2 when only 1 was ordered
	refs := []table.PositionRef{
		{PositionID: "pos-1", Menge: 2},
	}

	err := command.ZahlungKassieren(ctx, 1, "Test User", testActiveTisch.ID, refs, "")
	if err != ErrPositionNichtBezahlbar {
		t.Fatalf("expected ErrPositionNichtBezahlbar, got %v", err)
	}
}
