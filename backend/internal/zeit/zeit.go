// Package zeit hält die Zeitzone, in der jotti Zeitpunkte anzeigt, druckt und
// benennt: Die Datenbank speichert UTC, Beleg, Arbeitsbon, Archivname und
// Meldedatum tragen deutsche Ortszeit.
package zeit

import (
	"time"
	// Die eingebettete Zonendatenbank muss registriert sein, bevor der init dieses
	// Pakets die Zone lädt; nur ein Import hier erzwingt diese Reihenfolge (ein
	// Blank-Import in main.go steht in keiner Abhängigkeit zu diesem Paket).
	_ "time/tzdata"
)

// Berlin ist die Zone dieser Umrechnung, einmal beim Start geladen. Die
// Zonendatenbank liegt im Binary, das Laden schlägt nur bei kaputtem Build fehl.
var Berlin = mustLoad("Europe/Berlin")

func mustLoad(name string) *time.Location {
	ort, err := time.LoadLocation(name)
	if err != nil {
		panic("zeit: Zeitzone " + name + " nicht ladbar: " + err.Error())
	}
	return ort
}
