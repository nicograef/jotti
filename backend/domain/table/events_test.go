//go:build unit

package table

import (
	"testing"

	e "github.com/nicograef/jotti/backend/domain/event"
)

func mustCreateOrderEvent(t *testing.T, userID, tableID int, products []Position) e.Event {
	t.Helper()
	event, err := NewBestellungAufgegebenEvent(userID, tableID, products, "")
	if err != nil {
		t.Fatalf("failed to create order event: %v", err)
	}
	return event
}

func mustCreatePaymentEvent(t *testing.T, userID, tableID int, products []Position) e.Event {
	t.Helper()
	event, err := NewZahlungRegistriertEvent(userID, tableID, products, "")
	if err != nil {
		t.Fatalf("failed to create payment event: %v", err)
	}
	return event
}

func mustCreateCancelationEvent(t *testing.T, userID, tableID int, products []Position) e.Event {
	t.Helper()
	event, err := NewProdukteStorniertEvent(userID, tableID, products, "")
	if err != nil {
		t.Fatalf("failed to create cancelation event: %v", err)
	}
	return event
}

func mustCreateDeliveryEvent(t *testing.T, userID, tableID int, products []Position) e.Event {
	t.Helper()
	event, err := NewProdukteGeliefertEvent(userID, tableID, products, "")
	if err != nil {
		t.Fatalf("failed to create delivery event: %v", err)
	}
	return event
}

func mustCreateSnapshotEvent(t *testing.T, userID, tableID int, balance int, unpaid, undelivered []Position, totalPayments int) e.Event {
	t.Helper()
	event, err := NewSnapshotEvent(userID, tableID, balance, unpaid, undelivered, totalPayments)
	if err != nil {
		t.Fatalf("failed to create snapshot event: %v", err)
	}
	return event
}

func TestGetSaldoFromEvents_Empty(t *testing.T) {
	balance, err := GetSaldoFromEvents([]e.Event{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if balance != 0 {
		t.Fatalf("expected balance 0, got %d", balance)
	}
}

func TestGetSaldoFromEvents_OrderOnly(t *testing.T) {
	products := []Position{
		{ID: 1, Name: "Beer", PreisCents: 500, Quantity: 2},
	}
	events := []e.Event{
		mustCreateOrderEvent(t, 1, 1, products),
	}

	balance, err := GetSaldoFromEvents(events)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if balance != 1000 {
		t.Fatalf("expected balance 1000, got %d", balance)
	}
}

func TestGetSaldoFromEvents_OrderAndPayment(t *testing.T) {
	products := []Position{
		{ID: 1, Name: "Beer", PreisCents: 500, Quantity: 2},
	}
	paidProducts := []Position{
		{ID: 1, Name: "Beer", PreisCents: 500, Quantity: 1},
	}
	events := []e.Event{
		mustCreateOrderEvent(t, 1, 1, products),
		mustCreatePaymentEvent(t, 1, 1, paidProducts),
	}

	balance, err := GetSaldoFromEvents(events)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if balance != 500 {
		t.Fatalf("expected balance 500, got %d", balance)
	}
}

func TestGetSaldoFromEvents_OrderPaymentAndCancelation(t *testing.T) {
	products := []Position{
		{ID: 1, Name: "Beer", PreisCents: 500, Quantity: 2},
	}
	paidProducts := []Position{
		{ID: 1, Name: "Beer", PreisCents: 500, Quantity: 1},
	}
	canceledProducts := []Position{
		{ID: 1, Name: "Beer", PreisCents: 500, Quantity: 1},
	}
	events := []e.Event{
		mustCreateOrderEvent(t, 1, 1, products),
		mustCreatePaymentEvent(t, 1, 1, paidProducts),
		mustCreateCancelationEvent(t, 1, 1, canceledProducts),
	}

	balance, err := GetSaldoFromEvents(events)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if balance != 0 {
		t.Fatalf("expected balance 0, got %d", balance)
	}
}

func TestGetHistoryFromEvents_Empty(t *testing.T) {
	history, err := GetHistoryFromEvents([]e.Event{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(history) != 0 {
		t.Fatalf("expected empty history, got %d items", len(history))
	}
}

func TestGetHistoryFromEvents_ReturnsAllEventTypes(t *testing.T) {
	products := []Position{
		{ID: 1, Name: "Beer", PreisCents: 500, Quantity: 1},
	}
	events := []e.Event{
		mustCreateOrderEvent(t, 1, 1, products),
		mustCreatePaymentEvent(t, 1, 1, products),
		mustCreateCancelationEvent(t, 1, 1, products),
		mustCreateDeliveryEvent(t, 1, 1, products),
	}

	history, err := GetHistoryFromEvents(events)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(history) != 4 {
		t.Fatalf("expected 4 history items, got %d", len(history))
	}
}

func TestGetHistoryFromEvents_ReversesOrder(t *testing.T) {
	products := []Position{
		{ID: 1, Name: "Beer", PreisCents: 500, Quantity: 1},
	}
	events := []e.Event{
		mustCreateOrderEvent(t, 1, 1, products),
		mustCreatePaymentEvent(t, 1, 1, products),
	}

	history, err := GetHistoryFromEvents(events)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	// First item should be payment (last event), second should be order (first event)
	if _, ok := history[0].(Zahlung); !ok {
		t.Fatalf("expected first item to be Zahlung, got %T", history[0])
	}
	if _, ok := history[1].(Bestellung); !ok {
		t.Fatalf("expected second item to be Bestellung, got %T", history[1])
	}
}

func TestGetUnbezahltePositionenFromEvents_Empty(t *testing.T) {
	products, err := GetUnbezahltePositionenFromEvents([]e.Event{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(products) != 0 {
		t.Fatalf("expected empty products, got %d", len(products))
	}
}

func TestGetUnbezahltePositionenFromEvents_OrderOnly(t *testing.T) {
	orderProducts := []Position{
		{ID: 1, Name: "Beer", PreisCents: 500, Quantity: 2},
		{ID: 2, Name: "Fries", PreisCents: 300, Quantity: 1},
	}
	events := []e.Event{
		mustCreateOrderEvent(t, 1, 1, orderProducts),
	}

	products, err := GetUnbezahltePositionenFromEvents(events)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(products) != 2 {
		t.Fatalf("expected 2 products, got %d", len(products))
	}
	if products[0].Quantity != 2 {
		t.Fatalf("expected quantity 2, got %d", products[0].Quantity)
	}
}

func TestGetUnbezahltePositionenFromEvents_PartialPayment(t *testing.T) {
	orderProducts := []Position{
		{ID: 1, Name: "Beer", PreisCents: 500, Quantity: 3},
	}
	paidProducts := []Position{
		{ID: 1, Name: "Beer", PreisCents: 500, Quantity: 1},
	}
	events := []e.Event{
		mustCreateOrderEvent(t, 1, 1, orderProducts),
		mustCreatePaymentEvent(t, 1, 1, paidProducts),
	}

	products, err := GetUnbezahltePositionenFromEvents(events)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(products) != 1 {
		t.Fatalf("expected 1 product, got %d", len(products))
	}
	if products[0].Quantity != 2 {
		t.Fatalf("expected quantity 2, got %d", products[0].Quantity)
	}
}

func TestGetUnbezahltePositionenFromEvents_FullPayment(t *testing.T) {
	orderProducts := []Position{
		{ID: 1, Name: "Beer", PreisCents: 500, Quantity: 2},
	}
	events := []e.Event{
		mustCreateOrderEvent(t, 1, 1, orderProducts),
		mustCreatePaymentEvent(t, 1, 1, orderProducts),
	}

	products, err := GetUnbezahltePositionenFromEvents(events)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(products) != 0 {
		t.Fatalf("expected 0 products, got %d", len(products))
	}
}

func TestGetUnbezahltePositionenFromEvents_WithCancelation(t *testing.T) {
	orderProducts := []Position{
		{ID: 1, Name: "Beer", PreisCents: 500, Quantity: 3},
	}
	canceledProducts := []Position{
		{ID: 1, Name: "Beer", PreisCents: 500, Quantity: 2},
	}
	events := []e.Event{
		mustCreateOrderEvent(t, 1, 1, orderProducts),
		mustCreateCancelationEvent(t, 1, 1, canceledProducts),
	}

	products, err := GetUnbezahltePositionenFromEvents(events)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(products) != 1 {
		t.Fatalf("expected 1 product, got %d", len(products))
	}
	if products[0].Quantity != 1 {
		t.Fatalf("expected quantity 1, got %d", products[0].Quantity)
	}
}

func TestGetUnbezahltePositionenFromEvents_AccumulatesMultipleOrders(t *testing.T) {
	products1 := []Position{
		{ID: 1, Name: "Beer", PreisCents: 500, Quantity: 2},
	}
	products2 := []Position{
		{ID: 1, Name: "Beer", PreisCents: 500, Quantity: 3},
	}
	events := []e.Event{
		mustCreateOrderEvent(t, 1, 1, products1),
		mustCreateOrderEvent(t, 1, 1, products2),
	}

	products, err := GetUnbezahltePositionenFromEvents(events)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(products) != 1 {
		t.Fatalf("expected 1 product (merged), got %d", len(products))
	}
	if products[0].Quantity != 5 {
		t.Fatalf("expected quantity 5, got %d", products[0].Quantity)
	}
}

func TestGetUngeliefertePositionenFromEvents_Empty(t *testing.T) {
	products, err := GetUngeliefertePositionenFromEvents([]e.Event{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(products) != 0 {
		t.Fatalf("expected empty products, got %d", len(products))
	}
}

func TestGetUngeliefertePositionenFromEvents_OrderOnly(t *testing.T) {
	orderProducts := []Position{
		{ID: 1, Name: "Beer", PreisCents: 500, Quantity: 2},
	}
	events := []e.Event{
		mustCreateOrderEvent(t, 1, 1, orderProducts),
	}

	products, err := GetUngeliefertePositionenFromEvents(events)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(products) != 1 {
		t.Fatalf("expected 1 product, got %d", len(products))
	}
	if products[0].Quantity != 2 {
		t.Fatalf("expected quantity 2, got %d", products[0].Quantity)
	}
}

func TestGetUngeliefertePositionenFromEvents_PartialDelivery(t *testing.T) {
	orderProducts := []Position{
		{ID: 1, Name: "Beer", PreisCents: 500, Quantity: 3},
	}
	deliveredProducts := []Position{
		{ID: 1, Name: "Beer", PreisCents: 500, Quantity: 1},
	}
	events := []e.Event{
		mustCreateOrderEvent(t, 1, 1, orderProducts),
		mustCreateDeliveryEvent(t, 1, 1, deliveredProducts),
	}

	products, err := GetUngeliefertePositionenFromEvents(events)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(products) != 1 {
		t.Fatalf("expected 1 product, got %d", len(products))
	}
	if products[0].Quantity != 2 {
		t.Fatalf("expected quantity 2, got %d", products[0].Quantity)
	}
}

func TestGetUngeliefertePositionenFromEvents_FullDelivery(t *testing.T) {
	orderProducts := []Position{
		{ID: 1, Name: "Beer", PreisCents: 500, Quantity: 2},
	}
	events := []e.Event{
		mustCreateOrderEvent(t, 1, 1, orderProducts),
		mustCreateDeliveryEvent(t, 1, 1, orderProducts),
	}

	products, err := GetUngeliefertePositionenFromEvents(events)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(products) != 0 {
		t.Fatalf("expected 0 products, got %d", len(products))
	}
}

func TestGetUngeliefertePositionenFromEvents_WithCancelation(t *testing.T) {
	orderProducts := []Position{
		{ID: 1, Name: "Beer", PreisCents: 500, Quantity: 3},
	}
	canceledProducts := []Position{
		{ID: 1, Name: "Beer", PreisCents: 500, Quantity: 2},
	}
	events := []e.Event{
		mustCreateOrderEvent(t, 1, 1, orderProducts),
		mustCreateCancelationEvent(t, 1, 1, canceledProducts),
	}

	products, err := GetUngeliefertePositionenFromEvents(events)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(products) != 1 {
		t.Fatalf("expected 1 product, got %d", len(products))
	}
	if products[0].Quantity != 1 {
		t.Fatalf("expected quantity 1, got %d", products[0].Quantity)
	}
}

func TestGetSaldoFromEvents_WithSnapshot(t *testing.T) {
	// Snapshot represents state: balance = 500 cents
	// Then new order adds 300 cents
	unpaid := []Position{{ID: 1, Name: "Beer", PreisCents: 500, Quantity: 1}}
	undelivered := []Position{{ID: 1, Name: "Beer", PreisCents: 500, Quantity: 1}}
	newLineItems := []Position{{ID: 2, Name: "Fries", PreisCents: 300, Quantity: 1}}

	events := []e.Event{
		mustCreateSnapshotEvent(t, 1, 1, 500, unpaid, undelivered, 0),
		mustCreateOrderEvent(t, 1, 1, newLineItems),
	}

	balance, err := GetSaldoFromEvents(events)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if balance != 800 {
		t.Fatalf("expected balance 800 (500 from snapshot + 300 from order), got %d", balance)
	}
}

func TestGetSaldoFromEvents_SnapshotResetsBalance(t *testing.T) {
	// Old events before snapshot should be ignored
	oldProducts := []Position{{ID: 1, Name: "Beer", PreisCents: 1000, Quantity: 1}}
	unpaid := []Position{}
	undelivered := []Position{}

	events := []e.Event{
		mustCreateOrderEvent(t, 1, 1, oldProducts), // This should be ignored
		mustCreateSnapshotEvent(t, 1, 1, 0, unpaid, undelivered, 1000),
	}

	balance, err := GetSaldoFromEvents(events)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if balance != 0 {
		t.Fatalf("expected balance 0 (from snapshot), got %d", balance)
	}
}

func TestGetUnbezahltePositionenFromEvents_WithSnapshot(t *testing.T) {
	// Snapshot has 2 beers unpaid
	// Then order adds 1 fries
	snapshotUnpaid := []Position{{ID: 1, Name: "Beer", PreisCents: 500, Quantity: 2}}
	newLineItems := []Position{{ID: 2, Name: "Fries", PreisCents: 300, Quantity: 1}}

	events := []e.Event{
		mustCreateSnapshotEvent(t, 1, 1, 1000, snapshotUnpaid, []Position{}, 0),
		mustCreateOrderEvent(t, 1, 1, newLineItems),
	}

	products, err := GetUnbezahltePositionenFromEvents(events)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(products) != 2 {
		t.Fatalf("expected 2 products, got %d", len(products))
	}

	// Find beer and fries
	var beerQty, friesQty int
	for _, p := range products {
		if p.ID == 1 {
			beerQty = p.Quantity
		}
		if p.ID == 2 {
			friesQty = p.Quantity
		}
	}
	if beerQty != 2 {
		t.Fatalf("expected 2 beers, got %d", beerQty)
	}
	if friesQty != 1 {
		t.Fatalf("expected 1 fries, got %d", friesQty)
	}
}

func TestGetUnbezahltePositionenFromEvents_SnapshotResetsState(t *testing.T) {
	// Old order before snapshot should be ignored
	oldProducts := []Position{{ID: 1, Name: "Beer", PreisCents: 500, Quantity: 10}}
	snapshotUnpaid := []Position{{ID: 2, Name: "Fries", PreisCents: 300, Quantity: 1}}

	events := []e.Event{
		mustCreateOrderEvent(t, 1, 1, oldProducts), // This should be ignored
		mustCreateSnapshotEvent(t, 1, 1, 300, snapshotUnpaid, []Position{}, 5000),
	}

	products, err := GetUnbezahltePositionenFromEvents(events)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(products) != 1 {
		t.Fatalf("expected 1 product from snapshot, got %d", len(products))
	}
	if products[0].ID != 2 {
		t.Fatalf("expected product ID 2 (fries), got %d", products[0].ID)
	}
}

func TestGetUngeliefertePositionenFromEvents_WithSnapshot(t *testing.T) {
	// Snapshot has 3 beers undelivered
	// Then order adds 2 fries
	snapshotUndelivered := []Position{{ID: 1, Name: "Beer", PreisCents: 500, Quantity: 3}}
	newLineItems := []Position{{ID: 2, Name: "Fries", PreisCents: 300, Quantity: 2}}

	events := []e.Event{
		mustCreateSnapshotEvent(t, 1, 1, 1500, []Position{}, snapshotUndelivered, 0),
		mustCreateOrderEvent(t, 1, 1, newLineItems),
	}

	products, err := GetUngeliefertePositionenFromEvents(events)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(products) != 2 {
		t.Fatalf("expected 2 products, got %d", len(products))
	}

	// Verify quantities
	var beerQty, friesQty int
	for _, p := range products {
		if p.ID == 1 {
			beerQty = p.Quantity
		}
		if p.ID == 2 {
			friesQty = p.Quantity
		}
	}
	if beerQty != 3 {
		t.Fatalf("expected 3 beers, got %d", beerQty)
	}
	if friesQty != 2 {
		t.Fatalf("expected 2 fries, got %d", friesQty)
	}
}

func TestGetUngeliefertePositionenFromEvents_SnapshotResetsState(t *testing.T) {
	// Old order before snapshot should be ignored
	oldProducts := []Position{{ID: 1, Name: "Beer", PreisCents: 500, Quantity: 10}}
	snapshotUndelivered := []Position{{ID: 2, Name: "Fries", PreisCents: 300, Quantity: 2}}

	events := []e.Event{
		mustCreateOrderEvent(t, 1, 1, oldProducts), // This should be ignored
		mustCreateSnapshotEvent(t, 1, 1, 0, []Position{}, snapshotUndelivered, 5000),
	}

	products, err := GetUngeliefertePositionenFromEvents(events)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(products) != 1 {
		t.Fatalf("expected 1 product from snapshot, got %d", len(products))
	}
	if products[0].ID != 2 {
		t.Fatalf("expected product ID 2 (fries), got %d", products[0].ID)
	}
}

func TestGetGesamtZahlungenFromEvents_WithSnapshot(t *testing.T) {
	// Snapshot has totalPayments = 1000
	// Then payment adds 500
	paidProducts := []Position{{ID: 1, Name: "Beer", PreisCents: 500, Quantity: 1}}

	events := []e.Event{
		mustCreateSnapshotEvent(t, 1, 1, 500, []Position{}, []Position{}, 1000),
		mustCreatePaymentEvent(t, 1, 1, paidProducts),
	}

	totalPayments, err := GetGesamtZahlungenFromEvents(events)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if totalPayments != 1500 {
		t.Fatalf("expected totalPayments 1500 (1000 from snapshot + 500 from payment), got %d", totalPayments)
	}
}

func TestGetGesamtZahlungenFromEvents_SnapshotResetsState(t *testing.T) {
	// Old payment before snapshot should be ignored
	oldPaidProducts := []Position{{ID: 1, Name: "Beer", PreisCents: 1000, Quantity: 10}}

	events := []e.Event{
		mustCreatePaymentEvent(t, 1, 1, oldPaidProducts), // This should be ignored
		mustCreateSnapshotEvent(t, 1, 1, 0, []Position{}, []Position{}, 5000),
	}

	totalPayments, err := GetGesamtZahlungenFromEvents(events)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if totalPayments != 5000 {
		t.Fatalf("expected totalPayments 5000 (from snapshot), got %d", totalPayments)
	}
}

func TestNewSnapshotEvent(t *testing.T) {
	unpaid := []Position{{ID: 1, Name: "Beer", PreisCents: 500, Quantity: 2}}
	undelivered := []Position{{ID: 2, Name: "Fries", PreisCents: 300, Quantity: 1}}

	event, err := NewSnapshotEvent(1, 42, 1000, unpaid, undelivered, 500)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if event.Type != string(EventTypeSnapshotV1) {
		t.Fatalf("expected type %s, got %s", EventTypeSnapshotV1, event.Type)
	}
	if event.Subject != "tisch:42" {
		t.Fatalf("expected subject tisch:42, got %s", event.Subject)
	}
	if event.UserID != 1 {
		t.Fatalf("expected userID 1, got %d", event.UserID)
	}
}
