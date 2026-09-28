package application

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/nicograef/jotti/backend/domain/tse/tsetest"
)

// blockierenderSetupClient parks a lifecycle in ListTSS, where an unlocked second run would still
// see an empty account. gestartet signals the park, weiter releases it.
type blockierenderSetupClient struct {
	*tsetest.FakeSetupClient
	gestartet chan struct{}
	weiter    chan struct{}
}

func (c *blockierenderSetupClient) ListTSS(ctx context.Context) (tse.Umgebung, []tse.TSSInfo, error) {
	close(c.gestartet)
	<-c.weiter
	return c.FakeSetupClient.ListTSS(ctx)
}

// laufendeEinrichtung holds the lock while parked in ListTSS; freigeben lets it finish, fertig
// yields its result.
type laufendeEinrichtung struct {
	repo      *stubCommandRepo
	client    *blockierenderSetupClient
	fertig    <-chan error
	freigeben func()
}

// The release also runs in t.Cleanup and waits for the run, so a t.Fatalf never leaves the
// package-wide lock held or lets the run's own release unlock a later test. The lock reset,
// registered first, runs last as a safety net.
func starteBlockierteEinrichtung(t *testing.T) *laufendeEinrichtung {
	t.Helper()
	t.Cleanup(func() { einrichtungLaeuft.Store(false) })

	blockiert := &blockierenderSetupClient{
		FakeSetupClient: &tsetest.FakeSetupClient{
			UmgebungResponse:  tse.UmgebungTest,
			CreateTSSResponse: tse.TSSErstellt{ID: "tss-erste", PUK: "puk-123", State: "CREATED"},
		},
		gestartet: make(chan struct{}),
		weiter:    make(chan struct{}),
	}
	freigeben := sync.OnceFunc(func() { close(blockiert.weiter) })
	beendet := make(chan struct{})
	t.Cleanup(func() { freigeben(); <-beendet })

	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: uuid.New()}}
	erster := Command{
		TSERepo:             repo,
		KassensitzungenRepo: stubKassensitzungReader{},
		NewTSESetupClient: func(tse.SetupCredentials) (tse.SetupClient, error) {
			return blockiert, nil
		},
	}

	fertig := make(chan error, 1)
	go func() {
		// beendet closes only after RichteTSEEin and its deferred release have returned; the
		// cleanup waits for it.
		defer close(beendet)
		_, err := erster.RichteTSEEin(context.Background(), zugangsdaten(), tse.UmgebungTest, false)
		fertig <- err
	}()
	<-blockiert.gestartet

	return &laufendeEinrichtung{repo: repo, client: blockiert, fertig: fertig, freigeben: freigeben}
}

// A retry while an aborted lifecycle still runs must be refused before contacting fiskaly.
// Otherwise a second, paid TSS would be created and the saved configuration would not match the
// shown PUK/PIN.
func TestEinrichtung_ZweiterAufrufWaehrendLaufendemErstenAbgelehnt(t *testing.T) {
	lauf := starteBlockierteEinrichtung(t)

	// The second factory fails: building a client before the lock would yield
	// ErrTSEVerbindungFehlgeschlagen instead of ErrTSESetupLaeuftBereits.
	zweiterRepo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: uuid.New()}}
	zweiter := Command{
		TSERepo:             zweiterRepo,
		KassensitzungenRepo: stubKassensitzungReader{},
		NewTSESetupClient: func(tse.SetupCredentials) (tse.SetupClient, error) {
			return nil, errors.New("the rejected calls must not build a fiskaly client")
		},
	}

	if _, err := zweiter.RichteTSEEin(context.Background(), zugangsdaten(), tse.UmgebungTest, false); !errors.Is(err, ErrTSESetupLaeuftBereits) {
		t.Errorf("expected ErrTSESetupLaeuftBereits from the second setup, got %v", err)
	}
	if _, err := zweiter.UebernimmTSE(context.Background(), zugangsdaten(), tse.UmgebungTest, "tss-erste", "", ""); !errors.Is(err, ErrTSESetupLaeuftBereits) {
		t.Errorf("expected ErrTSESetupLaeuftBereits from a takeover while a setup runs, got %v", err)
	}
	if zweiterRepo.gespeichert != nil {
		t.Errorf("expected the rejected calls to save nothing, got %+v", zweiterRepo.gespeichert)
	}

	lauf.freigeben()
	if err := <-lauf.fertig; err != nil {
		t.Fatalf("unexpected error from the first setup: %v", err)
	}
	if len(lauf.client.ErstellteTSS) != 1 {
		t.Errorf("expected exactly one TSS to be created in total, got %+v", lauf.client.ErstellteTSS)
	}
	if lauf.repo.gespeichert == nil || lauf.repo.gespeichert.TssID != "tss-erste" {
		t.Errorf("expected the first setup to save its own configuration, got %+v", lauf.repo.gespeichert)
	}
}

// The manual credentials path must take the same lock, or the later writer wins
// and the instance signs against a TSS it was not set up with.
func TestUpdateTSEKonfiguration_WaehrendLaufenderEinrichtungAbgelehnt(t *testing.T) {
	lauf := starteBlockierteEinrichtung(t)

	manuellesRepo := &stubCommandRepo{}
	manuell := Command{TSERepo: manuellesRepo, KassensitzungenRepo: stubKassensitzungReader{}}
	konfiguration, err := tse.NewKonfiguration("api-key", "api-secret", "tss-von-hand", "client-von-hand")
	if err != nil {
		t.Fatalf("unexpected error building konfiguration: %v", err)
	}

	if err := manuell.UpdateTSEKonfiguration(context.Background(), konfiguration); !errors.Is(err, ErrTSESetupLaeuftBereits) {
		t.Errorf("expected ErrTSESetupLaeuftBereits while a setup runs, got %v", err)
	}
	if manuellesRepo.gespeichert != nil {
		t.Errorf("expected the rejected manual save to write nothing, got %+v", manuellesRepo.gespeichert)
	}

	lauf.freigeben()
	if err := <-lauf.fertig; err != nil {
		t.Fatalf("unexpected error from the running setup: %v", err)
	}
	if lauf.repo.gespeichert == nil || lauf.repo.gespeichert.TssID != "tss-erste" {
		t.Errorf("expected the setup to save its own configuration, got %+v", lauf.repo.gespeichert)
	}

	// After the run the lock is free again.
	if err := manuell.UpdateTSEKonfiguration(context.Background(), konfiguration); err != nil {
		t.Fatalf("expected the manual save after the setup to succeed, got %v", err)
	}
	if manuellesRepo.gespeichert == nil || manuellesRepo.gespeichert.TssID != "tss-von-hand" {
		t.Errorf("expected the manual save to store its configuration, got %+v", manuellesRepo.gespeichert)
	}
}

// The lock must be released after failure and success alike, or one fiskaly hiccup would lock setup
// for good.
func TestEinrichtung_SchlossIstNachFehlerUndNachErfolgWiederFrei(t *testing.T) {
	t.Cleanup(func() { einrichtungLaeuft.Store(false) })

	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: uuid.New()}}

	gescheitert := &tsetest.FakeSetupClient{TSSErr: errors.New("fiskaly nicht erreichbar")}
	if _, err := commandMit(repo, gescheitert).RichteTSEEin(context.Background(), zugangsdaten(), tse.UmgebungTest, false); !errors.Is(err, ErrTSEVerbindungFehlgeschlagen) {
		t.Errorf("expected ErrTSEVerbindungFehlgeschlagen, got %v", err)
	}

	erfolgreich := &tsetest.FakeSetupClient{
		UmgebungResponse:  tse.UmgebungTest,
		CreateTSSResponse: tse.TSSErstellt{ID: "tss-neu", PUK: "puk-123", State: "CREATED"},
	}
	if _, err := commandMit(repo, erfolgreich).RichteTSEEin(context.Background(), zugangsdaten(), tse.UmgebungTest, false); err != nil {
		t.Errorf("expected the setup after a failed run to start, got %v", err)
	}

	danach := &tsetest.FakeSetupClient{
		UmgebungResponse:  tse.UmgebungTest,
		CreateTSSResponse: tse.TSSErstellt{ID: "tss-danach", PUK: "puk-456", State: "CREATED"},
	}
	if _, err := commandMit(repo, danach).RichteTSEEin(context.Background(), zugangsdaten(), tse.UmgebungTest, false); err != nil {
		t.Errorf("expected the setup after a successful run to start, got %v", err)
	}
}
