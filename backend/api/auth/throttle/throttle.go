// Package throttle drosselt fehlgeschlagene Anmeldungen pro Konto im Speicher.
// Soft-Throttle: kein dauerhaftes Sperren (für ehrenamtliche Helfer im
// Event-Betrieb ein Footgun), sondern ein automatisch ablaufender, exponentiell
// wachsender Cooldown.
package throttle

import (
	"sync"
	"time"
)

const (
	defaultThreshold = 5
	defaultBase      = 1 * time.Second
	defaultMax       = 15 * time.Minute
	// defaultTTL: Inaktivitätsfrist eines Eintrags — die Map darf nicht unbegrenzt wachsen.
	defaultTTL = 1 * time.Hour
)

type entry struct {
	failures      int
	cooldownUntil time.Time
	lastSeen      time.Time
}

// LoginThrottle ist nebenläufig sicher: alle Methoden sind mutex-geschützt.
type LoginThrottle struct {
	mu      sync.Mutex
	entries map[string]*entry

	now       func() time.Time
	threshold int
	base      time.Duration
	max       time.Duration
	ttl       time.Duration
}

// NewLoginThrottle startet zusätzlich die Aufräum-Goroutine, die für die
// Prozesslebensdauer läuft.
func NewLoginThrottle() *LoginThrottle {
	t := newLoginThrottle(defaultThreshold, defaultBase, defaultMax, defaultTTL)
	go t.cleanupLoop()
	return t
}

// newLoginThrottle lässt die Aufräum-Goroutine weg: Tests steuern now, ohne mit
// ihr um das Feld zu rennen.
func newLoginThrottle(threshold int, base, max, ttl time.Duration) *LoginThrottle {
	return &LoginThrottle{
		entries:   make(map[string]*entry),
		now:       func() time.Time { return time.Now().UTC() },
		threshold: threshold,
		base:      base,
		max:       max,
		ttl:       ttl,
	}
}

func (t *LoginThrottle) Allow(username string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	e, ok := t.entries[username]
	if !ok {
		return true
	}

	now := t.now()
	if now.Sub(e.lastSeen) > t.ttl {
		delete(t.entries, username)
		return true
	}

	e.lastSeen = now
	return !now.Before(e.cooldownUntil)
}

func (t *LoginThrottle) RecordFailure(username string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := t.now()
	e, ok := t.entries[username]
	if !ok {
		e = &entry{}
		t.entries[username] = e
	}

	e.failures++
	e.lastSeen = now
	if e.failures >= t.threshold {
		e.cooldownUntil = now.Add(t.backoff(e.failures))
	}
}

func (t *LoginThrottle) Reset(username string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	delete(t.entries, username)
}

func (t *LoginThrottle) backoff(failures int) time.Duration {
	shift := failures - t.threshold
	if shift < 0 {
		shift = 0
	}
	if shift > 30 { // Schutz vor Überlauf beim Bit-Shift
		return t.max
	}
	d := t.base << uint(shift)
	if d <= 0 || d > t.max {
		return t.max
	}
	return d
}

func (t *LoginThrottle) cleanupLoop() {
	for {
		time.Sleep(t.ttl)
		t.mu.Lock()
		now := t.now()
		for username, e := range t.entries {
			if now.Sub(e.lastSeen) > t.ttl {
				delete(t.entries, username)
			}
		}
		t.mu.Unlock()
	}
}
