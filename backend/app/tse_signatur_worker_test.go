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
	geoeffnet     []string // Grund-Arten der geoeffneten Stoerungszeitraeume
	geschlossen   []string // Grund-Arten der geschlossenen Stoerungszeitraeume
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

func (m *mockTSESignaturStore) OeffneTSEStoerung(_ context.Context, grundArt string, _ string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.geoeffnet = append(m.geoeffnet, grundArt)
	return nil
}

func (m *mockTSESignaturStore) SchliesseTSEStoerung(_ context.Context, grundArt string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.geschlossen = append(m.geschlossen, grundArt)
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

// zaehlenderClient zaehlt die fiskaly-Aufrufe — fuer Tests, die belegen, dass
// der Durchlauf abbricht bzw. der Stoerungszustand fiskaly in Ruhe laesst.
type zaehlenderClient struct {
	tse.FakeClient
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

// txAbhaengigerClient signiert je nach txID erfolgreich oder lehnt mit dem
// hinterlegten Fehler ab — fuer Gift-Auftrag-Tests.
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

// Ein TSE-weiter Fehler (hier: Verbindungsfehler) bricht den Durchlauf beim
// ersten Auftrag ab: keine Auftrags-Fehlversuche, der zweite Auftrag wird gar
// nicht versucht, der Stoerungszeitraum tse_fehler wird geoeffnet und der
// Worker betritt den Stoerungszustand.
func TestTSESignaturWorker_ProcessOnce_TSEWeiterFehlerBrichtDurchlaufAb(t *testing.T) {
	store := &mockTSESignaturStore{offene: []tse_repo.OffenerSignaturauftrag{
		{ID: 2, TxID: "tx-2", ProcessType: "Kassenbeleg-V1", ProcessData: "Beleg^5.00"},
		{ID: 3, TxID: "tx-3", ProcessType: "Kassenbeleg-V1", ProcessData: "Beleg^6.00"},
	}}
	client := &zaehlenderClient{FakeClient: tse.FakeClient{RetrieveErr: errors.New("connection refused")}}
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
		t.Fatalf("expected abort after first auftrag (1 fiskaly call), got %d", client.anzahlCalls())
	}
	if len(store.fehlversuche) != 0 {
		t.Fatalf("expected no auftrags-fehlversuche on TSE-weitem Fehler, got %+v", store.fehlversuche)
	}
	if len(store.quittierungen) != 0 {
		t.Fatalf("expected no quittierungen, got %d", len(store.quittierungen))
	}
	if len(store.geoeffnet) != 1 || store.geoeffnet[0] != tse.StoerungGrundTSEFehler {
		t.Fatalf("expected geoeffneten tse_fehler-Zeitraum, got %v", store.geoeffnet)
	}
	if worker.stoerungSerie != 1 {
		t.Fatalf("expected fehlerserie 1, got %d", worker.stoerungSerie)
	}
	if !worker.stoerungNaechsterVersuch.Equal(jetzt.Add(5 * time.Second)) {
		t.Fatalf("expected naechsten Versuch nach 5s Backoff, got %v", worker.stoerungNaechsterVersuch)
	}
}

// Ein auftragsspezifischer Fehler (tse.AuftragsFehler, etwa eine
// 400-Ablehnung) verbucht einen Fehlversuch am Auftrag und ueberspringt ihn:
// Der nachfolgende Auftrag wird im selben Durchlauf signiert, es entsteht
// keine Stoerung — ein Gift-Auftrag staut nie die Queue.
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
		t.Fatalf("expected one fehlversuch for auftrag 10, got %+v", store.fehlversuche)
	}
	if len(store.quittierungen) != 1 || store.quittierungen[0].AuftragID != 11 {
		t.Fatalf("expected auftrag 11 signed in same run, got %+v", store.quittierungen)
	}
	if len(store.geoeffnet) != 0 {
		t.Fatalf("expected no stoerung on auftragsspezifischem Fehler, got %v", store.geoeffnet)
	}
	if worker.stoerungSerie != 0 {
		t.Fatalf("expected keine fehlerserie, got %d", worker.stoerungSerie)
	}
}

// Eine bei fiskaly stornierte Transaktion (unerwarteter Zustand CANCELLED)
// haengt an diesem einen Auftrag: Fehlversuch statt Durchlauf-Abbruch.
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
		newTSEClient: neuerWorkerClient(tse.FakeClient{
			RetrieveResponse: tse.RetrieveResult{State: tse.TransactionStateCancelled},
		}),
		now: time.Now,
	}

	if err := worker.processOnce(context.Background()); err != nil {
		t.Fatalf("expected no durchlauf error, got %v", err)
	}
	if len(store.fehlversuche) != 1 || store.fehlversuche[0].AuftragID != 12 {
		t.Fatalf("expected fehlversuch for auftrag 12, got %+v", store.fehlversuche)
	}
	if len(store.geoeffnet) != 0 {
		t.Fatalf("expected no stoerung, got %v", store.geoeffnet)
	}
}

// Im Stoerungszustand laesst der Worker fiskaly bis zum Backoff-Ablauf in
// Ruhe: Trigger und Ticks fuehren zu keinem fiskaly-Aufruf. Nach Ablauf ist
// der erste Auftrag die Half-Open-Probe: Scheitert sie TSE-weit, waechst der
// Backoff; gelingt sie, laeuft die volle Aufarbeitung und die erste
// erfolgreiche Signatur schliesst den Stoerungszeitraum.
func TestTSESignaturWorker_StoerungBackoffUndHalfOpenProbe(t *testing.T) {
	store := &mockTSESignaturStore{offene: []tse_repo.OffenerSignaturauftrag{
		{ID: 20, TxID: "tx-20", ProcessType: "Kassenbeleg-V1", ProcessData: "Beleg^1.00"},
		{ID: 21, TxID: "tx-21", ProcessType: "Kassenbeleg-V1", ProcessData: "Beleg^2.00"},
	}}
	client := &zaehlenderClient{FakeClient: tse.FakeClient{RetrieveErr: errors.New("503 service unavailable")}}
	jetzt := time.Date(2026, 6, 10, 18, 0, 0, 0, time.UTC)

	worker := &tseSignaturWorker{
		settingsRepo: &mockTSESettingsReader{conf: configuredTSE()},
		store:        store,
		newTSEClient: func(_ tse.Credentials) (tseWorkerClient, error) { return client, nil },
		now:          func() time.Time { return jetzt },
	}
	ctx := context.Background()

	// Durchlauf 1: TSE-weiter Fehler — Stoerungszustand, Backoff 5s.
	if err := worker.processOnce(ctx); err == nil {
		t.Fatal("expected TSE-weiten Fehler")
	}

	// Waehrend des Backoffs: kein einziger fiskaly-Aufruf.
	jetzt = jetzt.Add(2 * time.Second)
	callsVorher := client.anzahlCalls()
	if err := worker.processOnce(ctx); err != nil {
		t.Fatalf("expected gated durchlauf without error, got %v", err)
	}
	if client.anzahlCalls() != callsVorher {
		t.Fatalf("expected no fiskaly calls during stoerung, got %d new", client.anzahlCalls()-callsVorher)
	}

	// Probe nach Backoff-Ablauf scheitert TSE-weit: ein Aufruf, Serie und
	// Backoff wachsen (5s -> 10s).
	jetzt = jetzt.Add(4 * time.Second)
	callsVorher = client.anzahlCalls()
	if err := worker.processOnce(ctx); err == nil {
		t.Fatal("expected TSE-weiten Fehler der Probe")
	}
	if client.anzahlCalls() != callsVorher+1 {
		t.Fatalf("expected exactly one probe call, got %d", client.anzahlCalls()-callsVorher)
	}
	if worker.stoerungSerie != 2 {
		t.Fatalf("expected fehlerserie 2, got %d", worker.stoerungSerie)
	}
	if !worker.stoerungNaechsterVersuch.Equal(jetzt.Add(10 * time.Second)) {
		t.Fatalf("expected gewachsenen Backoff 10s, got %v", worker.stoerungNaechsterVersuch.Sub(jetzt))
	}

	// TSE erholt sich: Die Probe gelingt, die volle Aufarbeitung signiert
	// beide Auftraege, die erste erfolgreiche Signatur schliesst den
	// Stoerungszeitraum und setzt die Serie zurueck.
	client.FakeClient = tse.FakeClient{
		RetrieveErr:    tse.ErrTransactionNichtGefunden,
		StartResponse:  tse.StartResult{TransactionNumber: 70, LogTime: jetzt},
		FinishResponse: tse.FinishResult{TransactionNumber: 70, SignatureCounter: 900, SerialNumberTSE: "TSE-SN", Signature: "SIG"},
	}
	jetzt = jetzt.Add(11 * time.Second)
	if err := worker.processOnce(ctx); err != nil {
		t.Fatalf("expected recovery durchlauf without error, got %v", err)
	}
	if len(store.quittierungen) != 2 {
		t.Fatalf("expected volle Aufarbeitung (2 quittierungen), got %d", len(store.quittierungen))
	}
	if len(store.geschlossen) != 1 || store.geschlossen[0] != tse.StoerungGrundTSEFehler {
		t.Fatalf("expected geschlossenen tse_fehler-Zeitraum, got %v", store.geschlossen)
	}
	if worker.stoerungSerie != 0 || !worker.stoerungNaechsterVersuch.IsZero() {
		t.Fatalf("expected zurueckgesetzten Stoerungszustand, got serie=%d next=%v", worker.stoerungSerie, worker.stoerungNaechsterVersuch)
	}
}

// Der Stoerungs-Backoff ist deterministisch (ohne Jitter): Basis 5s,
// verdoppelt je Fehlerserie, gedeckelt auf 2 Minuten.
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

// Jeder Durchlauf hat eine Deadline: Ein haengender fiskaly-Aufruf wird
// abgebrochen und als TSE-weiter Fehler behandelt (Stoerungszustand statt
// blockiertem Worker).
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
		newTSEClient:      neuerWorkerClient(tse.FakeClient{ArtificialDelay: time.Minute}),
		durchlaufDeadline: 30 * time.Millisecond,
		now:               time.Now,
	}

	if err := worker.processOnce(context.Background()); err == nil {
		t.Fatal("expected deadline abort as durchlauf error")
	}
	if len(store.fehlversuche) != 0 {
		t.Fatalf("expected no fehlversuche on deadline abort, got %+v", store.fehlversuche)
	}
	if len(store.geoeffnet) != 1 || store.geoeffnet[0] != tse.StoerungGrundTSEFehler {
		t.Fatalf("expected geoeffneten tse_fehler-Zeitraum, got %v", store.geoeffnet)
	}
	if worker.stoerungSerie != 1 {
		t.Fatalf("expected fehlerserie 1, got %d", worker.stoerungSerie)
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
