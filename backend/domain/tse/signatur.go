package tse

import "time"

// Signatur ist die abgeschlossene TSE-Signatur, die der Signatur-Worker am
// Signaturauftrag quittiert. Beleg und DSFinV-K-Export lesen nur diese Quelle.
type Signatur struct {
	TransaktionNummer int
	SignaturZaehler   int
	TSESeriennummer   string
	LogTimeStart      time.Time
	LogTimeEnd        time.Time
	Signatur          string
	QRCodeData        string
}
