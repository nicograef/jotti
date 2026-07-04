package tse

import "time"

// Status eines Signaturauftrags (CHECK-Constraint der Tabelle
// tse_signaturauftraege).
const (
	StatusOffen                = "offen"
	StatusErledigt             = "erledigt"
	StatusFehlgeschlagen       = "fehlgeschlagen"
	StatusVerworfen            = "verworfen"
	StatusTSENichtKonfiguriert = "tse_nicht_konfiguriert"
)

// SignaturauftragStand ist der Signatur-Stand eines Events: Status des
// Auftrags plus Signatur, sobald quittiert.
type SignaturauftragStand struct {
	Status     string
	ErstelltAm time.Time
	Signatur   *Signatur
}
