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
	belegStatus application.BelegStatus
	called      bool
	last        application.KassenbelegDruckenCommand
}

func (m *mockCommand) KassenbelegDrucken(_ context.Context, cmd application.KassenbelegDruckenCommand) (application.BelegStatus, error) {
	m.called = true
	m.last = cmd
	if m.belegStatus != "" {
		return m.belegStatus, nil
	}
	return application.BelegStatusEingereiht, nil
}

func serveBelegDrucken(handler *CommandHandler, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/beleg-drucken", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.KassenbelegDruckenHandler().ServeHTTP(rec, req)
	return rec
}

// Each body form reaches the command with exactly its references; a combination outside the
// forms is a validation error that never reaches the command.
func TestKassenbelegDruckenHandler_Formen(t *testing.T) {
	const (
		zahlungID     = "11111111-1111-1111-1111-111111111111"
		verkaufID     = "22222222-2222-2222-2222-222222222222"
		stornierungID = "33333333-3333-3333-3333-333333333333"
		rueckgabeID   = "44444444-4444-4444-4444-444444444444"
	)
	tests := []struct {
		name       string
		body       string
		wantStatus int
		want       application.KassenbelegDruckenCommand
	}{
		{"Zahlung", `{"tischId":1,"zahlungId":"` + zahlungID + `"}`, http.StatusOK, application.KassenbelegDruckenCommand{TischID: 1, ZahlungID: zahlungID}},
		{"Tisch-Storno", `{"tischId":1,"stornierungId":"` + stornierungID + `"}`, http.StatusOK, application.KassenbelegDruckenCommand{TischID: 1, StornierungID: stornierungID}},
		{"Direktverkauf", `{"verkaufId":"` + verkaufID + `"}`, http.StatusOK, application.KassenbelegDruckenCommand{VerkaufID: verkaufID}},
		{"Direktverkauf-Storno", `{"verkaufId":"` + verkaufID + `","stornierungId":"` + stornierungID + `"}`, http.StatusOK, application.KassenbelegDruckenCommand{VerkaufID: verkaufID, StornierungID: stornierungID}},
		{"Pfandbeleg", `{"rueckgabeId":"` + rueckgabeID + `"}`, http.StatusOK, application.KassenbelegDruckenCommand{RueckgabeID: rueckgabeID}},
		{"Pfand-Storno", `{"rueckgabeId":"` + rueckgabeID + `","stornierungId":"` + stornierungID + `"}`, http.StatusOK, application.KassenbelegDruckenCommand{RueckgabeID: rueckgabeID, StornierungID: stornierungID}},
		{"Zahlung und Verkauf gemischt", `{"tischId":1,"zahlungId":"` + zahlungID + `","verkaufId":"` + verkaufID + `"}`, http.StatusBadRequest, application.KassenbelegDruckenCommand{}},
		{"Stornierung ohne Vorgang", `{"stornierungId":"` + stornierungID + `"}`, http.StatusBadRequest, application.KassenbelegDruckenCommand{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockCommand{}

			rec := serveBelegDrucken(&CommandHandler{Command: mock}, tt.body)

			if rec.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d: %s", tt.wantStatus, rec.Code, rec.Body.String())
			}
			if tt.wantStatus != http.StatusOK {
				if mock.called {
					t.Errorf("expected the command not to be called, got %+v", mock.last)
				}
				if !strings.Contains(rec.Body.String(), `"code":"validation_error"`) {
					t.Errorf("expected validation_error, got %s", rec.Body.String())
				}
				return
			}
			if mock.last != tt.want {
				t.Errorf("command = %+v, want %+v", mock.last, tt.want)
			}
			if got := rec.Body.String(); !strings.Contains(got, `"status":"eingereiht"`) {
				t.Errorf("expected status eingereiht in response, got %s", got)
			}
		})
	}
}

// Bei ausstehender TSE-Signatur antwortet der Endpunkt mit Status "ausstehend"
// (HTTP 200) — die UI fasst über denselben Endpunkt nach.
func TestKassenbelegDruckenHandler_Ausstehend(t *testing.T) {
	mock := &mockCommand{belegStatus: application.BelegStatusAusstehend}

	rec := serveBelegDrucken(&CommandHandler{Command: mock}, `{"tischId":1,"zahlungId":"11111111-1111-1111-1111-111111111111"}`)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	if got := rec.Body.String(); !strings.Contains(got, `"status":"ausstehend"`) {
		t.Errorf("expected status ausstehend in response, got %s", got)
	}
}
