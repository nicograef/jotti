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

// SignaturauftragStand ist der Signatur-Stand eines Events: Status des
// Auftrags plus Signatur, sobald quittiert.
type SignaturauftragStand struct {
	Status     string
	ErstelltAm time.Time
	Signatur   *Signatur
}

// SignaturQueueZustand ist der on demand berechnete Zustand der Signatur-Queue
// für das Admin-Monitoring: Rückstand (offene Aufträge, Alter des ältesten)
// und Leistung über ein gleitendes 15-Minuten-Fenster (Signaturen pro Minute,
// Signierdauer p95). So lässt sich ein wachsender von einem schrumpfenden
// Rückstand unterscheiden. FehlgeschlageneAuftraege und LetzterFehler sind
// sitzungsbezogen (nur die aktive Kassensitzung); mit dem Kassenabschluss
// verschwindet die Warnung.
type SignaturQueueZustand struct {
	OffeneAuftraege          int
	FehlgeschlageneAuftraege int
	LetzterFehler            string
	RueckstandSekunden       int
	SignaturenProMinute      float64
	SignierdauerP95Sekunden  float64
}
