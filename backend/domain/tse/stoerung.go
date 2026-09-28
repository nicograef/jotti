package tse

import "time"

// Grund-Art values, mirrored by the tse_stoerungen CHECK constraint.
// Each writer closes only periods of its own Grund-Art.
const (
	StoerungGrundTSEFehler          = "tse_fehler"
	StoerungGrundRueckstand         = "rueckstand"
	StoerungGrundKeineKonfiguration = "keine_konfiguration"
)

// Stoerung: at most one period is active, and open jobs during it count as Ausfall.
type Stoerung struct {
	Beginn     time.Time
	GrundArt   string
	Fehlertext string
}

// Stoerungszeitraum has a nil Ende while active.
type Stoerungszeitraum struct {
	ID         int
	Beginn     time.Time
	Ende       *time.Time
	GrundArt   string
	Fehlertext string
}
