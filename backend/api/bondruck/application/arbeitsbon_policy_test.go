//go:build unit

package application

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/nicograef/jotti/backend/api/bondruck/application/escpos"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
)

func makeBestellungEvent(id int, subject string, positionen []kasse.Position, kommentar string) event.Event {
	data, _ := json.Marshal(positionenMitKommentarData{
		Positionen: positionen,
		Kommentar:  kommentar,
	})
	return event.Event{
		ID:       id,
		Type:     string(kasse.EventTypeBestellungAufgenommenV1),
		UserName: "Maria",
		Subject:  subject,
		Time:     time.Date(2026, 3, 17, 19, 34, 0, 0, time.UTC),
		Data:     data,
	}
}

func makeDirektverkaufEvent(id int, positionen []kasse.Position, kommentar string) event.Event {
	data, _ := json.Marshal(positionenMitKommentarData{
		Positionen: positionen,
		Kommentar:  kommentar,
	})

	return event.Event{
		ID:       id,
		Type:     string(kasse.EventTypeDirektverkaufGetaetigtV1),
		UserName: "Maria",
		Subject:  fmt.Sprintf("kassensitzung-1/direktverkauf-%d", id),
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

// Direktverkauf-Ableitungsregel: ohne konfigurierte Abholbon-Station gehen die Bons
// an die Produktstationen je Kategorie.
func TestCreateArbeitsbonAuftraege_DirektverkaufAnProduktstationen(t *testing.T) {
	positionen := []kasse.Position{
		{ProduktName: "Pommes", VarianteName: "gross", Kategorie: "essen", Menge: 1},
		{ProduktName: "Bier", VarianteName: "0,5l", Kategorie: "getraenk", Menge: 2},
	}
	evt := makeDirektverkaufEvent(20, positionen, "schnell")
	stationen := map[string]Druckstation{
		"essen":    {IP: "192.168.1.51", Bonmodus: "pro_bestellung"},
		"getraenk": {IP: "192.168.1.52", Bonmodus: "pro_bestellung"},
	}

	auftraege := CreateArbeitsbonAuftraegeFromEvent(evt, stationen)

	if len(auftraege) != 2 {
		t.Fatalf("expected 2 auftraege (one per configured category), got %d", len(auftraege))
	}
	for _, auftrag := range auftraege {
		if auftrag.BonArt != "arbeitsbon" {
			t.Errorf("expected BonArt arbeitsbon, got %s", auftrag.BonArt)
		}
		if auftrag.Referenz != "direktverkauf-getaetigt:20" {
			t.Errorf("expected referenz direktverkauf-getaetigt:20, got %s", auftrag.Referenz)
		}
	}
}

// Ist die Abholbon-Station konfiguriert, hat sie Vorrang vor den Produktstationen.
func TestCreateArbeitsbonAuftraege_DirektverkaufAbholbon_ProBestellung(t *testing.T) {
	positionen := []kasse.Position{
		{ProduktName: "Pommes", VarianteName: "gross", Kategorie: "essen", Menge: 2},
		{ProduktName: "Bier", VarianteName: "0,5l", Kategorie: "getraenk", Menge: 1},
	}
	evt := makeDirektverkaufEvent(21, positionen, "ohne Senf")
	stationen := map[string]Druckstation{
		"essen":    {IP: "192.168.1.51", Bonmodus: "pro_position"},
		"abholbon": {IP: "192.168.1.77", Bonmodus: "pro_bestellung"},
	}

	auftraege := CreateArbeitsbonAuftraegeFromEvent(evt, stationen)

	if len(auftraege) != 1 {
		t.Fatalf("expected exactly 1 abholbon auftrag, got %d", len(auftraege))
	}
	if auftraege[0].ZielIP != "192.168.1.77" {
		t.Errorf("expected ZielIP 192.168.1.77, got %s", auftraege[0].ZielIP)
	}
	if auftraege[0].BonArt != "arbeitsbon" {
		t.Errorf("expected BonArt arbeitsbon, got %s", auftraege[0].BonArt)
	}
	if auftraege[0].Referenz != "direktverkauf-getaetigt:21" {
		t.Errorf("expected referenz direktverkauf-getaetigt:21, got %s", auftraege[0].Referenz)
	}

	expected := base64.StdEncoding.EncodeToString(escpos.FormatDirektverkaufAbholbon(positionen, evt.UserName, evt.Time, "ohne Senf"))
	if auftraege[0].Payload != expected {
		t.Fatal("abholbon payload is not byte-identical to formatter output")
	}
}

// Bonmodus pro_position der Abholbon-Station erzeugt einen Abholbon je Position.
func TestCreateArbeitsbonAuftraege_DirektverkaufAbholbon_ProPosition(t *testing.T) {
	positionen := []kasse.Position{
		{ProduktName: "Pommes", VarianteName: "gross", Kategorie: "essen", Menge: 2},
		{ProduktName: "Bier", VarianteName: "0,5l", Kategorie: "getraenk", Menge: 1},
	}
	evt := makeDirektverkaufEvent(22, positionen, "")
	stationen := map[string]Druckstation{
		"abholbon": {IP: "192.168.1.77", Bonmodus: "pro_position"},
	}

	auftraege := CreateArbeitsbonAuftraegeFromEvent(evt, stationen)

	if len(auftraege) != 2 {
		t.Fatalf("expected 2 abholbon auftraege (one per position), got %d", len(auftraege))
	}
	for i, auftrag := range auftraege {
		if auftrag.ZielIP != "192.168.1.77" {
			t.Errorf("expected ZielIP 192.168.1.77, got %s", auftrag.ZielIP)
		}
		expected := base64.StdEncoding.EncodeToString(escpos.FormatDirektverkaufAbholbon([]kasse.Position{positionen[i]}, evt.UserName, evt.Time, ""))
		if auftrag.Payload != expected {
			t.Errorf("abholbon payload %d is not byte-identical to formatter output", i)
		}
	}
}

// Ohne konfigurierte Druckstationen entstehen fuer einen Direktverkauf keine Auftraege.
func TestCreateArbeitsbonAuftraege_DirektverkaufOhneStationen(t *testing.T) {
	evt := makeDirektverkaufEvent(23, []kasse.Position{{ProduktName: "Pommes", VarianteName: "gross", Kategorie: "essen", Menge: 1}}, "")

	auftraege := CreateArbeitsbonAuftraegeFromEvent(evt, map[string]Druckstation{})

	if len(auftraege) != 0 {
		t.Fatalf("expected 0 auftraege without configured stations, got %d", len(auftraege))
	}
}
