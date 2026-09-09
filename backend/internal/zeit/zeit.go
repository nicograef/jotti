// Package zeit hält die Zeitzone, in der jotti Zeitpunkte anzeigt, druckt und
// benennt. Ein Zeitpunkt liegt als UTC in der Datenbank; Beleg, Arbeitsbon,
// Archivname und Meldedatum tragen deutsche Ortszeit, weil Gast, Betreiber und
// Prüfung sie am Wandkalender lesen.
package zeit

import (
	"time"
	// Die eingebettete Zonendatenbank muss registriert sein, bevor der init
	// dieses Pakets die Zone lädt. Nur ein Import hier erzwingt diese
	// Reihenfolge: die Init-Reihenfolge folgt dem Abhängigkeitsgraphen, und ein
	// Blank-Import in main.go steht in keiner Abhängigkeit zu diesem Paket.
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
