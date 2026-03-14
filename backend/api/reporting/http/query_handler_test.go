//go:build unit

package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/nicograef/jotti/backend/domain/reporting"
)

type mockQuery struct {
	data reporting.ReportingData
	err  error
}

func (m mockQuery) GetReporting(_ context.Context, _ reporting.Zeitraum) (reporting.ReportingData, error) {
	return m.data, m.err
}

func TestGetReportingHandler_InvalidJSON(t *testing.T) {
	handler := QueryHandler{}

	req := httptest.NewRequest(http.MethodPost, "/get-reporting", strings.NewReader("{invalid"))
	rec := httptest.NewRecorder()

	handler.GetReportingHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestGetReportingHandler_MissingVon(t *testing.T) {
	handler := QueryHandler{}

	body := `{"bis": "2026-03-14T03:00:00Z"}`
	req := httptest.NewRequest(http.MethodPost, "/get-reporting", strings.NewReader(body))
	rec := httptest.NewRecorder()

	handler.GetReportingHandler().ServeHTTP(rec, req)

	assertBadRequestCode(t, rec, "invalid_zeitraum")
}

func TestGetReportingHandler_MissingBis(t *testing.T) {
	handler := QueryHandler{}

	body := `{"von": "2026-03-13T17:00:00Z"}`
	req := httptest.NewRequest(http.MethodPost, "/get-reporting", strings.NewReader(body))
	rec := httptest.NewRecorder()

	handler.GetReportingHandler().ServeHTTP(rec, req)

	assertBadRequestCode(t, rec, "invalid_zeitraum")
}

func TestGetReportingHandler_VonAfterBis(t *testing.T) {
	handler := QueryHandler{}

	body := `{"von": "2026-03-14T03:00:00Z", "bis": "2026-03-13T17:00:00Z"}`
	req := httptest.NewRequest(http.MethodPost, "/get-reporting", strings.NewReader(body))
	rec := httptest.NewRecorder()

	handler.GetReportingHandler().ServeHTTP(rec, req)

	assertBadRequestCode(t, rec, "invalid_zeitraum")
}

func TestGetReportingHandler_VonEqualsBis(t *testing.T) {
	handler := QueryHandler{}

	body := `{"von": "2026-03-13T17:00:00Z", "bis": "2026-03-13T17:00:00Z"}`
	req := httptest.NewRequest(http.MethodPost, "/get-reporting", strings.NewReader(body))
	rec := httptest.NewRecorder()

	handler.GetReportingHandler().ServeHTTP(rec, req)

	assertBadRequestCode(t, rec, "invalid_zeitraum")
}

func TestGetReportingHandler_NonUTCZeitraum(t *testing.T) {
	handler := QueryHandler{}

	body := `{"von": "2026-03-13T17:00:00+02:00", "bis": "2026-03-13T18:00:00+02:00"}`
	req := httptest.NewRequest(http.MethodPost, "/get-reporting", strings.NewReader(body))
	rec := httptest.NewRecorder()

	handler.GetReportingHandler().ServeHTTP(rec, req)

	assertBadRequestCode(t, rec, "invalid_timezone")
}

func assertBadRequestCode(t *testing.T, rec *httptest.ResponseRecorder, expectedCode string) {
	t.Helper()

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}

	var payload struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("expected valid JSON body: %v", err)
	}

	if payload.Code != expectedCode {
		t.Fatalf("expected code %q, got %q", expectedCode, payload.Code)
	}
}

func TestGetReportingHandler_ValidRequest_ReturnsReportingData(t *testing.T) {
	mockData := reporting.ReportingData{
		Zeitraum: reporting.Zeitraum{
			Von: time.Date(2026, 3, 14, 0, 0, 0, 0, time.UTC),
			Bis: time.Date(2026, 3, 14, 23, 59, 0, 0, time.UTC),
		},
		Summary: reporting.Summary{
			GesamtUmsatzCents:  10000,
			AnzahlBestellungen: 3,
		},
		Breakdowns: reporting.Breakdowns{
			UmsatzProServicekraft: []reporting.UmsatzServicekraft{},
			UmsatzProTisch:        []reporting.UmsatzTisch{},
		},
		Stornierungen: []reporting.StornierungDetail{},
	}
	handler := QueryHandler{Query: mockQuery{data: mockData}}

	body := `{"von": "2026-03-14T00:00:00Z", "bis": "2026-03-14T23:59:00Z"}`
	req := httptest.NewRequest(http.MethodPost, "/get-reporting", strings.NewReader(body))
	rec := httptest.NewRecorder()

	handler.GetReportingHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Summary struct {
			GesamtUmsatzCents  int `json:"gesamtUmsatzCents"`
			AnzahlBestellungen int `json:"anzahlBestellungen"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("expected valid JSON response: %v", err)
	}
	if resp.Summary.GesamtUmsatzCents != 10000 {
		t.Errorf("expected gesamtUmsatzCents 10000, got %d", resp.Summary.GesamtUmsatzCents)
	}
	if resp.Summary.AnzahlBestellungen != 3 {
		t.Errorf("expected anzahlBestellungen 3, got %d", resp.Summary.AnzahlBestellungen)
	}
}

func TestGetReportingHandler_QueryError_Returns500(t *testing.T) {
	handler := QueryHandler{Query: mockQuery{err: errors.New("db error")}}

	body := `{"von": "2026-03-14T00:00:00Z", "bis": "2026-03-14T23:59:00Z"}`
	req := httptest.NewRequest(http.MethodPost, "/get-reporting", strings.NewReader(body))
	rec := httptest.NewRecorder()

	handler.GetReportingHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}
}
