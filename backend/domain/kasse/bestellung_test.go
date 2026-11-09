package kasse

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
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

// Ein persistiertes Event mit Menge 1000 bleibt lesbar und stornierbar: Die Obergrenze 999
// gilt nur auf dem Eingabeweg (PositionEingabeSchema), nicht in positionSchema.
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
		t.Errorf("expected menge 1000, got %d", bestellung.Positionen[0].Menge)
	}

	refs := []PositionRef{{PositionID: positionID, Menge: 1000}}
	positionen, gesamtCents, ok := ResolvePositionen(bestellung.Positionen, refs)
	if !ok {
		t.Fatal("expected the full menge to be stornierbar")
	}
	if _, err := NewStornierungErteiltEvent(testSubject, uuid.NewString(), 2, "Leitung", testZahlungID, positionen, gesamtCents, "Rueckgabe"); err != nil {
		t.Errorf("Stornierung einer Position mit Menge 1000 muss gelingen, got %v", err)
	}
}

// Eine Teilzahlung löst nur die angefragte Menge auf: Betrag und Menge folgen der Referenz,
// nicht der offenen Position.
func TestResolvePositionen(t *testing.T) {
	const bierID = "33333333-3333-4333-8333-333333333333"
	offen := []Position{{PositionID: bierID, ProduktName: "Bier", EinzelpreisCents: 350, Menge: 4}}

	tests := []struct {
		name       string
		refs       []PositionRef
		wantOK     bool
		wantGesamt int
		wantMengen []int
	}{
		{name: "Teilmenge", refs: []PositionRef{{PositionID: bierID, Menge: 1}}, wantOK: true, wantGesamt: 350, wantMengen: []int{1}},
		{name: "volle Menge", refs: []PositionRef{{PositionID: bierID, Menge: 4}}, wantOK: true, wantGesamt: 1400, wantMengen: []int{4}},
		{name: "mehr als offen", refs: []PositionRef{{PositionID: bierID, Menge: 5}}},
		{name: "unbekannte Position", refs: []PositionRef{{PositionID: "44444444-4444-4444-8444-444444444444", Menge: 1}}},
		{name: "Position doppelt", refs: []PositionRef{{PositionID: bierID, Menge: 1}, {PositionID: bierID, Menge: 1}}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			positionen, gesamt, ok := ResolvePositionen(offen, tc.refs)
			if ok != tc.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tc.wantOK)
			}
			if gesamt != tc.wantGesamt {
				t.Errorf("gesamtCents = %d, want %d", gesamt, tc.wantGesamt)
			}
			if len(positionen) != len(tc.wantMengen) {
				t.Fatalf("got %d Positionen, want %d", len(positionen), len(tc.wantMengen))
			}
			for i, want := range tc.wantMengen {
				if positionen[i].Menge != want {
					t.Errorf("Positionen[%d].Menge = %d, want %d", i, positionen[i].Menge, want)
				}
			}
		})
	}
	if offen[0].Menge != 4 {
		t.Errorf("offene Position verändert: Menge = %d, want 4", offen[0].Menge)
	}
}
