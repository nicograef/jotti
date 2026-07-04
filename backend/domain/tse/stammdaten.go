package tse

import (
	"strings"
	"time"
)

// Stammdaten sind die fiskalischen Stammdaten der TSS, die der DSFinV-K-Export
// fuer die tse.csv braucht: Signaturalgorithmus, Public Key, Zertifikat und
// Log-Time-Format. Sie aendern sich ueber die Lebensdauer einer TSS nicht und
// werden einmalig bei der Einrichtung von fiskaly gelesen und als Singleton
// gespeichert.
type Stammdaten struct {
	SignaturAlgorithmus string
	PublicKey           string
	Zertifikat          string
	LogTimeFormat       string
	UpdatedAt           time.Time
}

// NewStammdaten baut die Stammdaten aus den von fiskaly gelesenen Feldern und
// stempelt den Lesezeitpunkt. Validiert wird nicht: die Felder stammen aus der
// vertrauenswuerdigen TSS-Ressource, nicht aus Nutzereingaben.
func NewStammdaten(signaturAlgorithmus, publicKey, zertifikat, logTimeFormat string) Stammdaten {
	return Stammdaten{
		SignaturAlgorithmus: strings.TrimSpace(signaturAlgorithmus),
		PublicKey:           strings.TrimSpace(publicKey),
		Zertifikat:          strings.TrimSpace(zertifikat),
		LogTimeFormat:       strings.TrimSpace(logTimeFormat),
		UpdatedAt:           time.Now().UTC(),
	}
}
