//go:build unit

package application

import (
	"testing"
	"time"

	"github.com/nicograef/jotti/backend/domain/kasse"
)

func TestBuildKassenbelegProcessData_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		positionen  []kasse.Position
		zahlbetrag  int
		expected    string
		expectError bool
	}{
		{
			name:       "regelsteuersatz",
			positionen: []kasse.Position{{Einzelpreis: 1000, Menge: 1, Steuersatz: "regel"}},
			zahlbetrag: 1000,
			expected:   "Beleg^10.00_0.00_0.00_0.00_0.00^10.00:Bar",
		},
		{
			name: "ermaessigt und befreit",
			positionen: []kasse.Position{
				{Einzelpreis: 300, Menge: 1, Steuersatz: "ermaessigt"},
				{Einzelpreis: 200, Menge: 1, Steuersatz: "befreit"},
			},
			zahlbetrag: 500,
			expected:   "Beleg^0.00_3.00_0.00_0.00_2.00^5.00:Bar",
		},
		{
			name:       "kombi wird 70 30 aufgeteilt",
			positionen: []kasse.Position{{Einzelpreis: 1000, Menge: 1, Steuersatz: "kombi"}},
			zahlbetrag: 1000,
			expected:   "Beleg^3.00_7.00_0.00_0.00_0.00^10.00:Bar",
		},
		{
			name:       "ohne tausendertrennzeichen",
			positionen: []kasse.Position{{Einzelpreis: 123456, Menge: 1, Steuersatz: "regel"}},
			zahlbetrag: 123456,
			expected:   "Beleg^1234.56_0.00_0.00_0.00_0.00^1234.56:Bar",
		},
		{
			name:       "ohne positionen mit negativem betrag",
			positionen: nil,
			zahlbetrag: -1500,
			expected:   "Beleg^0.00_0.00_0.00_0.00_0.00^-15.00:Bar",
		},
		{
			name:       "zahlung 0.00 entfaellt",
			positionen: []kasse.Position{{Einzelpreis: 1000, Menge: 1, Steuersatz: "regel"}},
			zahlbetrag: 0,
			expected:   "Beleg^10.00_0.00_0.00_0.00_0.00^",
		},
		{
			name:        "ungueltiger steuersatz",
			positionen:  []kasse.Position{{Einzelpreis: 100, Menge: 1, Steuersatz: "foo"}},
			zahlbetrag:  100,
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := BuildKassenbelegProcessData(tc.positionen, tc.zahlbetrag)
			if tc.expectError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if got != tc.expected {
				t.Fatalf("unexpected processData\nwant: %q\ngot:  %q", tc.expected, got)
			}
		})
	}
}

func TestBuildKassenbelegProcessDataWithFaktor_NegativBeiStorno(t *testing.T) {
	got, err := BuildKassenbelegProcessDataWithFaktor(
		[]kasse.Position{{Einzelpreis: 350, Menge: 2, Steuersatz: "regel"}},
		-700,
		-1,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	want := "Beleg^-7.00_0.00_0.00_0.00_0.00^-7.00:Bar"
	if got != want {
		t.Fatalf("unexpected processData\nwant: %q\ngot:  %q", want, got)
	}
}

func TestBuildBestellungProcessData_CSVFormat(t *testing.T) {
	got, err := BuildBestellungProcessData([]kasse.Position{
		{ProduktName: "Maß Bier", VarianteName: "", Menge: 4, Einzelpreis: 950},
		{ProduktName: "Weißwurst", VarianteName: "normal", Menge: 2, Einzelpreis: 250},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	want := "4;\"Maß Bier\";9.50\r2;\"Weißwurst normal\";2.50"
	if got != want {
		t.Fatalf("unexpected processData\nwant: %q\ngot:  %q", want, got)
	}
}

func TestBuildBestellungProcessData_VerdoppeltAnfuehrungszeichen(t *testing.T) {
	got, err := BuildBestellungProcessData([]kasse.Position{
		{ProduktName: `Eisbecher "Himbeere"`, Menge: 2, Einzelpreis: 399},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Beispiel aus DSFinV-K Anhang I
	want := `2;"Eisbecher ""Himbeere""";3.99`
	if got != want {
		t.Fatalf("unexpected processData\nwant: %q\ngot:  %q", want, got)
	}
}

func TestBuildGeldtransitProcessData_Einlage(t *testing.T) {
	got, err := BuildGeldtransitProcessData("einlage", 1234)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	// Feld 5 (0 %, nicht steuerbar) trägt den Betrag und gleicht die Bar-Zahlung aus.
	want := "Beleg^0.00_0.00_0.00_0.00_12.34^12.34:Bar"
	if got != want {
		t.Fatalf("unexpected processData\nwant: %q\ngot:  %q", want, got)
	}
}

func TestBuildGeldtransitProcessData_Entnahme(t *testing.T) {
	got, err := BuildGeldtransitProcessData("entnahme", 1234)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	want := "Beleg^0.00_0.00_0.00_0.00_-12.34^-12.34:Bar"
	if got != want {
		t.Fatalf("unexpected processData\nwant: %q\ngot:  %q", want, got)
	}
}

func TestBuildGeldtransitProcessData_UngueltigeRichtung(t *testing.T) {
	if _, err := BuildGeldtransitProcessData("foo", 1234); err == nil {
		t.Fatal("expected error for invalid richtung, got nil")
	}
}

func TestBuildEigenbelegProcessData(t *testing.T) {
	tests := []struct {
		name            string
		zahlbetragCents int
		expected        string
	}{
		// Feld 5 (0 %, nicht steuerbar) trägt den Betrag, die Bar-Zahlung gleicht ihn aus.
		{name: "positiver betrag", zahlbetragCents: 250, expected: "Beleg^0.00_0.00_0.00_0.00_2.50^2.50:Bar"},
		{name: "negativer betrag", zahlbetragCents: -250, expected: "Beleg^0.00_0.00_0.00_0.00_-2.50^-2.50:Bar"},
		{name: "zahlung 0.00 entfaellt", zahlbetragCents: 0, expected: "Beleg^0.00_0.00_0.00_0.00_0.00^"},
		// Kassendifferenz: der Aufrufer übergibt die Bargeldbewegung (Ist − Soll).
		// Ein Fehlbetrag mindert den Bestand (negativ), ein Überschuss mehrt ihn.
		{name: "kassendifferenz fehlbetrag", zahlbetragCents: -100, expected: "Beleg^0.00_0.00_0.00_0.00_-1.00^-1.00:Bar"},
		{name: "kassendifferenz ueberschuss", zahlbetragCents: 100, expected: "Beleg^0.00_0.00_0.00_0.00_1.00^1.00:Bar"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := BuildEigenbelegProcessData(tc.zahlbetragCents)
			if got != tc.expected {
				t.Fatalf("unexpected processData\nwant: %q\ngot:  %q", tc.expected, got)
			}
		})
	}
}

func TestBuildTagesabschlussProcessData(t *testing.T) {
	von := time.Date(2026, 6, 10, 8, 0, 0, 0, time.UTC)
	bis := time.Date(2026, 6, 10, 22, 0, 0, 0, time.UTC)
	got := BuildTagesabschlussProcessData(7, von, bis)
	want := "Tagesabschluss^ZNr:7^Von:2026-06-10T08:00:00Z^Bis:2026-06-10T22:00:00Z"
	if got != want {
		t.Fatalf("unexpected processData\nwant: %q\ngot:  %q", want, got)
	}
}
