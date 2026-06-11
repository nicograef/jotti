//go:build unit

package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nicograef/jotti/backend/domain/settings"
)

type mockSettingsCommand struct {
	tse settings.TSEKonfiguration
	err error
}

func (m *mockSettingsCommand) UpdateBetreiber(_ context.Context, _ settings.Betreiber) error {
	return m.err
}

func (m *mockSettingsCommand) UpdateTSEKonfiguration(_ context.Context, b settings.TSEKonfiguration) error {
	if m.err != nil {
		return m.err
	}
	m.tse = b
	return nil
}

func TestUpdateTSEKonfigurationHandler_Success(t *testing.T) {
	mock := &mockSettingsCommand{}
	handler := &CommandHandler{Command: mock}

	body := `{"apiKey":"my-key","apiSecret":"my-secret","tssId":"tss-123","clientId":"client-123"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/update-tse-konfiguration", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.UpdateTSEKonfigurationHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if mock.tse.ApiKey != "my-key" {
		t.Fatalf("expected api key to be saved, got %q", mock.tse.ApiKey)
	}
	if mock.tse.ApiSecret != "my-secret" {
		t.Fatalf("expected api secret to be saved, got %q", mock.tse.ApiSecret)
	}
	if mock.tse.TssID != "tss-123" {
		t.Fatalf("expected tss id to be saved, got %q", mock.tse.TssID)
	}
	if mock.tse.ClientID != "client-123" {
		t.Fatalf("expected client id to be saved, got %q", mock.tse.ClientID)
	}
}

func TestUpdateTSEKonfigurationHandler_ClearAllowed(t *testing.T) {
	handler := &CommandHandler{Command: &mockSettingsCommand{}}

	body := `{"apiKey":"","apiSecret":"","tssId":"","clientId":""}`
	req := httptest.NewRequest(http.MethodPost, "/admin/update-tse-konfiguration", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.UpdateTSEKonfigurationHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateTSEKonfigurationHandler_TooLongAPIKey(t *testing.T) {
	handler := &CommandHandler{Command: &mockSettingsCommand{}}

	tooLong := strings.Repeat("a", 501)
	body := `{"apiKey":"` + tooLong + `","apiSecret":"secret","tssId":"tss-1","clientId":"client-1"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/update-tse-konfiguration", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.UpdateTSEKonfigurationHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestUpdateTSEKonfigurationHandler_PartialValuesRejected(t *testing.T) {
	handler := &CommandHandler{Command: &mockSettingsCommand{}}

	body := `{"apiKey":"my-key","apiSecret":"","tssId":"tss-1","clientId":"client-1"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/update-tse-konfiguration", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.UpdateTSEKonfigurationHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}
