//go:build unit

package application

import (
	"testing"
	"time"
)

func TestBuildGeldtransitProcessData_Einlage(t *testing.T) {
	got, err := buildGeldtransitProcessData("einlage", 1234)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	want := "Beleg^0.00_0.00_0.00_0.00_0.00^12.34:Bar"
	if got != want {
		t.Fatalf("unexpected processData\nwant: %q\ngot:  %q", want, got)
	}
}

func TestBuildGeldtransitProcessData_Entnahme(t *testing.T) {
	got, err := buildGeldtransitProcessData("entnahme", 1234)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	want := "Beleg^0.00_0.00_0.00_0.00_0.00^-12.34:Bar"
	if got != want {
		t.Fatalf("unexpected processData\nwant: %q\ngot:  %q", want, got)
	}
}

func TestBuildGeldtransitProcessData_UngueltigeRichtung(t *testing.T) {
	if _, err := buildGeldtransitProcessData("foo", 1234); err == nil {
		t.Fatal("expected error for invalid richtung, got nil")
	}
}

func TestBuildEigenbelegProcessData(t *testing.T) {
	tests := []struct {
		name            string
		zahlbetragCents int
		expected        string
	}{
		{name: "positiver betrag", zahlbetragCents: 250, expected: "Beleg^0.00_0.00_0.00_0.00_0.00^2.50:Bar"},
		{name: "negativer betrag", zahlbetragCents: -250, expected: "Beleg^0.00_0.00_0.00_0.00_0.00^-2.50:Bar"},
		{name: "zahlung 0.00 entfaellt", zahlbetragCents: 0, expected: "Beleg^0.00_0.00_0.00_0.00_0.00^"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := buildEigenbelegProcessData(tc.zahlbetragCents)
			if got != tc.expected {
				t.Fatalf("unexpected processData\nwant: %q\ngot:  %q", tc.expected, got)
			}
		})
	}
}

func TestBuildTagesabschlussProcessData(t *testing.T) {
	von := time.Date(2026, 6, 10, 8, 0, 0, 0, time.UTC)
	bis := time.Date(2026, 6, 10, 22, 0, 0, 0, time.UTC)
	got := buildTagesabschlussProcessData(7, von, bis)
	want := "Tagesabschluss^ZNr:7^Von:2026-06-10T08:00:00Z^Bis:2026-06-10T22:00:00Z"
	if got != want {
		t.Fatalf("unexpected processData\nwant: %q\ngot:  %q", want, got)
	}
}
