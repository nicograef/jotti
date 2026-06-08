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
