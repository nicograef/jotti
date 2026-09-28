package tse

import "time"

// Signatur is the only signature source that receipts and the DSFinV-K export read.
type Signatur struct {
	TransaktionNummer int
	SignaturZaehler   int
	TSESeriennummer   string
	LogTimeStart      time.Time
	LogTimeEnd        time.Time
	Signatur          string
	QRCodeData        string
}
