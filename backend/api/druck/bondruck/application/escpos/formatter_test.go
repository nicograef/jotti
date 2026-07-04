package escpos_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/nicograef/jotti/backend/api/druck/bondruck/application/escpos"
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
	if !strings.Contains(got, "3x Pommes gross") {
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
	if !strings.Contains(got, "3x Pommes gross") {
		t.Errorf("Sammelbon enthaelt nicht erste Position; got:\n%q", got)
	}
	if !strings.Contains(got, "1x Bratwurst mit Brot") {
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
	if !strings.Contains(got, "3x Pommes gross") {
		t.Errorf("Abholbon enthaelt nicht die erste Position; got:\n%q", got)
	}
	if !strings.Contains(got, "2x Bier 0,5l") {
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
		"3x Pommes gross",
		"GESAMT: 9,00 EUR",
		"Zahlungsart: bar",
	}

	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Fatalf("Kassenbeleg enthaelt %q nicht; got:\n%q", check, got)
		}
	}
}

func TestFormatKassenbeleg_Stornobeleg_ContainsStornoFelder(t *testing.T) {
	payload := escpos.FormatKassenbeleg(escpos.KassenbelegData{
		Vereinsname:        "SV Musterstadt",
		Strasse:            "Musterstrasse 1",
		Plz:                "12345",
		Ort:                "Musterstadt",
		KassenSeriennummer: "2e00c5d4-7adb-4f63-84d6-a34235f2b0f4",
		Belegnummer:        "43",
		Zeitpunkt:          testTime,
		Positionen: []kasse.Position{
			{PositionID: "pos-1", VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", Einzelpreis: -300, Menge: 3},
			{PositionID: "pos-2", VarianteID: 2, ProduktName: "Pfand", VarianteName: "Becher", Kategorie: "sonstiges", Steuersatz: "regel", Einzelpreis: -50, Menge: 1},
		},
		GesamtbetragCents:   -950,
		Zahlungsart:         "bar",
		Stornobeleg:         true,
		StornoZuBelegnummer: "42",
	})
	got := string(payload)

	checks := []string{
		"STORNOBELEG",
		"Bon-Nr: 43",
		"Storno zu Bon-Nr: 42",
		"-3,00 x 3 = -9,00 EUR",
		"-0,50 x 1 = -0,50 EUR",
		"GESAMT: -9,50 EUR",
	}
	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Fatalf("Stornobeleg enthaelt %q nicht; got:\n%q", check, got)
		}
	}

	if strings.Contains(got, "KASSENBELEG") {
		t.Fatalf("Stornobeleg darf nicht als KASSENBELEG betitelt sein; got:\n%q", got)
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

func TestFormatKassenbeleg_WithTSEAusfallvermerk_ContainsAusfallhinweis(t *testing.T) {
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
		TSEVermerk:        escpos.TSEVermerkVoruebergehend,
		GesamtbetragCents: 700,
		Zahlungsart:       "bar",
	})
	got := string(payload)

	if !strings.Contains(got, "TSE-Hinweis:") {
		t.Fatalf("Kassenbeleg mit TSE-Ausfall muss Hinweis enthalten; got:\n%q", got)
	}
	if !strings.Contains(got, "wird automatisch nachsigniert") {
		t.Fatalf("Kassenbeleg mit TSE-Ausfall muss Nachsignierhinweis enthalten; got:\n%q", got)
	}
}

func TestFormatKassenbeleg_WithKeineKonfiguration_ContainsHinweis(t *testing.T) {
	payload := escpos.FormatKassenbeleg(escpos.KassenbelegData{
		Vereinsname:        "SV Musterstadt",
		Strasse:            "Musterstrasse 1",
		Plz:                "12345",
		Ort:                "Musterstadt",
		KassenSeriennummer: "2e00c5d4-7adb-4f63-84d6-a34235f2b0f4",
		Belegnummer:        "43",
		Zeitpunkt:          testTime,
		Positionen:         []kasse.Position{{PositionID: "pos-1", VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", Einzelpreis: 700, Menge: 1}},
		Steuermatrix: []steuer.Aufteilung{
			{Satz: steuer.RegelSteuersatz, Brutto: 700, Netto: 588, Steuer: 112},
		},
		TSEVermerk:        escpos.TSEVermerkKeineKonfiguration,
		GesamtbetragCents: 700,
		Zahlungsart:       "bar",
	})
	got := string(payload)

	if !strings.Contains(got, "TSE-Hinweis:") {
		t.Fatalf("Kassenbeleg ohne TSE-Konfiguration muss Hinweis enthalten; got:\n%q", got)
	}
	if !strings.Contains(got, "keine TSE konfiguriert") {
		t.Fatalf("Kassenbeleg ohne TSE-Konfiguration muss den Konfigurationshinweis enthalten; got:\n%q", got)
	}
	if strings.Contains(got, "nachsigniert") {
		t.Fatalf("Kassenbeleg ohne TSE-Konfiguration darf keinen Nachsignierhinweis enthalten; got:\n%q", got)
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
			QRCodeData:      "V0;TSE:1003",
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

func TestFormatKassenbeleg_WithTSEQRCode_ContainsNativeESCPosQR(t *testing.T) {
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
			ZeitpunktBeginn: time.Date(2026, 5, 1, 20, 0, 12, 0, time.UTC),
			ZeitpunktEnde:   time.Date(2026, 5, 1, 20, 0, 14, 0, time.UTC),
			Signatur:        "ABCDEF0123456789",
			QRCodeData:      "V0;QR-TSE-TEST",
		},
	})

	if !bytes.Contains(payload, []byte(escpos.QRCodeModel2)) {
		t.Fatal("Kassenbeleg mit qr_code_data muss QR-Modellbefehl enthalten")
	}
	if !bytes.Contains(payload, []byte(escpos.QRCodePrint)) {
		t.Fatal("Kassenbeleg mit qr_code_data muss QR-Printbefehl enthalten")
	}
	if !bytes.Contains(payload, []byte("V0;QR-TSE-TEST")) {
		t.Fatal("Kassenbeleg mit qr_code_data muss QR-Payload enthalten")
	}
}

func TestFormatKassenbeleg_WithoutTSEQRCode_DoesNotContainNativeESCPosQR(t *testing.T) {
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
			ZeitpunktBeginn: time.Date(2026, 5, 1, 20, 0, 12, 0, time.UTC),
			ZeitpunktEnde:   time.Date(2026, 5, 1, 20, 0, 14, 0, time.UTC),
			Signatur:        "ABCDEF0123456789",
		},
	})

	if bytes.Contains(payload, []byte(escpos.QRCodePrint)) {
		t.Fatal("Kassenbeleg ohne qr_code_data darf keinen nativen QR-Printbefehl enthalten")
	}
}

func TestFormatKassenbeleg_WithErsteBestellungZeitpunkt_ContainsKlartext(t *testing.T) {
	ersteBestellung := time.Date(2026, 5, 1, 18, 1, 0, 0, time.UTC)
	payload := escpos.FormatKassenbeleg(escpos.KassenbelegData{
		Vereinsname:              "SV Musterstadt",
		Strasse:                  "Musterstrasse 1",
		Plz:                      "12345",
		Ort:                      "Musterstadt",
		KassenSeriennummer:       "2e00c5d4-7adb-4f63-84d6-a34235f2b0f4",
		Belegnummer:              "42",
		Zeitpunkt:                testTime,
		ErsteBestellungZeitpunkt: &ersteBestellung,
		Positionen:               []kasse.Position{testPos},
		GesamtbetragCents:        900,
		Zahlungsart:              "bar",
	})

	got := string(payload)
	if !strings.Contains(got, "Erste Bestellung: 01.05.2026 18:01:00") {
		t.Fatalf("Kassenbeleg mit erster Bestellung muss Klarschrift enthalten; got:\n%q", got)
	}
}

func TestFormatPositionBon_SetsWPC1252CodepageAfterInit(t *testing.T) {
	payload := string(escpos.FormatPositionBon(testPos, "Tisch 7", "Maria", testTime, "", false))

	cpIdx := strings.Index(payload, escpos.SetCodepageWPC1252)
	if cpIdx < 0 {
		t.Fatal("Bon setzt nicht die WPC1252-Codepage (ESC t 6)")
	}

	// ESC @ (Init) setzt die Codepage zurueck; sie muss danach gesetzt werden.
	initIdx := strings.Index(payload, escpos.Init)
	if initIdx < 0 || cpIdx < initIdx {
		t.Errorf("WPC1252-Codepage muss nach Init gesetzt werden; initIdx=%d cpIdx=%d", initIdx, cpIdx)
	}
}

func TestFormatPositionBon_TranscodesUmlautsAndEuroToWPC1252(t *testing.T) {
	// Umlaute und Euro stehen im Kommentar und muessen als WPC1252-Einzelbytes erscheinen.
	payload := escpos.FormatPositionBon(testPos, "Tisch 7", "Maria", testTime, "äöüÄÖÜß 1€", false)

	// WPC1252 (Windows-1252): ä=0xE4 ö=0xF6 ü=0xFC Ä=0xC4 Ö=0xD6 Ü=0xDC ß=0xDF €=0x80
	wantBytes := []byte{0xE4, 0xF6, 0xFC, 0xC4, 0xD6, 0xDC, 0xDF, 0x80}
	for _, b := range wantBytes {
		if !bytes.Contains(payload, []byte{b}) {
			t.Errorf("Bon enthaelt nicht das erwartete WPC1252-Byte 0x%02X", b)
		}
	}

	// Die UTF-8-Sequenz fuer ä (0xC3 0xA4) darf nach Transkodierung nicht mehr vorkommen.
	if bytes.Contains(payload, []byte{0xC3, 0xA4}) {
		t.Error("Bon enthaelt rohe UTF-8-Bytes statt WPC1252 (Transkodierung fehlt)")
	}
}
