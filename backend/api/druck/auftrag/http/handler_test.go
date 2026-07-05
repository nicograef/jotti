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

	"github.com/nicograef/jotti/backend/repository/druckauftrag_repo"
)

type mockDruckauftragQuery struct {
	result []druckauftrag_repo.FehlgeschlagenerDruckauftrag
	err    error
}

func (m *mockDruckauftragQuery) GetFehlgeschlageneDruckauftraege(context.Context) ([]druckauftrag_repo.FehlgeschlagenerDruckauftrag, error) {
	return m.result, m.err
}

type mockDruckauftragCommand struct {
	erneutID  int
	verworfen int
	err       error
}

func (m *mockDruckauftragCommand) RetryDruckauftrag(_ context.Context, id int) error {
	m.erneutID = id
	return m.err
}

func (m *mockDruckauftragCommand) DiscardDruckauftrag(_ context.Context, id int) error {
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

func TestGetFehlgeschlageneDruckauftraegeHandler_Success(t *testing.T) {
	query := &mockDruckauftragQuery{result: []druckauftrag_repo.FehlgeschlagenerDruckauftrag{
		{
			ID:            7,
			BonArt:        "arbeitsbon",
			ZielIP:        "192.168.1.51",
			Referenz:      "bestellung-aufgenommen:3",
			Versuche:      3,
			LetzterFehler: "drucker nicht erreichbar",
			ErstelltAm:    time.Now(),
		},
	}}
	handler := &QueryHandler{Query: query}

	rec := postJSON(t, handler.GetFehlgeschlageneDruckauftraegeHandler(), "/admin/get-fehlgeschlagene-druckauftraege", "{}")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Druckauftraege []struct {
			ID            int    `json:"id"`
			BonArt        string `json:"bonArt"`
			ZielIP        string `json:"zielIp"`
			Referenz      string `json:"referenz"`
			Versuche      int    `json:"versuche"`
			LetzterFehler string `json:"letzterFehler"`
		} `json:"druckauftraege"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp.Druckauftraege) != 1 {
		t.Fatalf("expected 1 auftrag, got %d", len(resp.Druckauftraege))
	}
	a := resp.Druckauftraege[0]
	if a.ID != 7 || a.BonArt != "arbeitsbon" || a.ZielIP != "192.168.1.51" || a.Versuche != 3 || a.LetzterFehler != "drucker nicht erreichbar" {
		t.Fatalf("unexpected auftrag DTO: %+v", a)
	}
}

func TestGetFehlgeschlageneDruckauftraegeHandler_EmptyList(t *testing.T) {
	handler := &QueryHandler{Query: &mockDruckauftragQuery{}}

	rec := postJSON(t, handler.GetFehlgeschlageneDruckauftraegeHandler(), "/admin/get-fehlgeschlagene-druckauftraege", "{}")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"druckauftraege":[]`) {
		t.Fatalf("expected empty array, got %s", rec.Body.String())
	}
}

func TestRetryDruckauftragHandler_Success(t *testing.T) {
	cmd := &mockDruckauftragCommand{}
	handler := &CommandHandler{Command: cmd}

	rec := postJSON(t, handler.RetryDruckauftragHandler(), "/admin/druckauftrag-erneut-versuchen", `{"id":42}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if cmd.erneutID != 42 {
		t.Fatalf("expected command called with id 42, got %d", cmd.erneutID)
	}
}

func TestDiscardDruckauftragHandler_Success(t *testing.T) {
	cmd := &mockDruckauftragCommand{}
	handler := &CommandHandler{Command: cmd}

	rec := postJSON(t, handler.DiscardDruckauftragHandler(), "/admin/druckauftrag-verwerfen", `{"id":42}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if cmd.verworfen != 42 {
		t.Fatalf("expected command called with id 42, got %d", cmd.verworfen)
	}
}

func TestDruckauftragCommandHandlers_RejectInvalidID(t *testing.T) {
	cmd := &mockDruckauftragCommand{}
	handlers := map[string]http.HandlerFunc{
		"erneut-versuchen": (&CommandHandler{Command: cmd}).RetryDruckauftragHandler(),
		"verwerfen":        (&CommandHandler{Command: cmd}).DiscardDruckauftragHandler(),
	}

	for name, handler := range handlers {
		for _, body := range []string{`{"id":0}`, `{}`} {
			rec := postJSON(t, handler, "/admin/"+name, body)
			if rec.Code != http.StatusBadRequest {
				t.Errorf("%s with body %s: expected status 400, got %d", name, body, rec.Code)
			}
		}
	}
	if cmd.erneutID != 0 || cmd.verworfen != 0 {
		t.Fatalf("expected command not called for invalid IDs, got erneut=%d verworfen=%d", cmd.erneutID, cmd.verworfen)
	}
}
