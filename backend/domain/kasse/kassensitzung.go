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
	KassensitzungOffen         KassensitzungStatus = "offen"
	KassensitzungAbgeschlossen KassensitzungStatus = "abgeschlossen"
)
