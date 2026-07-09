//go:build unit

package dsfinvkpruefung

import (
	"archive/zip"
	"bytes"
	"slices"
	"strings"
	"testing"
)

// --- Fixtures ---
//
// gutesArchiv baut ein minimales, strukturell konformes DSFinV-K-Archiv: eine
// index.xml, die zwei Tabellen (eine alphanumerische, eine mit numerischer Spalte)
// deklariert, die beiden passenden CSV-Dateien und die referenzierte DTD. Die
// kaputten Fixtures leiten sich durch gezielte Mutation davon ab.

const gutesIndexXML = `<?xml version="1.0" encoding="utf-8"?>
<!DOCTYPE DataSet SYSTEM "gdpdu-01-09-2004.dtd">
<DataSet>
  <Version>1.0</Version>
  <Media>
    <Name>CD Nummer 1</Name>
    <Table>
      <URL>cashregister.csv</URL>
      <Name>Stamm_Kassen</Name>
      <UTF8 />
      <DecimalSymbol>,</DecimalSymbol>
      <DigitGroupingSymbol>.</DigitGroupingSymbol>
      <Range><From>2</From></Range>
      <VariableLength>
        <ColumnDelimiter>;</ColumnDelimiter>
        <RecordDelimiter>&#xD;&#xA;</RecordDelimiter>
        <TextEncapsulator>"</TextEncapsulator>
        <VariableColumn><Name>Z_KASSE_ID</Name><AlphaNumeric /><MaxLength>50</MaxLength></VariableColumn>
        <VariableColumn><Name>KASSE_BRAND</Name><AlphaNumeric /><MaxLength>50</MaxLength></VariableColumn>
      </VariableLength>
    </Table>
    <Table>
      <URL>payment.csv</URL>
      <Name>Z_Zahlart</Name>
      <UTF8 />
      <DecimalSymbol>,</DecimalSymbol>
      <DigitGroupingSymbol>.</DigitGroupingSymbol>
      <Range><From>2</From></Range>
      <VariableLength>
        <ColumnDelimiter>;</ColumnDelimiter>
        <RecordDelimiter>&#xD;&#xA;</RecordDelimiter>
        <TextEncapsulator>"</TextEncapsulator>
        <VariableColumn><Name>ZAHLART_TYP</Name><AlphaNumeric /><MaxLength>25</MaxLength></VariableColumn>
        <VariableColumn><Name>Z_ZAHLART_BETRAG</Name><Numeric><Accuracy>2</Accuracy></Numeric></VariableColumn>
      </VariableLength>
    </Table>
  </Media>
</DataSet>`

type datei struct {
	name   string
	inhalt string
}

// baueZip verpackt eine Liste von Dateien in ein ZIP-Archiv.
func baueZip(t *testing.T, dateien []datei) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, d := range dateien {
		w, err := zw.Create(d.name)
		if err != nil {
			t.Fatalf("zip create %q: %v", d.name, err)
		}
		if _, err := w.Write([]byte(d.inhalt)); err != nil {
			t.Fatalf("zip write %q: %v", d.name, err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("zip close: %v", err)
	}
	return buf.Bytes()
}

// gutesArchiv liefert die Dateiliste eines konformen Archivs. Der Aufrufer kann
// einzelne Einträge vor dem Verpacken ersetzen, um einen Defekt einzubauen.
func gutesArchiv() []datei {
	return []datei{
		{name: "index.xml", inhalt: gutesIndexXML},
		{name: "gdpdu-01-09-2004.dtd", inhalt: "<!-- DTD -->"},
		{name: "cashregister.csv", inhalt: "Z_KASSE_ID;KASSE_BRAND\r\nKASSE-1;jotti\r\n"},
		{name: "payment.csv", inhalt: "ZAHLART_TYP;Z_ZAHLART_BETRAG\r\nBar;12,50\r\n"},
	}
}

// ersetze tauscht den Inhalt der Datei mit dem gegebenen Namen aus.
func ersetze(dateien []datei, name, inhalt string) []datei {
	out := make([]datei, len(dateien))
	copy(out, dateien)
	for i := range out {
		if out[i].name == name {
			out[i].inhalt = inhalt
		}
	}
	return out
}

// entferne löscht die Datei mit dem gegebenen Namen.
func entferne(dateien []datei, name string) []datei {
	var out []datei
	for _, d := range dateien {
		if d.name != name {
			out = append(out, d)
		}
	}
	return out
}

// hatBefund meldet, ob ein Befund der gegebenen Regel vorliegt.
func hatBefund(befunde []Befund, regel string) bool {
	for _, b := range befunde {
		if b.Regel == regel {
			return true
		}
	}
	return false
}

// muessePruefen prüft ein Archiv und schlägt fehl, wenn die Prüfung einen echten
// Fehler zurückgibt (kein Befund, sondern ein unlesbares ZIP). Rückgabe sind die
// Befunde — so entfällt der verworfene Fehler an jeder Aufrufstelle.
func muessePruefen(t *testing.T, archiv []byte) []Befund {
	t.Helper()
	befunde, err := PruefenBytes(archiv)
	if err != nil {
		t.Fatalf("PruefenBytes: %v", err)
	}
	return befunde
}

func befundText(befunde []Befund) string {
	var b strings.Builder
	for _, x := range befunde {
		b.WriteString(x.String())
		b.WriteByte('\n')
	}
	return b.String()
}

// --- Gute Fixture: befundfrei ---

func TestPruefen_GutesArchivBefundfrei(t *testing.T) {
	befunde := muessePruefen(t, baueZip(t, gutesArchiv()))
	if len(befunde) != 0 {
		t.Fatalf("erwartet befundfrei, erhielt:\n%s", befundText(befunde))
	}
}

func TestPruefen_KeinZip(t *testing.T) {
	if _, err := PruefenBytes([]byte("kein zip")); err == nil {
		t.Fatal("erwartet Fehler für Nicht-ZIP-Eingabe")
	}
}

// --- Kaputte Fixtures: Paket & Dateinamen ---

func TestPruefen_GrossgeschriebenerDateiname(t *testing.T) {
	// CSV-Dateinamen müssen kleingeschrieben sein: payment.csv → Payment.csv.
	d := entferne(gutesArchiv(), "payment.csv")
	d = append(d, datei{name: "Payment.csv", inhalt: "ZAHLART_TYP;Z_ZAHLART_BETRAG\r\nBar;12,50\r\n"})
	befunde := muessePruefen(t, baueZip(t, d))
	if !hatBefund(befunde, regelDateiname) {
		t.Fatalf("erwartet Befund %q, erhielt:\n%s", regelDateiname, befundText(befunde))
	}
}

func TestPruefen_DateiInUnterverzeichnis(t *testing.T) {
	d := append(gutesArchiv(), datei{name: "unterordner/extra.csv", inhalt: "A;B\r\n"})
	befunde := muessePruefen(t, baueZip(t, d))
	if !hatBefund(befunde, regelDateinamePfad) {
		t.Fatalf("erwartet Befund %q, erhielt:\n%s", regelDateinamePfad, befundText(befunde))
	}
}

func TestPruefen_UnerwartetesFremdformat(t *testing.T) {
	d := append(gutesArchiv(), datei{name: "readme.txt", inhalt: "hallo"})
	befunde := muessePruefen(t, baueZip(t, d))
	if !hatBefund(befunde, regelDateinameFremdformat) {
		t.Fatalf("erwartet Befund %q, erhielt:\n%s", regelDateinameFremdformat, befundText(befunde))
	}
}

func TestPruefen_FehlendeDTD(t *testing.T) {
	befunde := muessePruefen(t, baueZip(t, entferne(gutesArchiv(), "gdpdu-01-09-2004.dtd")))
	if !hatBefund(befunde, regelPaketpflicht) {
		t.Fatalf("erwartet Befund %q, erhielt:\n%s", regelPaketpflicht, befundText(befunde))
	}
}

func TestPruefen_FehlendeIndexXML(t *testing.T) {
	befunde := muessePruefen(t, baueZip(t, entferne(gutesArchiv(), "index.xml")))
	if !hatBefund(befunde, regelPaketpflicht) {
		t.Fatalf("erwartet Befund %q, erhielt:\n%s", regelPaketpflicht, befundText(befunde))
	}
}

// --- Kaputte Fixtures: index.xml / DTD-Struktur ---

func TestPruefen_IndexNichtWohlgeformt(t *testing.T) {
	befunde := muessePruefen(t, baueZip(t, ersetze(gutesArchiv(), "index.xml", "<DataSet><Version>1.0")))
	if !hatBefund(befunde, regelIndexParsbar) {
		t.Fatalf("erwartet Befund %q, erhielt:\n%s", regelIndexParsbar, befundText(befunde))
	}
}

func TestPruefen_IndexOhneDoctype(t *testing.T) {
	ohneDoctype := strings.Replace(gutesIndexXML,
		`<!DOCTYPE DataSet SYSTEM "gdpdu-01-09-2004.dtd">`, "", 1)
	befunde := muessePruefen(t, baueZip(t, ersetze(gutesArchiv(), "index.xml", ohneDoctype)))
	if !hatBefund(befunde, regelIndexDoctype) {
		t.Fatalf("erwartet Befund %q, erhielt:\n%s", regelIndexDoctype, befundText(befunde))
	}
}

func TestPruefen_IndexFalschesWurzelelement(t *testing.T) {
	xml := `<?xml version="1.0"?>` + "\n" +
		`<!DOCTYPE DataSet SYSTEM "gdpdu-01-09-2004.dtd">` + "\n" +
		`<Falsch><Version>1.0</Version></Falsch>`
	befunde := muessePruefen(t, baueZip(t, ersetze(gutesArchiv(), "index.xml", xml)))
	if !hatBefund(befunde, regelIndexWurzel) {
		t.Fatalf("erwartet Befund %q, erhielt:\n%s", regelIndexWurzel, befundText(befunde))
	}
}

func TestPruefen_IndexFehlendeVersion(t *testing.T) {
	ohneVersion := strings.Replace(gutesIndexXML, "<Version>1.0</Version>", "", 1)
	befunde := muessePruefen(t, baueZip(t, ersetze(gutesArchiv(), "index.xml", ohneVersion)))
	if !hatBefund(befunde, regelIndexVersion) {
		t.Fatalf("erwartet Befund %q, erhielt:\n%s", regelIndexVersion, befundText(befunde))
	}
}

func TestPruefen_IndexFalschesDezimalsymbol(t *testing.T) {
	// DecimalSymbol Punkt statt Komma verletzt die DSFinV-K-Formatvorgabe.
	kaputt := strings.Replace(gutesIndexXML, "<DecimalSymbol>,</DecimalSymbol>", "<DecimalSymbol>.</DecimalSymbol>", 1)
	befunde := muessePruefen(t, baueZip(t, ersetze(gutesArchiv(), "index.xml", kaputt)))
	if !hatBefund(befunde, regelIndexFormat) {
		t.Fatalf("erwartet Befund %q, erhielt:\n%s", regelIndexFormat, befundText(befunde))
	}
}

func TestPruefen_IndexFalscheKopfzeilenRange(t *testing.T) {
	kaputt := strings.Replace(gutesIndexXML, "<From>2</From>", "<From>1</From>", 1)
	befunde := muessePruefen(t, baueZip(t, ersetze(gutesArchiv(), "index.xml", kaputt)))
	if !hatBefund(befunde, regelIndexKopfzeile) {
		t.Fatalf("erwartet Befund %q, erhielt:\n%s", regelIndexKopfzeile, befundText(befunde))
	}
}

func TestPruefen_IndexSpalteOhneDatentyp(t *testing.T) {
	kaputt := strings.Replace(gutesIndexXML,
		"<VariableColumn><Name>KASSE_BRAND</Name><AlphaNumeric /><MaxLength>50</MaxLength></VariableColumn>",
		"<VariableColumn><Name>KASSE_BRAND</Name><MaxLength>50</MaxLength></VariableColumn>", 1)
	befunde := muessePruefen(t, baueZip(t, ersetze(gutesArchiv(), "index.xml", kaputt)))
	if !hatBefund(befunde, regelIndexSpalte) {
		t.Fatalf("erwartet Befund %q, erhielt:\n%s", regelIndexSpalte, befundText(befunde))
	}
}

// --- Kaputte Fixtures: index.xml ↔ CSV Abgleich ---

func TestPruefen_DeklarierteDateiFehlt(t *testing.T) {
	befunde := muessePruefen(t, baueZip(t, entferne(gutesArchiv(), "payment.csv")))
	if !hatBefund(befunde, regelIndexDatei) {
		t.Fatalf("erwartet Befund %q, erhielt:\n%s", regelIndexDatei, befundText(befunde))
	}
}

func TestPruefen_UndeklarierteCSV(t *testing.T) {
	d := append(gutesArchiv(), datei{name: "extra.csv", inhalt: "A;B\r\nx;y\r\n"})
	befunde := muessePruefen(t, baueZip(t, d))
	if !hatBefund(befunde, regelCsvUndeklar) {
		t.Fatalf("erwartet Befund %q, erhielt:\n%s", regelCsvUndeklar, befundText(befunde))
	}
}

// --- Kaputte Fixtures: CSV-Struktur ---

func TestPruefen_CSVFehlendeCRLF(t *testing.T) {
	// Unix-Zeilenenden statt CRLF.
	befunde := muessePruefen(t, baueZip(t, ersetze(gutesArchiv(), "cashregister.csv",
		"Z_KASSE_ID;KASSE_BRAND\nKASSE-1;jotti\n")))
	if !hatBefund(befunde, regelCsvCRLF) {
		t.Fatalf("erwartet Befund %q, erhielt:\n%s", regelCsvCRLF, befundText(befunde))
	}
}

func TestPruefen_CSVFalscheSpaltenreihenfolge(t *testing.T) {
	// Header mit vertauschten Spalten.
	befunde := muessePruefen(t, baueZip(t, ersetze(gutesArchiv(), "cashregister.csv",
		"KASSE_BRAND;Z_KASSE_ID\r\njotti;KASSE-1\r\n")))
	if !hatBefund(befunde, regelCsvKopfzeile) {
		t.Fatalf("erwartet Befund %q, erhielt:\n%s", regelCsvKopfzeile, befundText(befunde))
	}
}

func TestPruefen_CSVFalscheHeaderNamen(t *testing.T) {
	befunde := muessePruefen(t, baueZip(t, ersetze(gutesArchiv(), "cashregister.csv",
		"Z_KASSE_ID;FALSCH\r\nKASSE-1;jotti\r\n")))
	if !hatBefund(befunde, regelCsvKopfzeile) {
		t.Fatalf("erwartet Befund %q, erhielt:\n%s", regelCsvKopfzeile, befundText(befunde))
	}
}

func TestPruefen_CSVFalscheFeldanzahl(t *testing.T) {
	befunde := muessePruefen(t, baueZip(t, ersetze(gutesArchiv(), "cashregister.csv",
		"Z_KASSE_ID;KASSE_BRAND\r\nKASSE-1;jotti;ZUVIEL\r\n")))
	if !hatBefund(befunde, regelCsvSpaltenzahl) {
		t.Fatalf("erwartet Befund %q, erhielt:\n%s", regelCsvSpaltenzahl, befundText(befunde))
	}
}

func TestPruefen_CSVPunktStattKommaImNumerischenFeld(t *testing.T) {
	// payment.csv hat eine numerische Spalte Z_ZAHLART_BETRAG.
	befunde := muessePruefen(t, baueZip(t, ersetze(gutesArchiv(), "payment.csv",
		"ZAHLART_TYP;Z_ZAHLART_BETRAG\r\nBar;12.50\r\n")))
	if !hatBefund(befunde, regelCsvDezimal) {
		t.Fatalf("erwartet Befund %q, erhielt:\n%s", regelCsvDezimal, befundText(befunde))
	}
}

func TestPruefen_CSVLeer(t *testing.T) {
	befunde := muessePruefen(t, baueZip(t, ersetze(gutesArchiv(), "cashregister.csv", "")))
	if !hatBefund(befunde, regelCsvLeer) {
		t.Fatalf("erwartet Befund %q, erhielt:\n%s", regelCsvLeer, befundText(befunde))
	}
}

// splitFelder muss das Doublequote-Escaping und eingebettete Semikola beachten.
func TestSplitFelder_QuotedSemicolon(t *testing.T) {
	felder := splitFelder(`A;"B;mit;Semikolon";"C ""quote"""`)
	erwartet := []string{"A", "B;mit;Semikolon", `C "quote"`}
	if !slices.Equal(felder, erwartet) {
		t.Fatalf("splitFelder = %#v, erwartet %#v", felder, erwartet)
	}
}
