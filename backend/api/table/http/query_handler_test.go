//go:build unit

package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nicograef/jotti/backend/api/table/application"
	"github.com/nicograef/jotti/backend/domain/table"
)

type mockQuery struct {
	tisch    table.Tisch
	order    table.Bestellung
	position table.Position
	balance  int
	err      error
}

func (m mockQuery) GetAllTische(ctx context.Context) ([]table.Tisch, error) {
	return []table.Tisch{m.tisch}, m.err
}

func (m mockQuery) GetAktiveTische(ctx context.Context) ([]table.Tisch, error) {
	return []table.Tisch{m.tisch}, m.err
}

func (m mockQuery) GetTischHistorie(ctx context.Context, tischID int) ([]any, error) {
	return []any{m.order}, m.err
}

func (m mockQuery) GetTischState(ctx context.Context, tischID int) (table.TischState, error) {
	return table.TischState{
		SaldoCents:             m.balance,
		UnbezahltePositionen:   []table.Position{m.position},
		UngeliefertePositionen: []table.Position{m.position},
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
