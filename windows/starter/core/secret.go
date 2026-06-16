package core

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateSecret erzeugt 32 Bytes aus crypto/rand und gibt sie hex-kodiert
// zurueck (64 Zeichen) — identisch zu "openssl rand -hex 32" in
// scripts/init-env.sh. Hex statt Base64, damit das Secret die Postgres-URL im
// migrate-CMD (postgres://user:pass@...) nie durch +/= bricht.
func GenerateSecret() string {
	b := make([]byte, 32)
	// crypto/rand.Read fuellt den Puffer vollstaendig oder panict selbst; ein
	// zurueckgegebener Fehler ist hier praktisch ausgeschlossen.
	if _, err := rand.Read(b); err != nil {
		panic("crypto/rand nicht verfuegbar: " + err.Error())
	}
	return hex.EncodeToString(b)
}
