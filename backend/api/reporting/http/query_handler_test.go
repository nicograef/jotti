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
	"github.com/nicograef/jotti/backend/domain/steuer"
)

type mockQuery struct {
	data            reporting.ReportingData
	liveData        *reporting.LiveReportingData
	kassensitzungen []reporting.AbgeschlosseneSitzung
	err             error
}

func (m mockQuery) GetReporting(_ context.Context, _ int) (reporting.ReportingData, error) {
	return m.data, m.err
}

func (m mockQuery) GetEigeneUebersicht(_ context.Context, _ int) (reporting.EigeneUebersicht, error) {
	return reporting.EigeneUebersicht{}, m.err
}

func (m mockQuery) GetAbgeschlosseneKassensitzungen(_ context.Context) ([]reporting.AbgeschlosseneSitzung, error) {
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
	eroeffnetAm := time.Date(2026, 7, 5, 9, 58, 0, 0, time.UTC)
	abgeschlossenAm := time.Date(2026, 7, 5, 23, 12, 0, 0, time.UTC)
	kassensturzDifferenz := -150
	mockData := reporting.ReportingData{
		KassensitzungNr: 1,
		Metadaten: reporting.Metadaten{
			EroeffnetAm:               &eroeffnetAm,
			AbgeschlossenAm:           &abgeschlossenAm,
			AbgeschlossenVon:          "nico",
			KassensturzDifferenzCents: &kassensturzDifferenz,
		},
		Summary: reporting.Summary{
			GesamtUmsatzCents:        10000,
			AnzahlBestellungen:       3,
			AnzahlDirektverkaeufe:    4,
			DirektverkaufUmsatzCents: 2750,
		},
		Breakdowns: reporting.Breakdowns{
			AbrechnungProServicekraft: []reporting.AbrechnungServicekraft{
				{UserID: 3, UserName: "felix", Name: "Felix W.", KassiertCents: 5000, AnzahlZahlungen: 4, RuecknahmenCents: 800, AnzahlStornierungen: 2, AbzugebenCents: 4200},
			},
		},
		UmsatzProSteuersatz: []reporting.UmsatzSteuersatz{
			{Satz: steuer.RegelSteuersatz, BruttoCents: 1190, NettoCents: 1000, SteuerCents: 190},
		},
		Stornierungen: []reporting.StornierungDetail{},
		ProduktStatistik: []reporting.ProduktStatistik{
			{
				Kategorie: "essen", ProduktName: "Pommes", AusgegebeneMenge: 4, UmsatzCents: 900,
				Varianten: []reporting.VarianteStatistik{
					{VarianteID: 10, VarianteName: "groß", AusgegebeneMenge: 4, UmsatzCents: 900},
				},
			},
		},
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
		Metadaten struct {
			EroeffnetAm               *time.Time `json:"eroeffnetAm"`
			AbgeschlossenAm           *time.Time `json:"abgeschlossenAm"`
			AbgeschlossenVon          string     `json:"abgeschlossenVon"`
			KassensturzDifferenzCents *int       `json:"kassensturzDifferenzCents"`
		} `json:"metadaten"`
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
		Breakdowns struct {
			AbrechnungProServicekraft []struct {
				UserID              int `json:"userId"`
				KassiertCents       int `json:"kassiertCents"`
				AnzahlZahlungen     int `json:"anzahlZahlungen"`
				RuecknahmenCents    int `json:"ruecknahmenCents"`
				AnzahlStornierungen int `json:"anzahlStornierungen"`
				AbzugebenCents      int `json:"abzugebenCents"`
			} `json:"abrechnungProServicekraft"`
		} `json:"breakdowns"`
		ProduktStatistik []struct {
			Kategorie        string `json:"kategorie"`
			ProduktName      string `json:"produktName"`
			AusgegebeneMenge int    `json:"ausgegebeneMenge"`
			UmsatzCents      int    `json:"umsatzCents"`
			Varianten        []struct {
				VarianteID       int    `json:"varianteId"`
				VarianteName     string `json:"varianteName"`
				AusgegebeneMenge int    `json:"ausgegebeneMenge"`
				UmsatzCents      int    `json:"umsatzCents"`
			} `json:"varianten"`
		} `json:"produktStatistik"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("expected valid JSON response: %v", err)
	}
	if len(resp.ProduktStatistik) != 1 {
		t.Fatalf("expected 1 produktStatistik row, got %+v", resp.ProduktStatistik)
	}
	if p := resp.ProduktStatistik[0]; p.Kategorie != "essen" || p.ProduktName != "Pommes" || p.AusgegebeneMenge != 4 || p.UmsatzCents != 900 {
		t.Fatalf("unexpected produktStatistik row: %+v", resp.ProduktStatistik[0])
	}
	if v := resp.ProduktStatistik[0].Varianten; len(v) != 1 || v[0].VarianteID != 10 || v[0].VarianteName != "groß" {
		t.Fatalf("unexpected produktStatistik varianten: %+v", resp.ProduktStatistik[0].Varianten)
	}
	if len(resp.Breakdowns.AbrechnungProServicekraft) != 1 {
		t.Fatalf("expected 1 abrechnungProServicekraft row, got %+v", resp.Breakdowns.AbrechnungProServicekraft)
	}
	if sk := resp.Breakdowns.AbrechnungProServicekraft[0]; sk.UserID != 3 || sk.KassiertCents != 5000 || sk.AnzahlZahlungen != 4 || sk.RuecknahmenCents != 800 || sk.AnzahlStornierungen != 2 || sk.AbzugebenCents != 4200 {
		t.Fatalf("unexpected abrechnungProServicekraft row: %+v", resp.Breakdowns.AbrechnungProServicekraft[0])
	}
	if resp.Metadaten.EroeffnetAm == nil || !resp.Metadaten.EroeffnetAm.Equal(eroeffnetAm) {
		t.Errorf("expected eroeffnetAm %v, got %v", eroeffnetAm, resp.Metadaten.EroeffnetAm)
	}
	if resp.Metadaten.AbgeschlossenAm == nil || !resp.Metadaten.AbgeschlossenAm.Equal(abgeschlossenAm) {
		t.Errorf("expected abgeschlossenAm %v, got %v", abgeschlossenAm, resp.Metadaten.AbgeschlossenAm)
	}
	if resp.Metadaten.AbgeschlossenVon != "nico" {
		t.Errorf("expected abgeschlossenVon 'nico', got %q", resp.Metadaten.AbgeschlossenVon)
	}
	if resp.Metadaten.KassensturzDifferenzCents == nil || *resp.Metadaten.KassensturzDifferenzCents != -150 {
		t.Errorf("expected kassensturzDifferenzCents -150, got %v", resp.Metadaten.KassensturzDifferenzCents)
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
	abgeschlossenAm := time.Date(2026, 7, 5, 23, 12, 0, 0, time.UTC)
	handler := QueryHandler{Query: mockQuery{kassensitzungen: []reporting.AbgeschlosseneSitzung{
		{ZNr: 2, Datum: time.Date(2026, 7, 5, 0, 0, 0, 0, time.UTC), Bezeichnung: "Sommerfest Samstag", UmsatzGesamtCents: 341200, AbgeschlossenAm: &abgeschlossenAm},
	}}}

	req := httptest.NewRequest(http.MethodPost, "/get-abgeschlossene-kassensitzungen", strings.NewReader("{}"))
	rec := httptest.NewRecorder()

	handler.GetAbgeschlosseneKassensitzungenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Kassensitzungen []struct {
			ZNr               int        `json:"zNr"`
			Datum             string     `json:"datum"`
			Bezeichnung       string     `json:"bezeichnung"`
			UmsatzGesamtCents int        `json:"umsatzGesamtCents"`
			AbgeschlossenAm   *time.Time `json:"abgeschlossenAm"`
		} `json:"kassensitzungen"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("expected valid JSON response: %v", err)
	}
	if len(resp.Kassensitzungen) != 1 {
		t.Fatalf("expected 1 kassensitzung, got %d", len(resp.Kassensitzungen))
	}
	item := resp.Kassensitzungen[0]
	if item.ZNr != 2 || item.Datum != "2026-07-05" || item.UmsatzGesamtCents != 341200 {
		t.Fatalf("unexpected kassensitzung item: %+v", item)
	}
	if item.AbgeschlossenAm == nil || !item.AbgeschlossenAm.Equal(abgeschlossenAm) {
		t.Fatalf("expected abgeschlossenAm %v, got %v", abgeschlossenAm, item.AbgeschlossenAm)
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
			AbrechnungProServicekraft: []reporting.AbrechnungServicekraft{},
		},
		Servicekraefte: []reporting.ServicekraftLive{
			{
				UserID:              7,
				UserName:            "Anna",
				KassiertCents:       1500,
				RuecknahmenCents:    250,
				AnzahlStornierungen: 1,
				AbzugebenCents:      1250,
				OffenCents:          900,
				OffeneTische: []reporting.OffeneArbeitTisch{
					{TischID: 3, TischName: "Tisch 3", AnzahlOffen: 1, OffenCents: 900},
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
				UserID              int  `json:"userId"`
				KassiertCents       int  `json:"kassiertCents"`
				RuecknahmenCents    int  `json:"ruecknahmenCents"`
				AnzahlStornierungen int  `json:"anzahlStornierungen"`
				AbzugebenCents      int  `json:"abzugebenCents"`
				Erledigt            bool `json:"erledigt"`
				OffenCents          int  `json:"offenCents"`
				OffeneTische        []struct {
					TischName string `json:"tischName"`
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
	if sk.KassiertCents != 1500 || sk.RuecknahmenCents != 250 || sk.AnzahlStornierungen != 1 || sk.AbzugebenCents != 1250 {
		t.Errorf("expected the live row to carry the abrechnung, got %+v", sk)
	}
	if sk.OffenCents != 900 {
		t.Errorf("expected servicekraft offenCents 900, got %d", sk.OffenCents)
	}
	if len(sk.OffeneTische) != 1 || sk.OffeneTische[0].TischName != "Tisch 3" {
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
