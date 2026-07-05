//go:build unit

package throttle

import (
	"sync"
	"testing"
	"time"
)

// controllable baut eine Drosselung mit fester, testgesteuerter Uhr und OHNE
// Aufräum-Goroutine (die würde nebenläufig auf now zugreifen).
func controllable(threshold int, base, max, ttl time.Duration) (*LoginThrottle, *time.Time) {
	t := newLoginThrottle(threshold, base, max, ttl)
	clock := time.Now().UTC()
	t.now = func() time.Time { return clock }
	return t, &clock
}

func TestLoginThrottle_FreshUserAllowed(t *testing.T) {
	th, _ := controllable(3, time.Second, time.Minute, time.Hour)

	if !th.Allow("neu") {
		t.Fatal("ein Konto ohne Fehlversuche muss erlaubt sein")
	}
}

func TestLoginThrottle_ThrottlesAfterThreshold(t *testing.T) {
	th, _ := controllable(3, time.Second, time.Minute, time.Hour)

	for i := 0; i < 3; i++ {
		if !th.Allow("angriff") {
			t.Fatalf("Versuch %d sollte vor der Schwelle erlaubt sein", i+1)
		}
		th.RecordFailure("angriff")
	}

	if th.Allow("angriff") {
		t.Fatal("nach Erreichen der Schwelle muss der nächste Versuch gedrosselt sein")
	}
}

func TestLoginThrottle_SuccessResets(t *testing.T) {
	th, _ := controllable(3, time.Second, time.Minute, time.Hour)

	for i := 0; i < 3; i++ {
		th.RecordFailure("konto")
	}
	if th.Allow("konto") {
		t.Fatal("Vorbedingung: Konto sollte gedrosselt sein")
	}

	th.Reset("konto")

	if !th.Allow("konto") {
		t.Fatal("nach Reset (erfolgreicher Login) muss das Konto wieder erlaubt sein")
	}
}

func TestLoginThrottle_CooldownExpires(t *testing.T) {
	th, clock := controllable(3, time.Second, time.Minute, time.Hour)

	for i := 0; i < 3; i++ {
		th.RecordFailure("konto")
	}
	if th.Allow("konto") {
		t.Fatal("Vorbedingung: Konto sollte gedrosselt sein")
	}

	*clock = clock.Add(2 * time.Second) // Cooldown (1s) verstrichen

	if !th.Allow("konto") {
		t.Fatal("nach Ablauf des Cooldowns muss der Versuch wieder erlaubt sein")
	}
}

func TestLoginThrottle_IdleEntryEvicted(t *testing.T) {
	th, clock := controllable(3, time.Second, time.Minute, time.Hour)

	for i := 0; i < 5; i++ {
		th.RecordFailure("konto")
	}

	*clock = clock.Add(2 * time.Hour) // länger untätig als ttl

	if !th.Allow("konto") {
		t.Fatal("ein lange untätiges Konto muss frisch starten (Eintrag verworfen)")
	}
	th.mu.Lock()
	_, exists := th.entries["konto"]
	th.mu.Unlock()
	if exists {
		t.Fatal("der verwaiste Eintrag muss aus der Map entfernt sein")
	}
}

func TestLoginThrottle_ExponentialBackoff(t *testing.T) {
	th, _ := controllable(1, time.Second, time.Hour, 24*time.Hour)

	if got := th.backoff(1); got != time.Second {
		t.Errorf("erster Cooldown = %v, want 1s", got)
	}
	if got := th.backoff(2); got != 2*time.Second {
		t.Errorf("zweiter Cooldown = %v, want 2s", got)
	}
	if got := th.backoff(3); got != 4*time.Second {
		t.Errorf("dritter Cooldown = %v, want 4s", got)
	}
	if got := th.backoff(1000); got != time.Hour {
		t.Errorf("Cooldown muss auf max gedeckelt sein, got %v", got)
	}
}

func TestLoginThrottle_PerAccount(t *testing.T) {
	th, _ := controllable(3, time.Second, time.Minute, time.Hour)

	for i := 0; i < 5; i++ {
		th.RecordFailure("opfer")
	}

	if th.Allow("opfer") {
		t.Fatal("das gedrosselte Konto sollte blockiert sein")
	}
	if !th.Allow("unbeteiligt") {
		t.Fatal("ein anderes Konto darf nie von der Drosselung betroffen sein")
	}
}

// -race beweist die Nebenläufigkeitssicherheit der Map.
func TestLoginThrottle_ConcurrentAccessRaceSafe(t *testing.T) {
	th := newLoginThrottle(5, time.Millisecond, time.Second, time.Hour)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			user := "u" + string(rune('a'+n%5))
			th.Allow(user)
			th.RecordFailure(user)
			th.Allow(user)
			if n%3 == 0 {
				th.Reset(user)
			}
		}(i)
	}
	wg.Wait()
}
