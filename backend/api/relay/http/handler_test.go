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

	relayApp "github.com/nicograef/jotti/backend/api/relay/application"
)

type mockQueryRepo struct {
	offene []relayApp.DruckAuftrag
	err    error
}

func (m *mockQueryRepo) GetOffeneDruckauftraege(_ context.Context) ([]relayApp.DruckAuftrag, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.offene, nil
}

type mockCommandRepo struct {
	lastIDs []int
	err     error
}

func (m *mockCommandRepo) QuittiereGedruckteAuftraege(_ context.Context, ids []int) error {
	if m.err != nil {
		return m.err
	}
	m.lastIDs = append([]int(nil), ids...)
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

func makeHandler(relayToken string, offene []relayApp.DruckAuftrag, cmdErr error) (*Handler, *mockCommandRepo) {
	cmdRepo := &mockCommandRepo{err: cmdErr}

	h := &Handler{
		RelayToken: relayToken,
		Query: relayApp.Query{
			DruckauftragRepo: &mockQueryRepo{offene: offene},
		},
		Command: relayApp.Command{
			DruckauftragRepo: cmdRepo,
		},
	}

	return h, cmdRepo
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
	h, _ := makeHandler("relay-secret", []relayApp.DruckAuftrag{{ID: 7, ZielIP: "192.168.1.20", Payload: "AAA="}}, nil)

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
	h, _ := makeHandler("relay-secret", nil, nil)

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
	h, _ := makeHandler("", nil, nil)

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

func TestQuittierenHandler_AcceptsMatchingToken(t *testing.T) {
	h, cmdRepo := makeHandler("relay-secret", nil, nil)

	rr := performJSONRequest(t, h.QuittierenHandler(), map[string]any{"token": "relay-secret", "gedruckteIds": []int{4, 5}})
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	if len(cmdRepo.lastIDs) != 2 || cmdRepo.lastIDs[0] != 4 || cmdRepo.lastIDs[1] != 5 {
		t.Fatalf("expected ids [4 5], got %v", cmdRepo.lastIDs)
	}
}

func TestQuittierenHandler_RejectsEmptyAndWrongToken(t *testing.T) {
	h, cmdRepo := makeHandler("relay-secret", nil, errors.New("should not be called"))

	for _, token := range []string{"", "wrong-token"} {
		rr := performJSONRequest(t, h.QuittierenHandler(), map[string]any{"token": token, "gedruckteIds": []int{1}})
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("token %q: expected status 400, got %d", token, rr.Code)
		}
	}
	if len(cmdRepo.lastIDs) != 0 {
		t.Fatalf("expected command repo not called, got ids %v", cmdRepo.lastIDs)
	}
}
