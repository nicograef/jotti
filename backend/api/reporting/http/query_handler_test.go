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

	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/reporting"
)

type mockQuery struct {
	data     reporting.ReportingData
	liveData *reporting.LiveReportingData
	err      error
}

func (m mockQuery) GetReporting(_ context.Context, _ int) (reporting.ReportingData, error) {
	return m.data, m.err
}

func (m mockQuery) GetEigeneUebersicht(_ context.Context, _ int) (reporting.EigeneUebersicht, error) {
	return reporting.EigeneUebersicht{}, m.err
}

func (m mockQuery) GetAllKassensitzungen(_ context.Context) ([]kasse.Kassensitzung, error) {
	return nil, m.err
}

func (m mockQuery) GetLiveReporting(_ context.Context) (*reporting.LiveReportingData, error) {
	return m.liveData, m.err
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

func TestGetReportingHandler_MissingKassensitzungNr(t *testing.T) {
	handler := QueryHandler{}

	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/get-reporting", strings.NewReader(body))
	rec := httptest.NewRecorder()

	handler.GetReportingHandler().ServeHTTP(rec, req)

	assertBadRequestCode(t, rec, "invalid_kassensitzung_nr")
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
		KassensitzungNr: 1,
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

	body := `{"kassensitzungNr": 1}`
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

	body := `{"kassensitzungNr": 1}`
	req := httptest.NewRequest(http.MethodPost, "/get-reporting", strings.NewReader(body))
	rec := httptest.NewRecorder()

	handler.GetReportingHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}
}

func TestGetLiveReportingHandler_KeineOffeneSitzung_ReturnsNull(t *testing.T) {
	handler := QueryHandler{Query: mockQuery{liveData: nil}}

	req := httptest.NewRequest(http.MethodPost, "/get-live-reporting", strings.NewReader("{}"))
	rec := httptest.NewRecorder()

	handler.GetLiveReportingHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if strings.TrimSpace(rec.Body.String()) != "null" {
		t.Errorf("expected null response body, got %q", rec.Body.String())
	}
}

func TestGetLiveReportingHandler_OffeneSitzung_ReturnsDaten(t *testing.T) {
	liveData := &reporting.LiveReportingData{
		KassensitzungNr:  42,
		Bezeichnung:      "Sommerfest",
		OffeneSaldiCents: 1200,
		OffeneTische: []reporting.OffenerTisch{
			{TischID: 3, TischName: "Tisch 3", SaldoCents: 1200},
		},
		Summary: reporting.Summary{GesamtUmsatzCents: 45000},
		Breakdowns: reporting.Breakdowns{
			UmsatzProServicekraft: []reporting.UmsatzServicekraft{},
			UmsatzProTisch:        []reporting.UmsatzTisch{},
		},
		Stornierungen: []reporting.StornierungDetail{},
	}
	handler := QueryHandler{Query: mockQuery{liveData: liveData}}

	req := httptest.NewRequest(http.MethodPost, "/get-live-reporting", strings.NewReader("{}"))
	rec := httptest.NewRecorder()

	handler.GetLiveReportingHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		KassensitzungNr  int    `json:"kassensitzungNr"`
		Bezeichnung      string `json:"bezeichnung"`
		OffeneSaldiCents int    `json:"offeneSaldiCents"`
		OffeneTische     []struct {
			TischID    int    `json:"tischId"`
			TischName  string `json:"tischName"`
			SaldoCents int    `json:"saldoCents"`
		} `json:"offeneTische"`
		Summary struct {
			GesamtUmsatzCents int `json:"gesamtUmsatzCents"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("expected valid JSON response: %v", err)
	}
	if resp.KassensitzungNr != 42 {
		t.Errorf("expected kassensitzungNr 42, got %d", resp.KassensitzungNr)
	}
	if resp.Bezeichnung != "Sommerfest" {
		t.Errorf("expected Bezeichnung 'Sommerfest', got %q", resp.Bezeichnung)
	}
	if resp.OffeneSaldiCents != 1200 {
		t.Errorf("expected offeneSaldiCents 1200, got %d", resp.OffeneSaldiCents)
	}
	if len(resp.OffeneTische) != 1 || resp.OffeneTische[0].TischName != "Tisch 3" {
		t.Errorf("expected 1 offener Tisch named 'Tisch 3', got %+v", resp.OffeneTische)
	}
	if resp.Summary.GesamtUmsatzCents != 45000 {
		t.Errorf("expected gesamtUmsatzCents 45000, got %d", resp.Summary.GesamtUmsatzCents)
	}
}

func TestGetLiveReportingHandler_QueryError_Returns500(t *testing.T) {
	handler := QueryHandler{Query: mockQuery{err: errors.New("db error")}}

	req := httptest.NewRequest(http.MethodPost, "/get-live-reporting", strings.NewReader("{}"))
	rec := httptest.NewRecorder()

	handler.GetLiveReportingHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}
}
