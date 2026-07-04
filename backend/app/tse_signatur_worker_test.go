//go:build unit

package app

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/nicograef/jotti/backend/domain/settings"
	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/nicograef/jotti/backend/repository/tse_repo"
)

type mockTSESettingsReader struct {
	conf settings.TSEKonfiguration
	err  error
}

func (m *mockTSESettingsReader) GetTSEKonfiguration(_ context.Context) (settings.TSEKonfiguration, error) {
	if m.err != nil {
		return settings.TSEKonfiguration{}, m.err
	}
	return m.conf, nil
}

type fehlversuch struct {
	AuftragID int
	Fehler    string
}

type quittierung struct {
	AuftragID int
	Signatur  tse.Signatur
}

type mockTSESignaturStore struct {
	mu            sync.Mutex
	offene        []tse_repo.OffenerSignaturauftrag
	quittierungen []quittierung
	fehlversuche  []fehlversuch
	getErr        error
	quittiereErr  error
	// verarbeitet signalisiert jede Quittierung (fuer Run-Loop-Tests ohne Sleeps).
	verarbeitet chan struct{}
}

func (m *mockTSESignaturStore) GetOffeneTSESignaturauftraege(_ context.Context, _ int) ([]tse_repo.OffenerSignaturauftrag, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.offene, nil
}

func (m *mockTSESignaturStore) QuittiereTSESignaturauftrag(_ context.Context, auftragID int, signatur tse.Signatur) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.quittiereErr != nil {
		return m.quittiereErr
	}
	m.quittierungen = append(m.quittierungen, quittierung{AuftragID: auftragID, Signatur: signatur})
	if m.verarbeitet != nil {
		select {
		case m.verarbeitet <- struct{}{}:
		default:
		}
	}
	return nil
}

func (m *mockTSESignaturStore) TSESignaturauftragFehlversuch(_ context.Context, auftragID int, fehler string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.fehlversuche = append(m.fehlversuche, fehlversuch{AuftragID: auftragID, Fehler: fehler})
	return nil
}

func configuredTSE() settings.TSEKonfiguration {
	return settings.TSEKonfiguration{
		ApiKey:    "api-key",
		ApiSecret: "api-secret",
		TssID:     "tss-1",
		ClientID:  "client-1",
		UpdatedAt: time.Now(),
	}
}

func neuerWorkerClient(fake tse.FakeClient) tseClientFactory {
	return func(_ tse.Credentials) (tseWorkerClient, error) {
		return fake, nil
	}
}

func TestTSESignaturWorker_ProcessOnce_Success(t *testing.T) {
	store := &mockTSESignaturStore{offene: []tse_repo.OffenerSignaturauftrag{{
		ID:          1,
		TxID:        "tx-1",
		ProcessType: "Kassenbeleg-V1",
		ProcessData: "Beleg^3.50",
	}}}

	worker := &tseSignaturWorker{
		settingsRepo: &mockTSESettingsReader{conf: configuredTSE()},
		store:        store,
		newTSEClient: neuerWorkerClient(tse.FakeClient{
			RetrieveErr:   tse.ErrTransactionNichtGefunden,
			StartResponse: tse.StartResult{TransactionNumber: 41, LogTime: time.Date(2026, 6, 10, 18, 0, 1, 0, time.UTC)},
			FinishResponse: tse.FinishResult{
				TransactionNumber: 41,
				SignatureCounter:  700,
				SerialNumberTSE:   "TSE-SN-1",
				LogTimeEnd:        time.Date(2026, 6, 10, 18, 0, 2, 0, time.UTC),
				Signature:         "SIG-1",
				QRCodeData:        "V0;QR",
			},
		}),
		now: time.Now,
	}

	if err := worker.processOnce(context.Background()); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(store.quittierungen) != 1 {
		t.Fatalf("expected one quittierung, got %d", len(store.quittierungen))
	}
	if store.quittierungen[0].AuftragID != 1 {
		t.Fatalf("expected auftrag 1, got %d", store.quittierungen[0].AuftragID)
	}
	if store.quittierungen[0].Signatur.Signatur != "SIG-1" {
		t.Fatalf("expected SIG-1, got %q", store.quittierungen[0].Signatur.Signatur)
	}
	if !store.quittierungen[0].Signatur.LogTimeStart.Equal(time.Date(2026, 6, 10, 18, 0, 1, 0, time.UTC)) {
		t.Fatalf("expected log_time_start from start result, got %v", store.quittierungen[0].Signatur.LogTimeStart)
	}
	if len(store.fehlversuche) != 0 {
		t.Fatalf("expected no fehlversuche on success, got %d", len(store.fehlversuche))
	}
}

func TestTSESignaturWorker_ProcessOnce_TSEErrorVerbuchtFehlversuch(t *testing.T) {
	store := &mockTSESignaturStore{offene: []tse_repo.OffenerSignaturauftrag{{
		ID:          2,
		TxID:        "tx-2",
		ProcessType: "Kassenbeleg-V1",
		ProcessData: "Beleg^5.00",
	}}}

	worker := &tseSignaturWorker{
		settingsRepo: &mockTSESettingsReader{conf: configuredTSE()},
		store:        store,
		newTSEClient: neuerWorkerClient(tse.FakeClient{
			RetrieveErr: tse.ErrTransactionNichtGefunden,
			StartErr:    errors.New("fiskaly timeout"),
		}),
		now: time.Now,
	}

	if err := worker.processOnce(context.Background()); err != nil {
		t.Fatalf("expected no hard error, got %v", err)
	}

	if len(store.quittierungen) != 0 {
		t.Fatalf("expected no quittierung on TSE error, got %d", len(store.quittierungen))
	}
	if len(store.fehlversuche) != 1 {
		t.Fatalf("expected one fehlversuch, got %d", len(store.fehlversuche))
	}
	if store.fehlversuche[0].AuftragID != 2 {
		t.Fatalf("expected fehlversuch for auftrag 2, got %d", store.fehlversuche[0].AuftragID)
	}
	if store.fehlversuche[0].Fehler == "" {
		t.Fatal("expected fehlversuch to carry the error message")
	}
}

// Heilt das 409-Szenario: Die Transaktion wurde bei fiskaly bereits
// abgeschlossen (z. B. Abbruch zwischen Signierung und Quittierung). Der
// Worker uebernimmt die vorhandene Signatur, ohne neu zu signieren —
// Start/Finish wuerden in diesem Test fehlschlagen.
func TestTSESignaturWorker_ProcessOnce_BereitsFinishedWirdQuittiert(t *testing.T) {
	store := &mockTSESignaturStore{offene: []tse_repo.OffenerSignaturauftrag{{
		ID:          3,
		TxID:        "tx-3",
		ProcessType: "Kassenbeleg-V1",
		ProcessData: "Beleg^7.00",
	}}}

	worker := &tseSignaturWorker{
		settingsRepo: &mockTSESettingsReader{conf: configuredTSE()},
		store:        store,
		newTSEClient: neuerWorkerClient(tse.FakeClient{
			StartErr:  errors.New("409 E_TX_NO_TYPE_DEFINED"),
			FinishErr: errors.New("409 E_TX_NO_TYPE_DEFINED"),
			RetrieveResponse: tse.RetrieveResult{
				State: tse.TransactionStateFinished,
				FinishResult: tse.FinishResult{
					TransactionNumber: 43,
					SignatureCounter:  702,
					SerialNumberTSE:   "TSE-SN-3",
					LogTimeStart:      time.Date(2026, 6, 10, 18, 5, 1, 0, time.UTC),
					LogTimeEnd:        time.Date(2026, 6, 10, 18, 5, 2, 0, time.UTC),
					Signature:         "SIG-3",
					QRCodeData:        "V0;QR-3",
				},
			},
		}),
		now: time.Now,
	}

	if err := worker.processOnce(context.Background()); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(store.fehlversuche) != 0 {
		t.Fatalf("expected no fehlversuch, got %+v", store.fehlversuche)
	}
	if len(store.quittierungen) != 1 {
		t.Fatalf("expected one quittierung, got %d", len(store.quittierungen))
	}
	signatur := store.quittierungen[0].Signatur
	if signatur.Signatur != "SIG-3" || signatur.TransaktionNummer != 43 || signatur.SignaturZaehler != 702 {
		t.Fatalf("expected retrieved signature data, got %+v", signatur)
	}
	if !signatur.LogTimeStart.Equal(time.Date(2026, 6, 10, 18, 5, 1, 0, time.UTC)) {
		t.Fatalf("expected retrieved log_time_start, got %v", signatur.LogTimeStart)
	}
}

// Eine bei fiskaly noch aktive Transaktion (Start kam durch, Finish nicht)
// wird nur noch abgeschlossen — ein erneuter Start wuerde fehlschlagen.
func TestTSESignaturWorker_ProcessOnce_AktiveTransaktionWirdAbgeschlossen(t *testing.T) {
	store := &mockTSESignaturStore{offene: []tse_repo.OffenerSignaturauftrag{{
		ID:          4,
		TxID:        "tx-4",
		ProcessType: "Kassenbeleg-V1",
		ProcessData: "Beleg^9.00",
	}}}

	worker := &tseSignaturWorker{
		settingsRepo: &mockTSESettingsReader{conf: configuredTSE()},
		store:        store,
		newTSEClient: neuerWorkerClient(tse.FakeClient{
			StartErr: errors.New("409 transaction already started"),
			RetrieveResponse: tse.RetrieveResult{
				State: tse.TransactionStateActive,
				FinishResult: tse.FinishResult{
					TransactionNumber: 44,
					LogTimeStart:      time.Date(2026, 6, 10, 18, 7, 1, 0, time.UTC),
				},
			},
			FinishResponse: tse.FinishResult{
				TransactionNumber: 44,
				SignatureCounter:  710,
				SerialNumberTSE:   "TSE-SN-4",
				LogTimeEnd:        time.Date(2026, 6, 10, 18, 7, 5, 0, time.UTC),
				Signature:         "SIG-4",
				QRCodeData:        "V0;QR-4",
			},
		}),
		now: time.Now,
	}

	if err := worker.processOnce(context.Background()); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(store.fehlversuche) != 0 {
		t.Fatalf("expected no fehlversuch, got %+v", store.fehlversuche)
	}
	if len(store.quittierungen) != 1 {
		t.Fatalf("expected one quittierung, got %d", len(store.quittierungen))
	}
	signatur := store.quittierungen[0].Signatur
	if signatur.Signatur != "SIG-4" || signatur.TransaktionNummer != 44 {
		t.Fatalf("expected finish signature data, got %+v", signatur)
	}
	if !signatur.LogTimeStart.Equal(time.Date(2026, 6, 10, 18, 7, 1, 0, time.UTC)) {
		t.Fatalf("expected log_time_start from retrieved transaction, got %v", signatur.LogTimeStart)
	}
}

// Der TSE-Client wird ueber Durchlaeufe hinweg wiederverwendet (samt
// Auth-Token) und nur bei geaenderten Zugangsdaten neu gebaut.
func TestTSESignaturWorker_ClientWiederverwendung(t *testing.T) {
	settingsRepo := &mockTSESettingsReader{conf: configuredTSE()}
	factoryCalls := 0
	worker := &tseSignaturWorker{
		settingsRepo: settingsRepo,
		store:        &mockTSESignaturStore{},
		newTSEClient: func(_ tse.Credentials) (tseWorkerClient, error) {
			factoryCalls++
			return tse.FakeClient{}, nil
		},
		now: time.Now,
	}

	if err := worker.processOnce(context.Background()); err != nil {
		t.Fatalf("first run failed: %v", err)
	}
	if err := worker.processOnce(context.Background()); err != nil {
		t.Fatalf("second run failed: %v", err)
	}
	if factoryCalls != 1 {
		t.Fatalf("expected client to be reused (1 factory call), got %d", factoryCalls)
	}

	settingsRepo.conf.ApiSecret = "rotated-secret"
	if err := worker.processOnce(context.Background()); err != nil {
		t.Fatalf("third run failed: %v", err)
	}
	if factoryCalls != 2 {
		t.Fatalf("expected client rebuild after credential change, got %d factory calls", factoryCalls)
	}
}

// runWorker startet den Run-Loop und liefert cancel + done fuer den Abbau.
func runWorker(t *testing.T, worker *tseSignaturWorker) (context.CancelFunc, <-chan struct{}) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		worker.run(ctx)
		close(done)
	}()
	return cancel, done
}

func signierenderFakeClient() tseClientFactory {
	return neuerWorkerClient(tse.FakeClient{
		RetrieveErr:    tse.ErrTransactionNichtGefunden,
		StartResponse:  tse.StartResult{TransactionNumber: 50, LogTime: time.Date(2026, 6, 10, 19, 0, 1, 0, time.UTC)},
		FinishResponse: tse.FinishResult{TransactionNumber: 50, SignatureCounter: 800, SerialNumberTSE: "TSE-SN", Signature: "SIG"},
	})
}

// Der Sofort-Trigger nach einem Commit stoesst den Durchlauf ohne Warten auf
// den Polling-Tick an (Tick steht auf einer Stunde).
func TestTSESignaturWorker_Run_SofortTrigger(t *testing.T) {
	store := &mockTSESignaturStore{
		offene:      []tse_repo.OffenerSignaturauftrag{{ID: 6, TxID: "tx-6", ProcessType: "Kassenbeleg-V1", ProcessData: "Beleg^1.00"}},
		verarbeitet: make(chan struct{}, 1),
	}
	trigger := make(chan struct{}, 1)
	worker := &tseSignaturWorker{
		settingsRepo: &mockTSESettingsReader{conf: configuredTSE()},
		store:        store,
		newTSEClient: signierenderFakeClient(),
		trigger:      trigger,
		pollInterval: time.Hour,
		now:          time.Now,
	}

	cancel, done := runWorker(t, worker)
	defer func() { cancel(); <-done }()

	trigger <- struct{}{}

	select {
	case <-store.verarbeitet:
	case <-time.After(5 * time.Second):
		t.Fatal("Sofort-Trigger hat keinen Durchlauf angestossen")
	}
}

// Der Polling-Tick faengt verlorene Trigger (Crash zwischen Commit und
// Trigger): Ohne jeden Trigger wird der offene Auftrag am Tick verarbeitet.
func TestTSESignaturWorker_Run_PollingFallbackFaengtVerloreneTrigger(t *testing.T) {
	store := &mockTSESignaturStore{
		offene:      []tse_repo.OffenerSignaturauftrag{{ID: 7, TxID: "tx-7", ProcessType: "Kassenbeleg-V1", ProcessData: "Beleg^2.00"}},
		verarbeitet: make(chan struct{}, 1),
	}
	worker := &tseSignaturWorker{
		settingsRepo: &mockTSESettingsReader{conf: configuredTSE()},
		store:        store,
		newTSEClient: signierenderFakeClient(),
		trigger:      make(chan struct{}), // nie ausgeloest
		pollInterval: 10 * time.Millisecond,
		now:          time.Now,
	}

	cancel, done := runWorker(t, worker)
	defer func() { cancel(); <-done }()

	select {
	case <-store.verarbeitet:
	case <-time.After(5 * time.Second):
		t.Fatal("Polling-Fallback hat den offenen Auftrag nicht verarbeitet")
	}
}
