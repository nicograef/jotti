package signatur

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/nicograef/jotti/backend/domain/tse/tsetest"
	"github.com/nicograef/jotti/backend/repository/tse_repo"
)

type mockTSESettingsReader struct {
	conf tse.Konfiguration
	err  error
}

func (m *mockTSESettingsReader) GetTSEKonfiguration(_ context.Context) (tse.Konfiguration, error) {
	if m.err != nil {
		return tse.Konfiguration{}, m.err
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
	mu                sync.Mutex
	offene            []tse_repo.OffenerSignaturauftrag
	quittierungen     []quittierung
	fehlversuche      []fehlversuch
	geoeffnet         []string // Grund-Arten of opened periods
	geschlossen       []string // Grund-Arten of closed periods
	nichtKonfiguriert []tse_repo.OffenerSignaturauftrag
	getErr            error
	quittiereErr      error
	// verarbeitet signals each acknowledgement so run-loop tests need no sleeps.
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

func (m *mockTSESignaturStore) MarkOffeneAlsNichtKonfiguriert(_ context.Context) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	markiert := int64(len(m.offene))
	m.nichtKonfiguriert = append(m.nichtKonfiguriert, m.offene...)
	m.offene = nil
	return markiert, nil
}

func (m *mockTSESignaturStore) OpenTSEStoerung(_ context.Context, grundArt string, _ string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.geoeffnet = append(m.geoeffnet, grundArt)
	return nil
}

func (m *mockTSESignaturStore) CloseTSEStoerung(_ context.Context, grundArt string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.geschlossen = append(m.geschlossen, grundArt)
	return nil
}

func configuredTSE() tse.Konfiguration {
	return tse.Konfiguration{
		ApiKey:    "api-key",
		ApiSecret: "api-secret",
		TssID:     "tss-1",
		ClientID:  "client-1",
		UpdatedAt: time.Now(),
	}
}

func newWorkerClient(fake tsetest.FakeClient) tseClientFactory {
	return func(_ tse.Credentials) (tseWorkerClient, error) {
		return fake, nil
	}
}

// zaehlenderClient counts fiskaly calls to prove a run aborts or the outage state stays silent.
type zaehlenderClient struct {
	tsetest.FakeClient
	mu    sync.Mutex
	calls int
}

func (c *zaehlenderClient) zaehle() {
	c.mu.Lock()
	c.calls++
	c.mu.Unlock()
}

func (c *zaehlenderClient) anzahlCalls() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.calls
}

func (c *zaehlenderClient) RetrieveTransaction(ctx context.Context, txID string) (tse.RetrieveResult, error) {
	c.zaehle()
	return c.FakeClient.RetrieveTransaction(ctx, txID)
}

func (c *zaehlenderClient) StartTransaction(ctx context.Context, txID string) (tse.StartResult, error) {
	c.zaehle()
	return c.FakeClient.StartTransaction(ctx, txID)
}

func (c *zaehlenderClient) FinishTransaction(ctx context.Context, txID string, processType string, processData string) (tse.FinishResult, error) {
	c.zaehle()
	return c.FakeClient.FinishTransaction(ctx, txID, processType, processData)
}

// txAbhaengigerClient fails only the txIDs it holds an error for, for poison-job tests.
type txAbhaengigerClient struct {
	ablehnungen map[string]error
}

func (c txAbhaengigerClient) RetrieveTransaction(_ context.Context, _ string) (tse.RetrieveResult, error) {
	return tse.RetrieveResult{}, tse.ErrTransactionNichtGefunden
}

func (c txAbhaengigerClient) StartTransaction(_ context.Context, txID string) (tse.StartResult, error) {
	if err, ok := c.ablehnungen[txID]; ok {
		return tse.StartResult{}, err
	}
	return tse.StartResult{TransactionNumber: 60, LogTime: time.Date(2026, 6, 10, 20, 0, 1, 0, time.UTC)}, nil
}

func (c txAbhaengigerClient) FinishTransaction(_ context.Context, txID string, _ string, _ string) (tse.FinishResult, error) {
	if err, ok := c.ablehnungen[txID]; ok {
		return tse.FinishResult{}, err
	}
	return tse.FinishResult{TransactionNumber: 60, SignatureCounter: 900, SerialNumberTSE: "TSE-SN", Signature: "SIG-" + txID}, nil
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
		newTSEClient: newWorkerClient(tsetest.FakeClient{
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
		t.Errorf("expected auftrag 1, got %d", store.quittierungen[0].AuftragID)
	}
	if store.quittierungen[0].Signatur.Signatur != "SIG-1" {
		t.Errorf("expected SIG-1, got %q", store.quittierungen[0].Signatur.Signatur)
	}
	if !store.quittierungen[0].Signatur.LogTimeStart.Equal(time.Date(2026, 6, 10, 18, 0, 1, 0, time.UTC)) {
		t.Errorf("expected log_time_start from start result, got %v", store.quittierungen[0].Signatur.LogTimeStart)
	}
	if len(store.fehlversuche) != 0 {
		t.Errorf("expected no fehlversuche on success, got %d", len(store.fehlversuche))
	}
}

// A TSE-wide error aborts the run at the first job without counting attempts.
// It opens the tse_fehler period and enters the outage state.
func TestTSESignaturWorker_ProcessOnce_TSEWeiterFehlerBrichtDurchlaufAb(t *testing.T) {
	store := &mockTSESignaturStore{offene: []tse_repo.OffenerSignaturauftrag{
		{ID: 2, TxID: "tx-2", ProcessType: "Kassenbeleg-V1", ProcessData: "Beleg^5.00"},
		{ID: 3, TxID: "tx-3", ProcessType: "Kassenbeleg-V1", ProcessData: "Beleg^6.00"},
	}}
	client := &zaehlenderClient{FakeClient: tsetest.FakeClient{RetrieveErr: errors.New("connection refused")}}
	jetzt := time.Date(2026, 6, 10, 18, 0, 0, 0, time.UTC)

	worker := &tseSignaturWorker{
		settingsRepo: &mockTSESettingsReader{conf: configuredTSE()},
		store:        store,
		newTSEClient: func(_ tse.Credentials) (tseWorkerClient, error) { return client, nil },
		now:          func() time.Time { return jetzt },
	}

	if err := worker.processOnce(context.Background()); err == nil {
		t.Fatal("expected TSE-weiten Fehler als Durchlauf-Fehler")
	}

	if client.anzahlCalls() != 1 {
		t.Errorf("expected abort after first auftrag (1 fiskaly call), got %d", client.anzahlCalls())
	}
	if len(store.fehlversuche) != 0 {
		t.Errorf("expected no auftrags-fehlversuche on TSE-weitem Fehler, got %+v", store.fehlversuche)
	}
	if len(store.quittierungen) != 0 {
		t.Errorf("expected no quittierungen, got %d", len(store.quittierungen))
	}
	if len(store.geoeffnet) != 1 || store.geoeffnet[0] != tse.StoerungGrundTSEFehler {
		t.Errorf("expected geoeffneten tse_fehler-Zeitraum, got %v", store.geoeffnet)
	}
	if worker.stoerungSerie != 1 {
		t.Errorf("expected fehlerserie 1, got %d", worker.stoerungSerie)
	}
	if !worker.stoerungNaechsterVersuch.Equal(jetzt.Add(5 * time.Second)) {
		t.Errorf("expected naechsten Versuch nach 5s Backoff, got %v", worker.stoerungNaechsterVersuch)
	}
}

// A job-specific error counts an attempt and skips the job; the next job still signs in the same run.
// A poison job never blocks the queue or opens an outage.
func TestTSESignaturWorker_ProcessOnce_AuftragsFehlerUeberspringtUndSigniertWeiter(t *testing.T) {
	store := &mockTSESignaturStore{offene: []tse_repo.OffenerSignaturauftrag{
		{ID: 10, TxID: "tx-gift", ProcessType: "Kassenbeleg-V1", ProcessData: "kaputt"},
		{ID: 11, TxID: "tx-ok", ProcessType: "Kassenbeleg-V1", ProcessData: "Beleg^8.00"},
	}}
	client := txAbhaengigerClient{ablehnungen: map[string]error{
		"tx-gift": tse.AuftragsFehler{Err: errors.New("fiskaly api error 400 (E_FAILED_SCHEMA_VALIDATION)")},
	}}

	worker := &tseSignaturWorker{
		settingsRepo: &mockTSESettingsReader{conf: configuredTSE()},
		store:        store,
		newTSEClient: func(_ tse.Credentials) (tseWorkerClient, error) { return client, nil },
		now:          time.Now,
	}

	if err := worker.processOnce(context.Background()); err != nil {
		t.Fatalf("expected no durchlauf error, got %v", err)
	}

	if len(store.fehlversuche) != 1 || store.fehlversuche[0].AuftragID != 10 {
		t.Errorf("expected one fehlversuch for auftrag 10, got %+v", store.fehlversuche)
	}
	if len(store.quittierungen) != 1 || store.quittierungen[0].AuftragID != 11 {
		t.Errorf("expected auftrag 11 signed in same run, got %+v", store.quittierungen)
	}
	if len(store.geoeffnet) != 0 {
		t.Errorf("expected no stoerung on auftragsspezifischem Fehler, got %v", store.geoeffnet)
	}
	if worker.stoerungSerie != 0 {
		t.Errorf("expected keine fehlerserie, got %d", worker.stoerungSerie)
	}
}

// A CANCELLED transaction at fiskaly concerns only its job: a failed attempt, not an aborted run.
func TestTSESignaturWorker_ProcessOnce_UnerwarteterZustandIstAuftragsFehler(t *testing.T) {
	store := &mockTSESignaturStore{offene: []tse_repo.OffenerSignaturauftrag{{
		ID:          12,
		TxID:        "tx-cancelled",
		ProcessType: "Kassenbeleg-V1",
		ProcessData: "Beleg^1.00",
	}}}

	worker := &tseSignaturWorker{
		settingsRepo: &mockTSESettingsReader{conf: configuredTSE()},
		store:        store,
		newTSEClient: newWorkerClient(tsetest.FakeClient{
			RetrieveResponse: tse.RetrieveResult{State: tse.TransactionStateCancelled},
		}),
		now: time.Now,
	}

	if err := worker.processOnce(context.Background()); err != nil {
		t.Fatalf("expected no durchlauf error, got %v", err)
	}
	if len(store.fehlversuche) != 1 || store.fehlversuche[0].AuftragID != 12 {
		t.Errorf("expected fehlversuch for auftrag 12, got %+v", store.fehlversuche)
	}
	if len(store.geoeffnet) != 0 {
		t.Errorf("expected no stoerung, got %v", store.geoeffnet)
	}
}

// During backoff no trigger or tick calls fiskaly; afterwards the half-open probe grows the backoff on failure.
// On success the full backlog signs and the first signature closes the period.
func TestTSESignaturWorker_StoerungBackoffUndHalfOpenProbe(t *testing.T) {
	store := &mockTSESignaturStore{offene: []tse_repo.OffenerSignaturauftrag{
		{ID: 20, TxID: "tx-20", ProcessType: "Kassenbeleg-V1", ProcessData: "Beleg^1.00"},
		{ID: 21, TxID: "tx-21", ProcessType: "Kassenbeleg-V1", ProcessData: "Beleg^2.00"},
	}}
	client := &zaehlenderClient{FakeClient: tsetest.FakeClient{RetrieveErr: errors.New("503 service unavailable")}}
	jetzt := time.Date(2026, 6, 10, 18, 0, 0, 0, time.UTC)

	worker := &tseSignaturWorker{
		settingsRepo: &mockTSESettingsReader{conf: configuredTSE()},
		store:        store,
		newTSEClient: func(_ tse.Credentials) (tseWorkerClient, error) { return client, nil },
		now:          func() time.Time { return jetzt },
	}
	ctx := context.Background()

	// Run 1: TSE-wide error, outage state, 5s backoff.
	if err := worker.processOnce(ctx); err == nil {
		t.Fatal("expected TSE-weiten Fehler")
	}

	// No fiskaly call during the backoff.
	jetzt = jetzt.Add(2 * time.Second)
	callsVorher := client.anzahlCalls()
	if err := worker.processOnce(ctx); err != nil {
		t.Fatalf("expected gated durchlauf without error, got %v", err)
	}
	if client.anzahlCalls() != callsVorher {
		t.Errorf("expected no fiskaly calls during stoerung, got %d new", client.anzahlCalls()-callsVorher)
	}

	// The probe fails TSE-wide: one call, the backoff grows from 5s to 10s.
	jetzt = jetzt.Add(4 * time.Second)
	callsVorher = client.anzahlCalls()
	if err := worker.processOnce(ctx); err == nil {
		t.Fatal("expected TSE-weiten Fehler der Probe")
	}
	if client.anzahlCalls() != callsVorher+1 {
		t.Errorf("expected exactly one probe call, got %d", client.anzahlCalls()-callsVorher)
	}
	if worker.stoerungSerie != 2 {
		t.Errorf("expected fehlerserie 2, got %d", worker.stoerungSerie)
	}
	if !worker.stoerungNaechsterVersuch.Equal(jetzt.Add(10 * time.Second)) {
		t.Errorf("expected gewachsenen Backoff 10s, got %v", worker.stoerungNaechsterVersuch.Sub(jetzt))
	}

	// The probe succeeds: both jobs sign, the period closes and the series resets.
	client.FakeClient = tsetest.FakeClient{
		RetrieveErr:    tse.ErrTransactionNichtGefunden,
		StartResponse:  tse.StartResult{TransactionNumber: 70, LogTime: jetzt},
		FinishResponse: tse.FinishResult{TransactionNumber: 70, SignatureCounter: 900, SerialNumberTSE: "TSE-SN", Signature: "SIG"},
	}
	jetzt = jetzt.Add(11 * time.Second)
	if err := worker.processOnce(ctx); err != nil {
		t.Fatalf("expected recovery durchlauf without error, got %v", err)
	}
	if len(store.quittierungen) != 2 {
		t.Errorf("expected volle Aufarbeitung (2 quittierungen), got %d", len(store.quittierungen))
	}
	if len(store.geschlossen) != 1 || store.geschlossen[0] != tse.StoerungGrundTSEFehler {
		t.Errorf("expected geschlossenen tse_fehler-Zeitraum, got %v", store.geschlossen)
	}
	if worker.stoerungSerie != 0 || !worker.stoerungNaechsterVersuch.IsZero() {
		t.Errorf("expected zurueckgesetzten Stoerungszustand, got serie=%d next=%v", worker.stoerungSerie, worker.stoerungNaechsterVersuch)
	}
}

// The outage backoff is deterministic: 5s, doubled per series, capped at 2 minutes.
func TestTSEStoerungBackoff_DeterministischeKurve(t *testing.T) {
	tests := []struct {
		serie    int
		erwartet time.Duration
	}{
		{1, 5 * time.Second},
		{2, 10 * time.Second},
		{3, 20 * time.Second},
		{4, 40 * time.Second},
		{5, 80 * time.Second},
		{6, 2 * time.Minute},
		{7, 2 * time.Minute},
		{1000, 2 * time.Minute},
	}
	for _, tt := range tests {
		if got := tseStoerungBackoff(tt.serie); got != tt.erwartet {
			t.Errorf("serie %d: expected %v, got %v", tt.serie, tt.erwartet, got)
		}
	}
}

// A hanging fiskaly call hits the run deadline and counts as a TSE-wide error instead of blocking the worker.
func TestTSESignaturWorker_ProcessOnce_DurchlaufDeadlineBrichtAb(t *testing.T) {
	store := &mockTSESignaturStore{offene: []tse_repo.OffenerSignaturauftrag{{
		ID:          30,
		TxID:        "tx-30",
		ProcessType: "Kassenbeleg-V1",
		ProcessData: "Beleg^1.00",
	}}}

	worker := &tseSignaturWorker{
		settingsRepo:      &mockTSESettingsReader{conf: configuredTSE()},
		store:             store,
		newTSEClient:      newWorkerClient(tsetest.FakeClient{ArtificialDelay: time.Minute}),
		durchlaufDeadline: 30 * time.Millisecond,
		now:               time.Now,
	}

	if err := worker.processOnce(context.Background()); err == nil {
		t.Fatal("expected deadline abort as durchlauf error")
	}
	if len(store.fehlversuche) != 0 {
		t.Errorf("expected no fehlversuche on deadline abort, got %+v", store.fehlversuche)
	}
	if len(store.geoeffnet) != 1 || store.geoeffnet[0] != tse.StoerungGrundTSEFehler {
		t.Errorf("expected geoeffneten tse_fehler-Zeitraum, got %v", store.geoeffnet)
	}
	if worker.stoerungSerie != 1 {
		t.Errorf("expected fehlerserie 1, got %d", worker.stoerungSerie)
	}
}

// Heals the 409 case: a transaction already finished at fiskaly is adopted without re-signing.
// Start and Finish would fail in this test.
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
		newTSEClient: newWorkerClient(tsetest.FakeClient{
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
		t.Errorf("expected no fehlversuch, got %+v", store.fehlversuche)
	}
	if len(store.quittierungen) != 1 {
		t.Fatalf("expected one quittierung, got %d", len(store.quittierungen))
	}
	signatur := store.quittierungen[0].Signatur
	if signatur.Signatur != "SIG-3" || signatur.TransaktionNummer != 43 || signatur.SignaturZaehler != 702 {
		t.Errorf("expected retrieved signature data, got %+v", signatur)
	}
	if !signatur.LogTimeStart.Equal(time.Date(2026, 6, 10, 18, 5, 1, 0, time.UTC)) {
		t.Errorf("expected retrieved log_time_start, got %v", signatur.LogTimeStart)
	}
}

// A transaction still active at fiskaly is only finished; a second Start would fail.
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
		newTSEClient: newWorkerClient(tsetest.FakeClient{
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
		t.Errorf("expected no fehlversuch, got %+v", store.fehlversuche)
	}
	if len(store.quittierungen) != 1 {
		t.Fatalf("expected one quittierung, got %d", len(store.quittierungen))
	}
	signatur := store.quittierungen[0].Signatur
	if signatur.Signatur != "SIG-4" || signatur.TransaktionNummer != 44 {
		t.Errorf("expected finish signature data, got %+v", signatur)
	}
	if !signatur.LogTimeStart.Equal(time.Date(2026, 6, 10, 18, 7, 1, 0, time.UTC)) {
		t.Errorf("expected log_time_start from retrieved transaction, got %v", signatur.LogTimeStart)
	}
}

// The TSE client and its auth token survive runs and are rebuilt only on changed credentials.
func TestTSESignaturWorker_ClientWiederverwendung(t *testing.T) {
	settingsRepo := &mockTSESettingsReader{conf: configuredTSE()}
	worker := &tseSignaturWorker{
		settingsRepo: settingsRepo,
		store:        &mockTSESignaturStore{},
		newTSEClient: func(_ tse.Credentials) (tseWorkerClient, error) {
			return &tsetest.FakeClient{}, nil
		},
		now: time.Now,
	}

	if err := worker.processOnce(context.Background()); err != nil {
		t.Fatalf("first run failed: %v", err)
	}
	ersterClient := worker.client
	if err := worker.processOnce(context.Background()); err != nil {
		t.Fatalf("second run failed: %v", err)
	}
	if worker.client != ersterClient {
		t.Error("expected client to be reused across runs")
	}

	settingsRepo.conf.ApiSecret = "rotated-secret"
	if err := worker.processOnce(context.Background()); err != nil {
		t.Fatalf("third run failed: %v", err)
	}
	if worker.client == ersterClient || worker.clientCreds.ApiSecret != "rotated-secret" {
		t.Errorf("expected client rebuild after credential change, got creds %+v", worker.clientCreds)
	}
}

// Without a configuration, open jobs become tse_nicht_konfiguriert and keine_konfiguration opens.
// Covers both a missing row (db.ErrNotFound) and an empty one.
func TestTSESignaturWorker_ProcessOnce_OhneKonfigurationMarkiertEndgueltig(t *testing.T) {
	tests := []struct {
		name         string
		settingsRepo *mockTSESettingsReader
	}{
		{name: "keine Zeile", settingsRepo: &mockTSESettingsReader{err: db.ErrNotFound}},
		{name: "leere Konfiguration", settingsRepo: &mockTSESettingsReader{conf: tse.Konfiguration{}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &mockTSESignaturStore{offene: []tse_repo.OffenerSignaturauftrag{{ID: 1, TxID: "tx-1"}, {ID: 2, TxID: "tx-2"}}}
			worker := &tseSignaturWorker{
				settingsRepo: tt.settingsRepo,
				store:        store,
				newTSEClient: newWorkerClient(tsetest.FakeClient{}),
				now:          time.Now,
			}

			if err := worker.processOnce(context.Background()); err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if len(store.nichtKonfiguriert) != 2 || len(store.offene) != 0 {
				t.Errorf("expected both offene Auftraege marked, got marked %+v, offen %+v", store.nichtKonfiguriert, store.offene)
			}
			if len(store.geoeffnet) != 1 || store.geoeffnet[0] != tse.StoerungGrundKeineKonfiguration {
				t.Errorf("expected geoeffneten keine_konfiguration-Zeitraum, got %v", store.geoeffnet)
			}
			if len(store.quittierungen) != 0 {
				t.Errorf("expected no quittierungen without configuration, got %d", len(store.quittierungen))
			}
		})
	}
}

// With no open job, no period opens: keine_konfiguration records real transactions, not a fresh unused install.
func TestTSESignaturWorker_ProcessOnce_OhneKonfigurationOhneAuftraegeKeineStoerung(t *testing.T) {
	store := &mockTSESignaturStore{}
	worker := &tseSignaturWorker{
		settingsRepo: &mockTSESettingsReader{err: db.ErrNotFound},
		store:        store,
		newTSEClient: newWorkerClient(tsetest.FakeClient{}),
		now:          time.Now,
	}

	if err := worker.processOnce(context.Background()); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(store.geoeffnet) != 0 {
		t.Errorf("expected keinen Stoerungszeitraum ohne markierte Auftraege, got %v", store.geoeffnet)
	}
}

// A configuration read error is transient: the worker marks nothing and returns the error.
func TestTSESignaturWorker_ProcessOnce_NichtLesbareKonfigurationMarkiertNichts(t *testing.T) {
	store := &mockTSESignaturStore{offene: []tse_repo.OffenerSignaturauftrag{{ID: 1, TxID: "tx-1"}, {ID: 2, TxID: "tx-2"}, {ID: 3, TxID: "tx-3"}}}
	worker := &tseSignaturWorker{
		settingsRepo: &mockTSESettingsReader{err: errors.New("connection reset")},
		store:        store,
		newTSEClient: newWorkerClient(tsetest.FakeClient{}),
		now:          time.Now,
	}

	if err := worker.processOnce(context.Background()); err == nil {
		t.Fatal("expected error for unreadable configuration")
	}
	if len(store.nichtKonfiguriert) != 0 || len(store.offene) != 3 {
		t.Errorf("expected no Markierung on unreadable configuration, got marked %+v, offen %+v", store.nichtKonfiguriert, store.offene)
	}
	if len(store.geoeffnet) != 0 {
		t.Errorf("expected keinen Stoerungszeitraum, got %v", store.geoeffnet)
	}
}

func runWorker(t *testing.T, worker *tseSignaturWorker) (context.CancelFunc, <-chan struct{}) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		worker.Run(ctx)
		close(done)
	}()
	return cancel, done
}

func signierenderFakeClient() tseClientFactory {
	return newWorkerClient(tsetest.FakeClient{
		RetrieveErr:    tse.ErrTransactionNichtGefunden,
		StartResponse:  tse.StartResult{TransactionNumber: 50, LogTime: time.Date(2026, 6, 10, 19, 0, 1, 0, time.UTC)},
		FinishResponse: tse.FinishResult{TransactionNumber: 50, SignatureCounter: 800, SerialNumberTSE: "TSE-SN", Signature: "SIG"},
	})
}

// The post-commit trigger starts a run without waiting for the one-hour tick.
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
		t.Error("Sofort-Trigger hat keinen Durchlauf angestossen")
	}
}

// panicEinmalStore panics on the first load only.
type panicEinmalStore struct {
	*mockTSESignaturStore
	panicMu  sync.Mutex
	gepanict bool
}

func (s *panicEinmalStore) GetOffeneTSESignaturauftraege(ctx context.Context, limit int) ([]tse_repo.OffenerSignaturauftrag, error) {
	s.panicMu.Lock()
	erster := !s.gepanict
	s.gepanict = true
	s.panicMu.Unlock()
	if erster {
		panic("provozierter Panic im Durchlauf")
	}
	return s.mockTSESignaturStore.GetOffeneTSESignaturauftraege(ctx, limit)
}

// A panic in one run does not stop signing; the next trigger processes the open job.
func TestTSESignaturWorker_Run_PanicStopptSignierungNicht(t *testing.T) {
	store := &panicEinmalStore{mockTSESignaturStore: &mockTSESignaturStore{
		offene:      []tse_repo.OffenerSignaturauftrag{{ID: 8, TxID: "tx-8", ProcessType: "Kassenbeleg-V1", ProcessData: "Beleg^3.00"}},
		verarbeitet: make(chan struct{}, 1),
	}}
	trigger := make(chan struct{}, 2)
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

	trigger <- struct{}{} // first run panics
	trigger <- struct{}{} // second run signs

	select {
	case <-store.verarbeitet:
	case <-time.After(5 * time.Second):
		t.Error("Signierung lief nach dem Panic nicht weiter")
	}
}

// The polling tick catches lost triggers: with no trigger at all, the tick processes the open job.
func TestTSESignaturWorker_Run_PollingFallbackFaengtVerloreneTrigger(t *testing.T) {
	store := &mockTSESignaturStore{
		offene:      []tse_repo.OffenerSignaturauftrag{{ID: 7, TxID: "tx-7", ProcessType: "Kassenbeleg-V1", ProcessData: "Beleg^2.00"}},
		verarbeitet: make(chan struct{}, 1),
	}
	worker := &tseSignaturWorker{
		settingsRepo: &mockTSESettingsReader{conf: configuredTSE()},
		store:        store,
		newTSEClient: signierenderFakeClient(),
		trigger:      make(chan struct{}), // never fired
		pollInterval: 10 * time.Millisecond,
		now:          time.Now,
	}

	cancel, done := runWorker(t, worker)
	defer func() { cancel(); <-done }()

	select {
	case <-store.verarbeitet:
	case <-time.After(5 * time.Second):
		t.Error("Polling-Fallback hat den offenen Auftrag nicht verarbeitet")
	}
}
