package settings

import (
	"strings"
	"time"
)

// TSEStammdaten sind die fiskalischen Stammdaten der TSS, die der DSFinV-K-Export
// fuer die tse.csv braucht: Signaturalgorithmus, Public Key, Zertifikat,
// Log-Time-Format und die Versionsangabe. Sie aendern sich ueber die Lebensdauer
// einer TSS nicht und werden einmalig bei der Einrichtung von fiskaly gelesen und
// als Singleton gespeichert.
type TSEStammdaten struct {
	SignaturAlgorithmus string
	PublicKey           string
	Zertifikat          string
	LogTimeFormat       string
	Version             string
	UpdatedAt           time.Time
}

// NewTSEStammdaten baut die Stammdaten aus den von fiskaly gelesenen Feldern und
// stempelt den Lesezeitpunkt. Validiert wird nicht: die Felder stammen aus der
// vertrauenswuerdigen TSS-Ressource, nicht aus Nutzereingaben.
func NewTSEStammdaten(signaturAlgorithmus, publicKey, zertifikat, logTimeFormat, version string) TSEStammdaten {
	return TSEStammdaten{
		SignaturAlgorithmus: strings.TrimSpace(signaturAlgorithmus),
		PublicKey:           strings.TrimSpace(publicKey),
		Zertifikat:          strings.TrimSpace(zertifikat),
		LogTimeFormat:       strings.TrimSpace(logTimeFormat),
		Version:             strings.TrimSpace(version),
		UpdatedAt:           time.Now().UTC(),
	}
}
