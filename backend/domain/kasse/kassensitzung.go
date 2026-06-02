package kasse

import "time"

// Kassensitzung represents the CRUD entity for a Kassensitzung.
type Kassensitzung struct {
	ZNr         int
	Datum       time.Time
	Bezeichnung string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

const (
	KassensitzungOffen         = "offen"
	KassensitzungAbgeschlossen = "abgeschlossen"
)
