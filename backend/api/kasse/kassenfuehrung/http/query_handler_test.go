package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nicograef/jotti/backend/domain/kasse"
)

// mockQuery answers the Geldtransit list; the nil embedded interface panics on
// any other call, so a rejected body that still reaches the query fails the test.
type mockQuery struct {
	query
	geldtransit []kasse.Geldtransit
}

func (m *mockQuery) GetGeldtransitListe(context.Context, int) ([]kasse.Geldtransit, error) {
	return m.geldtransit, nil
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
