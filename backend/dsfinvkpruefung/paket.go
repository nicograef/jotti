package dsfinvkpruefung

import (
	"fmt"
	"strings"
)

const (
	regelDateiname            = "dateiname"
	regelPaketpflicht         = "paket-pflichtdatei"
	indexDatei                = "index.xml"
	dtdDatei                  = "gdpdu-01-09-2004.dtd"
	csvEndung                 = ".csv"
	regelDateinamePfad        = "dateiname-pfad"
	regelDateinameFremdformat = "dateiname-fremdformat"
)

// pruefePaketpflichtdateien requires index.xml and the GDPdU DTD it references; see docs/compliance.md §6.2.
func pruefePaketpflichtdateien(dateien map[string][]byte) []Befund {
	var befunde []Befund
	if _, ok := dateien[indexDatei]; !ok {
		befunde = append(befunde, Befund{
			Regel:   regelPaketpflicht,
			Meldung: fmt.Sprintf("Pflichtdatei %q fehlt im Archiv", indexDatei),
		})
	}
	if _, ok := dateien[dtdDatei]; !ok {
		befunde = append(befunde, Befund{
			Regel:   regelPaketpflicht,
			Meldung: fmt.Sprintf("Pflichtdatei %q fehlt im Archiv", dtdDatei),
		})
	}
	return befunde
}

// pruefeDateinamen allows only index.xml, the DTD and lowercase *.csv files, flat in the archive root.
// See docs/compliance.md §6.2.
func pruefeDateinamen(dateien map[string][]byte) []Befund {
	var befunde []Befund
	for _, name := range sortierteNamen(dateien) {
		if strings.ContainsAny(name, `/\`) {
			befunde = append(befunde, Befund{
				Datei:   name,
				Regel:   regelDateinamePfad,
				Meldung: "Dateiname enthält einen Verzeichnispfad; DSFinV-K-Dateien liegen flach im Wurzelverzeichnis",
			})
			continue
		}
		switch {
		case name == indexDatei, name == dtdDatei:
			// Mandatory description files.
		case strings.HasSuffix(name, csvEndung):
			// Case is deliberately not normalised: a ".CSV" file already violates the naming rule.
			if !istKleingeschrieben(name) {
				befunde = append(befunde, Befund{
					Datei:   name,
					Regel:   regelDateiname,
					Meldung: "CSV-Dateiname muss englisch und vollständig kleingeschrieben sein",
				})
			}
		default:
			befunde = append(befunde, Befund{
				Datei:   name,
				Regel:   regelDateinameFremdformat,
				Meldung: "unerwartete Datei im DSFinV-K-Archiv (erlaubt sind index.xml, die DTD und *.csv)",
			})
		}
	}
	return befunde
}

// istKleingeschrieben checks ASCII only: official names use just lowercase letters, digits, '_' and '.'.
func istKleingeschrieben(name string) bool {
	for i := 0; i < len(name); i++ {
		if name[i] >= 'A' && name[i] <= 'Z' {
			return false
		}
	}
	return true
}
