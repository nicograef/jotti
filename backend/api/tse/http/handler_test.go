//go:build unit

package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/nicograef/jotti/backend/repository/tse_repo"
)

type mockNachsignierQuery struct {
	result []tse_repo.NachsignierAuftrag
	err    error
}

func (m *mockNachsignierQuery) GetTSENachsignierAuftraege(context.Context) ([]tse_repo.NachsignierAuftrag, error) {
	return m.result, m.err
}

type mockNachsignierCommand struct {
	zurueckgesetzt int
	verworfen      int
	err            error
}

func (m *mockNachsignierCommand) TSENachsignierAuftragZuruecksetzen(_ context.Context, id int) error {
	m.zurueckgesetzt = id
	return m.err
}

func (m *mockNachsignierCommand) TSENachsignierAuftragVerwerfen(_ context.Context, id int) error {
	m.verworfen = id
	return m.err
}

func postJSON(t *testing.T, handler http.HandlerFunc, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func TestGetTSENachsignierAuftraegeHandler_Success(t *testing.T) {
	erledigtAm := time.Date(2026, 6, 11, 12, 30, 0, 0, time.UTC)
	query := &mockNachsignierQuery{result: []tse_repo.NachsignierAuftrag{
		{
			ID:            3,
			TxID:          "8c0f9c4e-3a52-4f5d-9e6b-2d1c7a8b4f01",
			ProcessType:   "Kassenbeleg-V1",
			Status:        "fehlgeschlagen",
			Versuche:      10,
			LetzterFehler: "fiskaly timeout",
			ErstelltAm:    time.Date(2026, 6, 11, 10, 0, 0, 0, time.UTC),
		},
		{
			ID:          2,
			TxID:        "11111111-2222-4333-8444-555555555555",
			ProcessType: "Bestellung-V1",
			Status:      "erledigt",
			ErstelltAm:  time.Date(2026, 6, 11, 9, 0, 0, 0, time.UTC),
			ErledigtAm:  &erledigtAm,
		},
	}}
	handler := &QueryHandler{Query: query}

	rec := postJSON(t, handler.GetTSENachsignierAuftraegeHandler(), "/admin/get-tse-nachsignier-auftraege", "{}")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Auftraege []struct {
			ID            int     `json:"id"`
			TxID          string  `json:"txId"`
			ProcessType   string  `json:"processType"`
			Status        string  `json:"status"`
			Versuche      int     `json:"versuche"`
			LetzterFehler string  `json:"letzterFehler"`
			ErstelltAm    string  `json:"erstelltAm"`
			ErledigtAm    *string `json:"erledigtAm"`
		} `json:"auftraege"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp.Auftraege) != 2 {
		t.Fatalf("expected 2 auftraege, got %d", len(resp.Auftraege))
	}
	a := resp.Auftraege[0]
	if a.ID != 3 || a.Status != "fehlgeschlagen" || a.Versuche != 10 || a.LetzterFehler != "fiskaly timeout" {
		t.Fatalf("unexpected auftrag DTO: %+v", a)
	}
	if a.ErledigtAm != nil {
		t.Fatalf("expected null erledigtAm for fehlgeschlagenen auftrag, got %v", *a.ErledigtAm)
	}
	if resp.Auftraege[1].ErledigtAm == nil {
		t.Fatal("expected erledigtAm for erledigten auftrag")
	}
}

func TestGetTSENachsignierAuftraegeHandler_EmptyList(t *testing.T) {
	handler := &QueryHandler{Query: &mockNachsignierQuery{}}

	rec := postJSON(t, handler.GetTSENachsignierAuftraegeHandler(), "/admin/get-tse-nachsignier-auftraege", "{}")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"auftraege":[]`) {
		t.Fatalf("expected empty array, got %s", rec.Body.String())
	}
}

func TestTSENachsignierAuftragZuruecksetzenHandler_Success(t *testing.T) {
	cmd := &mockNachsignierCommand{}
	handler := &CommandHandler{Command: cmd}

	rec := postJSON(t, handler.TSENachsignierAuftragZuruecksetzenHandler(), "/admin/tse-nachsignier-auftrag-zuruecksetzen", `{"id":42}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if cmd.zurueckgesetzt != 42 {
		t.Fatalf("expected command called with id 42, got %d", cmd.zurueckgesetzt)
	}
}

func TestTSENachsignierAuftragVerwerfenHandler_Success(t *testing.T) {
	cmd := &mockNachsignierCommand{}
	handler := &CommandHandler{Command: cmd}

	rec := postJSON(t, handler.TSENachsignierAuftragVerwerfenHandler(), "/admin/tse-nachsignier-auftrag-verwerfen", `{"id":42}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if cmd.verworfen != 42 {
		t.Fatalf("expected command called with id 42, got %d", cmd.verworfen)
	}
}

func TestTSENachsignierCommandHandlers_RejectInvalidID(t *testing.T) {
	cmd := &mockNachsignierCommand{}
	handlers := map[string]http.HandlerFunc{
		"zuruecksetzen": (&CommandHandler{Command: cmd}).TSENachsignierAuftragZuruecksetzenHandler(),
		"verwerfen":     (&CommandHandler{Command: cmd}).TSENachsignierAuftragVerwerfenHandler(),
	}

	for name, handler := range handlers {
		for _, body := range []string{`{"id":0}`, `{}`} {
			rec := postJSON(t, handler, "/admin/tse-nachsignier-auftrag-"+name, body)
			if rec.Code != http.StatusBadRequest {
				t.Errorf("%s with body %s: expected status 400, got %d", name, body, rec.Code)
			}
		}
	}
	if cmd.zurueckgesetzt != 0 || cmd.verworfen != 0 {
		t.Fatalf("expected command not called for invalid IDs, got zurueckgesetzt=%d verworfen=%d", cmd.zurueckgesetzt, cmd.verworfen)
	}
}
