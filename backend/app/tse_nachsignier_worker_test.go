//go:build unit

package app

import (
	"context"
	"errors"
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

func (m mockTSESettingsReader) GetTSEKonfiguration(_ context.Context) (settings.TSEKonfiguration, error) {
	if m.err != nil {
		return settings.TSEKonfiguration{}, m.err
	}
	return m.conf, nil
}

type fehlversuch struct {
	AuftragID int
	Fehler    string
}

type mockTSENachsignierStore struct {
	offene            []tse_repo.OffenerNachsignierAuftrag
	quittierungen     []tse_repo.Signatur
	distinctByTxID    map[string]tse_repo.Signatur
	fehlversuche      []fehlversuch
	getErr            error
	quittiereErr      error
	quittierteAuftrag []int
}

func (m *mockTSENachsignierStore) GetOffeneTSENachsignierAuftraege(_ context.Context, _ int) ([]tse_repo.OffenerNachsignierAuftrag, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.offene, nil
}

func (m *mockTSENachsignierStore) QuittiereTSENachsignierAuftrag(_ context.Context, auftragID int, signatur tse_repo.Signatur) error {
	if m.quittiereErr != nil {
		return m.quittiereErr
	}
	if m.distinctByTxID == nil {
		m.distinctByTxID = make(map[string]tse_repo.Signatur)
	}
	m.quittierteAuftrag = append(m.quittierteAuftrag, auftragID)
	m.quittierungen = append(m.quittierungen, signatur)
	m.distinctByTxID[signatur.TxID] = signatur
	return nil
}

func (m *mockTSENachsignierStore) TSENachsignierAuftragFehlversuch(_ context.Context, auftragID int, fehler string) error {
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

func TestTSENachsignierWorker_ProcessOnce_Success(t *testing.T) {
	store := &mockTSENachsignierStore{offene: []tse_repo.OffenerNachsignierAuftrag{{
		ID:          1,
		TxID:        "tx-1",
		ProcessType: "Kassenbeleg-V1",
		ProcessData: "Beleg^3.50",
	}}}

	worker := tseNachsignierWorker{
		settingsRepo: mockTSESettingsReader{conf: configuredTSE()},
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
	if store.quittierungen[0].TxID != "tx-1" {
		t.Fatalf("expected tx-1, got %q", store.quittierungen[0].TxID)
	}
	if store.quittierungen[0].Signatur != "SIG-1" {
		t.Fatalf("expected SIG-1, got %q", store.quittierungen[0].Signatur)
	}
	if len(store.fehlversuche) != 0 {
		t.Fatalf("expected no fehlversuche on success, got %d", len(store.fehlversuche))
	}
}

func TestTSENachsignierWorker_ProcessOnce_TSEErrorVerbuchtFehlversuch(t *testing.T) {
	store := &mockTSENachsignierStore{offene: []tse_repo.OffenerNachsignierAuftrag{{
		ID:          2,
		TxID:        "tx-2",
		ProcessType: "Kassenbeleg-V1",
		ProcessData: "Beleg^5.00",
	}}}

	worker := tseNachsignierWorker{
		settingsRepo: mockTSESettingsReader{conf: configuredTSE()},
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
func TestTSENachsignierWorker_ProcessOnce_BereitsFinishedWirdQuittiert(t *testing.T) {
	store := &mockTSENachsignierStore{offene: []tse_repo.OffenerNachsignierAuftrag{{
		ID:          3,
		TxID:        "tx-3",
		ProcessType: "Kassenbeleg-V1",
		ProcessData: "Beleg^7.00",
	}}}

	worker := tseNachsignierWorker{
		settingsRepo: mockTSESettingsReader{conf: configuredTSE()},
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
	signatur := store.quittierungen[0]
	if signatur.Signatur != "SIG-3" || signatur.TransaktionNummer != 43 || signatur.SignaturZaehler != 702 {
		t.Fatalf("expected retrieved signature data, got %+v", signatur)
	}
	if !signatur.LogTimeStart.Equal(time.Date(2026, 6, 10, 18, 5, 1, 0, time.UTC)) {
		t.Fatalf("expected retrieved log_time_start, got %v", signatur.LogTimeStart)
	}
}

// Eine bei fiskaly noch aktive Transaktion (Start kam durch, Finish nicht)
// wird nur noch abgeschlossen — ein erneuter Start wuerde fehlschlagen.
func TestTSENachsignierWorker_ProcessOnce_AktiveTransaktionWirdAbgeschlossen(t *testing.T) {
	store := &mockTSENachsignierStore{offene: []tse_repo.OffenerNachsignierAuftrag{{
		ID:          4,
		TxID:        "tx-4",
		ProcessType: "Kassenbeleg-V1",
		ProcessData: "Beleg^9.00",
	}}}

	worker := tseNachsignierWorker{
		settingsRepo: mockTSESettingsReader{conf: configuredTSE()},
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
	signatur := store.quittierungen[0]
	if signatur.Signatur != "SIG-4" || signatur.TransaktionNummer != 44 {
		t.Fatalf("expected finish signature data, got %+v", signatur)
	}
	if !signatur.LogTimeStart.Equal(time.Date(2026, 6, 10, 18, 7, 1, 0, time.UTC)) {
		t.Fatalf("expected log_time_start from retrieved transaction, got %v", signatur.LogTimeStart)
	}
}

func TestTSENachsignierWorker_ProcessOnce_IdempotentByTxID(t *testing.T) {
	store := &mockTSENachsignierStore{offene: []tse_repo.OffenerNachsignierAuftrag{{
		ID:          5,
		TxID:        "tx-5",
		ProcessType: "Kassenbeleg-V1",
		ProcessData: "Beleg^7.00",
	}}}

	worker := tseNachsignierWorker{
		settingsRepo: mockTSESettingsReader{conf: configuredTSE()},
		store:        store,
		newTSEClient: neuerWorkerClient(tse.FakeClient{
			RetrieveErr:   tse.ErrTransactionNichtGefunden,
			StartResponse: tse.StartResult{TransactionNumber: 45, LogTime: time.Date(2026, 6, 10, 18, 5, 1, 0, time.UTC)},
			FinishResponse: tse.FinishResult{
				TransactionNumber: 45,
				SignatureCounter:  702,
				SerialNumberTSE:   "TSE-SN-5",
				LogTimeEnd:        time.Date(2026, 6, 10, 18, 5, 2, 0, time.UTC),
				Signature:         "SIG-5",
			},
		}),
		now: time.Now,
	}

	if err := worker.processOnce(context.Background()); err != nil {
		t.Fatalf("first run failed: %v", err)
	}
	if err := worker.processOnce(context.Background()); err != nil {
		t.Fatalf("second run failed: %v", err)
	}

	if len(store.distinctByTxID) != 1 {
		t.Fatalf("expected one distinct tx_id signature, got %d", len(store.distinctByTxID))
	}
}
