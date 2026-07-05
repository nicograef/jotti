// Package throttle bietet eine In-Memory-Drosselung fehlgeschlagener Anmeldungen
// pro Konto. Sie ist ein Soft-Throttle: kein dauerhaftes Sperren eines Kontos
// (das wäre für nicht-technische Vereinshelfer im Event-Betrieb ein Footgun),
// sondern ein automatisch ablaufender, exponentiell wachsender Cooldown. Kein
// Schema, keine persistente Sicherheits-Statushaltung — gespiegelt am In-Memory-
// Muster von middleware.RateLimitMiddleware.
package throttle

import (
	"sync"
	"time"
)

const (
	// defaultThreshold: so viele aufeinanderfolgende Fehlversuche sind erlaubt,
	// bevor der Cooldown greift.
	defaultThreshold = 5
	// defaultBase: Cooldown nach dem ersten Überschreiten der Schwelle. Danach
	// verdoppelt sich der Cooldown je weiterem Fehlversuch (exponentieller Backoff).
	defaultBase = 1 * time.Second
	// defaultMax: Obergrenze für den Cooldown.
	defaultMax = 15 * time.Minute
	// defaultTTL: nach so langer Inaktivität wird ein Eintrag verworfen, damit die
	// Map nicht unbegrenzt wächst und ein Konto irgendwann wieder frisch startet.
	defaultTTL = 1 * time.Hour
)

// entry hält den Drosselungszustand eines einzelnen Kontos.
type entry struct {
	failures      int
	cooldownUntil time.Time
	lastSeen      time.Time
}

// LoginThrottle drosselt Anmeldeversuche pro Benutzername. Alle Methoden sind
// nebenläufig sicher (mutex-geschützt).
type LoginThrottle struct {
	mu      sync.Mutex
	entries map[string]*entry

	now       func() time.Time
	threshold int
	base      time.Duration
	max       time.Duration
	ttl       time.Duration
}

// NewLoginThrottle erzeugt eine gebrauchsfertige Drosselung mit den Standardwerten
// und startet die Aufräum-Goroutine, die verwaiste Einträge periodisch entfernt.
func NewLoginThrottle() *LoginThrottle {
	t := newLoginThrottle(defaultThreshold, defaultBase, defaultMax, defaultTTL)
	go t.cleanupLoop()
	return t
}

// newLoginThrottle baut die Drosselung OHNE Aufräum-Goroutine — für Tests, die
// die Uhr (now) kontrollieren, ohne mit der Goroutine um das now-Feld zu rennen.
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

// Allow meldet, ob für username gerade ein Anmeldeversuch erlaubt ist. Ein Konto
// ohne Eintrag ist immer erlaubt; ein verwaister Eintrag wird verworfen (frischer
// Start). Während eines aktiven Cooldowns liefert Allow false.
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

// RecordFailure verbucht einen Fehlversuch für username und setzt ab der Schwelle
// einen exponentiell wachsenden Cooldown.
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

// Reset löscht den Drosselungszustand eines Kontos — aufzurufen nach einer
// erfolgreichen Anmeldung.
func (t *LoginThrottle) Reset(username string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	delete(t.entries, username)
}

// backoff liefert den Cooldown für die gegebene Fehlversuchszahl: base verdoppelt
// je Fehlversuch über der Schwelle, gedeckelt auf max.
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

// cleanupLoop entfernt periodisch Einträge, die länger als ttl nicht gesehen
// wurden. Läuft für die Lebensdauer der Anwendung (Singleton).
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
