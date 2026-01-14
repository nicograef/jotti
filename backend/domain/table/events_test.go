//go:build unit

package table

import (
	"testing"

	e "github.com/nicograef/jotti/backend/domain/event"
)

func mustCreateOrderEvent(t *testing.T, userID, tableID int, products []OrderProduct) e.Event {
	t.Helper()
	event, err := NewOrderPlacedEvent(userID, tableID, products, "")
	if err != nil {
		t.Fatalf("failed to create order event: %v", err)
	}
	return event
}

func mustCreatePaymentEvent(t *testing.T, userID, tableID int, products []OrderProduct) e.Event {
	t.Helper()
	event, err := NewPaymentRegisteredEvent(userID, tableID, products, "")
	if err != nil {
		t.Fatalf("failed to create payment event: %v", err)
	}
	return event
}

func mustCreateCancelationEvent(t *testing.T, userID, tableID int, products []OrderProduct) e.Event {
	t.Helper()
	event, err := NewProductsCanceledEvent(userID, tableID, products, "")
	if err != nil {
		t.Fatalf("failed to create cancelation event: %v", err)
	}
	return event
}

func mustCreateDeliveryEvent(t *testing.T, userID, tableID int, products []OrderProduct) e.Event {
	t.Helper()
	event, err := NewProductsDeliveredEvent(userID, tableID, products, "")
	if err != nil {
		t.Fatalf("failed to create delivery event: %v", err)
	}
	return event
}

func mustCreateSnapshotEvent(t *testing.T, userID, tableID int, balance int, unpaid, undelivered []OrderProduct, totalPayments int) e.Event {
	t.Helper()
	event, err := NewSnapshotEvent(userID, tableID, balance, unpaid, undelivered, totalPayments)
	if err != nil {
		t.Fatalf("failed to create snapshot event: %v", err)
	}
	return event
}

func TestGetBalanceFromEvents_Empty(t *testing.T) {
	balance, err := GetBalanceFromEvents([]e.Event{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if balance != 0 {
		t.Fatalf("expected balance 0, got %d", balance)
	}
}

func TestGetBalanceFromEvents_OrderOnly(t *testing.T) {
	products := []OrderProduct{
		{ID: 1, Name: "Beer", NetPriceCents: 500, Quantity: 2},
	}
	events := []e.Event{
		mustCreateOrderEvent(t, 1, 1, products),
	}

	balance, err := GetBalanceFromEvents(events)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if balance != 1000 {
		t.Fatalf("expected balance 1000, got %d", balance)
	}
}

func TestGetBalanceFromEvents_OrderAndPayment(t *testing.T) {
	products := []OrderProduct{
		{ID: 1, Name: "Beer", NetPriceCents: 500, Quantity: 2},
	}
	paidProducts := []OrderProduct{
		{ID: 1, Name: "Beer", NetPriceCents: 500, Quantity: 1},
	}
	events := []e.Event{
		mustCreateOrderEvent(t, 1, 1, products),
		mustCreatePaymentEvent(t, 1, 1, paidProducts),
	}

	balance, err := GetBalanceFromEvents(events)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if balance != 500 {
		t.Fatalf("expected balance 500, got %d", balance)
	}
}

func TestGetBalanceFromEvents_OrderPaymentAndCancelation(t *testing.T) {
	products := []OrderProduct{
		{ID: 1, Name: "Beer", NetPriceCents: 500, Quantity: 2},
	}
	paidProducts := []OrderProduct{
		{ID: 1, Name: "Beer", NetPriceCents: 500, Quantity: 1},
	}
	canceledProducts := []OrderProduct{
		{ID: 1, Name: "Beer", NetPriceCents: 500, Quantity: 1},
	}
	events := []e.Event{
		mustCreateOrderEvent(t, 1, 1, products),
		mustCreatePaymentEvent(t, 1, 1, paidProducts),
		mustCreateCancelationEvent(t, 1, 1, canceledProducts),
	}

	balance, err := GetBalanceFromEvents(events)
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
	products := []OrderProduct{
		{ID: 1, Name: "Beer", NetPriceCents: 500, Quantity: 1},
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
	products := []OrderProduct{
		{ID: 1, Name: "Beer", NetPriceCents: 500, Quantity: 1},
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
	if _, ok := history[0].(Payment); !ok {
		t.Fatalf("expected first item to be Payment, got %T", history[0])
	}
	if _, ok := history[1].(Order); !ok {
		t.Fatalf("expected second item to be Order, got %T", history[1])
	}
}

func TestGetUnpaidProductsFromEvents_Empty(t *testing.T) {
	products, err := GetUnpaidProductsFromEvents([]e.Event{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(products) != 0 {
		t.Fatalf("expected empty products, got %d", len(products))
	}
}

func TestGetUnpaidProductsFromEvents_OrderOnly(t *testing.T) {
	orderProducts := []OrderProduct{
		{ID: 1, Name: "Beer", NetPriceCents: 500, Quantity: 2},
		{ID: 2, Name: "Fries", NetPriceCents: 300, Quantity: 1},
	}
	events := []e.Event{
		mustCreateOrderEvent(t, 1, 1, orderProducts),
	}

	products, err := GetUnpaidProductsFromEvents(events)
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

func TestGetUnpaidProductsFromEvents_PartialPayment(t *testing.T) {
	orderProducts := []OrderProduct{
		{ID: 1, Name: "Beer", NetPriceCents: 500, Quantity: 3},
	}
	paidProducts := []OrderProduct{
		{ID: 1, Name: "Beer", NetPriceCents: 500, Quantity: 1},
	}
	events := []e.Event{
		mustCreateOrderEvent(t, 1, 1, orderProducts),
		mustCreatePaymentEvent(t, 1, 1, paidProducts),
	}

	products, err := GetUnpaidProductsFromEvents(events)
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

func TestGetUnpaidProductsFromEvents_FullPayment(t *testing.T) {
	orderProducts := []OrderProduct{
		{ID: 1, Name: "Beer", NetPriceCents: 500, Quantity: 2},
	}
	events := []e.Event{
		mustCreateOrderEvent(t, 1, 1, orderProducts),
		mustCreatePaymentEvent(t, 1, 1, orderProducts),
	}

	products, err := GetUnpaidProductsFromEvents(events)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(products) != 0 {
		t.Fatalf("expected 0 products, got %d", len(products))
	}
}

func TestGetUnpaidProductsFromEvents_WithCancelation(t *testing.T) {
	orderProducts := []OrderProduct{
		{ID: 1, Name: "Beer", NetPriceCents: 500, Quantity: 3},
	}
	canceledProducts := []OrderProduct{
		{ID: 1, Name: "Beer", NetPriceCents: 500, Quantity: 2},
	}
	events := []e.Event{
		mustCreateOrderEvent(t, 1, 1, orderProducts),
		mustCreateCancelationEvent(t, 1, 1, canceledProducts),
	}

	products, err := GetUnpaidProductsFromEvents(events)
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

func TestGetUnpaidProductsFromEvents_AccumulatesMultipleOrders(t *testing.T) {
	products1 := []OrderProduct{
		{ID: 1, Name: "Beer", NetPriceCents: 500, Quantity: 2},
	}
	products2 := []OrderProduct{
		{ID: 1, Name: "Beer", NetPriceCents: 500, Quantity: 3},
	}
	events := []e.Event{
		mustCreateOrderEvent(t, 1, 1, products1),
		mustCreateOrderEvent(t, 1, 1, products2),
	}

	products, err := GetUnpaidProductsFromEvents(events)
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

func TestGetUndeliveredProductsFromEvents_Empty(t *testing.T) {
	products, err := GetUndeliveredProductsFromEvents([]e.Event{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(products) != 0 {
		t.Fatalf("expected empty products, got %d", len(products))
	}
}

func TestGetUndeliveredProductsFromEvents_OrderOnly(t *testing.T) {
	orderProducts := []OrderProduct{
		{ID: 1, Name: "Beer", NetPriceCents: 500, Quantity: 2},
	}
	events := []e.Event{
		mustCreateOrderEvent(t, 1, 1, orderProducts),
	}

	products, err := GetUndeliveredProductsFromEvents(events)
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

func TestGetUndeliveredProductsFromEvents_PartialDelivery(t *testing.T) {
	orderProducts := []OrderProduct{
		{ID: 1, Name: "Beer", NetPriceCents: 500, Quantity: 3},
	}
	deliveredProducts := []OrderProduct{
		{ID: 1, Name: "Beer", NetPriceCents: 500, Quantity: 1},
	}
	events := []e.Event{
		mustCreateOrderEvent(t, 1, 1, orderProducts),
		mustCreateDeliveryEvent(t, 1, 1, deliveredProducts),
	}

	products, err := GetUndeliveredProductsFromEvents(events)
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

func TestGetUndeliveredProductsFromEvents_FullDelivery(t *testing.T) {
	orderProducts := []OrderProduct{
		{ID: 1, Name: "Beer", NetPriceCents: 500, Quantity: 2},
	}
	events := []e.Event{
		mustCreateOrderEvent(t, 1, 1, orderProducts),
		mustCreateDeliveryEvent(t, 1, 1, orderProducts),
	}

	products, err := GetUndeliveredProductsFromEvents(events)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(products) != 0 {
		t.Fatalf("expected 0 products, got %d", len(products))
	}
}

func TestGetUndeliveredProductsFromEvents_WithCancelation(t *testing.T) {
	orderProducts := []OrderProduct{
		{ID: 1, Name: "Beer", NetPriceCents: 500, Quantity: 3},
	}
	canceledProducts := []OrderProduct{
		{ID: 1, Name: "Beer", NetPriceCents: 500, Quantity: 2},
	}
	events := []e.Event{
		mustCreateOrderEvent(t, 1, 1, orderProducts),
		mustCreateCancelationEvent(t, 1, 1, canceledProducts),
	}

	products, err := GetUndeliveredProductsFromEvents(events)
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

func TestGetBalanceFromEvents_WithSnapshot(t *testing.T) {
	// Snapshot represents state: balance = 500 cents
	// Then new order adds 300 cents
	unpaid := []OrderProduct{{ID: 1, Name: "Beer", NetPriceCents: 500, Quantity: 1}}
	undelivered := []OrderProduct{{ID: 1, Name: "Beer", NetPriceCents: 500, Quantity: 1}}
	newOrderProducts := []OrderProduct{{ID: 2, Name: "Fries", NetPriceCents: 300, Quantity: 1}}

	events := []e.Event{
		mustCreateSnapshotEvent(t, 1, 1, 500, unpaid, undelivered, 0),
		mustCreateOrderEvent(t, 1, 1, newOrderProducts),
	}

	balance, err := GetBalanceFromEvents(events)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if balance != 800 {
		t.Fatalf("expected balance 800 (500 from snapshot + 300 from order), got %d", balance)
	}
}

func TestGetBalanceFromEvents_SnapshotResetsBalance(t *testing.T) {
	// Old events before snapshot should be ignored
	oldProducts := []OrderProduct{{ID: 1, Name: "Beer", NetPriceCents: 1000, Quantity: 1}}
	unpaid := []OrderProduct{}
	undelivered := []OrderProduct{}

	events := []e.Event{
		mustCreateOrderEvent(t, 1, 1, oldProducts), // This should be ignored
		mustCreateSnapshotEvent(t, 1, 1, 0, unpaid, undelivered, 1000),
	}

	balance, err := GetBalanceFromEvents(events)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if balance != 0 {
		t.Fatalf("expected balance 0 (from snapshot), got %d", balance)
	}
}

func TestGetUnpaidProductsFromEvents_WithSnapshot(t *testing.T) {
	// Snapshot has 2 beers unpaid
	// Then order adds 1 fries
	snapshotUnpaid := []OrderProduct{{ID: 1, Name: "Beer", NetPriceCents: 500, Quantity: 2}}
	newOrderProducts := []OrderProduct{{ID: 2, Name: "Fries", NetPriceCents: 300, Quantity: 1}}

	events := []e.Event{
		mustCreateSnapshotEvent(t, 1, 1, 1000, snapshotUnpaid, []OrderProduct{}, 0),
		mustCreateOrderEvent(t, 1, 1, newOrderProducts),
	}

	products, err := GetUnpaidProductsFromEvents(events)
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

func TestGetUnpaidProductsFromEvents_SnapshotResetsState(t *testing.T) {
	// Old order before snapshot should be ignored
	oldProducts := []OrderProduct{{ID: 1, Name: "Beer", NetPriceCents: 500, Quantity: 10}}
	snapshotUnpaid := []OrderProduct{{ID: 2, Name: "Fries", NetPriceCents: 300, Quantity: 1}}

	events := []e.Event{
		mustCreateOrderEvent(t, 1, 1, oldProducts), // This should be ignored
		mustCreateSnapshotEvent(t, 1, 1, 300, snapshotUnpaid, []OrderProduct{}, 5000),
	}

	products, err := GetUnpaidProductsFromEvents(events)
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

func TestGetUndeliveredProductsFromEvents_WithSnapshot(t *testing.T) {
	// Snapshot has 3 beers undelivered
	// Then order adds 2 fries
	snapshotUndelivered := []OrderProduct{{ID: 1, Name: "Beer", NetPriceCents: 500, Quantity: 3}}
	newOrderProducts := []OrderProduct{{ID: 2, Name: "Fries", NetPriceCents: 300, Quantity: 2}}

	events := []e.Event{
		mustCreateSnapshotEvent(t, 1, 1, 1500, []OrderProduct{}, snapshotUndelivered, 0),
		mustCreateOrderEvent(t, 1, 1, newOrderProducts),
	}

	products, err := GetUndeliveredProductsFromEvents(events)
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

func TestGetUndeliveredProductsFromEvents_SnapshotResetsState(t *testing.T) {
	// Old order before snapshot should be ignored
	oldProducts := []OrderProduct{{ID: 1, Name: "Beer", NetPriceCents: 500, Quantity: 10}}
	snapshotUndelivered := []OrderProduct{{ID: 2, Name: "Fries", NetPriceCents: 300, Quantity: 2}}

	events := []e.Event{
		mustCreateOrderEvent(t, 1, 1, oldProducts), // This should be ignored
		mustCreateSnapshotEvent(t, 1, 1, 0, []OrderProduct{}, snapshotUndelivered, 5000),
	}

	products, err := GetUndeliveredProductsFromEvents(events)
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

func TestGetTotalPaymentsFromEvents_WithSnapshot(t *testing.T) {
	// Snapshot has totalPayments = 1000
	// Then payment adds 500
	paidProducts := []OrderProduct{{ID: 1, Name: "Beer", NetPriceCents: 500, Quantity: 1}}

	events := []e.Event{
		mustCreateSnapshotEvent(t, 1, 1, 500, []OrderProduct{}, []OrderProduct{}, 1000),
		mustCreatePaymentEvent(t, 1, 1, paidProducts),
	}

	totalPayments, err := GetTotalPaymentsFromEvents(events)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if totalPayments != 1500 {
		t.Fatalf("expected totalPayments 1500 (1000 from snapshot + 500 from payment), got %d", totalPayments)
	}
}

func TestGetTotalPaymentsFromEvents_SnapshotResetsState(t *testing.T) {
	// Old payment before snapshot should be ignored
	oldPaidProducts := []OrderProduct{{ID: 1, Name: "Beer", NetPriceCents: 1000, Quantity: 10}}

	events := []e.Event{
		mustCreatePaymentEvent(t, 1, 1, oldPaidProducts), // This should be ignored
		mustCreateSnapshotEvent(t, 1, 1, 0, []OrderProduct{}, []OrderProduct{}, 5000),
	}

	totalPayments, err := GetTotalPaymentsFromEvents(events)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if totalPayments != 5000 {
		t.Fatalf("expected totalPayments 5000 (from snapshot), got %d", totalPayments)
	}
}

func TestNewSnapshotEvent(t *testing.T) {
	unpaid := []OrderProduct{{ID: 1, Name: "Beer", NetPriceCents: 500, Quantity: 2}}
	undelivered := []OrderProduct{{ID: 2, Name: "Fries", NetPriceCents: 300, Quantity: 1}}

	event, err := NewSnapshotEvent(1, 42, 1000, unpaid, undelivered, 500)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if event.Type != string(EventTypeSnapshotV1) {
		t.Fatalf("expected type %s, got %s", EventTypeSnapshotV1, event.Type)
	}
	if event.Subject != "table:42" {
		t.Fatalf("expected subject table:42, got %s", event.Subject)
	}
	if event.UserID != 1 {
		t.Fatalf("expected userID 1, got %d", event.UserID)
	}
}
