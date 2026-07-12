//go:build unit

package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nicograef/jotti/backend/api/stammdaten/tisch/application"
	t "github.com/nicograef/jotti/backend/domain/tisch"
)

type mockQuery struct {
	err error
}

func (m mockQuery) GetAllTische(ctx context.Context) ([]application.TischMitSaldo, error) {
	return []application.TischMitSaldo{
		{Tisch: t.Tisch{ID: 1, Name: "Tisch 1", Status: t.ActiveStatus}, SaldoCents: 9850},
		{Tisch: t.Tisch{ID: 2, Name: "Tisch 2", Status: t.ActiveStatus}, SaldoCents: 0},
	}, m.err
}

func TestGetAllTischeHandler_Success(t *testing.T) {
	handler := &QueryHandler{Query: mockQuery{}}

	req := httptest.NewRequest(http.MethodPost, "/get-all-tische", nil)
	rec := httptest.NewRecorder()

	handler.GetAllTischeHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	// saldoCents muss pro Tisch im JSON serialisiert werden — fängt eine
	// Regression, die das Feld beim DTO-Mapping wieder verliert.
	var body struct {
		Tische []struct {
			ID         int `json:"id"`
			SaldoCents int `json:"saldoCents"`
		} `json:"tische"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if len(body.Tische) != 2 {
		t.Fatalf("expected 2 tische, got %d", len(body.Tische))
	}
	if body.Tische[0].SaldoCents != 9850 {
		t.Errorf("expected tisch %d saldoCents=9850, got %d", body.Tische[0].ID, body.Tische[0].SaldoCents)
	}
	if body.Tische[1].SaldoCents != 0 {
		t.Errorf("expected tisch %d saldoCents=0, got %d", body.Tische[1].ID, body.Tische[1].SaldoCents)
	}
}

func TestGetAllTischeHandler_Failure(t *testing.T) {
	handler := &QueryHandler{Query: mockQuery{err: application.ErrDatabase}}

	req := httptest.NewRequest(http.MethodPost, "/get-all-tische", nil)
	rec := httptest.NewRecorder()

	handler.GetAllTischeHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}
}
