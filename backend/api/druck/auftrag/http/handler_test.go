package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/nicograef/jotti/backend/domain/druckstation"
)

type mockDruckauftragQuery struct {
	result []druckstation.FehlgeschlagenerDruckauftrag
	err    error
}

func (m *mockDruckauftragQuery) GetFehlgeschlageneDruckauftraege(context.Context) ([]druckstation.FehlgeschlagenerDruckauftrag, error) {
	return m.result, m.err
}

// mockDruckauftragCommand records the IDs a handler forwards.
type mockDruckauftragCommand struct {
	erneutID  int
	verworfen int
}

func (m *mockDruckauftragCommand) RetryDruckauftrag(_ context.Context, id int) error {
	m.erneutID = id
	return nil
}

func (m *mockDruckauftragCommand) DiscardDruckauftrag(_ context.Context, id int) error {
	m.verworfen = id
	return nil
}

func (m *mockDruckauftragCommand) DiscardAlleFehlgeschlagenen(context.Context) (int64, error) {
	return 0, nil
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
	query := &mockDruckauftragQuery{result: []druckstation.FehlgeschlagenerDruckauftrag{
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
			ID              int    `json:"id"`
			BonArt          string `json:"bonArt"`
			ZielIP          string `json:"zielIp"`
			Referenz        string `json:"referenz"`
			ReferenzAnzeige string `json:"referenzAnzeige"`
			Versuche        int    `json:"versuche"`
			FehlerMeldung   string `json:"fehlerMeldung"`
		} `json:"druckauftraege"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp.Druckauftraege) != 1 {
		t.Fatalf("expected 1 auftrag, got %d", len(resp.Druckauftraege))
	}
	a := resp.Druckauftraege[0]
	if a.ID != 7 || a.BonArt != "arbeitsbon" || a.ZielIP != "192.168.1.51" || a.Versuche != 3 ||
		a.ReferenzAnzeige != "Bestellung Nr. 3" || a.FehlerMeldung != "Drucker nicht erreichbar" {
		t.Errorf("unexpected auftrag DTO: %+v", a)
	}
}

func TestGetFehlgeschlageneDruckauftraegeHandler_EmptyList(t *testing.T) {
	handler := &QueryHandler{Query: &mockDruckauftragQuery{}}

	rec := postJSON(t, handler.GetFehlgeschlageneDruckauftraegeHandler(), "/admin/get-fehlgeschlagene-druckauftraege", "{}")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"druckauftraege":[]`) {
		t.Errorf("expected empty array, got %s", rec.Body.String())
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
		t.Errorf("expected command not called for invalid IDs, got erneut=%d verworfen=%d", cmd.erneutID, cmd.verworfen)
	}
}
