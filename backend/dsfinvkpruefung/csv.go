package dsfinvkpruefung

import (
	"fmt"
	"slices"
	"strings"
)

const (
	regelIndexDatei     = "index-datei-fehlt"
	regelCsvUndeklar    = "csv-nicht-deklariert"
	regelCsvLeer        = "csv-leer"
	regelCsvCRLF        = "csv-crlf"
	regelCsvKopfzeile   = "csv-kopfzeile"
	regelCsvSpaltenzahl = "csv-spaltenzahl"
	regelCsvDezimal     = "csv-dezimal"
	crlf                = "\r\n"
)

// pruefeTabellenGegenIndex gleicht index.xml-Deklaration und vorhandene CSV-Dateien in
// beide Richtungen ab.
//
// Referenz: DSFinV-K 2.4 Tz. 1 „Erstellung der index.xml“ (die index.xml beschreibt den
// bereitgestellten Datenkranz — eine undeklarierte CSV ist deshalb ein Verstoß) und die
// GoBD-Anlage „Ergänzende Informationen zur Datenträgerüberlassung“ (Element URL je Table
// verweist auf eine vorhandene Datei).
func pruefeTabellenGegenIndex(dateien map[string][]byte, tabellen []indexTabelle) []Befund {
	var befunde []Befund

	deklariert := make(map[string]indexTabelle, len(tabellen))
	for _, t := range tabellen {
		deklariert[t.URL] = t
	}

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

	for _, name := range sortierteNamen(dateien) {
		if !strings.HasSuffix(name, csvEndung) {
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

// pruefeCSV prüft eine CSV gegen ihre index.xml-Deklaration.
//
// Referenz: DSFinV-K 2.4 Tz. 1 und die amtliche index.xml (Range/From = 2 ⇒ Header in
// Zeile 1; ColumnDelimiter „;“; RecordDelimiter CRLF; DecimalSymbol „,“). Die
// Spaltenreihenfolge folgt exakt der Reihenfolge der <VariableColumn>-Elemente der
// jeweiligen Table (Anhänge A–E der DSFinV-K 2.4).
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

	if verletztCRLF(text) {
		add(regelCsvCRLF, "Zeilenenden sind nicht durchgängig CRLF (\\r\\n)")
	}

	zeilen := zerlegeCRLF(text)
	if len(zeilen) == 0 {
		add(regelCsvKopfzeile, "keine Kopfzeile vorhanden")
		return befunde
	}

	header := splitFelder(zeilen[0])
	erwartet := spaltenNamen(tab)
	if !slices.Equal(header, erwartet) {
		add(regelCsvKopfzeile, fmt.Sprintf(
			"Kopfzeile weicht von der index.xml-Deklaration ab\n  erwartet: %s\n  gefunden: %s",
			strings.Join(erwartet, ";"), strings.Join(header, ";")))
		// Ohne passende Kopfzeile sind Spalten-bezogene Datenprüfungen nicht sinnvoll.
		return befunde
	}

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

func spaltenNamen(tab indexTabelle) []string {
	namen := make([]string, len(tab.Spalten))
	for i, s := range tab.Spalten {
		namen[i] = s.Name
	}
	return namen
}

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

// zerlegeCRLF verwirft die leere Schlusszeile nach dem finalen CRLF (jede CSV-Zeile endet auf CRLF).
func zerlegeCRLF(text string) []string {
	zeilen := strings.Split(text, crlf)
	if n := len(zeilen); n > 0 && zeilen[n-1] == "" {
		zeilen = zeilen[:n-1]
	}
	return zeilen
}

// splitFelder zerlegt am Semikolon mit Doublequote-Textbegrenzer: ein Semikolon in
// Anführungszeichen trennt nicht, "" ist ein literales Anführungszeichen im Feld.
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

// verletztDezimalKomma: DSFinV-K führt Zahlen mit Komma-Dezimaltrenner und ohne
// Tausenderpunkt; ein Punkt im numerischen Feld ist daher ein Verstoß. Ein leeres Feld
// ist zulässig.
func verletztDezimalKomma(feld string) bool {
	return strings.ContainsRune(feld, '.')
}
