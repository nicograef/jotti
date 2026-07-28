//go:build unit

package application

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/nicograef/jotti/backend/domain/tse"
)

// blockierenderSetupClient haelt einen laufenden Lebenszyklus in ListTSS fest —
// genau an der Stelle, an der ein zweiter Lauf ohne Schloss noch das leere Konto
// saehe und eine zweite, bezahlte TSS anlegte. gestartet meldet, dass der Lauf
// steht; weiter laesst ihn zu Ende laufen.
type blockierenderSetupClient struct {
	*tse.FakeSetupClient
	gestartet chan struct{}
	weiter    chan struct{}
}

func (c *blockierenderSetupClient) ListTSS(ctx context.Context) (tse.Umgebung, []tse.TSSInfo, error) {
	close(c.gestartet)
	<-c.weiter
	return c.FakeSetupClient.ListTSS(ctx)
}

// laufendeEinrichtung ist eine gestartete Einrichtung, die in ListTSS steht und
// dabei das Schloss haelt. freigeben laesst sie zu Ende laufen, fertig liefert
// danach ihr Ergebnis.
type laufendeEinrichtung struct {
	repo      *stubCommandRepo
	client    *blockierenderSetupClient
	fertig    <-chan error
	freigeben func()
}

// starteBlockierteEinrichtung startet eine Einrichtung und kehrt zurueck, sobald
// sie in ListTSS steht.
//
// Die Freigabe haengt zusaetzlich in t.Cleanup: Ein t.Fatalf zwischen Start und
// Freigabe beendet die Test-Goroutine per runtime.Goexit, der blockierte Lauf
// haenge sonst fuer immer und hielte das paketweite Schloss — jeder folgende
// Test des Pakets schluege dann mit ErrTSESetupLaeuftBereits fehl und
// verschleierte die eigentliche Ursache. sync.OnceFunc macht den doppelten
// Aufruf (regulaer im Test und aus dem Cleanup) unschaedlich; das Zuruecksetzen
// des Schlosses bleibt als letztes Sicherheitsnetz stehen.
func starteBlockierteEinrichtung(t *testing.T) *laufendeEinrichtung {
	t.Helper()
	t.Cleanup(func() { einrichtungLaeuft.Store(false) })

	blockiert := &blockierenderSetupClient{
		FakeSetupClient: &tse.FakeSetupClient{
			UmgebungResponse:  tse.UmgebungTest,
			CreateTSSResponse: tse.TSSErstellt{ID: "tss-erste", PUK: "puk-123", State: "CREATED"},
		},
		gestartet: make(chan struct{}),
		weiter:    make(chan struct{}),
	}
	freigeben := sync.OnceFunc(func() { close(blockiert.weiter) })
	t.Cleanup(freigeben)

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
		_, err := erster.RichteTSEEin(context.Background(), zugangsdaten(), tse.UmgebungTest, false)
		fertig <- err
	}()
	<-blockiert.gestartet

	return &laufendeEinrichtung{repo: repo, client: blockiert, fertig: fertig, freigeben: freigeben}
}

// Seit der Lebenszyklus vom Client-Abbruch entkoppelt ist, laeuft er nach einem
// Abbruch im Hintergrund weiter — der Admin sieht derweil eine Fehlermeldung und
// kann sofort erneut starten. Der zweite Aufruf muss deshalb sofort abgelehnt
// werden, ohne fiskaly auch nur anzusprechen: Sonst entstuende eine zweite,
// bezahlte TSS, und die zuletzt gespeicherte Konfiguration passte nicht mehr zu
// den angezeigten PUK/PIN. Neuanlage und Uebernahme teilen sich das Schloss.
func TestEinrichtung_ZweiterAufrufWaehrendLaufendemErstenAbgelehnt(t *testing.T) {
	lauf := starteBlockierteEinrichtung(t)

	// Der zweite Lauf darf fiskaly nicht einmal ansprechen. Die Fabrik zaehlt
	// jeden Versuch, einen Setup-Client zu bauen — der erste Schritt jeder
	// fiskaly-Sequenz und damit der schaerfste Nachweis.
	fabrikAufrufe := 0
	zweiterClient := &tse.FakeSetupClient{
		UmgebungResponse:  tse.UmgebungTest,
		CreateTSSResponse: tse.TSSErstellt{ID: "tss-zweite", PUK: "puk-456", State: "CREATED"},
	}
	zweiterRepo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: uuid.New()}}
	zweiter := Command{
		TSERepo:             zweiterRepo,
		KassensitzungenRepo: stubKassensitzungReader{},
		NewTSESetupClient: func(tse.SetupCredentials) (tse.SetupClient, error) {
			fabrikAufrufe++
			return zweiterClient, nil
		},
	}

	if _, err := zweiter.RichteTSEEin(context.Background(), zugangsdaten(), tse.UmgebungTest, false); !errors.Is(err, ErrTSESetupLaeuftBereits) {
		t.Fatalf("expected ErrTSESetupLaeuftBereits from the second setup, got %v", err)
	}
	if _, err := zweiter.UebernimmTSE(context.Background(), zugangsdaten(), tse.UmgebungTest, "tss-erste", "", ""); !errors.Is(err, ErrTSESetupLaeuftBereits) {
		t.Fatalf("expected ErrTSESetupLaeuftBereits from a takeover while a setup runs, got %v", err)
	}
	if fabrikAufrufe != 0 {
		t.Fatalf("expected the rejected calls to never build a fiskaly client, got %d", fabrikAufrufe)
	}
	if zweiterClient.CreateTSSCalls != 0 {
		t.Fatalf("expected no second TSS to be created, got %d calls", zweiterClient.CreateTSSCalls)
	}
	if zweiterRepo.gespeichert != nil {
		t.Fatalf("expected the rejected calls to save nothing, got %+v", zweiterRepo.gespeichert)
	}

	lauf.freigeben()
	if err := <-lauf.fertig; err != nil {
		t.Fatalf("unexpected error from the first setup: %v", err)
	}
	if lauf.client.CreateTSSCalls != 1 {
		t.Fatalf("expected exactly one TSS to be created in total, got %d", lauf.client.CreateTSSCalls)
	}
	if lauf.repo.gespeichert == nil || lauf.repo.gespeichert.TssID != "tss-erste" {
		t.Fatalf("expected the first setup to save its own configuration, got %+v", lauf.repo.gespeichert)
	}
}

// Der manuelle Zugangsdaten-Wechsel schreibt ueber denselben SaveEinrichtung wie
// die Einrichtung und liegt in der Oberflaeche direkt unter dem Wizard. Er muss
// deshalb dasselbe Schloss nehmen: Sonst speicherte der Admin waehrend eines
// laufenden Einrichtungslaufs von Hand eine Konfiguration, der spaetere
// Schreiber gewaenne, und die Instanz signierte anschliessend gegen eine
// TSS/Client-Kombination, die nicht die eingerichtete ist.
func TestUpdateTSEKonfiguration_WaehrendLaufenderEinrichtungAbgelehnt(t *testing.T) {
	lauf := starteBlockierteEinrichtung(t)

	manuellesRepo := &stubCommandRepo{}
	manuell := Command{TSERepo: manuellesRepo, KassensitzungenRepo: stubKassensitzungReader{}}
	konfiguration, err := tse.NewKonfiguration("api-key", "api-secret", "tss-von-hand", "client-von-hand")
	if err != nil {
		t.Fatalf("unexpected error building konfiguration: %v", err)
	}

	if err := manuell.UpdateTSEKonfiguration(context.Background(), konfiguration); !errors.Is(err, ErrTSESetupLaeuftBereits) {
		t.Fatalf("expected ErrTSESetupLaeuftBereits while a setup runs, got %v", err)
	}
	if manuellesRepo.gespeichert != nil {
		t.Fatalf("expected the rejected manual save to write nothing, got %+v", manuellesRepo.gespeichert)
	}

	lauf.freigeben()
	if err := <-lauf.fertig; err != nil {
		t.Fatalf("unexpected error from the running setup: %v", err)
	}
	if lauf.repo.gespeichert == nil || lauf.repo.gespeichert.TssID != "tss-erste" {
		t.Fatalf("expected the setup to save its own configuration, got %+v", lauf.repo.gespeichert)
	}

	// Nach dem Lauf ist das Schloss frei, der manuelle Pfad schreibt wieder.
	if err := manuell.UpdateTSEKonfiguration(context.Background(), konfiguration); err != nil {
		t.Fatalf("expected the manual save after the setup to succeed, got %v", err)
	}
	if manuellesRepo.gespeichert == nil || manuellesRepo.gespeichert.TssID != "tss-von-hand" {
		t.Fatalf("expected the manual save to store its configuration, got %+v", manuellesRepo.gespeichert)
	}
}

// Das Schloss darf keinen Pfad ueberdauern: Nach einem gescheiterten wie nach
// einem erfolgreichen Lauf muss die naechste Einrichtung wieder starten koennen.
// Sonst waere ein einziger fiskaly-Aussetzer eine dauerhafte Sperre.
func TestEinrichtung_SchlossIstNachFehlerUndNachErfolgWiederFrei(t *testing.T) {
	t.Cleanup(func() { einrichtungLaeuft.Store(false) })

	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: uuid.New()}}

	gescheitert := &tse.FakeSetupClient{TSSErr: errors.New("fiskaly nicht erreichbar")}
	if _, err := commandMit(repo, gescheitert).RichteTSEEin(context.Background(), zugangsdaten(), tse.UmgebungTest, false); !errors.Is(err, ErrTSEVerbindungFehlgeschlagen) {
		t.Fatalf("expected ErrTSEVerbindungFehlgeschlagen, got %v", err)
	}

	erfolgreich := &tse.FakeSetupClient{
		UmgebungResponse:  tse.UmgebungTest,
		CreateTSSResponse: tse.TSSErstellt{ID: "tss-neu", PUK: "puk-123", State: "CREATED"},
	}
	if _, err := commandMit(repo, erfolgreich).RichteTSEEin(context.Background(), zugangsdaten(), tse.UmgebungTest, false); err != nil {
		t.Fatalf("expected the setup after a failed run to start, got %v", err)
	}

	danach := &tse.FakeSetupClient{
		UmgebungResponse:  tse.UmgebungTest,
		CreateTSSResponse: tse.TSSErstellt{ID: "tss-danach", PUK: "puk-456", State: "CREATED"},
	}
	if _, err := commandMit(repo, danach).RichteTSEEin(context.Background(), zugangsdaten(), tse.UmgebungTest, false); err != nil {
		t.Fatalf("expected the setup after a successful run to start, got %v", err)
	}
}
