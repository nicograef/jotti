package tse

import "time"

// Signaturauftrag status values, mirrored by the tse_signaturauftraege CHECK constraint.
const (
	StatusOffen                = "offen"
	StatusErledigt             = "erledigt"
	StatusFehlgeschlagen       = "fehlgeschlagen"
	StatusTSENichtKonfiguriert = "tse_nicht_konfiguriert"
)

type SignaturauftragStand struct {
	Status     string
	ErstelltAm time.Time
	Signatur   *Signatur
}

// SignaturQueueZustand measures throughput over a sliding 15-minute window to tell a growing from a shrinking backlog.
// FehlgeschlageneAuftraege and LetzterFehler cover only the active Kassensitzung.
type SignaturQueueZustand struct {
	OffeneAuftraege          int
	FehlgeschlageneAuftraege int
	LetzterFehler            string
	RueckstandSekunden       int
	SignaturenProMinute      float64
	SignierdauerP95Sekunden  float64
}
