package application

import (
	"crypto/sha256"
	"sync"

	"github.com/nicograef/jotti/backend/domain/tse"
)

// UmgebungsCache remembers the fiskaly environment per key pair, which a key pair never changes.
// The status display then skips the auth round trip; the key is a hash, never the secret.
// The zero value is ready; a nil cache caches nothing.
type UmgebungsCache struct {
	mu         sync.Mutex
	umgebungen map[[sha256.Size]byte]tse.Umgebung
}

func umgebungsSchluessel(c tse.Credentials) [sha256.Size]byte {
	return sha256.Sum256([]byte(c.ApiKey + "\x00" + c.ApiSecret))
}

func (c *UmgebungsCache) lies(credentials tse.Credentials) (tse.Umgebung, bool) {
	if c == nil {
		return "", false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	umgebung, ok := c.umgebungen[umgebungsSchluessel(credentials)]
	return umgebung, ok
}

func (c *UmgebungsCache) merke(credentials tse.Credentials, umgebung tse.Umgebung) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.umgebungen == nil {
		c.umgebungen = map[[sha256.Size]byte]tse.Umgebung{}
	}
	c.umgebungen[umgebungsSchluessel(credentials)] = umgebung
}
