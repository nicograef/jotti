package tse

import (
	"strings"
	"time"
)

// Stammdaten feed the DSFinV-K tse.csv and never change over a TSS lifetime.
// The setup reads them once and stores them as a singleton.
type Stammdaten struct {
	// Seriennummer is fiskaly's TSS serial_number, the hex SHA-256 of the public key (DSFinV-K TSE_SERIAL).
	Seriennummer        string
	SignaturAlgorithmus string
	PublicKey           string
	Zertifikat          string
	LogTimeFormat       string
	UpdatedAt           time.Time
}

// NewStammdaten does not validate because the fields come from the TSS resource, not user input.
func NewStammdaten(seriennummer, signaturAlgorithmus, publicKey, zertifikat, logTimeFormat string) Stammdaten {
	return Stammdaten{
		Seriennummer:        strings.TrimSpace(seriennummer),
		SignaturAlgorithmus: strings.TrimSpace(signaturAlgorithmus),
		PublicKey:           strings.TrimSpace(publicKey),
		Zertifikat:          strings.TrimSpace(zertifikat),
		LogTimeFormat:       strings.TrimSpace(logTimeFormat),
		UpdatedAt:           time.Now().UTC(),
	}
}
