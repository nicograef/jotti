//go:build unit

package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockRepo struct {
	offene       []OffenerDruckauftrag
	gedruckteIDs []int
	fehlversuche []Fehlversuch
	err          error
}

func (m *mockRepo) GetOffeneDruckauftraege(_ context.Context) ([]OffenerDruckauftrag, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.offene, nil
}

func (m *mockRepo) ReportDruckergebnis(_ context.Context, gedruckteIDs []int, fehlversuche []Fehlversuch) error {
	if m.err != nil {
		return m.err
	}
	m.gedruckteIDs = append(m.gedruckteIDs, gedruckteIDs...)
	m.fehlversuche = append(m.fehlversuche, fehlversuche...)
	return nil
}

type apiError struct {
	Code string `json:"code"`
}

type pollResponseBody struct {
	Auftraege []struct {
		ID      int    `json:"id"`
		ZielIP  string `json:"zielIp"`
		Payload string `json:"payload"`
	} `json:"auftraege"`
}

func makeHandler(relayToken string, repo *mockRepo) *Handler {
	return &Handler{
		RelayToken: relayToken,
		Repo:       repo,
	}
}

func performJSONRequest(t *testing.T, handler http.HandlerFunc, body any) *httptest.ResponseRecorder {
	t.Helper()

	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler(rr, req)
	return rr
}

func TestPollHandler_AcceptsMatchingToken(t *testing.T) {
	h := makeHandler("relay-secret", &mockRepo{offene: []OffenerDruckauftrag{{ID: 7, ZielIP: "192.168.1.20", Payload: "AAA="}}})

	rr := performJSONRequest(t, h.PollHandler(), map[string]any{"token": "relay-secret"})
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var resp pollResponseBody
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp.Auftraege) != 1 {
		t.Fatalf("expected 1 auftrag, got %d", len(resp.Auftraege))
	}
	if resp.Auftraege[0].ID != 7 {
		t.Fatalf("expected auftrag id 7, got %d", resp.Auftraege[0].ID)
	}
}

func TestPollHandler_RejectsEmptyAndWrongToken(t *testing.T) {
	h := makeHandler("relay-secret", &mockRepo{})

	for _, token := range []string{"", "wrong-token"} {
		rr := performJSONRequest(t, h.PollHandler(), map[string]any{"token": token})
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("token %q: expected status 400, got %d", token, rr.Code)
		}

		var errResp apiError
		if err := json.Unmarshal(rr.Body.Bytes(), &errResp); err != nil {
			t.Fatalf("token %q: failed to decode error response: %v", token, err)
		}
		if errResp.Code != "unauthorized" {
			t.Fatalf("token %q: expected code unauthorized, got %q", token, errResp.Code)
		}
	}
}

func TestPollHandler_RejectsWhenConfiguredTokenIsEmpty(t *testing.T) {
	h := makeHandler("", &mockRepo{})

	rr := performJSONRequest(t, h.PollHandler(), map[string]any{"token": ""})
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}

	var errResp apiError
	if err := json.Unmarshal(rr.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}
	if errResp.Code != "unauthorized" {
		t.Fatalf("expected code unauthorized, got %q", errResp.Code)
	}
}

func TestErgebnisHandler_AcceptsMatchingToken(t *testing.T) {
	repo := &mockRepo{}
	h := makeHandler("relay-secret", repo)

	rr := performJSONRequest(t, h.ErgebnisHandler(), map[string]any{
		"token":        "relay-secret",
		"gedruckteIds": []int{4, 5},
		"fehlversuche": []map[string]any{{"id": 9, "fehler": "drucker nicht erreichbar"}},
	})
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	if len(repo.gedruckteIDs) != 2 || repo.gedruckteIDs[0] != 4 || repo.gedruckteIDs[1] != 5 {
		t.Fatalf("expected gedruckte ids [4 5], got %v", repo.gedruckteIDs)
	}
	if len(repo.fehlversuche) != 1 || repo.fehlversuche[0].ID != 9 || repo.fehlversuche[0].Fehler != "drucker nicht erreichbar" {
		t.Fatalf("expected one fehlversuch for id 9, got %v", repo.fehlversuche)
	}
}

func TestErgebnisHandler_RejectsEmptyAndWrongToken(t *testing.T) {
	repo := &mockRepo{err: errors.New("should not be called")}
	h := makeHandler("relay-secret", repo)

	for _, token := range []string{"", "wrong-token"} {
		rr := performJSONRequest(t, h.ErgebnisHandler(), map[string]any{"token": token, "gedruckteIds": []int{1}})
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("token %q: expected status 400, got %d", token, rr.Code)
		}
	}
	if len(repo.gedruckteIDs) != 0 {
		t.Fatalf("expected repo not called, got gedruckte ids %v", repo.gedruckteIDs)
	}
}
