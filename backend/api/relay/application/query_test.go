//go:build unit

package application

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
)

// --- Mocks ---

type mockEventRepo struct {
	events []event.Event
	err    error
}

func (m *mockEventRepo) GetBestellungEventsSinceCursor(_ context.Context, cursor int) ([]event.Event, error) {
	if m.err != nil {
		return nil, m.err
	}
	var result []event.Event
	for _, e := range m.events {
		if e.ID > cursor {
			result = append(result, e)
		}
	}
	return result, nil
}

type mockDruckerRepo struct {
	konfigs map[string]DruckerKonfig
	err     error
}

func (m *mockDruckerRepo) GetKonfigurierteKategorieDrucker(_ context.Context) (map[string]DruckerKonfig, error) {
	return m.konfigs, m.err
}

// --- Helpers ---

func makeBestellungEvent(id int, subject string, positionen []kasse.Position, kommentar string) event.Event {
	data, _ := json.Marshal(bestellungEventData{
		Positionen: positionen,
		Kommentar:  kommentar,
	})
	return event.Event{
		ID:       id,
		UserName: "Maria",
		Subject:  subject,
		Time:     time.Date(2026, 3, 17, 19, 34, 0, 0, time.UTC),
		Data:     data,
	}
}

// --- GetDruckAuftraege Tests ---

func TestGetDruckAuftraege_NoEvents_ReturnsNil(t *testing.T) {
	q := Query{
		EventRepo:   &mockEventRepo{events: nil},
		DruckerRepo: &mockDruckerRepo{konfigs: map[string]DruckerKonfig{}},
	}

	auftraege, err := q.GetDruckAuftraege(context.Background(), 0)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if auftraege != nil {
		t.Errorf("expected nil, got %v", auftraege)
	}
}

func TestGetDruckAuftraege_EventRepoError(t *testing.T) {
	q := Query{
		EventRepo:   &mockEventRepo{err: errors.New("db error")},
		DruckerRepo: &mockDruckerRepo{konfigs: map[string]DruckerKonfig{}},
	}

	_, err := q.GetDruckAuftraege(context.Background(), 0)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetDruckAuftraege_DruckerRepoError(t *testing.T) {
	positionen := []kasse.Position{
		{ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Menge: 1},
	}
	events := []event.Event{makeBestellungEvent(1, "tisch:3", positionen, "")}

	q := Query{
		EventRepo:   &mockEventRepo{events: events},
		DruckerRepo: &mockDruckerRepo{err: errors.New("db error")},
	}

	_, err := q.GetDruckAuftraege(context.Background(), 0)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetDruckAuftraege_WithEvents_GeneratesAuftraege(t *testing.T) {
	positionen := []kasse.Position{
		{ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Menge: 2},
	}
	events := []event.Event{makeBestellungEvent(5, "tisch:3", positionen, "")}

	konfigs := map[string]DruckerKonfig{
		"getraenk": {IP: "192.168.1.50", Bonmodus: "pro_position"},
	}
	q := Query{
		EventRepo:   &mockEventRepo{events: events},
		DruckerRepo: &mockDruckerRepo{konfigs: konfigs},
	}

	auftraege, err := q.GetDruckAuftraege(context.Background(), 0)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(auftraege) != 1 {
		t.Fatalf("expected 1 auftrag, got %d", len(auftraege))
	}
	if auftraege[0].EventID != 5 {
		t.Errorf("expected EventID 5, got %d", auftraege[0].EventID)
	}
	if auftraege[0].DruckerIP != "192.168.1.50" {
		t.Errorf("expected DruckerIP 192.168.1.50, got %s", auftraege[0].DruckerIP)
	}
	if auftraege[0].Payload == "" {
		t.Error("expected non-empty payload")
	}
}

func TestGetDruckAuftraege_CursorFiltering(t *testing.T) {
	positionen := []kasse.Position{
		{ProduktName: "Bier", VarianteName: "0,5l", Kategorie: "getraenk", Menge: 1},
	}
	events := []event.Event{
		makeBestellungEvent(3, "tisch:1", positionen, ""),
		makeBestellungEvent(7, "tisch:2", positionen, ""),
	}
	konfigs := map[string]DruckerKonfig{
		"getraenk": {IP: "192.168.1.50", Bonmodus: "pro_position"},
	}
	q := Query{
		EventRepo:   &mockEventRepo{events: events},
		DruckerRepo: &mockDruckerRepo{konfigs: konfigs},
	}

	// cursor=3 → only event with id>3 (id=7)
	auftraege, err := q.GetDruckAuftraege(context.Background(), 3)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(auftraege) != 1 {
		t.Fatalf("expected 1 auftrag, got %d", len(auftraege))
	}
	if auftraege[0].EventID != 7 {
		t.Errorf("expected EventID 7, got %d", auftraege[0].EventID)
	}
}

// --- createDruckAuftraegeFromEvent Tests ---

func TestCreateDruckAuftraege_ProPosition(t *testing.T) {
	positionen := []kasse.Position{
		{ProduktName: "Pommes", VarianteName: "groß", Kategorie: "essen", Menge: 2},
		{ProduktName: "Bratwurst", VarianteName: "mit Brot", Kategorie: "essen", Menge: 1},
	}
	evt := makeBestellungEvent(1, "tisch:7", positionen, "ohne Ketchup")
	konfig := map[string]DruckerKonfig{
		"essen": {IP: "192.168.1.51", Bonmodus: "pro_position"},
	}

	auftraege := createDruckAuftraegeFromEvent(evt, konfig)

	if len(auftraege) != 2 {
		t.Fatalf("expected 2 auftraege (one per position), got %d", len(auftraege))
	}
	for _, a := range auftraege {
		if a.DruckerIP != "192.168.1.51" {
			t.Errorf("expected DruckerIP 192.168.1.51, got %s", a.DruckerIP)
		}
		if a.EventID != 1 {
			t.Errorf("expected EventID 1, got %d", a.EventID)
		}
		if a.Payload == "" {
			t.Error("expected non-empty payload")
		}
	}
}

func TestCreateDruckAuftraege_ProBestellung(t *testing.T) {
	positionen := []kasse.Position{
		{ProduktName: "Pommes", VarianteName: "groß", Kategorie: "essen", Menge: 1},
		{ProduktName: "Schnitzel", VarianteName: "mit Salat", Kategorie: "essen", Menge: 1},
	}
	evt := makeBestellungEvent(2, "tisch:5", positionen, "")
	konfig := map[string]DruckerKonfig{
		"essen": {IP: "192.168.1.51", Bonmodus: "pro_bestellung"},
	}

	auftraege := createDruckAuftraegeFromEvent(evt, konfig)

	// pro_bestellung → 1 Sammelbon für alle Positionen der Kategorie
	if len(auftraege) != 1 {
		t.Fatalf("expected 1 sammelbon, got %d", len(auftraege))
	}
	if auftraege[0].Payload == "" {
		t.Error("expected non-empty payload")
	}
}

func TestCreateDruckAuftraege_NoDruckerFuerKategorie(t *testing.T) {
	positionen := []kasse.Position{
		{ProduktName: "Kaffee", VarianteName: "klein", Kategorie: "sonstiges", Menge: 1},
	}
	evt := makeBestellungEvent(3, "tisch:2", positionen, "")
	// Nur "essen" konfiguriert, nicht "sonstiges"
	konfig := map[string]DruckerKonfig{
		"essen": {IP: "192.168.1.51", Bonmodus: "pro_position"},
	}

	auftraege := createDruckAuftraegeFromEvent(evt, konfig)

	if len(auftraege) != 0 {
		t.Errorf("expected 0 auftraege (no printer for kategorie), got %d", len(auftraege))
	}
}

func TestCreateDruckAuftraege_MehrereKategorien(t *testing.T) {
	positionen := []kasse.Position{
		{ProduktName: "Pommes", VarianteName: "groß", Kategorie: "essen", Menge: 1},
		{ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Menge: 1},
	}
	evt := makeBestellungEvent(4, "tisch:3", positionen, "")
	konfig := map[string]DruckerKonfig{
		"essen":    {IP: "192.168.1.51", Bonmodus: "pro_position"},
		"getraenk": {IP: "192.168.1.50", Bonmodus: "pro_position"},
	}

	auftraege := createDruckAuftraegeFromEvent(evt, konfig)

	// 1 Bon pro Kategorie × 1 Position je = 2 Bons
	if len(auftraege) != 2 {
		t.Fatalf("expected 2 auftraege, got %d", len(auftraege))
	}
}

func TestCreateDruckAuftraege_InvalidJSON(t *testing.T) {
	evt := event.Event{
		ID:   1,
		Data: []byte("not json"),
	}
	auftraege := createDruckAuftraegeFromEvent(evt, map[string]DruckerKonfig{})

	if auftraege != nil {
		t.Errorf("expected nil on invalid JSON, got %v", auftraege)
	}
}

// --- parseTischName Tests ---

func TestParseTischName_Valid(t *testing.T) {
	result := parseTischName("tisch:7")
	if result != "Tisch 7" {
		t.Errorf("expected 'Tisch 7', got %q", result)
	}
}

func TestParseTischName_WithLargerID(t *testing.T) {
	result := parseTischName("tisch:42")
	if result != "Tisch 42" {
		t.Errorf("expected 'Tisch 42', got %q", result)
	}
}

func TestParseTischName_Invalid(t *testing.T) {
	// Falls kein Doppelpunkt vorhanden → subject zurückgeben
	result := parseTischName("invalid")
	if result != "invalid" {
		t.Errorf("expected 'invalid', got %q", result)
	}
}
