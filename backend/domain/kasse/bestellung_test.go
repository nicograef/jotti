//go:build unit

package kasse

import (
	"encoding/json"
	"testing"
	"time"

	e "github.com/nicograef/jotti/backend/domain/event"
)

func TestPositionBezeichnung(t *testing.T) {
	tests := []struct {
		name         string
		produktName  string
		varianteName string
		want         string
	}{
		{name: "Normalfall", produktName: "Pommes", varianteName: "mit Ketchup", want: "Pommes mit Ketchup"},
		{name: "gleichlautend ohne Dedup", produktName: "Cola", varianteName: "Cola", want: "Cola Cola"},
		{name: "leerer Variantenname ohne Trailing-Space", produktName: "Maß Bier", varianteName: "", want: "Maß Bier"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pos := Position{ProduktName: tc.produktName, VarianteName: tc.varianteName}
			if got := pos.Bezeichnung(); got != tc.want {
				t.Errorf("Bezeichnung() = %q, want %q", got, tc.want)
			}
		})
	}
}

// Ein persistiertes Event mit Menge 1000 bleibt lesbar und stornierbar: Die
// Obergrenze 999 gilt nur auf dem Eingabeweg (PositionEingabeSchema), während
// positionSchema jedes gelesene Event validiert.
func TestPersistiertesEventMitMenge1000_BleibtLesbarUndStornierbar(t *testing.T) {
	const positionID = "22222222-2222-4222-8222-222222222222"
	persisted := e.Event{
		ID:       1,
		UserID:   1,
		UserName: "Maria",
		Type:     string(EventTypeBestellungAufgenommenV1),
		Time:     time.Date(2026, 7, 1, 18, 0, 0, 0, time.UTC),
		Subject:  testSubject,
		Version:  1,
		Data: json.RawMessage(`{
			"bestellungId": "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
			"positionen": [{
				"positionId": "` + positionID + `",
				"varianteId": 1,
				"produktName": "Bier",
				"varianteName": "0,5l",
				"kategorie": "getraenk",
				"steuersatz": "regel",
				"einzelpreisCents": 350,
				"menge": 1000
			}],
			"gesamtPreisCents": 350000,
			"kommentar": ""
		}`),
	}

	bestellung, err := buildBestellungFromEvent(persisted)
	if err != nil {
		t.Fatalf("persistiertes Event mit Menge 1000 muss lesbar bleiben, got %v", err)
	}
	if bestellung.Positionen[0].Menge != 1000 {
		t.Fatalf("expected menge 1000, got %d", bestellung.Positionen[0].Menge)
	}

	refs := []PositionRef{{PositionID: positionID, Menge: 1000}}
	if !ValidatePositionRefs(bestellung.Positionen, refs) {
		t.Fatal("expected the full menge to be stornierbar")
	}

	positionen, gesamtCents := ResolvePositionen(bestellung.Positionen, refs)
	if _, err := NewStornierungErteiltEvent(testSubject, 2, "Leitung", testZahlungID, positionen, gesamtCents, "Rueckgabe"); err != nil {
		t.Fatalf("Stornierung einer Position mit Menge 1000 muss gelingen, got %v", err)
	}
}
