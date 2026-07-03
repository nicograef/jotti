package kasse

import "time"

type KassensitzungStatus string

// Kassensitzung represents the CRUD entity for a Kassensitzung.
type Kassensitzung struct {
	ZNr         int
	Datum       time.Time
	Bezeichnung string
	Status      KassensitzungStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

const (
	KassensitzungOffen KassensitzungStatus = "offen"
	// KassensitzungWirdAbgeschlossen ist der transiente Zwischenstatus während des
	// Abschlusses: Ab ihm lehnt der Status-Guard alle Buchungs-Events ab, nur die
	// Abschluss-Events selbst dürfen noch geschrieben werden.
	KassensitzungWirdAbgeschlossen KassensitzungStatus = "wird_abgeschlossen"
	KassensitzungAbgeschlossen     KassensitzungStatus = "abgeschlossen"
)
