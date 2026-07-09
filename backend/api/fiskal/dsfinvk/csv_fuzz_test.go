//go:build unit

package dsfinvk

import (
	"strings"
	"testing"
)

// FuzzSerializeCSV prüft den DSFinV-K-CSV-Encoder gegen beliebige Feldinhalte
// (Semikolon, Anführungszeichen, CR, LF, Unicode, Steuerzeichen). Zu haltende
// CSV-Invarianten:
//   - kein Panic;
//   - jede erzeugte Zeile hat exakt so viele Felder wie der Header (kein Feld
//     zerbricht durch ein rohes Trennzeichen in mehrere Spalten);
//   - jedes Feld übersteht einen Roundtrip durch einen RFC-4180-Parser (des hier
//     verwendeten Semikolon-Dialekts) unverändert (korrektes Quoting/Escaping).
//
// Ein defekter Encoder würde die Spaltenzuordnung einer amtlichen DSFinV-K-Datei
// zerstören und den gesamten Export bei der Kassennachschau unbrauchbar machen.
//
// Der Roundtrip nutzt einen eigenen, deterministischen Parser statt encoding/csv:
// Go's csv-Reader normalisiert CR/LF innerhalb von Feldern eigenmächtig, was den
// Encoder nicht betrifft, aber den Wertvergleich verfälschen würde.
func FuzzSerializeCSV(f *testing.F) {
	// Seeds aus dem echten Testfall in table_test.go plus Sonderzeichen-Kanten.
	f.Add("plain", "5.00", "ok")
	f.Add("semi;colon", "1.50", `inner"quote`)
	f.Add("line\nbreak", "0.00", "ende")
	f.Add("carriage\rreturn", `"leading-quote`, "\t\x00")
	f.Add("Ümläüte €", "-12,34", "")

	f.Fuzz(func(t *testing.T, a, b, c string) {
		cols := []column{alpha("A"), num("B", 2), alpha("C")}
		table := Table{Columns: cols, Records: [][]string{{a, b, c}}}
		out := string(serializeCSV(table))

		// Der Encoder terminiert jede Zeile mit CRLF. Die letzte Zeile endet daher
		// mit einem abschließenden CRLF, das keine leere Zeile einleitet.
		if !strings.HasSuffix(out, csvNewline) {
			t.Fatalf("Ausgabe endet nicht mit CRLF: %q", out)
		}
		lines := splitCSVRows(strings.TrimSuffix(out, csvNewline))
		if len(lines) != 2 {
			t.Fatalf("erwartet 2 Zeilen (Header + 1 Record), bekam %d: %q", len(lines), out)
		}

		header := parseCSVRow(lines[0])
		record := parseCSVRow(lines[1])
		if len(header) != len(cols) {
			t.Fatalf("Header-Feldanzahl %d != Spaltenanzahl %d: %q", len(header), len(cols), out)
		}
		if len(record) != len(cols) {
			t.Fatalf("Record-Feldanzahl %d != Spaltenanzahl %d (rohes Trennzeichen zerbrochen?): %q", len(record), len(cols), out)
		}

		want := []string{a, b, c}
		for i, got := range record {
			if got != want[i] {
				t.Fatalf("Feld %d verändert: got %q, want %q\noutput=%q", i, got, want[i], out)
			}
		}
	})
}

// splitCSVRows teilt einen CSV-Text in Zeilen, wobei CRLF innerhalb gequoteter
// Felder (Doublequote-Paare) keine Zeilengrenze bildet — genau die Regel, nach
// der escapeCSVField Felder mit Zeilenumbruch in Doublequotes fasst. Arbeitet
// byteweise: Semikolon, Doublequote, CR und LF sind ASCII und tauchen nie in
// einem UTF-8-Folgebyte auf, daher ist das für beliebige Bytes korrekt.
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
			i++ // das folgende \n überspringen
		default:
			cur.WriteByte(ch)
		}
	}
	rows = append(rows, cur.String())
	return rows
}

// parseCSVRow zerlegt eine Zeile in Felder nach den Regeln von escapeCSVField:
// Semikolon trennt, ein gequotetes Feld beginnt/endet mit Doublequote, ein
// verdoppeltes Doublequote ("") steht für ein wörtliches Anführungszeichen.
// Byteweise (siehe splitCSVRows).
func parseCSVRow(line string) []string {
	var fields []string
	var cur strings.Builder
	inQuotes := false
	for i := 0; i < len(line); i++ {
		ch := line[i]
		switch {
		case ch == '"' && inQuotes && i+1 < len(line) && line[i+1] == '"':
			cur.WriteByte('"')
			i++ // das zweite Anführungszeichen des Paares überspringen
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
