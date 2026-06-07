package escpos_test

import (
	"strings"
	"testing"
	"time"

	"github.com/nicograef/jotti/backend/api/bondruck/application/escpos"
	"github.com/nicograef/jotti/backend/domain/kasse"
)

var (
	testPos = kasse.Position{
		PositionID:   "pos-1",
		VarianteID:   1,
		ProduktName:  "Pommes",
		VarianteName: "gross",
		Kategorie:    "essen",
		Einzelpreis:  300,
		Menge:        3,
	}
	testTime = time.Date(2024, 6, 15, 19, 34, 0, 0, time.UTC)
)

func TestFormatPositionBon_ContainsTischName(t *testing.T) {
	payload := escpos.FormatPositionBon(testPos, "Tisch 7", "Maria", testTime, "", false)
	got := string(payload)
	if !strings.Contains(got, "Tisch 7") {
		t.Errorf("Bon enthaelt nicht den Tischnamen; got:\n%q", got)
	}
}

func TestFormatPositionBon_ContainsArtikel(t *testing.T) {
	payload := escpos.FormatPositionBon(testPos, "Tisch 7", "Maria", testTime, "", false)
	got := string(payload)
	if !strings.Contains(got, "3x Pommes (gross)") {
		t.Errorf("Bon enthaelt nicht den Artikel; got:\n%q", got)
	}
}

func TestFormatPositionBon_ContainsBedienung(t *testing.T) {
	payload := escpos.FormatPositionBon(testPos, "Tisch 7", "Maria", testTime, "", false)
	got := string(payload)
	if !strings.Contains(got, "Bedienung: Maria") {
		t.Errorf("Bon enthaelt nicht Bedienung; got:\n%q", got)
	}
}

func TestFormatPositionBon_ContainsZeit(t *testing.T) {
	payload := escpos.FormatPositionBon(testPos, "Tisch 7", "Maria", testTime, "", false)
	got := string(payload)
	if !strings.Contains(got, "19:34") {
		t.Errorf("Bon enthaelt nicht Zeitstempel; got:\n%q", got)
	}
}

func TestFormatPositionBon_WithKommentar(t *testing.T) {
	payload := escpos.FormatPositionBon(testPos, "Tisch 7", "Maria", testTime, "ohne Ketchup", false)
	got := string(payload)
	if !strings.Contains(got, "ohne Ketchup") {
		t.Errorf("Bon enthaelt nicht den Kommentar; got:\n%q", got)
	}
}

func TestFormatPositionBon_WithoutKommentar_NoEmptyKommentarLine(t *testing.T) {
	withKommentar := escpos.FormatPositionBon(testPos, "Tisch 7", "Maria", testTime, "Kommentar", false)
	without := escpos.FormatPositionBon(testPos, "Tisch 7", "Maria", testTime, "", false)
	if len(without) >= len(withKommentar) {
		t.Errorf("Bon ohne Kommentar sollte kuerzer sein als mit Kommentar")
	}
}

func TestFormatPositionBon_WithBeep_ContainsBeepBytes(t *testing.T) {
	withBeep := escpos.FormatPositionBon(testPos, "Tisch 7", "Maria", testTime, "", true)
	withoutBeep := escpos.FormatPositionBon(testPos, "Tisch 7", "Maria", testTime, "", false)
	if len(withBeep) <= len(withoutBeep) {
		t.Errorf("Bon mit Beep sollte laenger sein als ohne Beep")
	}
}

func TestFormatPositionBon_EndsWithCutPaper(t *testing.T) {
	payload := escpos.FormatPositionBon(testPos, "Tisch 7", "Maria", testTime, "", false)
	if !strings.HasSuffix(string(payload), escpos.CutPaper) {
		t.Errorf("Bon endet nicht mit CutPaper-Befehl")
	}
}

func TestFormatPositionBon_HasFiveNewlinesBeforeCut(t *testing.T) {
	payload := escpos.FormatPositionBon(testPos, "Tisch 7", "Maria", testTime, "", false)
	got := string(payload)
	cutIdx := strings.LastIndex(got, escpos.CutPaper)
	if cutIdx < 0 {
		t.Fatal("CutPaper nicht gefunden")
	}
	before := got[:cutIdx]
	if !strings.HasSuffix(before, "\n\n\n\n\n") {
		start := len(before) - 10
		if start < 0 {
			start = 0
		}
		t.Errorf("Erwartet 5 Leerzeilen vor CutPaper; suffix: %q", before[start:])
	}
}

func TestFormatSammelBon_ContainsAllPositionen(t *testing.T) {
	pos2 := kasse.Position{
		PositionID:   "pos-2",
		VarianteID:   2,
		ProduktName:  "Bratwurst",
		VarianteName: "mit Brot",
		Kategorie:    "essen",
		Einzelpreis:  250,
		Menge:        1,
	}
	payload := escpos.FormatSammelBon([]kasse.Position{testPos, pos2}, "Tisch 7", "Maria", testTime, "", false)
	got := string(payload)
	if !strings.Contains(got, "3x Pommes (gross)") {
		t.Errorf("Sammelbon enthaelt nicht erste Position; got:\n%q", got)
	}
	if !strings.Contains(got, "1x Bratwurst (mit Brot)") {
		t.Errorf("Sammelbon enthaelt nicht zweite Position; got:\n%q", got)
	}
}

func TestFormatSammelBon_ContainsTischName(t *testing.T) {
	payload := escpos.FormatSammelBon([]kasse.Position{testPos}, "Tisch 12", "Felix", testTime, "", false)
	got := string(payload)
	if !strings.Contains(got, "Tisch 12") {
		t.Errorf("Sammelbon enthaelt nicht Tischnamen; got:\n%q", got)
	}
}

func TestFormatSammelBon_EndsWithCutPaper(t *testing.T) {
	payload := escpos.FormatSammelBon([]kasse.Position{testPos}, "Tisch 7", "Maria", testTime, "", false)
	if !strings.HasSuffix(string(payload), escpos.CutPaper) {
		t.Errorf("Sammelbon endet nicht mit CutPaper-Befehl")
	}
}

func TestFormatSammelBon_WithKommentar(t *testing.T) {
	payload := escpos.FormatSammelBon([]kasse.Position{testPos}, "Tisch 7", "Maria", testTime, "ohne Ketchup fuer die Pommes", false)
	got := string(payload)
	if !strings.Contains(got, "ohne Ketchup") {
		t.Errorf("Sammelbon enthaelt nicht Kommentar; got:\n%q", got)
	}
}

func TestFormatKassenbeleg_ContainsPflichtfelder(t *testing.T) {
	payload := escpos.FormatKassenbeleg(escpos.KassenbelegData{
		Vereinsname:        "SV Musterstadt",
		Strasse:            "Musterstrasse 1",
		Plz:                "12345",
		Ort:                "Musterstadt",
		KassenSeriennummer: "2e00c5d4-7adb-4f63-84d6-a34235f2b0f4",
		Belegnummer:        "42",
		Zeitpunkt:          testTime,
		Positionen:         []kasse.Position{testPos},
		GesamtbetragCents:  900,
		Zahlungsart:        "bar",
	})
	got := string(payload)

	checks := []string{
		"KASSENBELEG",
		"SV Musterstadt",
		"Kassen-ID: 2e00c5d4-7adb-4f63-84d6-a34235f2b0f4",
		"Bon-Nr: 42",
		"3x Pommes (gross)",
		"GESAMT: 9,00 EUR",
		"Zahlungsart: bar",
	}

	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Fatalf("Kassenbeleg enthaelt %q nicht; got:\n%q", check, got)
		}
	}
}

func TestFormatKassenbeleg_EndsWithCutPaper(t *testing.T) {
	payload := escpos.FormatKassenbeleg(escpos.KassenbelegData{
		Vereinsname:        "SV Musterstadt",
		Strasse:            "Musterstrasse 1",
		Plz:                "12345",
		Ort:                "Musterstadt",
		KassenSeriennummer: "2e00c5d4-7adb-4f63-84d6-a34235f2b0f4",
		Belegnummer:        "42",
		Zeitpunkt:          testTime,
		Positionen:         []kasse.Position{testPos},
		GesamtbetragCents:  900,
		Zahlungsart:        "bar",
	})

	if !strings.HasSuffix(string(payload), escpos.CutPaper) {
		t.Errorf("Kassenbeleg endet nicht mit CutPaper-Befehl")
	}
}
