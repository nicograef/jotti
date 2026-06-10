//go:build unit

package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/nicograef/jotti/backend/api/settings/application"
	"github.com/nicograef/jotti/backend/domain/settings"
)

type mockSettingsQuery struct {
	tse settings.TSEKonfiguration
	err error
}

func (m *mockSettingsQuery) GetKassenidentitaet(_ context.Context) (settings.Kassenidentitaet, error) {
	return settings.Kassenidentitaet{}, errors.New("not implemented")
}

func (m *mockSettingsQuery) GetBetreiber(_ context.Context) (settings.Betreiber, error) {
	return settings.Betreiber{}, errors.New("not implemented")
}

func (m *mockSettingsQuery) GetBondruckEinstellungen(_ context.Context) (settings.BondruckEinstellungen, error) {
	return settings.BondruckEinstellungen{}, errors.New("not implemented")
}

func (m *mockSettingsQuery) GetTSEKonfiguration(_ context.Context) (settings.TSEKonfiguration, error) {
	if m.err != nil {
		return settings.TSEKonfiguration{}, m.err
	}
	return m.tse, nil
}

func TestGetTSEKonfigurationHandler_MaskedResponse(t *testing.T) {
	h := &QueryHandler{Query: &mockSettingsQuery{tse: settings.TSEKonfiguration{
		ApiKey:    "my-api-key",
		ApiSecret: "my-api-secret",
		TssID:     "tss-123",
		ClientID:  "client-123",
		UpdatedAt: time.Now(),
	}}}

	req := httptest.NewRequest(http.MethodPost, "/admin/get-tse-konfiguration", nil)
	rec := httptest.NewRecorder()

	h.GetTSEKonfigurationHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var body struct {
		ApiKeyGesetzt    bool   `json:"apiKeyGesetzt"`
		ApiSecretGesetzt bool   `json:"apiSecretGesetzt"`
		TssID            string `json:"tssId"`
		ClientID         string `json:"clientId"`
		IstKonfiguriert  bool   `json:"istKonfiguriert"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !body.ApiKeyGesetzt {
		t.Fatal("expected apiKeyGesetzt to be true")
	}
	if !body.ApiSecretGesetzt {
		t.Fatal("expected apiSecretGesetzt to be true")
	}
	if body.TssID != "tss-123" {
		t.Fatalf("expected tss id, got %q", body.TssID)
	}
	if body.ClientID != "client-123" {
		t.Fatalf("expected client id, got %q", body.ClientID)
	}
	if !body.IstKonfiguriert {
		t.Fatal("expected istKonfiguriert to be true")
	}
}

func TestGetTSEKonfigurationHandler_NotFoundReturnsEmpty(t *testing.T) {
	h := &QueryHandler{Query: &mockSettingsQuery{err: application.ErrNotFound}}

	req := httptest.NewRequest(http.MethodPost, "/admin/get-tse-konfiguration", nil)
	rec := httptest.NewRecorder()

	h.GetTSEKonfigurationHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var body struct {
		ApiKeyGesetzt    bool   `json:"apiKeyGesetzt"`
		ApiSecretGesetzt bool   `json:"apiSecretGesetzt"`
		TssID            string `json:"tssId"`
		ClientID         string `json:"clientId"`
		IstKonfiguriert  bool   `json:"istKonfiguriert"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body.ApiKeyGesetzt || body.ApiSecretGesetzt || body.IstKonfiguriert {
		t.Fatal("expected empty response flags to be false")
	}
	if body.TssID != "" || body.ClientID != "" {
		t.Fatal("expected empty response values")
	}
}
