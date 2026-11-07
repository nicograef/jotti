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
			t.Errorf("Kassenbeleg enthaelt %q nicht; got:\n%q", check, got)
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
			{PositionID: "pos-1", VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", EinzelpreisCents: -300, Menge: 3},
			{PositionID: "pos-2", VarianteID: 2, ProduktName: "Pfand", VarianteName: "Becher", Kategorie: "sonstiges", Steuersatz: "regel", EinzelpreisCents: -50, Menge: 1},
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
			t.Errorf("Stornobeleg enthaelt %q nicht; got:\n%q", check, got)
		}
	}

	if strings.Contains(got, "KASSENBELEG") {
		t.Errorf("Stornobeleg darf nicht als KASSENBELEG betitelt sein; got:\n%q", got)
	}
}

func TestFormatKassenbeleg_ContainsSteuerkennzeichenProPosition(t *testing.T) {
	positionen := []kasse.Position{
		{PositionID: "pos-1", VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", EinzelpreisCents: 350, Menge: 1},
		{PositionID: "pos-2", VarianteID: 2, ProduktName: "Wasser", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "ermaessigt", EinzelpreisCents: 300, Menge: 1},
		{PositionID: "pos-3", VarianteID: 3, ProduktName: "Spende", VarianteName: "frei", Kategorie: "sonstiges", Steuersatz: "befreit", EinzelpreisCents: 200, Menge: 1},
		{PositionID: "pos-4", VarianteID: 4, ProduktName: "Kombi-Menue", VarianteName: "Standard", Kategorie: "essen", Steuersatz: "kombi", EinzelpreisCents: 500, Menge: 1},
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
			t.Errorf("Kassenbeleg enthaelt Steuerkennzeichen %q nicht; got:\n%q", check, got)
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
		Positionen:         []kasse.Position{{PositionID: "pos-1", VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", EinzelpreisCents: 700, Menge: 1}},
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
		"A (19 %): Netto 5,88 EUR, Steuer 1,12 EUR, Brutto 7,00 EUR",
		"B (7 %): Netto 2,80 EUR, Steuer 0,20 EUR, Brutto 3,00 EUR",
	}

	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Errorf("Kassenbeleg enthaelt Steuermatrix-Zeile %q nicht; got:\n%q", check, got)
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
		Positionen:         []kasse.Position{{PositionID: "pos-1", VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", EinzelpreisCents: 700, Menge: 1}},
		Steuermatrix: []steuer.Aufteilung{
			{Satz: steuer.RegelSteuersatz, Brutto: 700, Netto: 588, Steuer: 112},
		},
		GesamtbetragCents: 700,
		Zahlungsart:       "bar",
	})
	got := string(payload)

	if strings.Contains(got, "TSE-Daten:") {
		t.Errorf("Kassenbeleg darf ohne TSE-Daten keinen TSE-Block enthalten; got:\n%q", got)
	}
}

func TestFormatKassenbeleg_TSEVermerk(t *testing.T) {
	tests := []struct {
		name        string
		vermerk     escpos.TSEBelegvermerk
		wantText    []string
		wantOhneTxt string
	}{
		{name: "voruebergehend", vermerk: escpos.TSEVermerkVoruebergehend, wantText: []string{"TSE-Hinweis:", "wird automatisch nachsigniert"}},
		{name: "keine Konfiguration", vermerk: escpos.TSEVermerkKeineKonfiguration, wantText: []string{"TSE-Hinweis:", "keine TSE konfiguriert"}, wantOhneTxt: "nachsigniert"},
		{name: "Signatur fehlgeschlagen", vermerk: escpos.TSEVermerkSignaturFehlgeschlagen, wantText: []string{"TSE-Hinweis:", "Signatur fehlgeschlagen.", "Vorgang ohne TSE-Signatur erfasst."}, wantOhneTxt: "nachsigniert"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := escpos.FormatKassenbeleg(escpos.KassenbelegData{
				Vereinsname:        "SV Musterstadt",
				Strasse:            "Musterstrasse 1",
				Plz:                "12345",
				Ort:                "Musterstadt",
				KassenSeriennummer: "2e00c5d4-7adb-4f63-84d6-a34235f2b0f4",
				Belegnummer:        "42",
				Zeitpunkt:          testTime,
				Positionen:         []kasse.Position{{PositionID: "pos-1", VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", EinzelpreisCents: 700, Menge: 1}},
				Steuermatrix: []steuer.Aufteilung{
					{Satz: steuer.RegelSteuersatz, Brutto: 700, Netto: 588, Steuer: 112},
				},
				TSEVermerk:        tt.vermerk,
				GesamtbetragCents: 700,
				Zahlungsart:       "bar",
			})
			got := string(payload)

			for _, want := range tt.wantText {
				if !strings.Contains(got, want) {
					t.Errorf("Kassenbeleg muss %q enthalten; got:\n%q", want, got)
				}
			}
			if tt.wantOhneTxt != "" && strings.Contains(got, tt.wantOhneTxt) {
				t.Errorf("Kassenbeleg darf %q nicht enthalten; got:\n%q", tt.wantOhneTxt, got)
			}
		})
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
		Positionen:         []kasse.Position{{PositionID: "pos-1", VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", EinzelpreisCents: 700, Menge: 1}},
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
		"TSE-Start: 01.05.2026 22:00:12",
		"TSE-Ende: 01.05.2026 22:00:14",
		"Signatur: ABCDEF0123456789",
	}

	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Errorf("Kassenbeleg enthaelt TSE-Pflichtfeld %q nicht; got:\n%q", check, got)
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
		Positionen:         []kasse.Position{{PositionID: "pos-1", VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", EinzelpreisCents: 700, Menge: 1}},
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
		t.Error("Kassenbeleg mit qr_code_data muss QR-Modellbefehl enthalten")
	}
	if !bytes.Contains(payload, []byte(escpos.QRCodePrint)) {
		t.Error("Kassenbeleg mit qr_code_data muss QR-Printbefehl enthalten")
	}
	if !bytes.Contains(payload, []byte("V0;QR-TSE-TEST")) {
		t.Error("Kassenbeleg mit qr_code_data muss QR-Payload enthalten")
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
		Positionen:         []kasse.Position{{PositionID: "pos-1", VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", EinzelpreisCents: 700, Menge: 1}},
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
		t.Error("Kassenbeleg ohne qr_code_data darf keinen nativen QR-Printbefehl enthalten")
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
	if !strings.Contains(got, "Erste Bestellung: 01.05.2026 20:01:00") {
		t.Errorf("Kassenbeleg mit erster Bestellung muss Klarschrift enthalten; got:\n%q", got)
	}
}

// A 0 % rate needs the exemption notice of § 6 Satz 1 Nr. 5 KassenSichV
// ("Hinweis darauf, dass eine Steuerbefreiung gilt").
func TestFormatKassenbeleg_SteuermatrixBefreitSatz_ZeigtBefreiungshinweis(t *testing.T) {
	payload := escpos.FormatKassenbeleg(escpos.KassenbelegData{
		Vereinsname:        "SV Musterstadt",
		Strasse:            "Musterstrasse 1",
		Plz:                "12345",
		Ort:                "Musterstadt",
		KassenSeriennummer: "2e00c5d4-7adb-4f63-84d6-a34235f2b0f4",
		Belegnummer:        "42",
		Zeitpunkt:          testTime,
		Positionen: []kasse.Position{
			{PositionID: "pos-1", VarianteID: 1, ProduktName: "Spende", VarianteName: "", Kategorie: "sonstiges", Steuersatz: "befreit", EinzelpreisCents: 500, Menge: 1},
		},
		Steuermatrix: []steuer.Aufteilung{
			{Satz: steuer.BefreitSteuersatz, Brutto: 500, Netto: 500, Steuer: 0},
		},
		GesamtbetragCents: 500,
		Zahlungsart:       "bar",
	})
	got := string(payload)

	if !strings.Contains(got, "C (umsatzsteuerfrei):") {
		t.Errorf("Beleg mit 0%%-Satz muss Befreiungshinweis enthalten; got:\n%q", got)
	}
}

// Zeitpunkte arrive as UTC but print in German local time, else the Beleg shows the wrong date.
// 2026-07-01T23:30:00Z is already 02.07.2026, 01:30 in Europe/Berlin (UTC+2).
func TestFormatKassenbeleg_ZeitpunkteInDeutscherOrtszeit(t *testing.T) {
	utcNacht := time.Date(2026, 7, 1, 23, 30, 0, 0, time.UTC)
	ersteBestellung := time.Date(2026, 7, 1, 22, 5, 9, 0, time.UTC)

	payload := escpos.FormatKassenbeleg(escpos.KassenbelegData{
		Vereinsname:              "SV Musterstadt",
		Strasse:                  "Musterstrasse 1",
		Plz:                      "12345",
		Ort:                      "Musterstadt",
		KassenSeriennummer:       "2e00c5d4-7adb-4f63-84d6-a34235f2b0f4",
		Belegnummer:              "77",
		Zeitpunkt:                utcNacht,
		ErsteBestellungZeitpunkt: &ersteBestellung,
		Positionen:               []kasse.Position{testPos},
		GesamtbetragCents:        900,
		Zahlungsart:              "bar",
		TSE: &escpos.TSEAbschnitt{
			TransaktionNr:   77,
			Signaturzaehler: 12,
			TSESeriennummer: "SW-TSE-SN-0042",
			ZeitpunktBeginn: utcNacht,
			ZeitpunktEnde:   utcNacht,
			Signatur:        "ABCDEF0123456789",
		},
	})
	got := string(payload)

	checks := []string{
		"Datum: 02.07.2026 01:30",
		"Erste Bestellung: 02.07.2026 00:05:09",
		"TSE-Start: 02.07.2026 01:30:00",
		"TSE-Ende: 02.07.2026 01:30:00",
	}

	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Errorf("Kassenbeleg enthaelt %q nicht; got:\n%q", check, got)
		}
	}
}
