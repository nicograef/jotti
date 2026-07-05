//go:build unit

package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nicograef/jotti/backend/api/druck/beleg/application"
)

type mockCommand struct {
	err         error
	belegStatus application.BelegStatus
	last        application.KassenbelegDruckenCommand
}

func (m *mockCommand) KassenbelegDrucken(_ context.Context, cmd application.KassenbelegDruckenCommand) (application.BelegStatus, error) {
	m.last = cmd
	if m.err != nil {
		return "", m.err
	}
	if m.belegStatus != "" {
		return m.belegStatus, nil
	}
	return application.BelegStatusEingereiht, nil
}

func TestKassenbelegDruckenHandler_Success(t *testing.T) {
	mock := &mockCommand{}
	handler := &CommandHandler{Command: mock}

	body := `{"tischId":1,"zahlungId":"11111111-1111-1111-1111-111111111111"}`
	req := httptest.NewRequest(http.MethodPost, "/beleg-drucken", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.KassenbelegDruckenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	if got := rec.Body.String(); !strings.Contains(got, `"status":"eingereiht"`) {
		t.Errorf("expected status eingereiht in response, got %s", got)
	}
	if mock.last.TischID != 1 {
		t.Errorf("expected tischId 1, got %d", mock.last.TischID)
	}
	if mock.last.ZahlungID != "11111111-1111-1111-1111-111111111111" {
		t.Errorf("expected zahlungId to be forwarded")
	}
	if mock.last.VerkaufID != "" {
		t.Errorf("expected verkaufId empty for zahlung path, got %q", mock.last.VerkaufID)
	}
}

// Bei ausstehender TSE-Signatur antwortet der Endpunkt mit Status "ausstehend"
// (HTTP 200) — die UI fasst über denselben Endpunkt nach.
func TestKassenbelegDruckenHandler_Ausstehend(t *testing.T) {
	mock := &mockCommand{belegStatus: application.BelegStatusAusstehend}
	handler := &CommandHandler{Command: mock}

	body := `{"tischId":1,"zahlungId":"11111111-1111-1111-1111-111111111111"}`
	req := httptest.NewRequest(http.MethodPost, "/beleg-drucken", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.KassenbelegDruckenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	if got := rec.Body.String(); !strings.Contains(got, `"status":"ausstehend"`) {
		t.Errorf("expected status ausstehend in response, got %s", got)
	}
}

func TestKassenbelegDruckenHandler_DirektverkaufSuccess(t *testing.T) {
	mock := &mockCommand{}
	handler := &CommandHandler{Command: mock}

	body := `{"verkaufId":"11111111-1111-1111-1111-111111111111"}`
	req := httptest.NewRequest(http.MethodPost, "/beleg-drucken", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.KassenbelegDruckenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	if mock.last.VerkaufID != "11111111-1111-1111-1111-111111111111" {
		t.Errorf("expected verkaufId to be forwarded")
	}
	if mock.last.TischID != 0 || mock.last.ZahlungID != "" {
		t.Errorf("expected zahlung reference to be empty for direktverkauf path")
	}
}

func TestKassenbelegDruckenHandler_DruckerNichtKonfiguriert(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{err: application.ErrKassenbelegDruckerNichtKonfiguriert}}

	body := `{"tischId":1,"zahlungId":"11111111-1111-1111-1111-111111111111"}`
	req := httptest.NewRequest(http.MethodPost, "/beleg-drucken", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.KassenbelegDruckenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestKassenbelegDruckenHandler_DirektverkaufNichtGefunden(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{err: application.ErrVerkaufNichtGefunden}}

	body := `{"verkaufId":"11111111-1111-1111-1111-111111111111"}`
	req := httptest.NewRequest(http.MethodPost, "/beleg-drucken", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.KassenbelegDruckenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestKassenbelegDruckenHandler_MixedReferenceValidationError(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{}}

	body := `{"tischId":1,"zahlungId":"11111111-1111-1111-1111-111111111111","verkaufId":"22222222-2222-2222-2222-222222222222"}`
	req := httptest.NewRequest(http.MethodPost, "/beleg-drucken", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.KassenbelegDruckenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestKassenbelegDruckenHandler_StornobelegSuccess(t *testing.T) {
	mock := &mockCommand{}
	handler := &CommandHandler{Command: mock}

	body := `{"verkaufId":"11111111-1111-1111-1111-111111111111","stornierungId":"33333333-3333-3333-3333-333333333333"}`
	req := httptest.NewRequest(http.MethodPost, "/beleg-drucken", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.KassenbelegDruckenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	if mock.last.VerkaufID != "11111111-1111-1111-1111-111111111111" {
		t.Errorf("expected verkaufId to be forwarded")
	}
	if mock.last.StornierungID != "33333333-3333-3333-3333-333333333333" {
		t.Errorf("expected stornierungId to be forwarded, got %q", mock.last.StornierungID)
	}
}

func TestKassenbelegDruckenHandler_TischStornoSuccess(t *testing.T) {
	mock := &mockCommand{}
	handler := &CommandHandler{Command: mock}

	body := `{"tischId":1,"stornierungId":"33333333-3333-3333-3333-333333333333"}`
	req := httptest.NewRequest(http.MethodPost, "/beleg-drucken", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.KassenbelegDruckenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	if mock.last.TischID != 1 {
		t.Errorf("expected tischId 1, got %d", mock.last.TischID)
	}
	if mock.last.StornierungID != "33333333-3333-3333-3333-333333333333" {
		t.Errorf("expected stornierungId to be forwarded, got %q", mock.last.StornierungID)
	}
	if mock.last.ZahlungID != "" || mock.last.VerkaufID != "" {
		t.Errorf("expected no zahlung/verkauf forwarded, got zahlung=%q verkauf=%q", mock.last.ZahlungID, mock.last.VerkaufID)
	}
}

func TestKassenbelegDruckenHandler_StornierungOhneVerkaufValidationError(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{}}

	body := `{"stornierungId":"33333333-3333-3333-3333-333333333333"}`
	req := httptest.NewRequest(http.MethodPost, "/beleg-drucken", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.KassenbelegDruckenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}
