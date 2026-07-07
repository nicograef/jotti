//go:build unit

package signatur

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

func (m *mockRueckstandStore) OpenTSEStoerung(_ context.Context, grundArt string, fehlertext string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.geoeffnet = append(m.geoeffnet, stoerungAufruf{GrundArt: grundArt, Fehlertext: fehlertext})
	m.signalCheck()
	return nil
}

func (m *mockRueckstandStore) CloseTSEStoerung(_ context.Context, grundArt string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.geschlossen = append(m.geschlossen, grundArt)
	m.signalCheck()
	return nil
}

func (m *mockRueckstandStore) signalCheck() {
	if m.geprueft != nil {
		select {
		case m.geprueft <- struct{}{}:
		default:
		}
	}
}

var watchdogJetzt = time.Date(2026, 6, 10, 18, 0, 0, 0, time.UTC)

func newTestWatchdog(store *mockRueckstandStore) *tseRueckstandWatchdog {
	return &tseRueckstandWatchdog{store: store, now: func() time.Time { return watchdogJetzt }}
}

// Der Watchdog oeffnet den Rueckstands-Zeitraum an der Schwelle allein anhand
// des Auftragsalters — ohne Mitwirkung des Signatur-Workers. Ein haengender
// Worker wird damit genauso dokumentiert wie eine langsame TSE.
func TestRueckstandWatchdog_OeffnetAbSchwelle_AuchOhneWorker(t *testing.T) {
	alt := watchdogJetzt.Add(-tse.RueckstandSchwelle)
	store := &mockRueckstandStore{aeltester: &alt}

	if err := newTestWatchdog(store).checkRueckstand(context.Background()); err != nil {
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

	if err := newTestWatchdog(store).checkRueckstand(context.Background()); err != nil {
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

	if err := newTestWatchdog(store).checkRueckstand(context.Background()); err != nil {
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

	if err := newTestWatchdog(store).checkRueckstand(context.Background()); err == nil {
		t.Fatal("expected error, got nil")
	}
	if len(store.geoeffnet) != 0 || len(store.geschlossen) != 0 {
		t.Error("expected keine Protokoll-Schreibzugriffe bei Store-Fehler")
	}
}

// panicEinmalRueckstandStore panict beim ersten Abfragen des aeltesten offenen
// Auftrags und funktioniert danach normal.
type panicEinmalRueckstandStore struct {
	*mockRueckstandStore
	panicMu  sync.Mutex
	gepanict bool
}

func (s *panicEinmalRueckstandStore) GetAeltesterOffenerTSESignaturauftrag(ctx context.Context) (*time.Time, error) {
	s.panicMu.Lock()
	erster := !s.gepanict
	s.gepanict = true
	s.panicMu.Unlock()
	if erster {
		panic("provozierter Panic im Durchlauf")
	}
	return s.mockRueckstandStore.GetAeltesterOffenerTSESignaturauftrag(ctx)
}

// Ein Panic im Durchlauf beendet den Watchdog nicht: Der Run-Loop faengt ihn
// ab und prueft am naechsten Tick weiter.
func TestRueckstandWatchdog_Run_PanicStopptUeberwachungNicht(t *testing.T) {
	alt := watchdogJetzt.Add(-tse.RueckstandSchwelle - time.Second)
	store := &panicEinmalRueckstandStore{mockRueckstandStore: &mockRueckstandStore{aeltester: &alt, geprueft: make(chan struct{}, 1)}}

	watchdog := &tseRueckstandWatchdog{store: store, now: func() time.Time { return watchdogJetzt }}
	watchdog.tickInterval = time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() {
		watchdog.Run(ctx)
		close(done)
	}()

	select {
	case <-store.geprueft:
	case <-time.After(2 * time.Second):
		t.Fatal("Watchdog hat nach dem Panic nicht weiter geprueft")
	}
	cancel()
	<-done
}

// Die Rueckstands-Schwelle materialisiert nur am Tick: Der Run-Loop prueft im
// Tick-Intervall und oeffnet den Zeitraum am naechsten Tick nach der
// Schwellen-Ueberschreitung.
func TestRueckstandWatchdog_Run_OeffnetAmTick(t *testing.T) {
	alt := watchdogJetzt.Add(-tse.RueckstandSchwelle - time.Second)
	store := &mockRueckstandStore{aeltester: &alt, geprueft: make(chan struct{}, 1)}

	watchdog := newTestWatchdog(store)
	watchdog.tickInterval = time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() {
		watchdog.Run(ctx)
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
