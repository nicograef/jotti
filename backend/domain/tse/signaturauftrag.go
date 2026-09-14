package tse

import "time"

// Status eines Signaturauftrags (CHECK-Constraint der Tabelle
// tse_signaturauftraege).
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

// SignaturQueueZustand ist der on demand berechnete Zustand der Signatur-Queue
// für das Admin-Monitoring: Rückstand (offene Aufträge, Alter des ältesten) und
// Leistung über ein gleitendes 15-Minuten-Fenster — so unterscheidet sich ein
// wachsender von einem schrumpfenden Rückstand. FehlgeschlageneAuftraege und
// LetzterFehler gelten nur für die aktive Kassensitzung.
type SignaturQueueZustand struct {
	OffeneAuftraege          int
	FehlgeschlageneAuftraege int
	LetzterFehler            string
	RueckstandSekunden       int
	SignaturenProMinute      float64
	SignierdauerP95Sekunden  float64
}
