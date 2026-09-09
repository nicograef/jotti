// Package zeit hält die Zeitzone, in der jotti Zeitpunkte anzeigt, druckt und
// benennt. Ein Zeitpunkt liegt als UTC in der Datenbank; Beleg, Arbeitsbon,
// Archivname und Meldedatum tragen deutsche Ortszeit, weil Gast, Betreiber und
// Prüfung sie am Wandkalender lesen.
package zeit

import "time"

// Berlin ist die Zone dieser Umrechnung, einmal beim Start geladen. tzdata ist
// ins Binary eingebettet (backend/main.go), das Laden schlägt nur bei kaputtem
// Build fehl.
var Berlin = mustLoad("Europe/Berlin")

func mustLoad(name string) *time.Location {
	ort, err := time.LoadLocation(name)
	if err != nil {
		panic("zeit: Zeitzone " + name + " nicht ladbar: " + err.Error())
	}
	return ort
}
