package dsfinvkpruefung

import (
	"bytes"
	"encoding/xml"
	"fmt"
)

// Regel-Kennungen der index.xml-/DTD-Prüfung.
const (
	regelIndexParsbar    = "index-parsbar"
	regelIndexWurzel     = "index-wurzel"
	regelIndexVersion    = "index-version"
	regelIndexMedia      = "index-media"
	regelIndexTabelle    = "index-tabelle"
	regelIndexSpalte     = "index-spalte"
	regelIndexFormat     = "index-format"
	regelIndexKopfzeile  = "index-kopfzeile-range"
	regelIndexDoctype    = "index-doctype"
	dtdDecimalSymbol     = ","      // DSFinV-K: Dezimal-KOMMA (index.xml DecimalSymbol)
	dtdColumnDelimiter   = ";"      // Semikolon als Spaltentrenner
	dtdRecordDelimiterCR = "\r"     // CR ist Teil des Zeilentrenners (CRLF)
	dtdRecordDelimiterLF = "\n"     // LF ist Teil des Zeilentrenners (CRLF)
	dtdKopfzeileFrom     = "2"      // Range/From = 2: Datenzeilen ab Zeile 2 (Zeile 1 = Header)
	dtdTextEncapsulator  = "\""     // Doublequote als Text-Begrenzer
	doctypeMarker        = dtdDatei // die index.xml referenziert die GDPdU-DTD im DOCTYPE
)

// indexSpalte ist eine geparste VariableColumn-Deklaration der index.xml.
type indexSpalte struct {
	Name         string
	Numeric      bool // AlphaNumeric vs. Numeric (bestimmt das Dezimalformat der CSV)
	DezimalKomma bool
}

// indexTabelle ist eine geparste Table-Deklaration: die referenzierte CSV-Datei,
// ihre Spalten in deklarierter Reihenfolge und die für die CSV-Prüfung relevanten
// Formatangaben.
type indexTabelle struct {
	URL             string
	Name            string
	Spalten         []indexSpalte
	columnDelimiter string
	recordDelimiter string
	decimalSymbol   string
	rangeFrom       string
	textEncap       string
	utf8            bool
}

// --- XML-Bindings (spiegeln die DTD-Struktur der index.xml) ---

type xmlDataSet struct {
	XMLName xml.Name
	Version string `xml:"Version"`
	Media   []struct {
		Name   string     `xml:"Name"`
		Tables []xmlTable `xml:"Table"`
	} `xml:"Media"`
}

type xmlTable struct {
	URL           string    `xml:"URL"`
	Name          string    `xml:"Name"`
	UTF8          *struct{} `xml:"UTF8"`
	DecimalSymbol string    `xml:"DecimalSymbol"`
	Range         *struct {
		From string `xml:"From"`
	} `xml:"Range"`
	VariableLength struct {
		ColumnDelimiter  string      `xml:"ColumnDelimiter"`
		RecordDelimiter  string      `xml:"RecordDelimiter"`
		TextEncapsulator string      `xml:"TextEncapsulator"`
		Columns          []xmlColumn `xml:"VariableColumn"`
	} `xml:"VariableLength"`
}

type xmlColumn struct {
	Name         string    `xml:"Name"`
	AlphaNumeric *struct{} `xml:"AlphaNumeric"`
	Numeric      *struct{} `xml:"Numeric"`
}

// pruefeIndexXML parst die index.xml und wendet die DTD-Strukturregeln als eigene
// Prüfungen an (kein Fremd-Parser, keine libxml-Bindung). Rückgabe sind die
// erfolgreich geparsten Tabellen (für die nachgelagerte CSV-Prüfung) und die
// dabei gefundenen Befunde.
//
// Referenz: gdpdu-01-09-2004.dtd (im Archiv beiliegend) für die Element- und
// Reihenfolge-Regeln (DataSet → Version, Media+; Media → Name, Table*; Table →
// URL, Name?, …, VariableLength; VariableLength → VariableColumn+) sowie
// DSFinV-K 2.4 Tz. 1 „Erstellung der index.xml“ für die Formatvorgaben (UTF-8,
// Dezimal-Komma, Semikolon-Spaltentrenner, CRLF-Zeilentrenner, Header in Zeile 1
// ⇒ Range/From = 2).
func pruefeIndexXML(inhalt []byte) ([]indexTabelle, []Befund) {
	if inhalt == nil {
		// Fehlende index.xml meldet bereits pruefePaketpflichtdateien.
		return nil, nil
	}

	var befunde []Befund

	// DTD-Referenz: die index.xml muss die beiliegende DTD im DOCTYPE referenzieren
	// (SYSTEM "gdpdu-01-09-2004.dtd"), sonst ist die Beschreibung nicht an ihre
	// Grammatik gebunden.
	if !bytes.Contains(inhalt, []byte(doctypeMarker)) {
		befunde = append(befunde, Befund{
			Datei:   indexDatei,
			Regel:   regelIndexDoctype,
			Meldung: fmt.Sprintf("DOCTYPE referenziert nicht die beiliegende DTD %q", dtdDatei),
		})
	}

	var ds xmlDataSet
	if err := xml.Unmarshal(inhalt, &ds); err != nil {
		return nil, append(befunde, Befund{
			Datei:   indexDatei,
			Regel:   regelIndexParsbar,
			Meldung: fmt.Sprintf("index.xml ist kein wohlgeformtes XML: %v", err),
		})
	}

	// DTD: <!ELEMENT DataSet (Extension*, Version, DataSupplier?, Command*, Media+ …)>
	if ds.XMLName.Local != "DataSet" {
		befunde = append(befunde, Befund{
			Datei:   indexDatei,
			Regel:   regelIndexWurzel,
			Meldung: fmt.Sprintf("Wurzelelement ist %q, erwartet \"DataSet\"", ds.XMLName.Local),
		})
	}
	if ds.Version == "" {
		befunde = append(befunde, Befund{
			Datei:   indexDatei,
			Regel:   regelIndexVersion,
			Meldung: "Pflichtelement <Version> fehlt oder ist leer",
		})
	}
	if len(ds.Media) == 0 {
		befunde = append(befunde, Befund{
			Datei:   indexDatei,
			Regel:   regelIndexMedia,
			Meldung: "kein <Media>-Element; die DTD verlangt mindestens eines (Media+)",
		})
		return nil, befunde
	}

	var tabellen []indexTabelle
	for _, m := range ds.Media {
		for i := range m.Tables {
			t := &m.Tables[i]
			tab, tabBefunde := pruefeTabelleDeklaration(t)
			befunde = append(befunde, tabBefunde...)
			if tab.URL != "" {
				tabellen = append(tabellen, tab)
			}
		}
	}
	return tabellen, befunde
}

// pruefeTabelleDeklaration prüft eine einzelne <Table>-Deklaration gegen die
// DTD-Pflichtstruktur und die DSFinV-K-Formatvorgaben und liefert die geparste
// Tabelle für die CSV-Prüfung.
func pruefeTabelleDeklaration(t *xmlTable) (indexTabelle, []Befund) {
	var befunde []Befund

	// DTD: <!ELEMENT Table (URL, Name?, …)> — URL ist zwingend.
	if t.URL == "" {
		befunde = append(befunde, Befund{
			Datei:   indexDatei,
			Regel:   regelIndexTabelle,
			Meldung: "eine <Table> ohne Pflichtelement <URL> übersprungen",
		})
		return indexTabelle{}, befunde
	}

	tab := indexTabelle{
		URL:             t.URL,
		Name:            t.Name,
		columnDelimiter: t.VariableLength.ColumnDelimiter,
		recordDelimiter: t.VariableLength.RecordDelimiter,
		decimalSymbol:   t.DecimalSymbol,
		textEncap:       t.VariableLength.TextEncapsulator,
		utf8:            t.UTF8 != nil,
	}
	if t.Range != nil {
		tab.rangeFrom = t.Range.From
	}

	// DTD: <!ELEMENT VariableLength (…, VariableColumn+ …)> — mindestens eine Spalte.
	if len(t.VariableLength.Columns) == 0 {
		befunde = append(befunde, Befund{
			Datei:   indexDatei,
			Regel:   regelIndexSpalte,
			Meldung: fmt.Sprintf("Tabelle %q deklariert keine <VariableColumn> (VariableColumn+ verlangt)", t.URL),
		})
	}
	for _, c := range t.VariableLength.Columns {
		if c.Name == "" {
			befunde = append(befunde, Befund{
				Datei:   indexDatei,
				Regel:   regelIndexSpalte,
				Meldung: fmt.Sprintf("Tabelle %q: eine <VariableColumn> ohne Pflichtelement <Name>", t.URL),
			})
			continue
		}
		// DTD: <!ELEMENT VariableColumn (Name, …, (Numeric | (AlphaNumeric, MaxLength?) | Date) …)>
		// Genau einer der Datentypen muss angegeben sein.
		if (c.Numeric == nil) == (c.AlphaNumeric == nil) {
			befunde = append(befunde, Befund{
				Datei:   indexDatei,
				Regel:   regelIndexSpalte,
				Meldung: fmt.Sprintf("Tabelle %q Spalte %q: genau ein Datentyp (Numeric oder AlphaNumeric) verlangt", t.URL, c.Name),
			})
		}
		tab.Spalten = append(tab.Spalten, indexSpalte{Name: c.Name, Numeric: c.Numeric != nil})
	}

	befunde = append(befunde, pruefeTabelleFormat(tab)...)
	return tab, befunde
}

// pruefeTabelleFormat prüft die DSFinV-K-Formatvorgaben einer Tabellendeklaration:
// UTF-8, Dezimal-Komma, Semikolon-Spaltentrenner, CRLF-Zeilentrenner, Range/From=2
// und Doublequote-Text-Begrenzer.
//
// Referenz: DSFinV-K 2.4 Tz. 1 „Erstellung der index.xml“ und die amtliche index.xml
// (docs/rechtsquellen/…/DSFinV-K-2.4/02_index.xml): DecimalSymbol „,“,
// DigitGroupingSymbol „.“, ColumnDelimiter „;“, RecordDelimiter CRLF (&#xD;&#xA;),
// UTF8, TextEncapsulator „"“, Range/From 2 (Kopfzeile in Zeile 1).
func pruefeTabelleFormat(tab indexTabelle) []Befund {
	var befunde []Befund
	add := func(regel, meldung string) {
		befunde = append(befunde, Befund{Datei: indexDatei, Regel: regel, Meldung: fmt.Sprintf("Tabelle %q: %s", tab.URL, meldung)})
	}

	if !tab.utf8 {
		add(regelIndexFormat, "UTF8-Kodierung nicht deklariert")
	}
	if tab.decimalSymbol != dtdDecimalSymbol {
		add(regelIndexFormat, fmt.Sprintf("DecimalSymbol = %q, erwartet %q (Dezimal-Komma)", tab.decimalSymbol, dtdDecimalSymbol))
	}
	if tab.columnDelimiter != dtdColumnDelimiter {
		add(regelIndexFormat, fmt.Sprintf("ColumnDelimiter = %q, erwartet %q (Semikolon)", tab.columnDelimiter, dtdColumnDelimiter))
	}
	if tab.recordDelimiter != dtdRecordDelimiterCR+dtdRecordDelimiterLF {
		add(regelIndexFormat, "RecordDelimiter ist nicht CRLF")
	}
	if tab.textEncap != dtdTextEncapsulator {
		add(regelIndexFormat, fmt.Sprintf("TextEncapsulator = %q, erwartet %q (Doublequote)", tab.textEncap, dtdTextEncapsulator))
	}
	if tab.rangeFrom != dtdKopfzeileFrom {
		add(regelIndexKopfzeile, fmt.Sprintf("Range/From = %q, erwartet %q (Kopfzeile in Zeile 1, Daten ab Zeile 2)", tab.rangeFrom, dtdKopfzeileFrom))
	}
	return befunde
}
