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

// pruefePaketpflichtdateien: index.xml und die referenzierte GDPdU-DTD sind zwingend.
//
// Referenz: DSFinV-K 2.4 Tz. 1 „Erstellung der index.xml“ und die GoBD-Anlage
// „Ergänzende Informationen zur Datenträgerüberlassung“ (Beschreibungsstandard).
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

// pruefeDateinamen: DSFinV-K-CSVs sind englisch, kleingeschrieben und liegen flach im
// Wurzelverzeichnis; außer index.xml, DTD und *.csv gehört nichts ins Archiv.
//
// Referenz: DSFinV-K 2.4 Anhänge A–E und die Dateiübersicht (Tz. 6, csv-Dateinamen wie
// cashpointclosing.csv, transactions.csv …), durchgängig kleingeschrieben und englisch.
// Die GDPdU-URL-Regel lässt nur relative Namen zu — hier als flache Wurzeldatei geprüft.
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
			// Pflicht-Beschreibungsdateien: erlaubt.
		case strings.HasSuffix(name, csvEndung):
			// Groß-/Kleinschreibung wird bewusst nicht normalisiert: eine
			// ".CSV"-Datei ist bereits ein Dateinamensverstoß.
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

// istKleingeschrieben prüft auf ASCII-Großbuchstaben; die amtlichen Dateinamen bestehen
// nur aus Kleinbuchstaben, Ziffern, Unterstrich und dem Punkt der Endung.
func istKleingeschrieben(name string) bool {
	for i := 0; i < len(name); i++ {
		if name[i] >= 'A' && name[i] <= 'Z' {
			return false
		}
	}
	return true
}
