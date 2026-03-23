package kasse

import "time"

// KassensitzungState represents the CRUD entity for a Kassensitzung.
type KassensitzungState struct {
	ZNr         int
	Datum       time.Time
	Bezeichnung string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

const (
	KassensitzungStatusOffen         = "offen"
	KassensitzungStatusAbgeschlossen = "abgeschlossen"
)
