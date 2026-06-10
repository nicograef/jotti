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
	bondruck settings.BondruckEinstellungen
	tse      settings.TSEKonfiguration
	err      error
}

func (m *mockSettingsCommand) UpdateBetreiber(_ context.Context, _ settings.Betreiber) error {
	return m.err
}

func (m *mockSettingsCommand) UpdateBondruckEinstellungen(_ context.Context, b settings.BondruckEinstellungen) error {
	if m.err != nil {
		return m.err
	}
	m.bondruck = b
	return nil
}

func (m *mockSettingsCommand) UpdateTSEKonfiguration(_ context.Context, b settings.TSEKonfiguration) error {
	if m.err != nil {
		return m.err
	}
	m.tse = b
	return nil
}

func TestUpdateBondruckEinstellungenHandler_Success(t *testing.T) {
	mock := &mockSettingsCommand{}
	handler := &CommandHandler{Command: mock}

	body := `{"kassenbelegDruckerIp":"192.168.1.80","direktverkaufModus":"abholbon","abholbonDruckerIp":"192.168.1.81"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/update-bondruck-einstellungen", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.UpdateBondruckEinstellungenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if mock.bondruck.KassenbelegDruckerIP != "192.168.1.80" {
		t.Errorf("expected kassenbeleg ip 192.168.1.80, got %q", mock.bondruck.KassenbelegDruckerIP)
	}
	if mock.bondruck.DirektverkaufModus != settings.DirektverkaufModusAbholbon {
		t.Errorf("expected direktverkaufModus abholbon, got %q", mock.bondruck.DirektverkaufModus)
	}
	if mock.bondruck.AbholbonDruckerIP != "192.168.1.81" {
		t.Errorf("expected abholbon ip 192.168.1.81, got %q", mock.bondruck.AbholbonDruckerIP)
	}
}

func TestUpdateBondruckEinstellungenHandler_EmptyIPsAllowed(t *testing.T) {
	handler := &CommandHandler{Command: &mockSettingsCommand{}}

	body := `{"kassenbelegDruckerIp":"","direktverkaufModus":"kein_bon","abholbonDruckerIp":""}`
	req := httptest.NewRequest(http.MethodPost, "/admin/update-bondruck-einstellungen", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.UpdateBondruckEinstellungenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateBondruckEinstellungenHandler_InvalidDirektverkaufModus(t *testing.T) {
	handler := &CommandHandler{Command: &mockSettingsCommand{}}

	body := `{"kassenbelegDruckerIp":"192.168.1.80","direktverkaufModus":"invalid","abholbonDruckerIp":"192.168.1.81"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/update-bondruck-einstellungen", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.UpdateBondruckEinstellungenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestUpdateBondruckEinstellungenHandler_InvalidAbholbonIP(t *testing.T) {
	handler := &CommandHandler{Command: &mockSettingsCommand{}}

	body := `{"kassenbelegDruckerIp":"192.168.1.80","direktverkaufModus":"abholbon","abholbonDruckerIp":"999.999.999.999"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/update-bondruck-einstellungen", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.UpdateBondruckEinstellungenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
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
