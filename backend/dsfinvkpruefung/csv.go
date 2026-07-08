package dsfinvkpruefung

import (
	"fmt"
	"strings"
)

// Regel-Kennungen der CSV- und Deklarations-Abgleichprüfung.
const (
	regelIndexDatei     = "index-datei-fehlt"    // index.xml deklariert eine nicht vorhandene CSV
	regelCsvUndeklar    = "csv-nicht-deklariert" // CSV im Archiv ohne index.xml-Deklaration
	regelCsvLeer        = "csv-leer"
	regelCsvCRLF        = "csv-crlf"
	regelCsvKopfzeile   = "csv-kopfzeile"
	regelCsvSpaltenzahl = "csv-spaltenzahl"
	regelCsvDezimal     = "csv-dezimal"
	crlf                = "\r\n"
)

// pruefeTabellenGegenIndex verknüpft die index.xml-Deklaration mit den tatsächlich
// vorhandenen CSV-Dateien: jede deklarierte Tabelle muss als Datei existieren, und
// jede vorhandene CSV muss deklariert sein (die index.xml deklariert nur vorhandene
// Tabellen). Für jede deklarierte-und-vorhandene CSV folgt die Strukturprüfung.
//
// Referenz: DSFinV-K 2.4 Tz. 1 „Erstellung der index.xml“ (die index.xml beschreibt
// den bereitgestellten Datenkranz) und die GoBD-Anlage „Ergänzende Informationen
// zur Datenträgerüberlassung“ (Element URL je Table verweist auf eine vorhandene
// Datei).
func pruefeTabellenGegenIndex(dateien map[string][]byte, tabellen []indexTabelle) []Befund {
	var befunde []Befund

	deklariert := make(map[string]indexTabelle, len(tabellen))
	for _, t := range tabellen {
		deklariert[t.URL] = t
	}

	// (a) Jede deklarierte Tabelle muss als CSV vorhanden sein.
	for _, t := range tabellen {
		inhalt, ok := dateien[t.URL]
		if !ok {
			befunde = append(befunde, Befund{
				Datei:   t.URL,
				Regel:   regelIndexDatei,
				Meldung: "in index.xml deklariert, aber nicht im Archiv vorhanden",
			})
			continue
		}
		befunde = append(befunde, pruefeCSV(t.URL, inhalt, t)...)
	}

	// (b) Jede vorhandene CSV muss in der index.xml deklariert sein.
	for _, name := range sortierteNamen(dateien) {
		if !hatEndung(name, csvEndung) {
			continue
		}
		if _, ok := deklariert[name]; !ok {
			befunde = append(befunde, Befund{
				Datei:   name,
				Regel:   regelCsvUndeklar,
				Meldung: "CSV im Archiv, aber nicht in index.xml deklariert",
			})
		}
	}

	return befunde
}

// pruefeCSV prüft eine einzelne CSV-Datei gegen ihre index.xml-Deklaration:
// CRLF-Zeilenenden, Header-Zeile mit exakt der deklarierten Spaltenreihenfolge,
// gleiche Feldanzahl je Datenzeile und Komma-Dezimaltrennung numerischer Felder.
//
// Referenz: DSFinV-K 2.4 Tz. 1 „Erstellung der index.xml“ und die amtliche index.xml
// (Range/From = 2 ⇒ Header in Zeile 1; ColumnDelimiter „;“; RecordDelimiter CRLF;
// DecimalSymbol „,“). Die Spaltenreihenfolge folgt exakt der Reihenfolge der
// <VariableColumn>-Elemente der jeweiligen Table (Anhänge A–E der DSFinV-K 2.4).
func pruefeCSV(name string, inhalt []byte, tab indexTabelle) []Befund {
	var befunde []Befund
	add := func(regel, meldung string) {
		befunde = append(befunde, Befund{Datei: name, Regel: regel, Meldung: meldung})
	}

	if len(inhalt) == 0 {
		add(regelCsvLeer, "Datei ist leer (mindestens die Kopfzeile wird erwartet)")
		return befunde
	}

	text := string(inhalt)

	// CRLF-Zeilenenden: jedes LF muss von einem CR unmittelbar vorangegangen sein.
	// Ein einzelnes LF (Unix-Zeilenende) verletzt die Vorgabe.
	if verletztCRLF(text) {
		add(regelCsvCRLF, "Zeilenenden sind nicht durchgängig CRLF (\\r\\n)")
	}

	zeilen := zerlegeCRLF(text)
	if len(zeilen) == 0 {
		add(regelCsvKopfzeile, "keine Kopfzeile vorhanden")
		return befunde
	}

	// Kopfzeile: exakte Spaltennamen in exakter Reihenfolge.
	header := splitFelder(zeilen[0])
	erwartet := spaltenNamen(tab)
	if !gleicheReihenfolge(header, erwartet) {
		add(regelCsvKopfzeile, fmt.Sprintf(
			"Kopfzeile weicht von der index.xml-Deklaration ab\n  erwartet: %s\n  gefunden: %s",
			strings.Join(erwartet, ";"), strings.Join(header, ";")))
		// Ohne passende Kopfzeile sind Spalten-bezogene Datenprüfungen nicht sinnvoll.
		return befunde
	}

	// Datenzeilen: gleiche Feldanzahl und Komma-Dezimalformat je numerischer Spalte.
	for i := 1; i < len(zeilen); i++ {
		felder := splitFelder(zeilen[i])
		if len(felder) != len(erwartet) {
			add(regelCsvSpaltenzahl, fmt.Sprintf("Zeile %d hat %d Felder, erwartet %d", i+1, len(felder), len(erwartet)))
			continue
		}
		for si, sp := range tab.Spalten {
			if sp.Numeric && verletztDezimalKomma(felder[si]) {
				add(regelCsvDezimal, fmt.Sprintf("Zeile %d, Spalte %q: numerisches Feld %q nutzt einen Punkt statt Komma als Dezimaltrenner", i+1, sp.Name, felder[si]))
			}
		}
	}

	return befunde
}

// spaltenNamen liefert die deklarierten Spaltennamen einer Tabelle in Reihenfolge.
func spaltenNamen(tab indexTabelle) []string {
	namen := make([]string, len(tab.Spalten))
	for i, s := range tab.Spalten {
		namen[i] = s.Name
	}
	return namen
}

// gleicheReihenfolge meldet, ob zwei String-Slices elementweise identisch sind.
func gleicheReihenfolge(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// verletztCRLF meldet, ob im Text ein LF ohne unmittelbar vorangehendes CR steht
// (ein reines Unix-Zeilenende).
func verletztCRLF(text string) bool {
	for i := 0; i < len(text); i++ {
		if text[i] == '\n' {
			if i == 0 || text[i-1] != '\r' {
				return true
			}
		}
	}
	return false
}

// zerlegeCRLF zerlegt den Text an CRLF-Grenzen in Zeilen und verwirft eine leere
// Schlusszeile nach dem finalen CRLF (jede CSV-Zeile endet auf CRLF).
func zerlegeCRLF(text string) []string {
	zeilen := strings.Split(text, crlf)
	if n := len(zeilen); n > 0 && zeilen[n-1] == "" {
		zeilen = zeilen[:n-1]
	}
	return zeilen
}

// splitFelder zerlegt eine CSV-Zeile am Semikolon unter Beachtung des Doublequote-
// Text-Begrenzers: ein Semikolon innerhalb von Anführungszeichen trennt nicht, ein
// verdoppeltes Anführungszeichen ("") ist ein literales Zeichen im Feld.
func splitFelder(zeile string) []string {
	var felder []string
	var b strings.Builder
	inQuotes := false
	for i := 0; i < len(zeile); i++ {
		c := zeile[i]
		switch {
		case c == '"':
			if inQuotes && i+1 < len(zeile) && zeile[i+1] == '"' {
				b.WriteByte('"')
				i++
				continue
			}
			inQuotes = !inQuotes
		case c == ';' && !inQuotes:
			felder = append(felder, b.String())
			b.Reset()
		default:
			b.WriteByte(c)
		}
	}
	felder = append(felder, b.String())
	return felder
}

// verletztDezimalKomma meldet, ob ein numerisches Feld einen Punkt als Dezimal-
// oder Gruppierungstrenner nutzt. DSFinV-K führt Zahlen mit Komma-Dezimaltrenner
// und ohne Tausenderpunkt; ein Punkt im numerischen Feld ist daher ein Verstoß.
// Ein leeres Feld ist zulässig (optionale numerische Angabe).
func verletztDezimalKomma(feld string) bool {
	return strings.ContainsRune(feld, '.')
}
