package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// mockCommand panics through its nil embedded interface on any call, so a rejected
// body that still reaches the command fails the test.
type mockCommand struct {
	betreiberCommand
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
