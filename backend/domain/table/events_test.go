//go:build unit

package table

import (
	"testing"

	e "github.com/nicograef/jotti/backend/domain/event"
)

func mustCreateOrderEvent(t *testing.T, userID, tableID int, products []LineItem) e.Event {
	t.Helper()
	event, err := NewOrderPlacedEvent(userID, tableID, products, "")
	if err != nil {
		t.Fatalf("failed to create order event: %v", err)
	}
	return event
}

func mustCreatePaymentEvent(t *testing.T, userID, tableID int, products []LineItem) e.Event {
	t.Helper()
	event, err := NewPaymentRegisteredEvent(userID, tableID, products, "")
	if err != nil {
		t.Fatalf("failed to create payment event: %v", err)
	}
	return event
}

func mustCreateCancelationEvent(t *testing.T, userID, tableID int, products []LineItem) e.Event {
	t.Helper()
	event, err := NewVariantsCanceledEvent(userID, tableID, products, "")
	if err != nil {
		t.Fatalf("failed to create cancelation event: %v", err)
	}
	return event
}

func mustCreateDeliveryEvent(t *testing.T, userID, tableID int, products []LineItem) e.Event {
	t.Helper()
	event, err := NewVariantsDeliveredEvent(userID, tableID, products, "")
	if err != nil {
		t.Fatalf("failed to create delivery event: %v", err)
	}
	return event
}

func mustCreateSnapshotEvent(t *testing.T, userID, tableID int, balance int, unpaid, undelivered []LineItem, totalPayments int) e.Event {
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
	products := []LineItem{
		{ID: 1, Name: "Beer", PriceCents: 500, Quantity: 2},
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
	products := []LineItem{
		{ID: 1, Name: "Beer", PriceCents: 500, Quantity: 2},
	}
	paidProducts := []LineItem{
		{ID: 1, Name: "Beer", PriceCents: 500, Quantity: 1},
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
	products := []LineItem{
		{ID: 1, Name: "Beer", PriceCents: 500, Quantity: 2},
	}
	paidProducts := []LineItem{
		{ID: 1, Name: "Beer", PriceCents: 500, Quantity: 1},
	}
	canceledProducts := []LineItem{
		{ID: 1, Name: "Beer", PriceCents: 500, Quantity: 1},
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
	products := []LineItem{
		{ID: 1, Name: "Beer", PriceCents: 500, Quantity: 1},
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
	products := []LineItem{
		{ID: 1, Name: "Beer", PriceCents: 500, Quantity: 1},
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

func TestGetUnpaidVariantsFromEvents_Empty(t *testing.T) {
	products, err := GetUnpaidVariantsFromEvents([]e.Event{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(products) != 0 {
		t.Fatalf("expected empty products, got %d", len(products))
	}
}

func TestGetUnpaidVariantsFromEvents_OrderOnly(t *testing.T) {
	orderProducts := []LineItem{
		{ID: 1, Name: "Beer", PriceCents: 500, Quantity: 2},
		{ID: 2, Name: "Fries", PriceCents: 300, Quantity: 1},
	}
	events := []e.Event{
		mustCreateOrderEvent(t, 1, 1, orderProducts),
	}

	products, err := GetUnpaidVariantsFromEvents(events)
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

func TestGetUnpaidVariantsFromEvents_PartialPayment(t *testing.T) {
	orderProducts := []LineItem{
		{ID: 1, Name: "Beer", PriceCents: 500, Quantity: 3},
	}
	paidProducts := []LineItem{
		{ID: 1, Name: "Beer", PriceCents: 500, Quantity: 1},
	}
	events := []e.Event{
		mustCreateOrderEvent(t, 1, 1, orderProducts),
		mustCreatePaymentEvent(t, 1, 1, paidProducts),
	}

	products, err := GetUnpaidVariantsFromEvents(events)
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

func TestGetUnpaidVariantsFromEvents_FullPayment(t *testing.T) {
	orderProducts := []LineItem{
		{ID: 1, Name: "Beer", PriceCents: 500, Quantity: 2},
	}
	events := []e.Event{
		mustCreateOrderEvent(t, 1, 1, orderProducts),
		mustCreatePaymentEvent(t, 1, 1, orderProducts),
	}

	products, err := GetUnpaidVariantsFromEvents(events)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(products) != 0 {
		t.Fatalf("expected 0 products, got %d", len(products))
	}
}

func TestGetUnpaidVariantsFromEvents_WithCancelation(t *testing.T) {
	orderProducts := []LineItem{
		{ID: 1, Name: "Beer", PriceCents: 500, Quantity: 3},
	}
	canceledProducts := []LineItem{
		{ID: 1, Name: "Beer", PriceCents: 500, Quantity: 2},
	}
	events := []e.Event{
		mustCreateOrderEvent(t, 1, 1, orderProducts),
		mustCreateCancelationEvent(t, 1, 1, canceledProducts),
	}

	products, err := GetUnpaidVariantsFromEvents(events)
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

func TestGetUnpaidVariantsFromEvents_AccumulatesMultipleOrders(t *testing.T) {
	products1 := []LineItem{
		{ID: 1, Name: "Beer", PriceCents: 500, Quantity: 2},
	}
	products2 := []LineItem{
		{ID: 1, Name: "Beer", PriceCents: 500, Quantity: 3},
	}
	events := []e.Event{
		mustCreateOrderEvent(t, 1, 1, products1),
		mustCreateOrderEvent(t, 1, 1, products2),
	}

	products, err := GetUnpaidVariantsFromEvents(events)
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

func TestGetUndeliveredVariantsFromEvents_Empty(t *testing.T) {
	products, err := GetUndeliveredVariantsFromEvents([]e.Event{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(products) != 0 {
		t.Fatalf("expected empty products, got %d", len(products))
	}
}

func TestGetUndeliveredVariantsFromEvents_OrderOnly(t *testing.T) {
	orderProducts := []LineItem{
		{ID: 1, Name: "Beer", PriceCents: 500, Quantity: 2},
	}
	events := []e.Event{
		mustCreateOrderEvent(t, 1, 1, orderProducts),
	}

	products, err := GetUndeliveredVariantsFromEvents(events)
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

func TestGetUndeliveredVariantsFromEvents_PartialDelivery(t *testing.T) {
	orderProducts := []LineItem{
		{ID: 1, Name: "Beer", PriceCents: 500, Quantity: 3},
	}
	deliveredProducts := []LineItem{
		{ID: 1, Name: "Beer", PriceCents: 500, Quantity: 1},
	}
	events := []e.Event{
		mustCreateOrderEvent(t, 1, 1, orderProducts),
		mustCreateDeliveryEvent(t, 1, 1, deliveredProducts),
	}

	products, err := GetUndeliveredVariantsFromEvents(events)
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

func TestGetUndeliveredVariantsFromEvents_FullDelivery(t *testing.T) {
	orderProducts := []LineItem{
		{ID: 1, Name: "Beer", PriceCents: 500, Quantity: 2},
	}
	events := []e.Event{
		mustCreateOrderEvent(t, 1, 1, orderProducts),
		mustCreateDeliveryEvent(t, 1, 1, orderProducts),
	}

	products, err := GetUndeliveredVariantsFromEvents(events)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(products) != 0 {
		t.Fatalf("expected 0 products, got %d", len(products))
	}
}

func TestGetUndeliveredVariantsFromEvents_WithCancelation(t *testing.T) {
	orderProducts := []LineItem{
		{ID: 1, Name: "Beer", PriceCents: 500, Quantity: 3},
	}
	canceledProducts := []LineItem{
		{ID: 1, Name: "Beer", PriceCents: 500, Quantity: 2},
	}
	events := []e.Event{
		mustCreateOrderEvent(t, 1, 1, orderProducts),
		mustCreateCancelationEvent(t, 1, 1, canceledProducts),
	}

	products, err := GetUndeliveredVariantsFromEvents(events)
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
	unpaid := []LineItem{{ID: 1, Name: "Beer", PriceCents: 500, Quantity: 1}}
	undelivered := []LineItem{{ID: 1, Name: "Beer", PriceCents: 500, Quantity: 1}}
	newLineItems := []LineItem{{ID: 2, Name: "Fries", PriceCents: 300, Quantity: 1}}

	events := []e.Event{
		mustCreateSnapshotEvent(t, 1, 1, 500, unpaid, undelivered, 0),
		mustCreateOrderEvent(t, 1, 1, newLineItems),
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
	oldProducts := []LineItem{{ID: 1, Name: "Beer", PriceCents: 1000, Quantity: 1}}
	unpaid := []LineItem{}
	undelivered := []LineItem{}

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

func TestGetUnpaidVariantsFromEvents_WithSnapshot(t *testing.T) {
	// Snapshot has 2 beers unpaid
	// Then order adds 1 fries
	snapshotUnpaid := []LineItem{{ID: 1, Name: "Beer", PriceCents: 500, Quantity: 2}}
	newLineItems := []LineItem{{ID: 2, Name: "Fries", PriceCents: 300, Quantity: 1}}

	events := []e.Event{
		mustCreateSnapshotEvent(t, 1, 1, 1000, snapshotUnpaid, []LineItem{}, 0),
		mustCreateOrderEvent(t, 1, 1, newLineItems),
	}

	products, err := GetUnpaidVariantsFromEvents(events)
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

func TestGetUnpaidVariantsFromEvents_SnapshotResetsState(t *testing.T) {
	// Old order before snapshot should be ignored
	oldProducts := []LineItem{{ID: 1, Name: "Beer", PriceCents: 500, Quantity: 10}}
	snapshotUnpaid := []LineItem{{ID: 2, Name: "Fries", PriceCents: 300, Quantity: 1}}

	events := []e.Event{
		mustCreateOrderEvent(t, 1, 1, oldProducts), // This should be ignored
		mustCreateSnapshotEvent(t, 1, 1, 300, snapshotUnpaid, []LineItem{}, 5000),
	}

	products, err := GetUnpaidVariantsFromEvents(events)
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

func TestGetUndeliveredVariantsFromEvents_WithSnapshot(t *testing.T) {
	// Snapshot has 3 beers undelivered
	// Then order adds 2 fries
	snapshotUndelivered := []LineItem{{ID: 1, Name: "Beer", PriceCents: 500, Quantity: 3}}
	newLineItems := []LineItem{{ID: 2, Name: "Fries", PriceCents: 300, Quantity: 2}}

	events := []e.Event{
		mustCreateSnapshotEvent(t, 1, 1, 1500, []LineItem{}, snapshotUndelivered, 0),
		mustCreateOrderEvent(t, 1, 1, newLineItems),
	}

	products, err := GetUndeliveredVariantsFromEvents(events)
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

func TestGetUndeliveredVariantsFromEvents_SnapshotResetsState(t *testing.T) {
	// Old order before snapshot should be ignored
	oldProducts := []LineItem{{ID: 1, Name: "Beer", PriceCents: 500, Quantity: 10}}
	snapshotUndelivered := []LineItem{{ID: 2, Name: "Fries", PriceCents: 300, Quantity: 2}}

	events := []e.Event{
		mustCreateOrderEvent(t, 1, 1, oldProducts), // This should be ignored
		mustCreateSnapshotEvent(t, 1, 1, 0, []LineItem{}, snapshotUndelivered, 5000),
	}

	products, err := GetUndeliveredVariantsFromEvents(events)
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
	paidProducts := []LineItem{{ID: 1, Name: "Beer", PriceCents: 500, Quantity: 1}}

	events := []e.Event{
		mustCreateSnapshotEvent(t, 1, 1, 500, []LineItem{}, []LineItem{}, 1000),
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
	oldPaidProducts := []LineItem{{ID: 1, Name: "Beer", PriceCents: 1000, Quantity: 10}}

	events := []e.Event{
		mustCreatePaymentEvent(t, 1, 1, oldPaidProducts), // This should be ignored
		mustCreateSnapshotEvent(t, 1, 1, 0, []LineItem{}, []LineItem{}, 5000),
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
	unpaid := []LineItem{{ID: 1, Name: "Beer", PriceCents: 500, Quantity: 2}}
	undelivered := []LineItem{{ID: 2, Name: "Fries", PriceCents: 300, Quantity: 1}}

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
