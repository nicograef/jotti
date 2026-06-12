//go:build unit

package kasse

import (
	"testing"

	e "github.com/nicograef/jotti/backend/domain/event"
)

func TestGetHistorieFromEvents_Empty(t *testing.T) {
	history, err := GetHistorieFromEvents([]e.Event{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(history) != 0 {
		t.Fatalf("expected empty history, got %d items", len(history))
	}
}

func TestGetHistorieFromEvents_ReturnsAllEventTypes(t *testing.T) {
	products := []Position{
		testPosition(1, "Beer", "Pils 0.5l", "getraenk", 500, 1),
	}
	orderEvent := mustCreateOrderEvent(t, testSubject, 1, products)
	positions := positionsFromOrder(t, orderEvent, 1)

	events := []e.Event{
		orderEvent,
		mustCreatePaymentEvent(t, testSubject, 1, positions, 500),
		mustCreateCancelationEvent(t, testSubject, 1, positions, 500),
		mustCreateDeliveryEvent(t, testSubject, 1, positions),
	}

	history, err := GetHistorieFromEvents(events)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(history) != 4 {
		t.Fatalf("expected 4 history items, got %d", len(history))
	}
}

func TestGetHistorieFromEvents_ReversesOrder(t *testing.T) {
	products := []Position{
		testPosition(1, "Beer", "Pils 0.5l", "getraenk", 500, 1),
	}
	orderEvent := mustCreateOrderEvent(t, testSubject, 1, products)
	positions := positionsFromOrder(t, orderEvent, 1)

	events := []e.Event{
		orderEvent,
		mustCreatePaymentEvent(t, testSubject, 1, positions, 500),
	}

	history, err := GetHistorieFromEvents(events)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	// First item should be payment (last event), second should be order (first event)
	if history[0].Art != HistorieEintragZahlung || history[0].Zahlung == nil {
		t.Fatalf("expected first item to be Zahlung, got kind %q", history[0].Art)
	}
	if history[1].Art != HistorieEintragBestellung || history[1].Bestellung == nil {
		t.Fatalf("expected second item to be Bestellung, got kind %q", history[1].Art)
	}
}

func TestGetHistorieFromEvents_IncludesAuszahlung(t *testing.T) {
	products := []Position{
		testPosition(1, "Beer", "Pils 0.5l", "getraenk", 500, 1),
	}
	orderEvent := mustCreateOrderEvent(t, testSubject, 1, products)
	positions := positionsFromOrder(t, orderEvent, 1)

	events := []e.Event{
		orderEvent,
		mustCreatePaymentEvent(t, testSubject, 1, positions, 500),
		mustCreateCancelationEvent(t, testSubject, 1, positions, 500),
		mustCreateAuszahlungEvent(t, testSubject, 1, 500, "Rueckzahlung"),
	}

	history, err := GetHistorieFromEvents(events)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(history) != 4 {
		t.Fatalf("expected 4 history items, got %d", len(history))
	}
	// Most recent first: auszahlung should be at index 0
	if history[0].Art != HistorieEintragAuszahlung || history[0].Auszahlung == nil {
		t.Fatalf("expected first item to be Auszahlung, got kind %q", history[0].Art)
	}
	if history[0].Auszahlung.BetragCents != 500 {
		t.Fatalf("expected BetragCents 500, got %d", history[0].Auszahlung.BetragCents)
	}
	if history[0].Auszahlung.Kommentar != "Rueckzahlung" {
		t.Fatalf("expected Kommentar Rueckzahlung, got %s", history[0].Auszahlung.Kommentar)
	}
}

func TestNewAuszahlungGeleistetEvent_InvalidBetrag(t *testing.T) {
	_, err := NewAuszahlungGeleistetEvent(testSubject, 1, "TestUser", 0, "Rueckzahlung")
	if err == nil {
		t.Fatal("expected error for betragCents=0, got nil")
	}
}

func TestNewAuszahlungGeleistetEvent_InvalidKommentar(t *testing.T) {
	_, err := NewAuszahlungGeleistetEvent(testSubject, 1, "TestUser", 500, "ab")
	if err == nil {
		t.Fatal("expected error for kommentar too short, got nil")
	}
}
