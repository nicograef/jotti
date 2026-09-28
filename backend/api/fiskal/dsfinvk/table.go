package dsfinvk

import "strings"

// DSFinV-K CSV format rules: semicolon separator, CRLF line end, UTF-8, double quote
// as text delimiter.
const (
	csvSeparator        = ";"
	csvNewline          = "\r\n"
	csvTextEncapsulator = `"`
)

// Table is one DSFinV-K CSV file. It carries no field types or decimals because the
// archive ships the official index.xml unchanged.
type Table struct {
	File        string
	LogicalName string
	Description string
	Columns     []string
	Records     [][]string
}

// serializeCSV renders the table as DSFinV-K CSV: a header row, then one row per record.
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
