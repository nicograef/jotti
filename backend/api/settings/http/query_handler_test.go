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
	"github.com/nicograef/jotti/backend/domain/tse"
)

type mockSettingsQuery struct {
	tse              settings.TSEKonfiguration
	err              error
	verbindungStatus tse.VerbindungStatus
	verbindungErr    error
	tseStatus        application.TSEStatus
	tseStatusErr     error
}

func (m *mockSettingsQuery) GetKassenidentitaet(_ context.Context) (settings.Kassenidentitaet, error) {
	return settings.Kassenidentitaet{}, errors.New("not implemented")
}

func (m *mockSettingsQuery) GetBetreiber(_ context.Context) (settings.Betreiber, error) {
	return settings.Betreiber{}, errors.New("not implemented")
}

func (m *mockSettingsQuery) GetTSEKonfiguration(_ context.Context) (settings.TSEKonfiguration, error) {
	if m.err != nil {
		return settings.TSEKonfiguration{}, m.err
	}
	return m.tse, nil
}

func (m *mockSettingsQuery) TestTSEVerbindung(_ context.Context) (tse.VerbindungStatus, error) {
	if m.verbindungErr != nil {
		return tse.VerbindungStatus{}, m.verbindungErr
	}
	return m.verbindungStatus, nil
}

func (m *mockSettingsQuery) GetTSEStatus(_ context.Context) (application.TSEStatus, error) {
	if m.tseStatusErr != nil {
		return application.TSEStatus{}, m.tseStatusErr
	}
	return m.tseStatus, nil
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

func TestTestTSEVerbindungHandler_Success(t *testing.T) {
	h := &QueryHandler{Query: &mockSettingsQuery{verbindungStatus: tse.VerbindungStatus{
		Umgebung:            tse.UmgebungTest,
		TSSState:            "INITIALIZED",
		ClientState:         "REGISTERED",
		ClientSerialNumber:  "kasse-serial-1",
		SeriennummerKorrekt: true,
	}}}

	req := httptest.NewRequest(http.MethodPost, "/admin/test-tse-verbindung", nil)
	rec := httptest.NewRecorder()

	h.TestTSEVerbindungHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var body struct {
		Umgebung            string `json:"umgebung"`
		TSSState            string `json:"tssState"`
		ClientState         string `json:"clientState"`
		ClientSerialNumber  string `json:"clientSerialNumber"`
		SeriennummerKorrekt bool   `json:"seriennummerKorrekt"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body.Umgebung != "TEST" {
		t.Fatalf("expected TEST environment, got %q", body.Umgebung)
	}
	if body.TSSState != "INITIALIZED" {
		t.Fatalf("expected INITIALIZED state, got %q", body.TSSState)
	}
	if body.ClientState != "REGISTERED" {
		t.Fatalf("expected REGISTERED client state, got %q", body.ClientState)
	}
	if body.ClientSerialNumber != "kasse-serial-1" {
		t.Fatalf("expected client serial kasse-serial-1, got %q", body.ClientSerialNumber)
	}
	if !body.SeriennummerKorrekt {
		t.Fatal("expected seriennummerKorrekt to be true")
	}
}

func TestTestTSEVerbindungHandler_NotConfigured(t *testing.T) {
	h := &QueryHandler{Query: &mockSettingsQuery{verbindungErr: application.ErrTSENichtKonfiguriert}}

	req := httptest.NewRequest(http.MethodPost, "/admin/test-tse-verbindung", nil)
	rec := httptest.NewRecorder()

	h.TestTSEVerbindungHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}

	var body struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body.Code != "tse_nicht_konfiguriert" {
		t.Fatalf("expected code tse_nicht_konfiguriert, got %q", body.Code)
	}
}

func TestTestTSEVerbindungHandler_VerbindungFehlgeschlagen(t *testing.T) {
	h := &QueryHandler{Query: &mockSettingsQuery{verbindungErr: application.ErrTSEVerbindungFehlgeschlagen}}

	req := httptest.NewRequest(http.MethodPost, "/admin/test-tse-verbindung", nil)
	rec := httptest.NewRecorder()

	h.TestTSEVerbindungHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}

	var body struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body.Code != "tse_verbindung_fehlgeschlagen" {
		t.Fatalf("expected code tse_verbindung_fehlgeschlagen, got %q", body.Code)
	}
}

func TestGetTSEStatusHandler_Success(t *testing.T) {
	h := &QueryHandler{Query: &mockSettingsQuery{tseStatus: application.TSEStatus{
		Umgebung:               "TEST",
		OffeneNachsignierungen: 3,
		IstKonfiguriert:        true,
	}}}

	req := httptest.NewRequest(http.MethodPost, "/admin/get-tse-status", nil)
	rec := httptest.NewRecorder()

	h.GetTSEStatusHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var body struct {
		Umgebung               string `json:"umgebung"`
		OffeneNachsignierungen int    `json:"offeneNachsignierungen"`
		IstKonfiguriert        bool   `json:"istKonfiguriert"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body.Umgebung != "TEST" {
		t.Fatalf("expected TEST, got %q", body.Umgebung)
	}
	if body.OffeneNachsignierungen != 3 {
		t.Fatalf("expected 3 offeneNachsignierungen, got %d", body.OffeneNachsignierungen)
	}
	if !body.IstKonfiguriert {
		t.Fatal("expected istKonfiguriert true")
	}
}
