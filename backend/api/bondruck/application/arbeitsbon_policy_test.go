//go:build unit

package application

import (
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	"github.com/nicograef/jotti/backend/api/bondruck/application/escpos"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
)

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

func TestCreateArbeitsbonAuftraege_ProPosition(t *testing.T) {
	positionen := []kasse.Position{
		{ProduktName: "Pommes", VarianteName: "gross", Kategorie: "essen", Menge: 2},
		{ProduktName: "Bratwurst", VarianteName: "mit Brot", Kategorie: "essen", Menge: 1},
	}
	evt := makeBestellungEvent(1, "kassensitzung-1/tisch-7", positionen, "ohne Ketchup")
	konfig := map[string]Druckstation{
		"essen": {IP: "192.168.1.51", Bonmodus: "pro_position"},
	}

	auftraege := CreateArbeitsbonAuftraegeFromEvent(evt, konfig)

	if len(auftraege) != 2 {
		t.Fatalf("expected 2 auftraege (one per position), got %d", len(auftraege))
	}
	for _, auftrag := range auftraege {
		if auftrag.ZielIP != "192.168.1.51" {
			t.Errorf("expected ZielIP 192.168.1.51, got %s", auftrag.ZielIP)
		}
		if auftrag.BonArt != "arbeitsbon" {
			t.Errorf("expected BonArt arbeitsbon, got %s", auftrag.BonArt)
		}
		if auftrag.Payload == "" {
			t.Error("expected non-empty payload")
		}
	}
}

func TestCreateArbeitsbonAuftraege_ProBestellung(t *testing.T) {
	positionen := []kasse.Position{
		{ProduktName: "Pommes", VarianteName: "gross", Kategorie: "essen", Menge: 1},
		{ProduktName: "Schnitzel", VarianteName: "mit Salat", Kategorie: "essen", Menge: 1},
	}
	evt := makeBestellungEvent(2, "kassensitzung-1/tisch-5", positionen, "")
	konfig := map[string]Druckstation{
		"essen": {IP: "192.168.1.51", Bonmodus: "pro_bestellung"},
	}

	auftraege := CreateArbeitsbonAuftraegeFromEvent(evt, konfig)

	if len(auftraege) != 1 {
		t.Fatalf("expected 1 sammelbon, got %d", len(auftraege))
	}
	if auftraege[0].Payload == "" {
		t.Error("expected non-empty payload")
	}
}

func TestCreateArbeitsbonAuftraege_NoDruckerFuerKategorie(t *testing.T) {
	positionen := []kasse.Position{
		{ProduktName: "Kaffee", VarianteName: "klein", Kategorie: "sonstiges", Menge: 1},
	}
	evt := makeBestellungEvent(3, "kassensitzung-1/tisch-2", positionen, "")
	konfig := map[string]Druckstation{
		"essen": {IP: "192.168.1.51", Bonmodus: "pro_position"},
	}

	auftraege := CreateArbeitsbonAuftraegeFromEvent(evt, konfig)

	if len(auftraege) != 0 {
		t.Errorf("expected 0 auftraege (no printer for kategorie), got %d", len(auftraege))
	}
}

func TestCreateArbeitsbonAuftraege_ByteIdentischZumFormatter_ProPosition(t *testing.T) {
	positionen := []kasse.Position{
		{ProduktName: "Pommes", VarianteName: "gross", Kategorie: "essen", Menge: 2},
		{ProduktName: "Bratwurst", VarianteName: "mit Brot", Kategorie: "essen", Menge: 1},
	}
	evt := makeBestellungEvent(10, "kassensitzung-1/tisch-7", positionen, "ohne Ketchup")
	konfig := map[string]Druckstation{
		"essen": {IP: "192.168.1.51", Bonmodus: "pro_position"},
	}

	auftraege := CreateArbeitsbonAuftraegeFromEvent(evt, konfig)
	if len(auftraege) != 2 {
		t.Fatalf("expected 2 auftraege, got %d", len(auftraege))
	}

	expectedFirst := base64.StdEncoding.EncodeToString(escpos.FormatPositionBon(
		positionen[0],
		"Tisch 7",
		evt.UserName,
		evt.Time,
		"ohne Ketchup",
		true,
	))
	expectedSecond := base64.StdEncoding.EncodeToString(escpos.FormatPositionBon(
		positionen[1],
		"Tisch 7",
		evt.UserName,
		evt.Time,
		"ohne Ketchup",
		false,
	))

	if auftraege[0].Payload != expectedFirst {
		t.Fatal("first payload is not byte-identical to formatter output")
	}
	if auftraege[1].Payload != expectedSecond {
		t.Fatal("second payload is not byte-identical to formatter output")
	}
}
