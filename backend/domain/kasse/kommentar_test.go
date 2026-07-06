//go:build unit

package kasse

import (
	"strings"
	"testing"
)

var kommentarTests = []struct {
	name      string
	kommentar string
	wantErr   bool
}{
	{name: "leer", kommentar: "", wantErr: true},
	{name: "zu kurz (1 Zeichen)", kommentar: "A", wantErr: true},
	{name: "zu kurz (2 Zeichen)", kommentar: "AB", wantErr: true},
	{name: "gültig (3 Zeichen)", kommentar: "ABC", wantErr: false},
	{name: "gültig (Standardtext)", kommentar: "Reklamation", wantErr: false},
	{name: "maximale Länge (100 Zeichen)", kommentar: strings.Repeat("x", 100), wantErr: false},
	{name: "zu lang (101 Zeichen)", kommentar: strings.Repeat("x", 101), wantErr: true},
}

var optionalKommentarTests = []struct {
	name      string
	kommentar string
	wantErr   bool
}{
	{name: "leer (optional)", kommentar: "", wantErr: false},
	{name: "gültig", kommentar: "Anmerkung", wantErr: false},
	{name: "maximale Länge (100 Zeichen)", kommentar: strings.Repeat("x", 100), wantErr: false},
	{name: "zu lang (101 Zeichen)", kommentar: strings.Repeat("x", 101), wantErr: true},
}

var kommentarTestSubject = "kassensitzung-1/tisch-1"
var kommentarTestPositionen = []Position{
	{
		VarianteID:       1,
		ProduktName:      "Cola",
		VarianteName:     "0,5l",
		Kategorie:        "getraenk",
		Steuersatz:       "regel",
		EinzelpreisCents: 350,
		Menge:            1,
	},
}

// kommentarTestPositionenMitID is for events whose schema requires a PositionID
// (Zahlung, Stornierung, Ausgabe) — unlike Bestellung, which generates the IDs itself.
var kommentarTestPositionenMitID = []Position{
	{
		PositionID:       "a87f1b2c-3d4e-5f6a-7b8c-9d0e1f2a3b4c",
		VarianteID:       1,
		ProduktName:      "Cola",
		VarianteName:     "0,5l",
		Kategorie:        "getraenk",
		Steuersatz:       "regel",
		EinzelpreisCents: 350,
		Menge:            1,
	},
}

// --- Stornierung: Kommentar ist Pflichtfeld (min 3 Zeichen) ---

func TestNewStornierungErteiltEvent_Kommentar(t *testing.T) {
	for _, tt := range kommentarTests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewStornierungErteiltEvent(kommentarTestSubject, 1, "Servicekraft", "11111111-1111-1111-1111-111111111111", kommentarTestPositionenMitID, 350, tt.kommentar)
			if tt.wantErr && err == nil {
				t.Errorf("erwartete Fehler für Kommentar %q, aber kein Fehler", tt.kommentar)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("kein Fehler erwartet für Kommentar %q, aber: %v", tt.kommentar, err)
			}
		})
	}
}

// --- Bestellung: Kommentar ist optional ---

func TestNewBestellungAufgenommenEvent_Kommentar(t *testing.T) {
	for _, tt := range optionalKommentarTests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewBestellungAufgenommenEvent(kommentarTestSubject, 1, "Servicekraft", kommentarTestPositionen, tt.kommentar)
			if tt.wantErr && err == nil {
				t.Errorf("erwartete Fehler für Kommentar %q, aber kein Fehler", tt.kommentar)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("kein Fehler erwartet für Kommentar %q, aber: %v", tt.kommentar, err)
			}
		})
	}
}

// --- Zahlung: Kommentar ist optional ---

func TestNewZahlungKassiertEvent_Kommentar(t *testing.T) {
	for _, tt := range optionalKommentarTests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewZahlungKassiertEvent(kommentarTestSubject, 1, "Servicekraft", kommentarTestPositionenMitID, 350, tt.kommentar)
			if tt.wantErr && err == nil {
				t.Errorf("erwartete Fehler für Kommentar %q, aber kein Fehler", tt.kommentar)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("kein Fehler erwartet für Kommentar %q, aber: %v", tt.kommentar, err)
			}
		})
	}
}

// --- Ausgabe: Kommentar ist optional ---

func TestNewAusgabeBestaetigtEvent_Kommentar(t *testing.T) {
	for _, tt := range optionalKommentarTests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewAusgabeBestaetigtEvent(kommentarTestSubject, 1, "Servicekraft", kommentarTestPositionenMitID, tt.kommentar)
			if tt.wantErr && err == nil {
				t.Errorf("erwartete Fehler für Kommentar %q, aber kein Fehler", tt.kommentar)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("kein Fehler erwartet für Kommentar %q, aber: %v", tt.kommentar, err)
			}
		})
	}
}
