//go:build unit

package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/nicograef/jotti/backend/api/fiskal/setup/application"
	"github.com/nicograef/jotti/backend/domain/tse"
)

type mockSettingsQuery struct {
	tse              tse.Konfiguration
	err              error
	verbindungStatus tse.VerbindungStatus
	verbindungErr    error
	setupBefund      application.TSESetupBefund
	setupErr         error
	tseStatus        application.TSEStatus
	tseStatusErr     error
}

func (m *mockSettingsQuery) GetKassenidentitaet(_ context.Context) (tse.Kassenidentitaet, error) {
	return tse.Kassenidentitaet{}, errors.New("not implemented")
}

func (m *mockSettingsQuery) GetTSEKonfiguration(_ context.Context) (tse.Konfiguration, error) {
	if m.err != nil {
		return tse.Konfiguration{}, m.err
	}
	return m.tse, nil
}

func (m *mockSettingsQuery) TestTSEVerbindung(_ context.Context) (tse.VerbindungStatus, error) {
	if m.verbindungErr != nil {
		return tse.VerbindungStatus{}, m.verbindungErr
	}
	return m.verbindungStatus, nil
}

func (m *mockSettingsQuery) PruefeTSESetup(_ context.Context, _ tse.SetupCredentials) (application.TSESetupBefund, error) {
	if m.setupErr != nil {
		return application.TSESetupBefund{}, m.setupErr
	}
	return m.setupBefund, nil
}

func (m *mockSettingsQuery) GetTSEStatus(_ context.Context) (application.TSEStatus, error) {
	if m.tseStatusErr != nil {
		return application.TSEStatus{}, m.tseStatusErr
	}
	return m.tseStatus, nil
}

func TestGetTSEKonfigurationHandler_MaskedResponse(t *testing.T) {
	h := &QueryHandler{Query: &mockSettingsQuery{tse: tse.Konfiguration{
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

func TestPruefeTSESetupHandler_Success(t *testing.T) {
	h := &QueryHandler{Query: &mockSettingsQuery{setupBefund: application.TSESetupBefund{
		Umgebung: "TEST",
		VorhandeneTSS: []application.TSSBefund{
			{
				ID:    "tss-1",
				State: "INITIALIZED",
				PassenderClient: &application.ClientBefund{
					ID:           "client-1",
					SerialNumber: "kasse-serial-1",
					State:        "REGISTERED",
				},
			},
			{ID: "tss-2", State: "CREATED"},
		},
	}}}

	req := httptest.NewRequest(http.MethodPost, "/admin/tse-setup-pruefen",
		strings.NewReader(`{"apiKey":"api-key","apiSecret":"api-secret"}`))
	rec := httptest.NewRecorder()

	h.PruefeTSESetupHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var body struct {
		Umgebung      string `json:"umgebung"`
		VorhandeneTSS []struct {
			ID              string `json:"id"`
			State           string `json:"state"`
			PassenderClient *struct {
				ID           string `json:"id"`
				SerialNumber string `json:"serialNumber"`
				State        string `json:"state"`
			} `json:"passenderClient"`
		} `json:"vorhandeneTss"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body.Umgebung != "TEST" {
		t.Fatalf("expected TEST environment, got %q", body.Umgebung)
	}
	if len(body.VorhandeneTSS) != 2 {
		t.Fatalf("expected two TSS, got %d", len(body.VorhandeneTSS))
	}
	if body.VorhandeneTSS[0].PassenderClient == nil || body.VorhandeneTSS[0].PassenderClient.ID != "client-1" {
		t.Fatalf("expected matching client client-1, got %+v", body.VorhandeneTSS[0].PassenderClient)
	}
	if body.VorhandeneTSS[1].PassenderClient != nil {
		t.Fatalf("expected no matching client for tss-2, got %+v", body.VorhandeneTSS[1].PassenderClient)
	}
}

func TestPruefeTSESetupHandler_FalscheZugangsdaten(t *testing.T) {
	h := &QueryHandler{Query: &mockSettingsQuery{setupErr: application.ErrTSESetupZugangsdaten}}

	req := httptest.NewRequest(http.MethodPost, "/admin/tse-setup-pruefen",
		strings.NewReader(`{"apiKey":"wrong","apiSecret":"wrong"}`))
	rec := httptest.NewRecorder()

	h.PruefeTSESetupHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}

	var body struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body.Code != "tse_setup_zugangsdaten_ungueltig" {
		t.Fatalf("expected code tse_setup_zugangsdaten_ungueltig, got %q", body.Code)
	}
}

func TestPruefeTSESetupHandler_ValidationError(t *testing.T) {
	h := &QueryHandler{Query: &mockSettingsQuery{}}

	req := httptest.NewRequest(http.MethodPost, "/admin/tse-setup-pruefen",
		strings.NewReader(`{"apiKey":"","apiSecret":""}`))
	rec := httptest.NewRecorder()

	h.PruefeTSESetupHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}

	var body struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body.Code != "validation_error" {
		t.Fatalf("expected code validation_error, got %q", body.Code)
	}
}

func TestGetTSEStatusHandler_Success(t *testing.T) {
	h := &QueryHandler{Query: &mockSettingsQuery{tseStatus: application.TSEStatus{
		Umgebung:        "TEST",
		IstKonfiguriert: true,
	}}}

	req := httptest.NewRequest(http.MethodPost, "/admin/get-tse-status", nil)
	rec := httptest.NewRecorder()

	h.GetTSEStatusHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var body struct {
		Umgebung        string `json:"umgebung"`
		IstKonfiguriert bool   `json:"istKonfiguriert"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body.Umgebung != "TEST" {
		t.Fatalf("expected TEST, got %q", body.Umgebung)
	}
	if !body.IstKonfiguriert {
		t.Fatal("expected istKonfiguriert true")
	}
}
