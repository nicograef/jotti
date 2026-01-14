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
