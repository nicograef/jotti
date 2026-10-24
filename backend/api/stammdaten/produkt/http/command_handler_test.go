package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nicograef/jotti/backend/api/stammdaten/produkt/application"
	dom "github.com/nicograef/jotti/backend/domain/produkt"
	"github.com/nicograef/jotti/backend/domain/steuer"
)

// mockCommand stubs only UpdateProdukt; the nil embedded interface panics on any
// other call, so a rejected body that still reaches the command fails the test.
type mockCommand struct {
	command
	err error
}

func (m *mockCommand) UpdateProdukt(context.Context, int, string, dom.Kategorie, steuer.Steuersatz) error {
	return m.err
}

// Each row is a request the handler must refuse with 400 and its wire code: a body
// the schema rejects, or a domain error the handler maps.
func TestCommandHandler_Abweisungen(t *testing.T) {
	h := &CommandHandler{Command: &mockCommand{err: application.ErrProduktAlreadyExists}}
	tests := []struct {
		name     string
		handler  http.HandlerFunc
		body     string
		wantCode string
	}{
		{"update-produkt: Name vergeben", h.UpdateProduktHandler(), `{"id":1,"name":"French Fries","kategorie":"essen","steuersatz":"ermaessigt"}`, "produkt_already_exists"},
		{"create-variante: ungültige Werte", h.CreateVarianteHandler(), `{"produktId":0,"name":"ab","preisCents":-1}`, "validation_error"},
		{"update-variante: ungültige Werte", h.UpdateVarianteHandler(), `{"id":0,"name":"ab","preisCents":-1}`, "validation_error"},
		{"verschiebe-produkt: unbekannte Richtung", h.VerschiebeProduktHandler(), `{"id":1,"richtung":"links"}`, "validation_error"},
		{"verschiebe-variante: unbekannte Richtung", h.VerschiebeVarianteHandler(), `{"id":1,"richtung":"links"}`, "validation_error"},
		{"sortiere-varianten: ungültige Produkt-ID", h.SortiereVariantenHandler(), `{"produktId":0}`, "validation_error"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()

			tt.handler.ServeHTTP(rec, req)

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
