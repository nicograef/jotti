package dsfinvk

import "strings"

// dataSupplier sind die DataSupplier-Angaben der index.xml (Herkunft der Daten).
type dataSupplier struct {
	name     string
	location string
	comment  string
}

// buildIndexXML erzeugt den GDPdU-Descriptor (index.xml) nach der
// gdpdu-01-09-2004.dtd. Er deklariert genau die übergebenen Tabellen mit ihren
// Spalten, Feldtypen, Trennzeichen und überspringt je Tabelle die Header-Zeile
// (Range/From = 2). Manuell gebaut, um Elementreihenfolge, DOCTYPE und leere
// Typ-Elemente exakt zu steuern.
func buildIndexXML(supplier dataSupplier, tables []Table) []byte {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	b.WriteString(`<!DOCTYPE DataSet SYSTEM "` + DTDFilename + `">` + "\n")
	b.WriteString("<DataSet>\n")
	b.WriteString("  <Version>" + xmlEscape(Version) + "</Version>\n")
	b.WriteString("  <DataSupplier>\n")
	b.WriteString("    <Name>" + xmlEscape(supplier.name) + "</Name>\n")
	b.WriteString("    <Location>" + xmlEscape(supplier.location) + "</Location>\n")
	b.WriteString("    <Comment>" + xmlEscape(supplier.comment) + "</Comment>\n")
	b.WriteString("  </DataSupplier>\n")
	b.WriteString("  <Media>\n")
	b.WriteString("    <Name>DSFinV-K</Name>\n")
	for _, t := range tables {
		writeTableXML(&b, t)
	}
	b.WriteString("  </Media>\n")
	b.WriteString("</DataSet>\n")
	return []byte(b.String())
}

func writeTableXML(b *strings.Builder, t Table) {
	b.WriteString("    <Table>\n")
	b.WriteString("      <URL>" + xmlEscape(t.File) + "</URL>\n")
	b.WriteString("      <Name>" + xmlEscape(t.LogicalName) + "</Name>\n")
	if t.Description != "" {
		b.WriteString("      <Description>" + xmlEscape(t.Description) + "</Description>\n")
	}
	b.WriteString("      <UTF8/>\n")
	b.WriteString("      <DecimalSymbol>.</DecimalSymbol>\n")
	b.WriteString("      <DigitGroupingSymbol>,</DigitGroupingSymbol>\n")
	b.WriteString("      <Range><From>2</From></Range>\n")
	b.WriteString("      <VariableLength>\n")
	b.WriteString("        <ColumnDelimiter>;</ColumnDelimiter>\n")
	b.WriteString("        <RecordDelimiter>&#13;&#10;</RecordDelimiter>\n")
	b.WriteString("        <TextEncapsulator>\"</TextEncapsulator>\n")
	for _, c := range t.Columns {
		writeColumnXML(b, c)
	}
	b.WriteString("      </VariableLength>\n")
	b.WriteString("    </Table>\n")
}

func writeColumnXML(b *strings.Builder, c column) {
	b.WriteString("        <VariableColumn>\n")
	b.WriteString("          <Name>" + xmlEscape(c.name) + "</Name>\n")
	switch c.typ {
	case numeric:
		b.WriteString("          <Numeric><Accuracy>" + itoa(c.accuracy) + "</Accuracy></Numeric>\n")
	default:
		b.WriteString("          <AlphaNumeric/>\n")
	}
	b.WriteString("        </VariableColumn>\n")
}

var xmlReplacer = strings.NewReplacer(
	"&", "&amp;",
	"<", "&lt;",
	">", "&gt;",
	`"`, "&quot;",
	"'", "&apos;",
)

func xmlEscape(s string) string { return xmlReplacer.Replace(s) }
