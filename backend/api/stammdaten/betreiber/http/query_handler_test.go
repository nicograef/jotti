//go:build unit

package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/nicograef/jotti/backend/domain/betreiber"
)

type mockQuery struct {
	betreiber betreiber.Betreiber
	err       error
}

func (m *mockQuery) GetBetreiber(_ context.Context) (betreiber.Betreiber, error) {
	return m.betreiber, m.err
}

func decodeBetreiberResponse(t *testing.T, rec *httptest.ResponseRecorder) betreiberResponse {
	t.Helper()
	var resp betreiberResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	return resp
}

func TestGetBetreiberHandler_ElsterNichtGemeldet(t *testing.T) {
	handler := &QueryHandler{Query: &mockQuery{betreiber: betreiber.Betreiber{
		Vereinsname:      "Musterverein e.V.",
		Strasse:          "Musterstraße 1",
		Plz:              "12345",
		Ort:              "Musterstadt",
		ElsterGemeldetAm: nil,
	}}}

	req := httptest.NewRequest(http.MethodPost, "/get-betreiber", nil)
	rec := httptest.NewRecorder()

	handler.GetBetreiberHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	resp := decodeBetreiberResponse(t, rec)
	if resp.ElsterGemeldetAm != nil {
		t.Errorf("expected elsterGemeldetAm null, got %q", *resp.ElsterGemeldetAm)
	}
	if resp.Vereinsname != "Musterverein e.V." {
		t.Errorf("expected vereinsname to be returned, got %q", resp.Vereinsname)
	}
}

func TestGetBetreiberHandler_ElsterGemeldet(t *testing.T) {
	gemeldet := time.Date(2026, 7, 12, 0, 0, 0, 0, time.UTC)
	handler := &QueryHandler{Query: &mockQuery{betreiber: betreiber.Betreiber{
		Vereinsname:      "Musterverein e.V.",
		Strasse:          "Musterstraße 1",
		Plz:              "12345",
		Ort:              "Musterstadt",
		ElsterGemeldetAm: &gemeldet,
	}}}

	req := httptest.NewRequest(http.MethodPost, "/get-betreiber", nil)
	rec := httptest.NewRecorder()

	handler.GetBetreiberHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	resp := decodeBetreiberResponse(t, rec)
	if resp.ElsterGemeldetAm == nil {
		t.Fatal("expected elsterGemeldetAm to be set, got null")
	}
	if *resp.ElsterGemeldetAm != "2026-07-12" {
		t.Errorf("expected elsterGemeldetAm 2026-07-12, got %q", *resp.ElsterGemeldetAm)
	}
}
