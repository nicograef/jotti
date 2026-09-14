package core

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateSecret erzeugt 32 Bytes aus crypto/rand, hex-kodiert (64 Zeichen) —
// identisch zu "openssl rand -hex 32" in scripts/init-env.sh. Hex statt Base64,
// damit das Secret die Postgres-URL im migrate-CMD nie durch +/= bricht.
func GenerateSecret() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic("crypto/rand nicht verfuegbar: " + err.Error())
	}
	return hex.EncodeToString(b)
}
