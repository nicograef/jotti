//go:build unit

package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nicograef/jotti/backend/domain/kasse"
)

var errDB = errors.New("db error")

type mockQuery struct {
	kassensitzung *kasse.Kassensitzung
	bestand       kasse.Kassenbestand
	geldtransit   []kasse.Geldtransit
	err           error
}

func (m *mockQuery) GetOffeneKassensitzung(_ context.Context) (*kasse.Kassensitzung, error) {
	return m.kassensitzung, m.err
}

func (m *mockQuery) GetKassenbestand(_ context.Context, _ int) (kasse.Kassenbestand, error) {
	return m.bestand, m.err
}

func (m *mockQuery) GetGeldtransitListe(_ context.Context, _ int) ([]kasse.Geldtransit, error) {
	return m.geldtransit, m.err
}

// GetOffeneKassensitzung

func TestGetOffeneKassensitzungHandler_Success(t *testing.T) {
	ks := &kasse.Kassensitzung{ZNr: 1, Bezeichnung: "Maihock", Status: kasse.KassensitzungOffen}
	handler := &QueryHandler{Query: &mockQuery{kassensitzung: ks}}

	req := httptest.NewRequest(http.MethodPost, "/get-offene-kassensitzung", nil)
	rec := httptest.NewRecorder()

	handler.GetOffeneKassensitzungHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestGetOffeneKassensitzungHandler_NoneOpen(t *testing.T) {
	handler := &QueryHandler{Query: &mockQuery{kassensitzung: nil}}

	req := httptest.NewRequest(http.MethodPost, "/get-offene-kassensitzung", nil)
	rec := httptest.NewRecorder()

	handler.GetOffeneKassensitzungHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestGetOffeneKassensitzungHandler_DBError(t *testing.T) {
	handler := &QueryHandler{Query: &mockQuery{err: errDB}}

	req := httptest.NewRequest(http.MethodPost, "/get-offene-kassensitzung", nil)
	rec := httptest.NewRecorder()

	handler.GetOffeneKassensitzungHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}

// GetKassenbestand

func TestGetKassenbestandHandler_Success(t *testing.T) {
	handler := &QueryHandler{Query: &mockQuery{bestand: kasse.Kassenbestand{SollBestandCents: 5000}}}

	body := `{"kassensitzungNr":1}`
	req := httptest.NewRequest(http.MethodPost, "/get-kassenbestand", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.GetKassenbestandHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestGetKassenbestandHandler_InvalidNr(t *testing.T) {
	handler := &QueryHandler{Query: &mockQuery{}}

	body := `{"kassensitzungNr":0}`
	req := httptest.NewRequest(http.MethodPost, "/get-kassenbestand", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.GetKassenbestandHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

// GetGeldtransitListe

func TestGetGeldtransitListeHandler_Success(t *testing.T) {
	handler := &QueryHandler{Query: &mockQuery{geldtransit: []kasse.Geldtransit{
		{Richtung: kasse.GeldtransitRichtungEinlage, BetragCents: 20000, Kommentar: "Wechselgeld", GebuchtVon: "sophie"},
	}}}

	body := `{"kassensitzungNr":1}`
	req := httptest.NewRequest(http.MethodPost, "/get-geldtransit-liste", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.GetGeldtransitListeHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "sophie") || !strings.Contains(rec.Body.String(), "einlage") {
		t.Errorf("expected geldtransit item in body, got %s", rec.Body.String())
	}
}

func TestGetGeldtransitListeHandler_InvalidNr(t *testing.T) {
	handler := &QueryHandler{Query: &mockQuery{}}

	body := `{"kassensitzungNr":0}`
	req := httptest.NewRequest(http.MethodPost, "/get-geldtransit-liste", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.GetGeldtransitListeHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestGetGeldtransitListeHandler_DBError(t *testing.T) {
	handler := &QueryHandler{Query: &mockQuery{err: errDB}}

	body := `{"kassensitzungNr":1}`
	req := httptest.NewRequest(http.MethodPost, "/get-geldtransit-liste", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.GetGeldtransitListeHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}
