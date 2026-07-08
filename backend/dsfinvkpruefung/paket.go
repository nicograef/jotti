package dsfinvkpruefung

import "fmt"

// Regel-Kennungen der Paket- und Dateinamensprüfung.
const (
	regelDateiname       = "dateiname"
	regelPaketpflicht    = "paket-pflichtdatei"
	indexDatei           = "index.xml"
	dtdDatei             = "gdpdu-01-09-2004.dtd"
	csvEndung            = ".csv"
	regelDateinamePfad   = "dateiname-pfad"
	regelDateinameGrafik = "dateiname-fremdformat"
)

// pruefePaketpflichtdateien stellt sicher, dass die beiden zwingenden
// Beschreibungsdateien des Datenträgers vorhanden sind: die beschreibende
// index.xml und die referenzierte GDPdU-DTD.
//
// Referenz: DSFinV-K 2.4 Tz. 1 „Erstellung der index.xml“ sowie die GoBD-Anlage
// „Ergänzende Informationen zur Datenträgerüberlassung“ (Beschreibungsstandard):
// der Datenträger enthält eine index.xml, die die zugehörige DTD referenziert.
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

// pruefeDateinamen prüft die Dateinamensregeln des Datenträgers:
//   - Die CSV-Dateinamen der DSFinV-K sind englisch und kleingeschrieben und liegen
//     flach im Wurzelverzeichnis (kein Pfad-Anteil).
//   - Es dürfen keine unerwarteten Fremdformate enthalten sein (nur index.xml, die
//     DTD und *.csv).
//
// Referenz: DSFinV-K 2.4 Anhänge A–E und die Dateiübersicht (Tz. 6, Auflistung der
// csv-Dateinamen wie cashpointclosing.csv, transactions.csv …); die amtlichen
// Dateinamen sind durchgängig kleingeschrieben und einheitlich englisch. Die GDPdU-
// URL-Regel (gdpdu-01-09-2004.dtd, Element URL) lässt nur relative Namen zu — hier
// als flache Wurzeldatei geprüft.
func pruefeDateinamen(dateien map[string][]byte) []Befund {
	var befunde []Befund
	for _, name := range sortierteNamen(dateien) {
		if istPfad(name) {
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
		case hatEndung(name, csvEndung):
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
				Regel:   regelDateinameGrafik,
				Meldung: "unerwartete Datei im DSFinV-K-Archiv (erlaubt sind index.xml, die DTD und *.csv)",
			})
		}
	}
	return befunde
}

// istPfad meldet, ob der Name einen Verzeichnisanteil trägt.
func istPfad(name string) bool {
	for i := 0; i < len(name); i++ {
		if name[i] == '/' || name[i] == '\\' {
			return true
		}
	}
	return false
}

// hatEndung meldet, ob name auf endung endet (Groß-/Kleinschreibung wird bewusst
// nicht normalisiert: eine ".CSV"-Datei ist bereits ein Dateinamensverstoß).
func hatEndung(name, endung string) bool {
	if len(name) < len(endung) {
		return false
	}
	return name[len(name)-len(endung):] == endung
}

// istKleingeschrieben meldet, ob der Name keine Großbuchstaben enthält (ASCII).
// Die amtlichen Dateinamen bestehen ausschließlich aus Kleinbuchstaben, Ziffern,
// Unterstrich und dem Punkt der Endung.
func istKleingeschrieben(name string) bool {
	for i := 0; i < len(name); i++ {
		if name[i] >= 'A' && name[i] <= 'Z' {
			return false
		}
	}
	return true
}
