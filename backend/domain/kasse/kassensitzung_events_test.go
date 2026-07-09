//go:build unit

package kasse

import (
	"strings"
	"testing"
)

// Regression: zog wertet den Zero-Value 0 bei Required() als fehlend. Ein
// Anfangsbestand von 0 Cent ist fachlich gültig und darf die Event-Erstellung
// nicht ablehnen (die HTTP-Schicht erlaubt 0 via Ptr+NotNil ausdrücklich).
func TestNewKassensitzungEroeffnetEvent_ErlaubtNullBetrag(t *testing.T) {
	event, err := NewKassensitzungEroeffnetEvent(KassensitzungSubject(1), 1, "TestUser", "2026-07-09", "ops-smoke", 0)
	if err != nil {
		t.Fatalf("expected no error for betragCents 0, got %v", err)
	}
	if event.Type != string(EventTypeKassensitzungEroeffnetV1) {
		t.Errorf("expected type %s, got %s", EventTypeKassensitzungEroeffnetV1, event.Type)
	}
}

func TestNewKassensitzungEroeffnetEvent_LehntNegativenBetragAb(t *testing.T) {
	_, err := NewKassensitzungEroeffnetEvent(KassensitzungSubject(1), 1, "TestUser", "2026-07-09", "ops-smoke", -1)
	if err == nil {
		t.Fatal("expected error for negative betragCents, got nil")
	}
	if !strings.Contains(err.Error(), "BetragCents") {
		t.Errorf("expected BetragCents validation error, got %v", err)
	}
}

// Regression: eine leer gezählte Kasse (Ist-Bestand 0 Cent) ist ein gültiger
// Kassensturz.
func TestNewKassensturzDurchgefuehrtEvent_ErlaubtNullIstBestand(t *testing.T) {
	event, err := NewKassensturzDurchgefuehrtEvent(KassensitzungSubject(1), 1, "TestUser", 0, 0, 0)
	if err != nil {
		t.Fatalf("expected no error for istBestandCents 0, got %v", err)
	}
	if event.Type != string(EventTypeKassensturzDurchgefuehrtV1) {
		t.Errorf("expected type %s, got %s", EventTypeKassensturzDurchgefuehrtV1, event.Type)
	}
}

func TestNewKassensturzDurchgefuehrtEvent_LehntNegativenIstBestandAb(t *testing.T) {
	_, err := NewKassensturzDurchgefuehrtEvent(KassensitzungSubject(1), 1, "TestUser", 100, -1, -101)
	if err == nil {
		t.Fatal("expected error for negative istBestandCents, got nil")
	}
	if !strings.Contains(err.Error(), "IstBestandCents") {
		t.Errorf("expected IstBestandCents validation error, got %v", err)
	}
}
