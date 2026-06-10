//go:build unit

package application

import (
	"testing"
	"time"
)

func TestBuildGeldtransitProcessData(t *testing.T) {
	got := buildGeldtransitProcessData("einlage", 1234, "Wechselgeld")
	want := "Geldtransit^einlage^12.34^Wechselgeld"
	if got != want {
		t.Fatalf("unexpected processData\nwant: %q\ngot:  %q", want, got)
	}
}

func TestBuildDifferenzProcessData_Negativ(t *testing.T) {
	got := buildDifferenzProcessData(-250)
	want := "DifferenzSollIst^-2.50"
	if got != want {
		t.Fatalf("unexpected processData\nwant: %q\ngot:  %q", want, got)
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
