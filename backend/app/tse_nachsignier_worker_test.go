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

type mockTSENachsignierStore struct {
	offene            []tse_repo.OffenerNachsignierAuftrag
	quittierungen     []tse_repo.Signatur
	distinctByTxID    map[string]tse_repo.Signatur
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

func configuredTSE() settings.TSEKonfiguration {
	return settings.TSEKonfiguration{
		ApiKey:    "api-key",
		ApiSecret: "api-secret",
		TssID:     "tss-1",
		ClientID:  "client-1",
		UpdatedAt: time.Now(),
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
		newTSEClient: func(_ tse.Credentials) (tse.TSEClient, error) {
			return tse.FakeClient{
				StartResponse: tse.StartResult{TransactionNumber: 41, LogTime: time.Date(2026, 6, 10, 18, 0, 1, 0, time.UTC)},
				FinishResponse: tse.FinishResult{
					TransactionNumber: 41,
					SignatureCounter:  700,
					SerialNumberTSE:   "TSE-SN-1",
					LogTimeEnd:        time.Date(2026, 6, 10, 18, 0, 2, 0, time.UTC),
					Signature:         "SIG-1",
					QRCodeData:        "V0;QR",
				},
			}, nil
		},
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
}

func TestTSENachsignierWorker_ProcessOnce_TSEErrorKeepsAuftragOpen(t *testing.T) {
	store := &mockTSENachsignierStore{offene: []tse_repo.OffenerNachsignierAuftrag{{
		ID:          2,
		TxID:        "tx-2",
		ProcessType: "Kassenbeleg-V1",
		ProcessData: "Beleg^5.00",
	}}}

	worker := tseNachsignierWorker{
		settingsRepo: mockTSESettingsReader{conf: configuredTSE()},
		store:        store,
		newTSEClient: func(_ tse.Credentials) (tse.TSEClient, error) {
			return tse.FakeClient{StartErr: errors.New("timeout")}, nil
		},
		now: time.Now,
	}

	if err := worker.processOnce(context.Background()); err != nil {
		t.Fatalf("expected no hard error, got %v", err)
	}

	if len(store.quittierungen) != 0 {
		t.Fatalf("expected no quittierung on TSE error, got %d", len(store.quittierungen))
	}
}

func TestTSENachsignierWorker_ProcessOnce_IdempotentByTxID(t *testing.T) {
	store := &mockTSENachsignierStore{offene: []tse_repo.OffenerNachsignierAuftrag{{
		ID:          3,
		TxID:        "tx-3",
		ProcessType: "Kassenbeleg-V1",
		ProcessData: "Beleg^7.00",
	}}}

	worker := tseNachsignierWorker{
		settingsRepo: mockTSESettingsReader{conf: configuredTSE()},
		store:        store,
		newTSEClient: func(_ tse.Credentials) (tse.TSEClient, error) {
			return tse.FakeClient{
				StartResponse: tse.StartResult{TransactionNumber: 43, LogTime: time.Date(2026, 6, 10, 18, 5, 1, 0, time.UTC)},
				FinishResponse: tse.FinishResult{
					TransactionNumber: 43,
					SignatureCounter:  702,
					SerialNumberTSE:   "TSE-SN-3",
					LogTimeEnd:        time.Date(2026, 6, 10, 18, 5, 2, 0, time.UTC),
					Signature:         "SIG-3",
				},
			}, nil
		},
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
