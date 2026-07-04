//go:build unit

package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nicograef/jotti/backend/api/fiskal/setup/application"
	"github.com/nicograef/jotti/backend/domain/tse"
)

type mockSettingsCommand struct {
	tse          tse.Konfiguration
	err          error
	einrichten   application.TSESetupErgebnis
	einrichtErr  error
	uebernehmen  application.TSESetupErgebnis
	uebernehmErr error
}

func (m *mockSettingsCommand) UpdateTSEKonfiguration(_ context.Context, b tse.Konfiguration) error {
	if m.err != nil {
		return m.err
	}
	m.tse = b
	return nil
}

func (m *mockSettingsCommand) RichteTSEEin(_ context.Context, _ tse.SetupCredentials, _ tse.Umgebung, _ bool) (application.TSESetupErgebnis, error) {
	if m.einrichtErr != nil {
		return application.TSESetupErgebnis{}, m.einrichtErr
	}
	return m.einrichten, nil
}

func (m *mockSettingsCommand) UebernimmTSE(_ context.Context, _ tse.SetupCredentials, _ tse.Umgebung, _, _, _ string) (application.TSESetupErgebnis, error) {
	if m.uebernehmErr != nil {
		return application.TSESetupErgebnis{}, m.uebernehmErr
	}
	return m.uebernehmen, nil
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

// TestRichteTSEEinHandler_Success sichert, dass PUK und Admin-PIN genau einmal
// in der Antwort an die UI erscheinen.
func TestRichteTSEEinHandler_Success(t *testing.T) {
	mock := &mockSettingsCommand{einrichten: application.TSESetupErgebnis{
		TssID:    "tss-neu",
		ClientID: "kasse-serial",
		PUK:      "puk-xyz",
		AdminPIN: "1234567890",
		Umgebung: "TEST",
	}}
	handler := &CommandHandler{Command: mock}

	body := `{"apiKey":"my-key","apiSecret":"my-secret","umgebung":"TEST"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/tse-einrichten", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.RichteTSEEinHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		TssID    string `json:"tssId"`
		ClientID string `json:"clientId"`
		Puk      string `json:"puk"`
		AdminPin string `json:"adminPin"`
		Umgebung string `json:"umgebung"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Puk != "puk-xyz" || resp.AdminPin != "1234567890" {
		t.Fatalf("expected puk and admin pin in response, got %+v", resp)
	}
	if resp.TssID != "tss-neu" || resp.ClientID != "kasse-serial" || resp.Umgebung != "TEST" {
		t.Fatalf("unexpected response fields: %+v", resp)
	}
}

// TestRichteTSEEinHandler_InvalidUmgebung sichert, dass eine ungültige Umgebung
// abgewiesen wird, ohne den Orchestrator aufzurufen.
func TestRichteTSEEinHandler_InvalidUmgebung(t *testing.T) {
	handler := &CommandHandler{Command: &mockSettingsCommand{}}

	body := `{"apiKey":"my-key","apiSecret":"my-secret","umgebung":"STAGING"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/tse-einrichten", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.RichteTSEEinHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

// TestRichteTSEEinHandler_BereitsEingerichtet sichert die Übersetzung des
// Sentinels in den verständlichen Fehlercode für die UI.
func TestRichteTSEEinHandler_BereitsEingerichtet(t *testing.T) {
	handler := &CommandHandler{Command: &mockSettingsCommand{einrichtErr: application.ErrTSEBereitsEingerichtet}}

	body := `{"apiKey":"my-key","apiSecret":"my-secret","umgebung":"TEST"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/tse-einrichten", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.RichteTSEEinHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "tse_bereits_eingerichtet") {
		t.Fatalf("expected error code tse_bereits_eingerichtet, got %s", rec.Body.String())
	}
}

// TestUebernimmTSEHandler_Success sichert, dass die Uebernahme die TSS-ID
// entgegennimmt und das Ergebnis (inkl. ggf. neuer Geheimnisse) zurueckgibt.
func TestUebernimmTSEHandler_Success(t *testing.T) {
	mock := &mockSettingsCommand{uebernehmen: application.TSESetupErgebnis{
		TssID:    "tss-halb",
		ClientID: "client-neu",
		PUK:      "puk-refetch",
		AdminPIN: "1234567890",
		Umgebung: "TEST",
	}}
	handler := &CommandHandler{Command: mock}

	body := `{"apiKey":"my-key","apiSecret":"my-secret","umgebung":"TEST","tssId":"tss-halb"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/tse-uebernehmen", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.UebernimmTSEHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "tss-halb") || !strings.Contains(rec.Body.String(), "puk-refetch") {
		t.Fatalf("expected the takeover result in the response, got %s", rec.Body.String())
	}
}

// TestUebernimmTSEHandler_FehlendeTssID sichert, dass die Uebernahme ohne TSS-ID
// abgewiesen wird, ohne den Orchestrator aufzurufen.
func TestUebernimmTSEHandler_FehlendeTssID(t *testing.T) {
	handler := &CommandHandler{Command: &mockSettingsCommand{}}

	body := `{"apiKey":"my-key","apiSecret":"my-secret","umgebung":"TEST"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/tse-uebernehmen", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.UebernimmTSEHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

// TestUebernimmTSEHandler_UnbekanntePIN sichert die Uebersetzung der
// Sackgassen-Meldung in den verstaendlichen Fehlercode fuer die UI.
func TestUebernimmTSEHandler_UnbekanntePIN(t *testing.T) {
	handler := &CommandHandler{Command: &mockSettingsCommand{uebernehmErr: application.ErrTSESetupPINUnbekannt}}

	body := `{"apiKey":"my-key","apiSecret":"my-secret","umgebung":"TEST","tssId":"tss-init","pin":"0000000000"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/tse-uebernehmen", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.UebernimmTSEHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "tse_setup_pin_unbekannt") {
		t.Fatalf("expected error code tse_setup_pin_unbekannt, got %s", rec.Body.String())
	}
}

// TestUebernimmTSEHandler_UnbekannterPUK sichert die Uebersetzung des
// PUK-Reset-Fehlers in den verstaendlichen Fehlercode fuer die UI.
func TestUebernimmTSEHandler_UnbekannterPUK(t *testing.T) {
	handler := &CommandHandler{Command: &mockSettingsCommand{uebernehmErr: application.ErrTSESetupPUKUnbekannt}}

	body := `{"apiKey":"my-key","apiSecret":"my-secret","umgebung":"TEST","tssId":"tss-init","puk":"falscher-puk"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/tse-uebernehmen", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.UebernimmTSEHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "tse_setup_puk_unbekannt") {
		t.Fatalf("expected error code tse_setup_puk_unbekannt, got %s", rec.Body.String())
	}
}
