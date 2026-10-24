package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nicograef/jotti/backend/api/stammdaten/tisch/application"
)

// mockCommand stubs the calls a row reaches; the nil embedded interface panics on
// any other, so a rejected body that still reaches the command fails the test.
type mockCommand struct {
	command
	err error
}

func (m *mockCommand) TischAktualisieren(context.Context, int, string) error { return m.err }
func (m *mockCommand) TischAktivieren(context.Context, int) error            { return m.err }
func (m *mockCommand) TischDeaktivieren(context.Context, int) error          { return m.err }
func (m *mockCommand) TischLoeschen(context.Context, int) error              { return m.err }

// Each row is a request the handler must refuse with 400 and its wire code: a body
// the schema rejects, or a domain error the handler maps.
func TestCommandHandler_Abweisungen(t *testing.T) {
	tests := []struct {
		name     string
		handler  func(*CommandHandler) http.HandlerFunc
		err      error
		body     string
		wantCode string
	}{
		{"erstellen: Name zu kurz", (*CommandHandler).TischErstellenHandler, nil, `{"name":"ab"}`, "validation_error"},
		{"aktualisieren: ungültige Werte", (*CommandHandler).TischAktualisierenHandler, nil, `{"id":0,"name":"ab"}`, "validation_error"},
		{"aktualisieren: unbekannter Tisch", (*CommandHandler).TischAktualisierenHandler, application.ErrTischNotFound, `{"id":999,"name":"Aktualisierter Tisch"}`, "tisch_not_found"},
		{"aktualisieren: Name vergeben", (*CommandHandler).TischAktualisierenHandler, application.ErrTischAlreadyExists, `{"id":1,"name":"Bereits vergebener Name"}`, "tisch_already_exists"},
		{"aktivieren: unbekannter Tisch", (*CommandHandler).TischAktivierenHandler, application.ErrTischNotFound, `{"id":999}`, "tisch_not_found"},
		{"deaktivieren: unbekannter Tisch", (*CommandHandler).TischDeaktivierenHandler, application.ErrTischNotFound, `{"id":999}`, "tisch_not_found"},
		{"deaktivieren: Saldo offen", (*CommandHandler).TischDeaktivierenHandler, application.ErrTischSaldoOffen, `{"id":1}`, "tisch_saldo_offen"},
		{"löschen: Saldo offen", (*CommandHandler).TischLoeschenHandler, application.ErrTischSaldoOffen, `{"id":1}`, "tisch_saldo_offen"},
		{"favorit: ungültige Tisch-ID", (*CommandHandler).FavoritHinzufuegenHandler, nil, `{"tischId":0}`, "validation_error"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := tt.handler(&CommandHandler{Command: &mockCommand{err: tt.err}})
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected status 400, got %d", rec.Code)
			}
			var body struct {
				Code string `json:"code"`
			}
			if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
				t.Fatalf("failed to decode error body: %v", err)
			}
			if body.Code != tt.wantCode {
				t.Errorf("expected code %s, got %s", tt.wantCode, body.Code)
			}
		})
	}
}
