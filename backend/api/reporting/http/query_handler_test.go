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

	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/reporting"
	"github.com/nicograef/jotti/backend/domain/steuer"
)

type mockQuery struct {
	data            reporting.ReportingData
	liveData        *reporting.LiveReportingData
	kassensitzungen []kasse.Kassensitzung
	err             error
}

func (m mockQuery) GetReporting(_ context.Context, _ int) (reporting.ReportingData, error) {
	return m.data, m.err
}

func (m mockQuery) GetEigeneUebersicht(_ context.Context, _ int) (reporting.EigeneUebersicht, error) {
	return reporting.EigeneUebersicht{}, m.err
}

func (m mockQuery) GetAbgeschlosseneKassensitzungen(_ context.Context) ([]kasse.Kassensitzung, error) {
	return m.kassensitzungen, m.err
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
			GesamtUmsatzCents:        10000,
			AnzahlBestellungen:       3,
			AnzahlDirektverkaeufe:    4,
			DirektverkaufUmsatzCents: 2750,
		},
		Breakdowns: reporting.Breakdowns{
			UmsatzProServicekraft: []reporting.UmsatzServicekraft{},
		},
		UmsatzProSteuersatz: []reporting.UmsatzSteuersatz{
			{Satz: steuer.RegelSteuersatz, BruttoCents: 1190, NettoCents: 1000, SteuerCents: 190},
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
			GesamtUmsatzCents        int `json:"gesamtUmsatzCents"`
			AnzahlBestellungen       int `json:"anzahlBestellungen"`
			AnzahlDirektverkaeufe    int `json:"anzahlDirektverkaeufe"`
			DirektverkaufUmsatzCents int `json:"direktverkaufUmsatzCents"`
		} `json:"summary"`
		UmsatzProSteuersatz []struct {
			Satz        string `json:"satz"`
			BruttoCents int    `json:"bruttoCents"`
			NettoCents  int    `json:"nettoCents"`
			SteuerCents int    `json:"steuerCents"`
		} `json:"umsatzProSteuersatz"`
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
	if resp.Summary.AnzahlDirektverkaeufe != 4 {
		t.Errorf("expected anzahlDirektverkaeufe 4, got %d", resp.Summary.AnzahlDirektverkaeufe)
	}
	if resp.Summary.DirektverkaufUmsatzCents != 2750 {
		t.Errorf("expected direktverkaufUmsatzCents 2750, got %d", resp.Summary.DirektverkaufUmsatzCents)
	}
	if len(resp.UmsatzProSteuersatz) != 1 {
		t.Fatalf("expected 1 umsatzProSteuersatz row, got %d", len(resp.UmsatzProSteuersatz))
	}
	if resp.UmsatzProSteuersatz[0].Satz != "regel" || resp.UmsatzProSteuersatz[0].BruttoCents != 1190 {
		t.Fatalf("unexpected umsatzProSteuersatz row: %+v", resp.UmsatzProSteuersatz[0])
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

func TestGetAbgeschlosseneKassensitzungenHandler_ReturnsItems(t *testing.T) {
	handler := QueryHandler{Query: mockQuery{kassensitzungen: []kasse.Kassensitzung{
		{ZNr: 2, Datum: time.Date(2026, 7, 5, 0, 0, 0, 0, time.UTC), Bezeichnung: "Sommerfest Samstag", Status: kasse.KassensitzungAbgeschlossen},
	}}}

	req := httptest.NewRequest(http.MethodPost, "/get-abgeschlossene-kassensitzungen", strings.NewReader("{}"))
	rec := httptest.NewRecorder()

	handler.GetAbgeschlosseneKassensitzungenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Kassensitzungen []struct {
			ZNr         int    `json:"zNr"`
			Datum       string `json:"datum"`
			Bezeichnung string `json:"bezeichnung"`
			Status      string `json:"status"`
		} `json:"kassensitzungen"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("expected valid JSON response: %v", err)
	}
	if len(resp.Kassensitzungen) != 1 {
		t.Fatalf("expected 1 kassensitzung, got %d", len(resp.Kassensitzungen))
	}
	item := resp.Kassensitzungen[0]
	if item.ZNr != 2 || item.Datum != "2026-07-05" || item.Status != "abgeschlossen" {
		t.Fatalf("unexpected kassensitzung item: %+v", item)
	}
}

func TestGetAbgeschlosseneKassensitzungenHandler_QueryError_Returns500(t *testing.T) {
	handler := QueryHandler{Query: mockQuery{err: errors.New("db error")}}

	req := httptest.NewRequest(http.MethodPost, "/get-abgeschlossene-kassensitzungen", strings.NewReader("{}"))
	rec := httptest.NewRecorder()

	handler.GetAbgeschlosseneKassensitzungenHandler().ServeHTTP(rec, req)

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
		Summary: reporting.Summary{
			GesamtUmsatzCents:        45000,
			AnzahlDirektverkaeufe:    2,
			DirektverkaufUmsatzCents: 1800,
		},
		Breakdowns: reporting.Breakdowns{
			UmsatzProServicekraft: []reporting.UmsatzServicekraft{},
		},
		Servicekraefte: []reporting.ServicekraftLive{
			{
				UserID:         7,
				UserName:       "Anna",
				ZahlungenCents: 1500,
				OffeneTische: []reporting.OffeneArbeitTisch{
					{TischID: 3, TischName: "Tisch 3", AnzahlUnbezahlt: 1, AnzahlOffen: 1},
				},
				Erledigt: false,
			},
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
			GesamtUmsatzCents        int `json:"gesamtUmsatzCents"`
			AnzahlDirektverkaeufe    int `json:"anzahlDirektverkaeufe"`
			DirektverkaufUmsatzCents int `json:"direktverkaufUmsatzCents"`
		} `json:"summary"`
		Breakdowns struct {
			Servicekraefte []struct {
				UserID       int  `json:"userId"`
				Erledigt     bool `json:"erledigt"`
				OffeneTische []struct {
					TischName   string `json:"tischName"`
					AnzahlOffen int    `json:"anzahlOffen"`
				} `json:"offeneTische"`
			} `json:"servicekraefte"`
		} `json:"breakdowns"`
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
	if resp.Summary.AnzahlDirektverkaeufe != 2 {
		t.Errorf("expected anzahlDirektverkaeufe 2, got %d", resp.Summary.AnzahlDirektverkaeufe)
	}
	if resp.Summary.DirektverkaufUmsatzCents != 1800 {
		t.Errorf("expected direktverkaufUmsatzCents 1800, got %d", resp.Summary.DirektverkaufUmsatzCents)
	}
	if len(resp.Breakdowns.Servicekraefte) != 1 {
		t.Fatalf("expected 1 servicekraft, got %+v", resp.Breakdowns.Servicekraefte)
	}
	sk := resp.Breakdowns.Servicekraefte[0]
	if sk.UserID != 7 || sk.Erledigt {
		t.Errorf("expected servicekraft 7 with open work, got %+v", sk)
	}
	if len(sk.OffeneTische) != 1 || sk.OffeneTische[0].TischName != "Tisch 3" || sk.OffeneTische[0].AnzahlOffen != 1 {
		t.Errorf("expected open work at 'Tisch 3', got %+v", sk.OffeneTische)
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
