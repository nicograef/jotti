package tse

import "time"

// Signatur ist die abgeschlossene TSE-Signatur eines Vorgangs, wie sie der
// Signatur-Worker am Signaturauftrag quittiert. Beleg und DSFinV-K-Export
// lesen genau diese eine Quelle.
type Signatur struct {
	TransaktionNummer int
	SignaturZaehler   int
	TSESeriennummer   string
	LogTimeStart      time.Time
	LogTimeEnd        time.Time
	Signatur          string
	QRCodeData        string
}
