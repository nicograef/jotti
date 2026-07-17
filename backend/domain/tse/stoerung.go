package tse

import "time"

// Grund-Art eines Stoerungszeitraums (CHECK-Constraint der Tabelle
// tse_stoerungen). Jeder Schreiber des Stoerungsprotokolls schliesst nur
// Zeitraeume seiner Grund-Art.
const (
	StoerungGrundTSEFehler          = "tse_fehler"
	StoerungGrundRueckstand         = "rueckstand"
	StoerungGrundKeineKonfiguration = "keine_konfiguration"
)

// Stoerung ist ein Zeitraum im Stoerungsprotokoll der TSE-Signierung.
// Hoechstens ein Zeitraum ist aktiv; offene Signaturauftraege werden waehrend
// eines aktiven Zeitraums dem Ausfall zugerechnet (DetermineSignaturstatus).
type Stoerung struct {
	Beginn     time.Time
	GrundArt   string
	Fehlertext string
}

// Stoerungszeitraum ist ein Eintrag des Stoerungsprotokolls (Ausfalldokumentation):
// ein Zeitraum mit Beginn, Ende (nil solange aktiv) und Grund-Art.
type Stoerungszeitraum struct {
	ID         int
	Beginn     time.Time
	Ende       *time.Time
	GrundArt   string
	Fehlertext string
}
