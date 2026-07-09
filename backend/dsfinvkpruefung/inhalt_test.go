//go:build unit

package dsfinvkpruefung

import (
	"strings"
	"testing"
)

// --- Fixtures für die Inhaltsprüfung ---
//
// Anders als die Struktur-Fixture (gutesArchiv, minimal) deklariert dieses Archiv
// alle Tabellen, die die Inhaltsregeln betrachten: transactions, references,
// lines_vat, transactions_vat, tse und allocation_groups. Die kaputten Fixtures
// mutieren gezielt einzelne Felder.

// gutesInhaltIndexXML deklariert die sechs von den Inhaltsregeln geprüften Tabellen
// mit ihren realen Spalten (Reihenfolge wie im Erzeuger, api/fiskal/dsfinvk).
const gutesInhaltIndexXML = `<?xml version="1.0" encoding="utf-8"?>
<!DOCTYPE DataSet SYSTEM "gdpdu-01-09-2004.dtd">
<DataSet>
  <Version>1.0</Version>
  <Media>
    <Name>CD Nummer 1</Name>
    <Table>
      <URL>transactions.csv</URL>
      <Name>Bonkopf</Name>
      <UTF8 />
      <DecimalSymbol>,</DecimalSymbol>
      <Range><From>2</From></Range>
      <VariableLength>
        <ColumnDelimiter>;</ColumnDelimiter>
        <RecordDelimiter>&#xD;&#xA;</RecordDelimiter>
        <TextEncapsulator>"</TextEncapsulator>
        <VariableColumn><Name>BON_ID</Name><AlphaNumeric /></VariableColumn>
        <VariableColumn><Name>BON_TYP</Name><AlphaNumeric /></VariableColumn>
        <VariableColumn><Name>BON_NAME</Name><AlphaNumeric /></VariableColumn>
        <VariableColumn><Name>BON_STORNO</Name><AlphaNumeric /></VariableColumn>
        <VariableColumn><Name>BEDIENER_ID</Name><AlphaNumeric /></VariableColumn>
        <VariableColumn><Name>BEDIENER_NAME</Name><AlphaNumeric /></VariableColumn>
        <VariableColumn><Name>UMS_BRUTTO</Name><Numeric><Accuracy>2</Accuracy></Numeric></VariableColumn>
      </VariableLength>
    </Table>
    <Table>
      <URL>references.csv</URL>
      <Name>Bon_Referenzen</Name>
      <UTF8 />
      <DecimalSymbol>,</DecimalSymbol>
      <Range><From>2</From></Range>
      <VariableLength>
        <ColumnDelimiter>;</ColumnDelimiter>
        <RecordDelimiter>&#xD;&#xA;</RecordDelimiter>
        <TextEncapsulator>"</TextEncapsulator>
        <VariableColumn><Name>BON_ID</Name><AlphaNumeric /></VariableColumn>
        <VariableColumn><Name>POS_ZEILE</Name><AlphaNumeric /></VariableColumn>
        <VariableColumn><Name>REF_TYP</Name><AlphaNumeric /></VariableColumn>
        <VariableColumn><Name>REF_NAME</Name><AlphaNumeric /></VariableColumn>
        <VariableColumn><Name>REF_BON_ID</Name><AlphaNumeric /></VariableColumn>
      </VariableLength>
    </Table>
    <Table>
      <URL>lines.csv</URL>
      <Name>Bonpos</Name>
      <UTF8 />
      <DecimalSymbol>,</DecimalSymbol>
      <Range><From>2</From></Range>
      <VariableLength>
        <ColumnDelimiter>;</ColumnDelimiter>
        <RecordDelimiter>&#xD;&#xA;</RecordDelimiter>
        <TextEncapsulator>"</TextEncapsulator>
        <VariableColumn><Name>BON_ID</Name><AlphaNumeric /></VariableColumn>
        <VariableColumn><Name>ARTIKELTEXT</Name><AlphaNumeric /></VariableColumn>
        <VariableColumn><Name>GV_TYP</Name><AlphaNumeric /></VariableColumn>
      </VariableLength>
    </Table>
    <Table>
      <URL>lines_vat.csv</URL>
      <Name>Bonpos_USt</Name>
      <UTF8 />
      <DecimalSymbol>,</DecimalSymbol>
      <Range><From>2</From></Range>
      <VariableLength>
        <ColumnDelimiter>;</ColumnDelimiter>
        <RecordDelimiter>&#xD;&#xA;</RecordDelimiter>
        <TextEncapsulator>"</TextEncapsulator>
        <VariableColumn><Name>BON_ID</Name><AlphaNumeric /></VariableColumn>
        <VariableColumn><Name>UST_SCHLUESSEL</Name><Numeric><Accuracy>0</Accuracy></Numeric></VariableColumn>
        <VariableColumn><Name>POS_BRUTTO</Name><Numeric><Accuracy>5</Accuracy></Numeric></VariableColumn>
      </VariableLength>
    </Table>
    <Table>
      <URL>transactions_vat.csv</URL>
      <Name>Bonkopf_USt</Name>
      <UTF8 />
      <DecimalSymbol>,</DecimalSymbol>
      <Range><From>2</From></Range>
      <VariableLength>
        <ColumnDelimiter>;</ColumnDelimiter>
        <RecordDelimiter>&#xD;&#xA;</RecordDelimiter>
        <TextEncapsulator>"</TextEncapsulator>
        <VariableColumn><Name>BON_ID</Name><AlphaNumeric /></VariableColumn>
        <VariableColumn><Name>UST_SCHLUESSEL</Name><Numeric><Accuracy>0</Accuracy></Numeric></VariableColumn>
        <VariableColumn><Name>BON_BRUTTO</Name><Numeric><Accuracy>5</Accuracy></Numeric></VariableColumn>
      </VariableLength>
    </Table>
    <Table>
      <URL>tse.csv</URL>
      <Name>Stamm_TSE</Name>
      <UTF8 />
      <DecimalSymbol>,</DecimalSymbol>
      <Range><From>2</From></Range>
      <VariableLength>
        <ColumnDelimiter>;</ColumnDelimiter>
        <RecordDelimiter>&#xD;&#xA;</RecordDelimiter>
        <TextEncapsulator>"</TextEncapsulator>
        <VariableColumn><Name>TSE_ID</Name><Numeric><Accuracy>0</Accuracy></Numeric></VariableColumn>
        <VariableColumn><Name>TSE_SERIAL</Name><AlphaNumeric /></VariableColumn>
        <VariableColumn><Name>TSE_SIG_ALGO</Name><AlphaNumeric /></VariableColumn>
        <VariableColumn><Name>TSE_PUBLIC_KEY</Name><AlphaNumeric /></VariableColumn>
        <VariableColumn><Name>TSE_ZERTIFIKAT_I</Name><AlphaNumeric /></VariableColumn>
      </VariableLength>
    </Table>
    <Table>
      <URL>allocation_groups.csv</URL>
      <Name>Bonkopf_AbrKreis</Name>
      <UTF8 />
      <DecimalSymbol>,</DecimalSymbol>
      <Range><From>2</From></Range>
      <VariableLength>
        <ColumnDelimiter>;</ColumnDelimiter>
        <RecordDelimiter>&#xD;&#xA;</RecordDelimiter>
        <TextEncapsulator>"</TextEncapsulator>
        <VariableColumn><Name>BON_ID</Name><AlphaNumeric /></VariableColumn>
        <VariableColumn><Name>ABRECHNUNGSKREIS</Name><AlphaNumeric /></VariableColumn>
      </VariableLength>
    </Table>
  </Media>
</DataSet>`

// Kanonische, konsistente CSV-Inhalte: ein Kombi-Verkaufsbon (7 % + 19 %), seine
// Warenrücknahme (negativer Beleg mit Referenz, BON_STORNO 0) und ein
// Tagesabschlussbon (AVSonstige).
const (
	// transit-1 ist ein negativer Bargeldabfluss (Geldtransit-Entnahme): negativer
	// UMS_BRUTTO, aber GV_TYP "Geldtransit" (nicht "Umsatz") und ohne Referenz —
	// er darf die Storno-Referenzregel nicht auslösen.
	gutTransactionsCSV = "BON_ID;BON_TYP;BON_NAME;BON_STORNO;BEDIENER_ID;BEDIENER_NAME;UMS_BRUTTO\r\n" +
		"verkauf-1;Beleg;;0;4;maria;7,50\r\n" +
		"storno-1;Beleg;;0;3;felix;-3,00\r\n" +
		"transit-1;Beleg;;0;2;thomas;-500,00\r\n" +
		"abschluss-1;AVSonstige;Tagesabschluss;0;2;thomas;0,00\r\n"
	gutLinesCSV = "BON_ID;ARTIKELTEXT;GV_TYP\r\n" +
		"verkauf-1;Tagesgericht;Umsatz\r\n" +
		"verkauf-1;Bier;Umsatz\r\n" +
		"storno-1;Bier;Umsatz\r\n" +
		"transit-1;Geldtransit;Geldtransit\r\n"
	gutReferencesCSV = "BON_ID;POS_ZEILE;REF_TYP;REF_NAME;REF_BON_ID\r\n" +
		"storno-1;;Transaktion;;verkauf-1\r\n"
	gutLinesVatCSV = "BON_ID;UST_SCHLUESSEL;POS_BRUTTO\r\n" +
		"verkauf-1;2;3,50000\r\n" +
		"verkauf-1;1;4,00000\r\n" +
		"storno-1;1;-3,00000\r\n" +
		"transit-1;5;-500,00000\r\n"
	gutTransactionsVatCSV = "BON_ID;UST_SCHLUESSEL;BON_BRUTTO\r\n" +
		"verkauf-1;2;3,50000\r\n" +
		"verkauf-1;1;4,00000\r\n" +
		"storno-1;1;-3,00000\r\n" +
		"transit-1;5;-500,00000\r\n"
	gutTseCSV = "TSE_ID;TSE_SERIAL;TSE_SIG_ALGO;TSE_PUBLIC_KEY;TSE_ZERTIFIKAT_I\r\n" +
		"1;TSE-SN-1;ecdsa-plain-SHA256;PUBKEY==;CERTBASE64\r\n"
	gutAllocationGroupsCSV = "BON_ID;ABRECHNUNGSKREIS\r\n" +
		"verkauf-1;Tisch 1\r\n" +
		"storno-1;Tisch 1\r\n"
)

// gutesInhaltArchiv liefert die Dateiliste eines inhaltlich konsistenten Archivs.
func gutesInhaltArchiv() []datei {
	return []datei{
		{name: "index.xml", inhalt: gutesInhaltIndexXML},
		{name: "gdpdu-01-09-2004.dtd", inhalt: "<!-- DTD -->"},
		{name: "transactions.csv", inhalt: gutTransactionsCSV},
		{name: "lines.csv", inhalt: gutLinesCSV},
		{name: "references.csv", inhalt: gutReferencesCSV},
		{name: "lines_vat.csv", inhalt: gutLinesVatCSV},
		{name: "transactions_vat.csv", inhalt: gutTransactionsVatCSV},
		{name: "tse.csv", inhalt: gutTseCSV},
		{name: "allocation_groups.csv", inhalt: gutAllocationGroupsCSV},
	}
}

// --- Gute Fixture: befundfrei ---

func TestInhalt_GutesArchivBefundfrei(t *testing.T) {
	befunde, err := PruefenBytes(baueZip(t, gutesInhaltArchiv()))
	if err != nil {
		t.Fatalf("Pruefen: %v", err)
	}
	if len(befunde) != 0 {
		t.Fatalf("erwartet befundfrei, erhielt:\n%s", befundText(befunde))
	}
}

// --- Regel 1: Storno-Referenzen und BON_STORNO ---

func TestInhalt_StornoOhneReferenz(t *testing.T) {
	// Die references.csv-Zeile des Stornos fehlt: der Storno hat keine Referenz.
	d := ersetze(gutesInhaltArchiv(), "references.csv", "BON_ID;POS_ZEILE;REF_TYP;REF_NAME;REF_BON_ID\r\n")
	befunde, _ := PruefenBytes(baueZip(t, d))
	if !hatBefund(befunde, regelStornoReferenz) {
		t.Fatalf("erwartet Befund %q, erhielt:\n%s", regelStornoReferenz, befundText(befunde))
	}
}

func TestInhalt_StornoMitLeeremRefBonID(t *testing.T) {
	// REF_BON_ID leer: die Referenz benennt keinen Ursprungsbeleg.
	kaputt := "BON_ID;POS_ZEILE;REF_TYP;REF_NAME;REF_BON_ID\r\nstorno-1;;Transaktion;;\r\n"
	d := ersetze(gutesInhaltArchiv(), "references.csv", kaputt)
	befunde, _ := PruefenBytes(baueZip(t, d))
	if !hatBefund(befunde, regelStornoReferenz) {
		t.Fatalf("erwartet Befund %q, erhielt:\n%s", regelStornoReferenz, befundText(befunde))
	}
}

func TestInhalt_StornoMitBonStornoKennzeichen(t *testing.T) {
	// BON_STORNO = 1 auf dem Negativbeleg: jotti nutzt die Negativdarstellung,
	// nie die Vorgangsaufhebung (compliance.md 6.6).
	kaputt := strings.Replace(gutTransactionsCSV, "storno-1;Beleg;;0;", "storno-1;Beleg;;1;", 1)
	d := ersetze(gutesInhaltArchiv(), "transactions.csv", kaputt)
	befunde, _ := PruefenBytes(baueZip(t, d))
	if !hatBefund(befunde, regelStornoBonStorno) {
		t.Fatalf("erwartet Befund %q, erhielt:\n%s", regelStornoBonStorno, befundText(befunde))
	}
}

func TestInhalt_BargeldabflussIstKeinStorno(t *testing.T) {
	// transit-1 ist ein negativer Bargeldabfluss ohne Referenz (GV_TYP Geldtransit):
	// die gute Fixture ist befundfrei — die Storno-Referenzregel darf ihn nicht als
	// Storno werten. (Regressionstest zum ausgeschlossenen False Positive.)
	befunde, _ := PruefenBytes(baueZip(t, gutesInhaltArchiv()))
	for _, b := range befunde {
		if b.Regel == regelStornoReferenz {
			t.Fatalf("Bargeldabfluss fälschlich als Storno gewertet: %s", b.String())
		}
	}
}

// --- Regel 2: Kombi-Steueraufteilung ---

func TestInhalt_KombiOhneBonkopfAufteilung(t *testing.T) {
	// lines_vat.csv trägt für verkauf-1 beide Sätze, transactions_vat.csv verschmilzt
	// sie fälschlich zu einer 19-%-Zeile (der 7-%-Anteil fehlt im Bonkopf).
	kaputt := "BON_ID;UST_SCHLUESSEL;BON_BRUTTO\r\n" +
		"verkauf-1;1;7,50000\r\n" +
		"storno-1;1;-3,00000\r\n"
	d := ersetze(gutesInhaltArchiv(), "transactions_vat.csv", kaputt)
	befunde, _ := PruefenBytes(baueZip(t, d))
	if !hatBefund(befunde, regelKombiSteuer) {
		t.Fatalf("erwartet Befund %q, erhielt:\n%s", regelKombiSteuer, befundText(befunde))
	}
}

// --- Regel 3: Bediener-Felder ---

func TestInhalt_BedienerNameLeer(t *testing.T) {
	kaputt := strings.Replace(gutTransactionsCSV, "verkauf-1;Beleg;;0;4;maria;", "verkauf-1;Beleg;;0;4;;", 1)
	d := ersetze(gutesInhaltArchiv(), "transactions.csv", kaputt)
	befunde, _ := PruefenBytes(baueZip(t, d))
	if !hatBefund(befunde, regelBedienerLeer) {
		t.Fatalf("erwartet Befund %q, erhielt:\n%s", regelBedienerLeer, befundText(befunde))
	}
}

func TestInhalt_BedienerIDNichtNumerisch(t *testing.T) {
	// BEDIENER_ID trägt den Klarnamen statt der user_id.
	kaputt := strings.Replace(gutTransactionsCSV, "verkauf-1;Beleg;;0;4;maria;", "verkauf-1;Beleg;;0;maria;maria;", 1)
	d := ersetze(gutesInhaltArchiv(), "transactions.csv", kaputt)
	befunde, _ := PruefenBytes(baueZip(t, d))
	if !hatBefund(befunde, regelBedienerIDNumerisch) {
		t.Fatalf("erwartet Befund %q, erhielt:\n%s", regelBedienerIDNumerisch, befundText(befunde))
	}
}

// --- Regel 4: Tagesabschluss-Zeile ---

func TestInhalt_TagesabschlussFalscherBonName(t *testing.T) {
	kaputt := strings.Replace(gutTransactionsCSV, "abschluss-1;AVSonstige;Tagesabschluss;", "abschluss-1;AVSonstige;Kassenschnitt;", 1)
	d := ersetze(gutesInhaltArchiv(), "transactions.csv", kaputt)
	befunde, _ := PruefenBytes(baueZip(t, d))
	if !hatBefund(befunde, regelTagesabschlussName) {
		t.Fatalf("erwartet Befund %q, erhielt:\n%s", regelTagesabschlussName, befundText(befunde))
	}
}

// --- Regel 5: TSE-Stammdaten ---

func TestInhalt_TSEStammdatenUnvollstaendig(t *testing.T) {
	// Public Key und Zertifikat fehlen (der Default-Zustand einer nicht
	// eingerichteten TSE).
	kaputt := "TSE_ID;TSE_SERIAL;TSE_SIG_ALGO;TSE_PUBLIC_KEY;TSE_ZERTIFIKAT_I\r\n" +
		"1;TSE-SN-1;ecdsa-plain-SHA256;;\r\n"
	d := ersetze(gutesInhaltArchiv(), "tse.csv", kaputt)
	befunde, _ := PruefenBytes(baueZip(t, d))
	if !hatBefund(befunde, regelTSEStammdaten) {
		t.Fatalf("erwartet Befund %q, erhielt:\n%s", regelTSEStammdaten, befundText(befunde))
	}
}

// --- Regel 6: Abrechnungskreis ---

func TestInhalt_AbrechnungskreisLeer(t *testing.T) {
	kaputt := "BON_ID;ABRECHNUNGSKREIS\r\nverkauf-1;\r\nstorno-1;Tisch 1\r\n"
	d := ersetze(gutesInhaltArchiv(), "allocation_groups.csv", kaputt)
	befunde, _ := PruefenBytes(baueZip(t, d))
	if !hatBefund(befunde, regelAbrechnungskreis) {
		t.Fatalf("erwartet Befund %q, erhielt:\n%s", regelAbrechnungskreis, befundText(befunde))
	}
}

func TestInhalt_AbrechnungskreisOhneBonkopf(t *testing.T) {
	// Eine Abrechnungskreis-Zeile verweist auf eine BON_ID ohne Bonkopf.
	kaputt := "BON_ID;ABRECHNUNGSKREIS\r\nverkauf-1;Tisch 1\r\nunbekannt-9;Tisch 2\r\n"
	d := ersetze(gutesInhaltArchiv(), "allocation_groups.csv", kaputt)
	befunde, _ := PruefenBytes(baueZip(t, d))
	if !hatBefund(befunde, regelAbrechnungskreis) {
		t.Fatalf("erwartet Befund %q, erhielt:\n%s", regelAbrechnungskreis, befundText(befunde))
	}
}
