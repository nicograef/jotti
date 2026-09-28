package dsfinvk

import (
	"strings"
	"testing"
)

// FuzzSerializeCSV guards that no field content breaks the column layout of an official
// DSFinV-K file: no panic, header-width rows, lossless round trip. It uses its own parser
// because encoding/csv normalises CR/LF inside quoted fields.
func FuzzSerializeCSV(f *testing.F) {
	// Seeds: the real case from table_test.go plus special-character edges.
	f.Add("plain", "5.00", "ok")
	f.Add("semi;colon", "1.50", `inner"quote`)
	f.Add("line\nbreak", "0.00", "ende")
	f.Add("carriage\rreturn", `"leading-quote`, "\t\x00")
	f.Add("Ümläüte €", "-12,34", "")

	f.Fuzz(func(t *testing.T, a, b, c string) {
		cols := []string{"A", "B", "C"}
		table := Table{Columns: cols, Records: [][]string{{a, b, c}}}
		out := string(serializeCSV(table))

		// The encoder ends every row with CRLF, so the final CRLF opens no empty row.
		if !strings.HasSuffix(out, csvNewline) {
			t.Errorf("Ausgabe endet nicht mit CRLF: %q", out)
		}
		lines := splitCSVRows(strings.TrimSuffix(out, csvNewline))
		if len(lines) != 2 {
			t.Fatalf("erwartet 2 Zeilen (Header + 1 Record), bekam %d: %q", len(lines), out)
		}

		header := parseCSVRow(lines[0])
		record := parseCSVRow(lines[1])
		if len(header) != len(cols) {
			t.Errorf("Header-Feldanzahl %d != Spaltenanzahl %d: %q", len(header), len(cols), out)
		}
		if len(record) != len(cols) {
			t.Fatalf("Record-Feldanzahl %d != Spaltenanzahl %d (rohes Trennzeichen zerbrochen?): %q", len(record), len(cols), out)
		}

		want := []string{a, b, c}
		for i, got := range record {
			if got != want[i] {
				t.Errorf("Feld %d verändert: got %q, want %q\noutput=%q", i, got, want[i], out)
			}
		}
	})
}

// splitCSVRows splits at CRLF outside double-quoted fields, mirroring escapeCSVField.
// Byte-wise is safe: ; " CR LF are ASCII and never occur in a UTF-8 continuation byte.
func splitCSVRows(s string) []string {
	var rows []string
	var cur strings.Builder
	inQuotes := false
	for i := 0; i < len(s); i++ {
		ch := s[i]
		switch {
		case ch == '"':
			inQuotes = !inQuotes
			cur.WriteByte(ch)
		case ch == '\r' && !inQuotes && i+1 < len(s) && s[i+1] == '\n':
			rows = append(rows, cur.String())
			cur.Reset()
			i++ // skip the following \n
		default:
			cur.WriteByte(ch)
		}
	}
	rows = append(rows, cur.String())
	return rows
}

// parseCSVRow splits a row by escapeCSVField's rules: ";" separates, quotes wrap a
// field, "" is a literal quote. Byte-wise, see splitCSVRows.
func parseCSVRow(line string) []string {
	var fields []string
	var cur strings.Builder
	inQuotes := false
	for i := 0; i < len(line); i++ {
		ch := line[i]
		switch {
		case ch == '"' && inQuotes && i+1 < len(line) && line[i+1] == '"':
			cur.WriteByte('"')
			i++ // skip the pair's second quote
		case ch == '"':
			inQuotes = !inQuotes
		case ch == ';' && !inQuotes:
			fields = append(fields, cur.String())
			cur.Reset()
		default:
			cur.WriteByte(ch)
		}
	}
	fields = append(fields, cur.String())
	return fields
}
