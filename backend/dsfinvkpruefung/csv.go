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

// pruefeTabellenGegenIndex matches index.xml declarations and archive CSVs in both directions.
// See docs/compliance.md §6.2.
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

// pruefeCSV checks a CSV against its index.xml declaration: header in line 1, ';', CRLF,
// decimal comma, columns in <VariableColumn> order. See docs/compliance.md §6.2.
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
		// Without a matching header, per-column checks are meaningless.
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

// zerlegeCRLF drops the empty tail after the final CRLF, since every CSV line ends in CRLF.
func zerlegeCRLF(text string) []string {
	zeilen := strings.Split(text, crlf)
	if n := len(zeilen); n > 0 && zeilen[n-1] == "" {
		zeilen = zeilen[:n-1]
	}
	return zeilen
}

// splitFelder splits on ';' with '"' as text encapsulator: a quoted ';' does not split, "" is a literal quote.
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

// verletztDezimalKomma: DSFinV-K numbers use a decimal comma and no thousands separator, so any '.' violates.
func verletztDezimalKomma(feld string) bool {
	return strings.ContainsRune(feld, '.')
}
