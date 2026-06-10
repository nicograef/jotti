package escpos_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/nicograef/jotti/backend/api/bondruck/application/escpos"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/steuer"
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

func TestFormatDirektverkaufAbholbon_HasFixedLabelAndNoPrices(t *testing.T) {
	pos2 := kasse.Position{
		PositionID:   "pos-2",
		VarianteID:   2,
		ProduktName:  "Bier",
		VarianteName: "0,5l",
		Kategorie:    "getraenk",
		Einzelpreis:  400,
		Menge:        2,
	}

	payload := escpos.FormatDirektverkaufAbholbon([]kasse.Position{testPos, pos2}, "Maria", testTime, "abholen")
	got := string(payload)

	if !strings.Contains(got, "Direktverkauf") {
		t.Errorf("Abholbon enthaelt nicht das fixe Label Direktverkauf; got:\n%q", got)
	}
	if !strings.Contains(got, "3x Pommes (gross)") {
		t.Errorf("Abholbon enthaelt nicht die erste Position; got:\n%q", got)
	}
	if !strings.Contains(got, "2x Bier (0,5l)") {
		t.Errorf("Abholbon enthaelt nicht die zweite Position; got:\n%q", got)
	}
	if strings.Contains(got, "EUR") {
		t.Errorf("Abholbon darf keine Preise enthalten; got:\n%q", got)
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

func TestFormatKassenbeleg_ContainsSteuerkennzeichenProPosition(t *testing.T) {
	positionen := []kasse.Position{
		{PositionID: "pos-1", VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", Einzelpreis: 350, Menge: 1},
		{PositionID: "pos-2", VarianteID: 2, ProduktName: "Wasser", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "ermaessigt", Einzelpreis: 300, Menge: 1},
		{PositionID: "pos-3", VarianteID: 3, ProduktName: "Spende", VarianteName: "frei", Kategorie: "sonstiges", Steuersatz: "befreit", Einzelpreis: 200, Menge: 1},
		{PositionID: "pos-4", VarianteID: 4, ProduktName: "Kombi-Menue", VarianteName: "Standard", Kategorie: "essen", Steuersatz: "kombi", Einzelpreis: 500, Menge: 1},
	}

	payload := escpos.FormatKassenbeleg(escpos.KassenbelegData{
		Vereinsname:        "SV Musterstadt",
		Strasse:            "Musterstrasse 1",
		Plz:                "12345",
		Ort:                "Musterstadt",
		KassenSeriennummer: "2e00c5d4-7adb-4f63-84d6-a34235f2b0f4",
		Belegnummer:        "42",
		Zeitpunkt:          testTime,
		Positionen:         positionen,
		GesamtbetragCents:  1350,
		Zahlungsart:        "bar",
	})
	got := string(payload)

	checks := []string{
		"= 3,50 EUR (A)",
		"= 3,00 EUR (B)",
		"= 2,00 EUR (C)",
		"= 5,00 EUR (A/B)",
	}

	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Fatalf("Kassenbeleg enthaelt Steuerkennzeichen %q nicht; got:\n%q", check, got)
		}
	}
}

func TestFormatKassenbeleg_ContainsSteuermatrix(t *testing.T) {
	payload := escpos.FormatKassenbeleg(escpos.KassenbelegData{
		Vereinsname:        "SV Musterstadt",
		Strasse:            "Musterstrasse 1",
		Plz:                "12345",
		Ort:                "Musterstadt",
		KassenSeriennummer: "2e00c5d4-7adb-4f63-84d6-a34235f2b0f4",
		Belegnummer:        "42",
		Zeitpunkt:          testTime,
		Positionen:         []kasse.Position{{PositionID: "pos-1", VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", Einzelpreis: 700, Menge: 1}},
		Steuermatrix: []steuer.Aufteilung{
			{Satz: steuer.RegelSteuersatz, Brutto: 700, Netto: 588, Steuer: 112},
			{Satz: steuer.ErmaessigtSteuersatz, Brutto: 300, Netto: 280, Steuer: 20},
		},
		GesamtbetragCents: 1000,
		Zahlungsart:       "bar",
	})
	got := string(payload)

	checks := []string{
		"Steueraufteilung:",
		"A: Netto 5,88 EUR, Steuer 1,12 EUR, Brutto 7,00 EUR",
		"B: Netto 2,80 EUR, Steuer 0,20 EUR, Brutto 3,00 EUR",
	}

	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Fatalf("Kassenbeleg enthaelt Steuermatrix-Zeile %q nicht; got:\n%q", check, got)
		}
	}
}

func TestFormatKassenbeleg_WithoutTSE_DoesNotContainTSEBlock(t *testing.T) {
	payload := escpos.FormatKassenbeleg(escpos.KassenbelegData{
		Vereinsname:        "SV Musterstadt",
		Strasse:            "Musterstrasse 1",
		Plz:                "12345",
		Ort:                "Musterstadt",
		KassenSeriennummer: "2e00c5d4-7adb-4f63-84d6-a34235f2b0f4",
		Belegnummer:        "42",
		Zeitpunkt:          testTime,
		Positionen:         []kasse.Position{{PositionID: "pos-1", VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", Einzelpreis: 700, Menge: 1}},
		Steuermatrix: []steuer.Aufteilung{
			{Satz: steuer.RegelSteuersatz, Brutto: 700, Netto: 588, Steuer: 112},
		},
		GesamtbetragCents: 700,
		Zahlungsart:       "bar",
	})
	got := string(payload)

	if strings.Contains(got, "TSE-Daten:") {
		t.Fatalf("Kassenbeleg darf ohne TSE-Daten keinen TSE-Block enthalten; got:\n%q", got)
	}
}

func TestFormatKassenbeleg_WithTSE_ContainsTSEPflichtfelder(t *testing.T) {
	tseBeginn := time.Date(2026, 5, 1, 20, 0, 12, 0, time.UTC)
	tseEnde := time.Date(2026, 5, 1, 20, 0, 14, 0, time.UTC)

	payload := escpos.FormatKassenbeleg(escpos.KassenbelegData{
		Vereinsname:        "SV Musterstadt",
		Strasse:            "Musterstrasse 1",
		Plz:                "12345",
		Ort:                "Musterstadt",
		KassenSeriennummer: "2e00c5d4-7adb-4f63-84d6-a34235f2b0f4",
		Belegnummer:        "1003",
		Zeitpunkt:          testTime,
		Positionen:         []kasse.Position{{PositionID: "pos-1", VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", Einzelpreis: 700, Menge: 1}},
		GesamtbetragCents:  700,
		Zahlungsart:        "bar",
		TSE: &escpos.TSEAbschnitt{
			TransaktionNr:   1003,
			Signaturzaehler: 5871,
			TSESeriennummer: "SW-TSE-SN-0042",
			ZeitpunktBeginn: tseBeginn,
			ZeitpunktEnde:   tseEnde,
			Signatur:        "ABCDEF0123456789",
		},
	})
	got := string(payload)

	checks := []string{
		"TSE-Daten:",
		"TSE-Transaktion: 1003",
		"Signaturzaehler: 5871",
		"TSE-Seriennummer: SW-TSE-SN-0042",
		"TSE-Start: 01.05.2026 20:00:12",
		"TSE-Ende: 01.05.2026 20:00:14",
		"Signatur: ABCDEF0123456789",
	}

	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Fatalf("Kassenbeleg enthaelt TSE-Pflichtfeld %q nicht; got:\n%q", check, got)
		}
	}
}

func TestFormatPositionBon_SetsCP858CodepageAfterInit(t *testing.T) {
	payload := string(escpos.FormatPositionBon(testPos, "Tisch 7", "Maria", testTime, "", false))

	cpIdx := strings.Index(payload, escpos.SetCodepageCP858)
	if cpIdx < 0 {
		t.Fatal("Bon setzt nicht die CP858-Codepage (ESC t 19)")
	}

	// ESC @ (Init) setzt die Codepage zurueck; sie muss danach gesetzt werden.
	initIdx := strings.Index(payload, escpos.Init)
	if initIdx < 0 || cpIdx < initIdx {
		t.Errorf("CP858-Codepage muss nach Init gesetzt werden; initIdx=%d cpIdx=%d", initIdx, cpIdx)
	}
}

func TestFormatPositionBon_TranscodesUmlautsAndEuroToCP858(t *testing.T) {
	// Umlaute und Euro stehen im Kommentar und muessen als CP858-Einzelbytes erscheinen.
	payload := escpos.FormatPositionBon(testPos, "Tisch 7", "Maria", testTime, "äöüÄÖÜß 1€", false)

	// CP858: ä=0x84 ö=0x94 ü=0x81 Ä=0x8E Ö=0x99 Ü=0x9A ß=0xE1 €=0xD5
	wantBytes := []byte{0x84, 0x94, 0x81, 0x8E, 0x99, 0x9A, 0xE1, 0xD5}
	for _, b := range wantBytes {
		if !bytes.Contains(payload, []byte{b}) {
			t.Errorf("Bon enthaelt nicht das erwartete CP858-Byte 0x%02X", b)
		}
	}

	// Die UTF-8-Sequenz fuer ä (0xC3 0xA4) darf nach Transkodierung nicht mehr vorkommen.
	if bytes.Contains(payload, []byte{0xC3, 0xA4}) {
		t.Error("Bon enthaelt rohe UTF-8-Bytes statt CP858 (Transkodierung fehlt)")
	}
}
