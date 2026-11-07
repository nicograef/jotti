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

// mockCommand stubs only UpdateProdukt and CreateVariante; the nil embedded interface panics on
// any other call, so a rejected body that still reaches the command fails the test.
type mockCommand struct {
	command
	err error
	// pfandVarianteID records the argument of the last CreateVariante call.
	pfandVarianteID *int
	createVariante  bool
}

func (m *mockCommand) UpdateProdukt(context.Context, int, string, dom.Kategorie, steuer.Steuersatz, bool) error {
	return m.err
}

func (m *mockCommand) CreateVariante(_ context.Context, _ int, _ string, _ int, pfandVarianteID *int) (int, error) {
	m.createVariante = true
	m.pfandVarianteID = pfandVarianteID
	return 1, m.err
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
		{"deactivate-varianten: ungültige Produkt-ID", h.DeactivateAlleVariantenHandler(), `{"produktId":0}`, "validation_error"},
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

// pfandVarianteId is optional: missing and null mean no Pfand, 0 is no ID.
func TestCommandHandler_PfandVarianteId(t *testing.T) {
	sieben := 7
	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantPfand  *int
	}{
		{"ohne Feld", `{"produktId":1,"name":"0,5 l","preisCents":400}`, http.StatusOK, nil},
		{"null", `{"produktId":1,"name":"0,5 l","preisCents":400,"pfandVarianteId":null}`, http.StatusOK, nil},
		{"gesetzt", `{"produktId":1,"name":"0,5 l","preisCents":400,"pfandVarianteId":7}`, http.StatusOK, &sieben},
		{"null-ID", `{"produktId":1,"name":"0,5 l","preisCents":400,"pfandVarianteId":0}`, http.StatusBadRequest, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mockCommand{}
			h := &CommandHandler{Command: m}
			rec := httptest.NewRecorder()

			h.CreateVarianteHandler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body)))

			if rec.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d: %s", tt.wantStatus, rec.Code, rec.Body.String())
			}
			if tt.wantStatus != http.StatusOK {
				if m.createVariante {
					t.Error("expected the rejected body not to reach the command")
				}
				return
			}
			if (m.pfandVarianteID == nil) != (tt.wantPfand == nil) || (m.pfandVarianteID != nil && *m.pfandVarianteID != *tt.wantPfand) {
				t.Errorf("expected pfandVarianteID %v, got %v", tt.wantPfand, m.pfandVarianteID)
			}
		})
	}
}
