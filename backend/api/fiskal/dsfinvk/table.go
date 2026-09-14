package dsfinvk

import "strings"

// DSFinV-K-CSV-Formatregeln: Semikolon als Trennzeichen, CRLF als Zeilenende,
// UTF-8, Doublequote als Text-Begrenzer.
const (
	csvSeparator        = ";"
	csvNewline          = "\r\n"
	csvTextEncapsulator = `"`
)

// Table ist eine serialisierbare DSFinV-K-CSV-Datei. Feldtyp und
// Nachkommastellen fehlen bewusst: das Archiv liefert die amtliche index.xml
// unverändert mit (amtlicheIndexXML) und erzeugt keine eigene Felddeklaration.
type Table struct {
	File        string
	LogicalName string
	Description string
	Columns     []string
	Records     [][]string
}

// serializeCSV rendert die Tabelle als DSFinV-K-CSV: eine Header-Zeile mit den
// Spaltennamen, dann je Datensatz eine Zeile.
func serializeCSV(t Table) []byte {
	var b strings.Builder
	writeCSVRow(&b, t.Columns)
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

func escapeCSVField(field string) string {
	if !strings.ContainsAny(field, ";\"\r\n") {
		return field
	}
	return csvTextEncapsulator + strings.ReplaceAll(field, `"`, `""`) + csvTextEncapsulator
}
