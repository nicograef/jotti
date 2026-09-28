package dsfinvkpruefung

import (
	"bytes"
	"encoding/xml"
	"fmt"
)

const (
	regelIndexParsbar   = "index-parsbar"
	regelIndexWurzel    = "index-wurzel"
	regelIndexVersion   = "index-version"
	regelIndexMedia     = "index-media"
	regelIndexTabelle   = "index-tabelle"
	regelIndexSpalte    = "index-spalte"
	regelIndexFormat    = "index-format"
	regelIndexKopfzeile = "index-kopfzeile-range"
	regelIndexDoctype   = "index-doctype"
	dtdDecimalSymbol    = ","
	dtdColumnDelimiter  = ";"
	dtdRecordDelimiter  = crlf
	dtdKopfzeileFrom    = "2"
	dtdTextEncapsulator = "\""
	doctypeMarker       = dtdDatei
)

type indexSpalte struct {
	Name    string
	Numeric bool // Numeric columns must use the decimal comma in the CSV
}

type indexTabelle struct {
	URL             string
	Spalten         []indexSpalte
	columnDelimiter string
	recordDelimiter string
	decimalSymbol   string
	rangeFrom       string
	textEncap       string
	utf8            bool
}

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

// pruefeIndexXML checks index.xml against gdpdu-01-09-2004.dtd and the DSFinV-K 2.4 format rules.
func pruefeIndexXML(inhalt []byte) ([]indexTabelle, []Befund) {
	if inhalt == nil {
		// pruefePaketpflichtdateien reports the missing index.xml.
		return nil, nil
	}

	var befunde []Befund

	// Without a DOCTYPE naming the bundled DTD, the description is not bound to its grammar.
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

func pruefeTabelleDeklaration(t *xmlTable) (indexTabelle, []Befund) {
	var befunde []Befund

	// DTD: <!ELEMENT Table (URL, Name?, …)>, so URL is mandatory.
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
		columnDelimiter: t.VariableLength.ColumnDelimiter,
		recordDelimiter: t.VariableLength.RecordDelimiter,
		decimalSymbol:   t.DecimalSymbol,
		textEncap:       t.VariableLength.TextEncapsulator,
		utf8:            t.UTF8 != nil,
	}
	if t.Range != nil {
		tab.rangeFrom = t.Range.From
	}

	// DTD: <!ELEMENT VariableLength (…, VariableColumn+ …)>, so at least one column.
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
		// Exactly one data type is required.
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

// pruefeTabelleFormat checks a table declaration against the official index.xml
// (docs/rechtsquellen/technik-spezifikationen/DSFinV-K-2.4/02_index.xml); see docs/compliance.md §6.2.
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
	if tab.recordDelimiter != dtdRecordDelimiter {
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
