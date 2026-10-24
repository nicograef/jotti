package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// mockCommand records the ELSTER calls; the nil embedded interface panics on any
// other, so a rejected body that still reaches the command fails the test.
type mockCommand struct {
	betreiberCommand
	setzenCalled        bool
	zuruecknehmenCalled bool
}

func (m *mockCommand) SetzeElsterMeldung(context.Context) error {
	m.setzenCalled = true
	return nil
}

func (m *mockCommand) NimmElsterMeldungZurueck(context.Context) error {
	m.zuruecknehmenCalled = true
	return nil
}

func TestElsterMeldungHandler_RuftDenCommand(t *testing.T) {
	tests := []struct {
		name    string
		handler func(*CommandHandler) http.HandlerFunc
		called  func(*mockCommand) bool
	}{
		{"setzen", (*CommandHandler).SetzeElsterMeldungHandler, func(m *mockCommand) bool { return m.setzenCalled }},
		{"zurücknehmen", (*CommandHandler).NimmElsterMeldungZurueckHandler, func(m *mockCommand) bool { return m.zuruecknehmenCalled }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &mockCommand{}
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
			rec := httptest.NewRecorder()

			tt.handler(&CommandHandler{Command: cmd}).ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("expected status 200, got %d", rec.Code)
			}
			if !tt.called(cmd) {
				t.Error("expected the command to be called")
			}
		})
	}
}

// Der Handler prüft die amtlichen Maximallängen aus domain/betreiber: Ein zu
// langer Vereinsname ist eine ungültige Eingabe und damit 400, kein Serverfehler.
func TestUpdateBetreiberHandler_ZuLangerVereinsname(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{}}
	body := `{"vereinsname":"` + strings.Repeat("a", 61) + `","strasse":"Musterstraße 1","plz":"12345","ort":"Musterstadt","steuernummer":null,"ustId":null}`
	req := httptest.NewRequest(http.MethodPost, "/update-betreiber", strings.NewReader(body))
	rec := httptest.NewRecorder()

	handler.UpdateBetreiberHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}
