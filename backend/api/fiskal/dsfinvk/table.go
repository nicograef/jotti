package dsfinvk

import "strings"

// DSFinV-K-CSV-Formatregeln: Semikolon als Trennzeichen, CRLF als Zeilenende,
// UTF-8, Doublequote als Text-Begrenzer.
const (
	csvSeparator        = ";"
	csvNewline          = "\r\n"
	csvTextEncapsulator = `"`
)

// column beschreibt ein CSV-Feld über seinen Namen, die Spaltenüberschrift.
// Feldtyp und Nachkommastellen stehen nicht hier: das Archiv liefert die
// eingebettete amtliche index.xml unverändert mit (amtlicheIndexXML) und
// erzeugt keine eigene Felddeklaration.
type column struct {
	name string
}

func col(name string) column { return column{name: name} }

// Table ist eine serialisierbare DSFinV-K-CSV-Datei: offizieller Dateiname,
// logische (deutsche) Bezeichnung, Spaltenbeschreibung und die bereits als
// Strings formatierten Datenzeilen.
type Table struct {
	File        string
	LogicalName string
	Description string
	Columns     []column
	Records     [][]string
}

// header liefert die Spaltennamen in Reihenfolge.
func (t Table) header() []string {
	names := make([]string, len(t.Columns))
	for i, c := range t.Columns {
		names[i] = c.name
	}
	return names
}

// serializeCSV rendert eine Tabelle als DSFinV-K-konforme CSV-Bytes: eine
// Header-Zeile mit den Spaltennamen, dann je Datensatz eine Zeile, Felder per
// Semikolon getrennt, Zeilen per CRLF abgeschlossen, UTF-8. Felder mit
// Trennzeichen, Anführungszeichen oder Zeilenumbruch werden in Doublequotes
// gefasst (innere `"` verdoppelt).
func serializeCSV(t Table) []byte {
	var b strings.Builder
	writeCSVRow(&b, t.header())
	for _, record := range t.Records {
		writeCSVRow(&b, record)
	}
	return []byte(b.String())
}

func writeCSVRow(b *strings.Builder, fields []string) {
	for i, field := range fields {
		if i > 0 {
			b.WriteString(csvSeparator)
		}
		b.WriteString(escapeCSVField(field))
	}
	b.WriteString(csvNewline)
}

// escapeCSVField fasst ein Feld nur dann in Doublequotes, wenn es ein
// Sonderzeichen (Trennzeichen, Anführungszeichen, CR, LF) enthält.
func escapeCSVField(field string) string {
	if !strings.ContainsAny(field, ";\"\r\n") {
		return field
	}
	return csvTextEncapsulator + strings.ReplaceAll(field, `"`, `""`) + csvTextEncapsulator
}
