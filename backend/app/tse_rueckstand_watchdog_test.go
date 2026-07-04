//go:build unit

package app

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/nicograef/jotti/backend/domain/tse"
)

type stoerungAufruf struct {
	GrundArt   string
	Fehlertext string
}

type mockRueckstandStore struct {
	mu          sync.Mutex
	aeltester   *time.Time
	getErr      error
	geoeffnet   []stoerungAufruf
	geschlossen []string
	// geprueft signalisiert jeden Durchlauf (fuer Run-Loop-Tests ohne Sleeps).
	geprueft chan struct{}
}

func (m *mockRueckstandStore) GetAeltesterOffenerTSESignaturauftrag(_ context.Context) (*time.Time, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.aeltester, nil
}

func (m *mockRueckstandStore) OeffneTSEStoerung(_ context.Context, grundArt string, fehlertext string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.geoeffnet = append(m.geoeffnet, stoerungAufruf{GrundArt: grundArt, Fehlertext: fehlertext})
	m.signalisierePruefung()
	return nil
}

func (m *mockRueckstandStore) SchliesseTSEStoerung(_ context.Context, grundArt string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.geschlossen = append(m.geschlossen, grundArt)
	m.signalisierePruefung()
	return nil
}

func (m *mockRueckstandStore) signalisierePruefung() {
	if m.geprueft != nil {
		select {
		case m.geprueft <- struct{}{}:
		default:
		}
	}
}

var watchdogJetzt = time.Date(2026, 6, 10, 18, 0, 0, 0, time.UTC)

func neuerTestWatchdog(store *mockRueckstandStore) *tseRueckstandWatchdog {
	return &tseRueckstandWatchdog{store: store, now: func() time.Time { return watchdogJetzt }}
}

// Der Watchdog oeffnet den Rueckstands-Zeitraum an der Schwelle allein anhand
// des Auftragsalters — ohne Mitwirkung des Signatur-Workers. Ein haengender
// Worker wird damit genauso dokumentiert wie eine langsame TSE.
func TestRueckstandWatchdog_OeffnetAbSchwelle_AuchOhneWorker(t *testing.T) {
	alt := watchdogJetzt.Add(-tse.RueckstandSchwelle)
	store := &mockRueckstandStore{aeltester: &alt}

	if err := neuerTestWatchdog(store).pruefeRueckstand(context.Background()); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(store.geoeffnet) != 1 {
		t.Fatalf("expected 1 geoeffneten Zeitraum, got %d", len(store.geoeffnet))
	}
	if store.geoeffnet[0].GrundArt != tse.StoerungGrundRueckstand {
		t.Errorf("GrundArt = %q, erwartet %q", store.geoeffnet[0].GrundArt, tse.StoerungGrundRueckstand)
	}
	if store.geoeffnet[0].Fehlertext == "" {
		t.Error("Fehlertext ist leer")
	}
	if len(store.geschlossen) != 0 {
		t.Errorf("expected kein Schliessen, got %v", store.geschlossen)
	}
}

func TestRueckstandWatchdog_SchliesstUnterSchwelle(t *testing.T) {
	jung := watchdogJetzt.Add(-30 * time.Second)
	store := &mockRueckstandStore{aeltester: &jung}

	if err := neuerTestWatchdog(store).pruefeRueckstand(context.Background()); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(store.geoeffnet) != 0 {
		t.Errorf("expected kein Oeffnen, got %v", store.geoeffnet)
	}
	if len(store.geschlossen) != 1 || store.geschlossen[0] != tse.StoerungGrundRueckstand {
		t.Errorf("expected Schliessen der Grund-Art rueckstand, got %v", store.geschlossen)
	}
}

func TestRueckstandWatchdog_SchliesstOhneOffeneAuftraege(t *testing.T) {
	store := &mockRueckstandStore{}

	if err := neuerTestWatchdog(store).pruefeRueckstand(context.Background()); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(store.geoeffnet) != 0 {
		t.Errorf("expected kein Oeffnen, got %v", store.geoeffnet)
	}
	if len(store.geschlossen) != 1 {
		t.Errorf("expected 1 Schliessen, got %d", len(store.geschlossen))
	}
}

func TestRueckstandWatchdog_StoreFehlerWirdGemeldet(t *testing.T) {
	store := &mockRueckstandStore{getErr: errors.New("db down")}

	if err := neuerTestWatchdog(store).pruefeRueckstand(context.Background()); err == nil {
		t.Fatal("expected error, got nil")
	}
	if len(store.geoeffnet) != 0 || len(store.geschlossen) != 0 {
		t.Error("expected keine Protokoll-Schreibzugriffe bei Store-Fehler")
	}
}

// Die Rueckstands-Schwelle materialisiert nur am Tick: Der Run-Loop prueft im
// Tick-Intervall und oeffnet den Zeitraum am naechsten Tick nach der
// Schwellen-Ueberschreitung.
func TestRueckstandWatchdog_Run_OeffnetAmTick(t *testing.T) {
	alt := watchdogJetzt.Add(-tse.RueckstandSchwelle - time.Second)
	store := &mockRueckstandStore{aeltester: &alt, geprueft: make(chan struct{}, 1)}

	watchdog := neuerTestWatchdog(store)
	watchdog.tickInterval = time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() {
		watchdog.run(ctx)
		close(done)
	}()

	select {
	case <-store.geprueft:
	case <-time.After(2 * time.Second):
		t.Fatal("Watchdog hat am Tick nicht geprueft")
	}
	cancel()
	<-done

	store.mu.Lock()
	defer store.mu.Unlock()
	if len(store.geoeffnet) < 1 {
		t.Fatal("expected geoeffneten Rueckstands-Zeitraum nach Tick")
	}
}
