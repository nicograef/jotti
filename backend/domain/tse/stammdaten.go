package tse

import (
	"strings"
	"time"
)

// Stammdaten sind die fiskalischen Stammdaten der TSS für die tse.csv des
// DSFinV-K-Exports. Sie ändern sich über die Lebensdauer einer TSS nicht und
// werden einmalig bei der Einrichtung gelesen und als Singleton gespeichert.
type Stammdaten struct {
	// Seriennummer ist die TSS-Seriennummer (fiskaly: serial_number der
	// TSS-Ressource; SHA-256 des Public Key, hex-kodiert). DSFinV-K-Feld
	// TSE_SERIAL.
	Seriennummer        string
	SignaturAlgorithmus string
	PublicKey           string
	Zertifikat          string
	LogTimeFormat       string
	UpdatedAt           time.Time
}

// NewStammdaten validiert nicht: die Felder stammen aus der TSS-Ressource, nicht
// aus Nutzereingaben.
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
