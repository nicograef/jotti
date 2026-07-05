//go:build unit

package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nicograef/jotti/backend/api/stammdaten/tisch/application"
	t "github.com/nicograef/jotti/backend/domain/tisch"
)

type mockQuery struct {
	tisch t.Tisch
	err   error
}

func (m mockQuery) GetAllTische(ctx context.Context) ([]t.Tisch, error) {
	return []t.Tisch{m.tisch}, m.err
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
