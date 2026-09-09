package tse

import "time"

// Grund-Art eines Störungszeitraums (CHECK-Constraint der Tabelle
// tse_stoerungen). Jeder Schreiber des Störungsprotokolls schließt nur
// Zeiträume seiner Grund-Art.
const (
	StoerungGrundTSEFehler          = "tse_fehler"
	StoerungGrundRueckstand         = "rueckstand"
	StoerungGrundKeineKonfiguration = "keine_konfiguration"
)

// Stoerung ist ein Zeitraum im Störungsprotokoll der TSE-Signierung.
// Höchstens ein Zeitraum ist aktiv; offene Signaturaufträge werden während
// eines aktiven Zeitraums dem Ausfall zugerechnet (DetermineSignaturstatus).
type Stoerung struct {
	Beginn     time.Time
	GrundArt   string
	Fehlertext string
}

// Stoerungszeitraum ist ein Eintrag des Störungsprotokolls (Ausfalldokumentation):
// ein Zeitraum mit Beginn, Ende (nil solange aktiv) und Grund-Art.
type Stoerungszeitraum struct {
	ID         int
	Beginn     time.Time
	Ende       *time.Time
	GrundArt   string
	Fehlertext string
}
