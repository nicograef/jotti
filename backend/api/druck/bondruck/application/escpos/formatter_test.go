package escpos_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/nicograef/jotti/backend/api/druck/bondruck/application/escpos"
	"github.com/nicograef/jotti/backend/domain/kasse"
)

var (
	testPos = kasse.Position{
		PositionID:       "pos-1",
		VarianteID:       1,
		ProduktName:      "Pommes",
		VarianteName:     "gross",
		Kategorie:        "essen",
		EinzelpreisCents: 300,
		Menge:            3,
	}
	testTime = time.Date(2024, 6, 15, 19, 34, 0, 0, time.UTC)
)

func TestFormatPositionBon_ContainsKopfUndFuss(t *testing.T) {
	got := string(escpos.FormatPositionBon(testPos, "Tisch 7", "Maria", testTime, ""))
	for _, want := range []string{"Tisch 7", "3x Pommes gross", "Bedienung: Maria", "21:34"} {
		if !strings.Contains(got, want) {
			t.Errorf("Bon enthaelt nicht %q; got:\n%q", want, got)
		}
	}
}

func TestFormatPositionBon_WithKommentar(t *testing.T) {
	payload := escpos.FormatPositionBon(testPos, "Tisch 7", "Maria", testTime, "ohne Ketchup")
	got := string(payload)
	if !strings.Contains(got, "ohne Ketchup") {
		t.Errorf("Bon enthaelt nicht den Kommentar; got:\n%q", got)
	}
}

func TestFormatPositionBon_WithoutKommentar_NoEmptyKommentarLine(t *testing.T) {
	withKommentar := escpos.FormatPositionBon(testPos, "Tisch 7", "Maria", testTime, "Kommentar")
	without := escpos.FormatPositionBon(testPos, "Tisch 7", "Maria", testTime, "")
	if len(without) >= len(withKommentar) {
		t.Errorf("Bon ohne Kommentar sollte kuerzer sein als mit Kommentar")
	}
}

func TestFormat_EndsWithCutPaper(t *testing.T) {
	bons := map[string][]byte{
		"Positionsbon": escpos.FormatPositionBon(testPos, "Tisch 7", "Maria", testTime, ""),
		"Sammelbon":    escpos.FormatSammelBon([]kasse.Position{testPos}, "Tisch 7", "Maria", testTime, ""),
		"Kassenbeleg": escpos.FormatKassenbeleg(escpos.KassenbelegData{
			Vereinsname:       "SV Musterstadt",
			Belegnummer:       "42",
			Zeitpunkt:         testTime,
			Positionen:        []kasse.Position{testPos},
			GesamtbetragCents: 900,
			Zahlungsart:       "bar",
		}),
	}
	for name, payload := range bons {
		if !strings.HasSuffix(string(payload), escpos.CutPaper) {
			t.Errorf("%s endet nicht mit CutPaper-Befehl", name)
		}
	}
}

func TestFormatPositionBon_HasFiveNewlinesBeforeCut(t *testing.T) {
	payload := escpos.FormatPositionBon(testPos, "Tisch 7", "Maria", testTime, "")
	got := string(payload)
	cutIdx := strings.LastIndex(got, escpos.CutPaper)
	if cutIdx < 0 {
		t.Fatal("CutPaper nicht gefunden")
	}
	before := got[:cutIdx]
	if !strings.HasSuffix(before, "\n\n\n\n\n") {
		start := max(len(before)-10, 0)
		t.Errorf("Erwartet 5 Leerzeilen vor CutPaper; suffix: %q", before[start:])
	}
}

func TestFormatSammelBon_ContainsAllPositionen(t *testing.T) {
	pos2 := kasse.Position{
		PositionID:       "pos-2",
		VarianteID:       2,
		ProduktName:      "Bratwurst",
		VarianteName:     "mit Brot",
		Kategorie:        "essen",
		EinzelpreisCents: 250,
		Menge:            1,
	}
	payload := escpos.FormatSammelBon([]kasse.Position{testPos, pos2}, "Tisch 7", "Maria", testTime, "")
	got := string(payload)
	if !strings.Contains(got, "3x Pommes gross") {
		t.Errorf("Sammelbon enthaelt nicht erste Position; got:\n%q", got)
	}
	if !strings.Contains(got, "1x Bratwurst mit Brot") {
		t.Errorf("Sammelbon enthaelt nicht zweite Position; got:\n%q", got)
	}
}

func TestFormatSammelBon_ContainsTischName(t *testing.T) {
	payload := escpos.FormatSammelBon([]kasse.Position{testPos}, "Tisch 12", "Felix", testTime, "")
	got := string(payload)
	if !strings.Contains(got, "Tisch 12") {
		t.Errorf("Sammelbon enthaelt nicht Tischnamen; got:\n%q", got)
	}
}

func TestFormatSammelBon_WithKommentar(t *testing.T) {
	payload := escpos.FormatSammelBon([]kasse.Position{testPos}, "Tisch 7", "Maria", testTime, "ohne Ketchup fuer die Pommes")
	got := string(payload)
	if !strings.Contains(got, "ohne Ketchup") {
		t.Errorf("Sammelbon enthaelt nicht Kommentar; got:\n%q", got)
	}
}

const (
	arbeitsbonKopfTisch7 = escpos.Init + escpos.SetCodepagePC437 + escpos.AlignCenter + escpos.TextDoubleAll + escpos.BoldOn +
		"== Tisch 7 ==\n" + escpos.BoldOff + escpos.TextNormal + "\n"
	arbeitsbonFussMaria = "------------------------------------------------\n" + "  21:34  Bedienung: Maria\n" + "\n\n\n\n\n" + escpos.CutPaper
)

func TestFormatSammelBon_PositionskommentarUnterSeinerPosition(t *testing.T) {
	kommentiert := testPos
	kommentiert.Kommentar = "bitte ohne Zwiebeln und ohne Kuemmel, dafuer mit extra viel Senf"
	ohne := kasse.Position{PositionID: "pos-2", ProduktName: "Bratwurst", VarianteName: "mit Brot", Kategorie: "essen", Menge: 1}

	got := escpos.FormatSammelBon([]kasse.Position{kommentiert, ohne}, "Tisch 7", "Maria", testTime, "Tisch wartet")

	want := arbeitsbonKopfTisch7 +
		escpos.AlignLeft + escpos.TextDoubleHigh + escpos.BoldOn +
		"3x Pommes gross\n" +
		escpos.TextNormal +
		"  bitte ohne Zwiebeln und ohne Kuemmel, dafuer\n" +
		"  mit extra viel Senf\n" +
		escpos.TextDoubleHigh +
		"1x Bratwurst mit Brot\n" +
		escpos.BoldOff + escpos.TextNormal +
		"\n" + escpos.BoldOn + "Tisch wartet\n" + escpos.BoldOff +
		arbeitsbonFussMaria
	if string(got) != want {
		t.Errorf("Sammelbon:\n got %q\nwant %q", got, want)
	}
}

func TestFormatPositionBon_PositionskommentarVorBestellkommentar(t *testing.T) {
	kommentiert := testPos
	kommentiert.Kommentar = "ohne Ketchup"

	got := escpos.FormatPositionBon(kommentiert, "Tisch 7", "Maria", testTime, "Tisch wartet")

	want := arbeitsbonKopfTisch7 +
		escpos.TextDoubleHigh + escpos.BoldOn + "3x Pommes gross\n" + escpos.BoldOff + escpos.TextNormal +
		escpos.BoldOn + "ohne Ketchup\n" + escpos.BoldOff +
		"\n" + escpos.AlignLeft + escpos.BoldOn + "Tisch wartet\n" + escpos.BoldOff +
		escpos.AlignLeft + arbeitsbonFussMaria
	if string(got) != want {
		t.Errorf("Einzelbon:\n got %q\nwant %q", got, want)
	}
}

func TestFormatDirektverkaufAbholbon_HasFixedLabelAndNoPrices(t *testing.T) {
	pos2 := kasse.Position{
		PositionID:       "pos-2",
		VarianteID:       2,
		ProduktName:      "Bier",
		VarianteName:     "0,5l",
		Kategorie:        "getraenk",
		EinzelpreisCents: 400,
		Menge:            2,
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

// ESC t 6 is Hiragana on Epson (#144); ESC t 0 is PC437 on Epson, MUNBYN and most clones.
func TestFormatPositionBon_SetsPC437CodepageAfterInit(t *testing.T) {
	payload := string(escpos.FormatPositionBon(testPos, "Tisch 7", "Maria", testTime, ""))

	cpIdx := strings.Index(payload, "\x1b\x74\x00")
	if cpIdx < 0 {
		t.Fatal("Bon setzt nicht die PC437-Codepage (ESC t 0)")
	}

	// ESC @ (Init) setzt die Codepage zurück; sie muss danach gesetzt werden.
	initIdx := strings.Index(payload, escpos.Init)
	if initIdx < 0 || cpIdx < initIdx {
		t.Errorf("PC437-Codepage muss nach Init gesetzt werden; initIdx=%d cpIdx=%d", initIdx, cpIdx)
	}
}

// Handy-Tastaturen setzen typografische Zeichen, die CP437 fehlt; sie werden lesbares ASCII.
func TestFormatPositionBon_KommentarAlsCP437(t *testing.T) {
	payload := escpos.FormatPositionBon(testPos, "Tisch 7", "Maria", testTime, "„Weißwein“ für Jörg – ÄÖÜ, 2 € zurück…")

	want := []byte("\"Wei\xe1wein\" f\x81r J\x94rg - \x8e\x99\x9a, 2 EUR zur\x81ck...")
	if !bytes.Contains(payload, want) {
		t.Errorf("Bon enthaelt nicht den CP437-Kommentar %q; got:\n%q", want, payload)
	}
}

// Steuerbytes im Kommentar führt der Drucker als ESC/POS-Befehle aus; printable entfernt sie.
func TestFormatPositionBon_KommentarOhneSteuerbytes(t *testing.T) {
	payload := escpos.FormatPositionBon(testPos, "Tisch 7", "Maria", testTime, "Ohne\x1b\x64\x05Zwiebeln\x1d\x56\x00!")

	for _, injiziert := range []string{"\x1b\x64\x05", "\x1d\x56\x00"} {
		if bytes.Contains(payload, []byte(injiziert)) {
			t.Errorf("Bon enthaelt den injizierten Befehl %q; got:\n%q", injiziert, payload)
		}
	}
	if !bytes.Contains(payload, []byte("OhnedZwiebelnV!\n")) {
		t.Errorf("Bon enthaelt den bereinigten Kommentar nicht; got:\n%q", payload)
	}
}

func TestFormatTestbon_DrucktUmlautzeileInCP437(t *testing.T) {
	payload := escpos.FormatTestbon("Getränke", testTime)

	for _, want := range []string{"Station: Getr\x84nke\n", "Umlaute: \x84\x94\x81 \x8e\x99\x9a \xe1\n"} {
		if !bytes.Contains(payload, []byte(want)) {
			t.Errorf("Testbon enthaelt %q nicht; got:\n%q", want, payload)
		}
	}
}

func TestFormatPositionBon_ZeitpunktInDeutscherOrtszeit(t *testing.T) {
	utcNacht := time.Date(2026, 7, 1, 23, 30, 0, 0, time.UTC)

	payload := escpos.FormatPositionBon(testPos, "Tisch 7", "Maria", utcNacht, "")
	got := string(payload)

	if !strings.Contains(got, "01:30") {
		t.Errorf("Arbeitsbon zeigt nicht die deutsche Ortszeit 01:30; got:\n%q", got)
	}
}
