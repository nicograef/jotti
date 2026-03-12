//go:build unit

package application

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/event"
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
	Kategorie: product.BeverageKategorie,
	Status:    product.ActiveStatus,
	Variants:  []product.Variante{},
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

func TestBestellungAufgeben_WithOCC(t *testing.T) {
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

	err := command.BestellungAufgeben(ctx, 1, "Test User", 1, inputs, "Testkommentar")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestBestellungAufgeben_OCCRetrySuccess(t *testing.T) {
	ctx := context.Background()
	productMock := product_repo.NewMock([]product.Produkt{testProduct}, nil)
	productMock.AddVariant(testProduct.ID, testVariant)
	// First WriteEvent call fails with ErrAlreadyExists, second succeeds
	eventMock := event_repo.NewMockWithWriteErr(nil, db.ErrAlreadyExists, 1)
	command := Command{
		TableRepo:   table_repo.NewMock([]table.Tisch{testActiveTisch}, nil),
		EventRepo:   eventMock,
		ProductRepo: productMock,
	}

	inputs := []BestellPositionInput{
		{ProduktID: testProduct.ID, VarianteID: testVariant.ID, Menge: 1},
	}

	err := command.BestellungAufgeben(ctx, 1, "Test User", 1, inputs, "")
	if err != nil {
		t.Fatalf("expected no error after OCC retry, got %v", err)
	}
}

func TestBestellungAufgeben_OCCConflictAfterRetries(t *testing.T) {
	ctx := context.Background()
	productMock := product_repo.NewMock([]product.Produkt{testProduct}, nil)
	productMock.AddVariant(testProduct.ID, testVariant)
	// All WriteEvent calls fail with ErrAlreadyExists
	eventMock := event_repo.NewMockWithWriteErr(nil, db.ErrAlreadyExists, 10)
	command := Command{
		TableRepo:   table_repo.NewMock([]table.Tisch{testActiveTisch}, nil),
		EventRepo:   eventMock,
		ProductRepo: productMock,
	}

	inputs := []BestellPositionInput{
		{ProduktID: testProduct.ID, VarianteID: testVariant.ID, Menge: 1},
	}

	err := command.BestellungAufgeben(ctx, 1, "Test User", 1, inputs, "")
	if err != ErrConflict {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

// --- Invariant Tests ---

// createOrderEvent creates an order event and assigns it an ID for use in mock repos.
func createOrderEvent(t *testing.T, tischID int, positions []table.Position) event.Event {
	t.Helper()
	e, err := table.NewBestellungAufgegebenEvent(1, "TestUser", tischID, positions, "")
	if err != nil {
		t.Fatalf("failed to create order event: %v", err)
	}
	return e
}

// extractPositionRefs parses the event data to get PositionRefs with the generated UUIDs.
func extractPositionRefs(t *testing.T, orderEvent event.Event, menge int) []table.PositionRef {
	t.Helper()
	var data struct {
		Positionen []struct {
			PositionID string `json:"positionId"`
		} `json:"positionen"`
	}
	if err := json.Unmarshal(orderEvent.Data, &data); err != nil {
		t.Fatalf("failed to parse order event data: %v", err)
	}
	refs := make([]table.PositionRef, len(data.Positionen))
	for i, p := range data.Positionen {
		refs[i] = table.PositionRef{PositionID: p.PositionID, Menge: menge}
	}
	return refs
}

func createPaymentEvent(t *testing.T, tischID int, refs []table.PositionRef, amountCents int) event.Event {
	t.Helper()
	e, err := table.NewZahlungRegistriertEvent(1, "TestUser", tischID, refs, amountCents, "")
	if err != nil {
		t.Fatalf("failed to create payment event: %v", err)
	}
	return e
}

func createDeliveryEvent(t *testing.T, tischID int, refs []table.PositionRef) event.Event {
	t.Helper()
	e, err := table.NewProdukteGeliefertEvent(1, "TestUser", tischID, refs, "")
	if err != nil {
		t.Fatalf("failed to create delivery event: %v", err)
	}
	return e
}

func createCancelEvent(t *testing.T, tischID int, refs []table.PositionRef, amountCents int) event.Event {
	t.Helper()
	e, err := table.NewProdukteStorniertEvent(1, "TestUser", tischID, refs, amountCents, "")
	if err != nil {
		t.Fatalf("failed to create cancel event: %v", err)
	}
	return e
}

// assignIDs assigns sequential IDs to events for use in mock repos.
func assignIDs(events []event.Event) []event.Event {
	for i := range events {
		events[i].ID = i + 1
	}
	return events
}

func TestBestellungAufgeben_InactiveTisch(t *testing.T) {
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

	err := command.BestellungAufgeben(ctx, 1, "Test User", testInactiveTisch.ID, inputs, "")
	if err != ErrTischNotActive {
		t.Fatalf("expected ErrTischNotActive, got %v", err)
	}
}

func TestZahlungRegistrieren_NonOrderedPosition(t *testing.T) {
	ctx := context.Background()
	// No order events exist — paying a non-existent position should fail
	command := Command{
		TableRepo: table_repo.NewMock([]table.Tisch{testActiveTisch}, nil),
		EventRepo: event_repo.NewMock(nil, nil),
	}

	fakeRefs := []table.PositionRef{
		{PositionID: "00000000-0000-0000-0000-000000000001", Menge: 1},
	}

	err := command.ZahlungRegistrieren(ctx, 1, "Test User", testActiveTisch.ID, fakeRefs, 350, "")
	if err != ErrPositionNichtBezahlbar {
		t.Fatalf("expected ErrPositionNichtBezahlbar, got %v", err)
	}
}

func TestZahlungRegistrieren_DoublePayment(t *testing.T) {
	ctx := context.Background()
	positions := []table.Position{
		{VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "beverage", Einzelpreis: 350, Menge: 1},
	}
	orderEvent := createOrderEvent(t, testActiveTisch.ID, positions)
	refs := extractPositionRefs(t, orderEvent, 1)
	paymentEvent := createPaymentEvent(t, testActiveTisch.ID, refs, 350)

	events := assignIDs([]event.Event{orderEvent, paymentEvent})
	command := Command{
		TableRepo: table_repo.NewMock([]table.Tisch{testActiveTisch}, nil),
		EventRepo: event_repo.NewMock(events, nil),
	}

	// Try to pay again — should fail
	err := command.ZahlungRegistrieren(ctx, 1, "Test User", testActiveTisch.ID, refs, 350, "")
	if err != ErrPositionNichtBezahlbar {
		t.Fatalf("expected ErrPositionNichtBezahlbar, got %v", err)
	}
}

func TestProdukteLiefern_NonOrderedPosition(t *testing.T) {
	ctx := context.Background()
	// No order events exist — delivering a non-existent position should fail
	command := Command{
		TableRepo: table_repo.NewMock([]table.Tisch{testActiveTisch}, nil),
		EventRepo: event_repo.NewMock(nil, nil),
	}

	fakeRefs := []table.PositionRef{
		{PositionID: "00000000-0000-0000-0000-000000000001", Menge: 1},
	}

	err := command.ProdukteLiefern(ctx, 1, "Test User", testActiveTisch.ID, fakeRefs, "")
	if err != ErrPositionNichtLieferbar {
		t.Fatalf("expected ErrPositionNichtLieferbar, got %v", err)
	}
}

func TestProdukteStornieren_AlreadyPaidPosition(t *testing.T) {
	ctx := context.Background()
	positions := []table.Position{
		{VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "beverage", Einzelpreis: 350, Menge: 1},
	}
	orderEvent := createOrderEvent(t, testActiveTisch.ID, positions)
	refs := extractPositionRefs(t, orderEvent, 1)
	paymentEvent := createPaymentEvent(t, testActiveTisch.ID, refs, 350)

	events := assignIDs([]event.Event{orderEvent, paymentEvent})
	command := Command{
		TableRepo: table_repo.NewMock([]table.Tisch{testActiveTisch}, nil),
		EventRepo: event_repo.NewMock(events, nil),
	}

	// Stornierung of already-paid position should fail (unbezahlt is empty)
	err := command.ProdukteStornieren(ctx, 1, "Test User", testActiveTisch.ID, refs, 350, "")
	if err != ErrPositionNichtStornierbar {
		t.Fatalf("expected ErrPositionNichtStornierbar, got %v", err)
	}
}

func TestZahlungRegistrieren_ExceedsAvailableMenge(t *testing.T) {
	ctx := context.Background()
	positions := []table.Position{
		{VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "beverage", Einzelpreis: 350, Menge: 1},
	}
	orderEvent := createOrderEvent(t, testActiveTisch.ID, positions)
	// Try to pay for Menge 2 when only 1 was ordered
	refs := extractPositionRefs(t, orderEvent, 2)

	events := assignIDs([]event.Event{orderEvent})
	command := Command{
		TableRepo: table_repo.NewMock([]table.Tisch{testActiveTisch}, nil),
		EventRepo: event_repo.NewMock(events, nil),
	}

	err := command.ZahlungRegistrieren(ctx, 1, "Test User", testActiveTisch.ID, refs, 700, "")
	if err != ErrPositionNichtBezahlbar {
		t.Fatalf("expected ErrPositionNichtBezahlbar, got %v", err)
	}
}
