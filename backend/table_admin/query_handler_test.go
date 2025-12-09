//go:build unit

package table_admin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockQuery struct {
	err error
}

func (m *mockQuery) GetAllTables(ctx context.Context) ([]Table, error) {
	return []Table{{ID: 1, Name: "Table 1", Status: ActiveStatus}}, m.err
}

func TestGetAllTablesHandler_Success(t *testing.T) {
	handler := &QueryHandler{Query: &mockQuery{}}

	req := httptest.NewRequest(http.MethodPost, "/get-all-tables", nil)
	rec := httptest.NewRecorder()

	handler.GetAllTablesHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestGetAllTablesHandler_Failure(t *testing.T) {
	handler := &QueryHandler{Query: &mockQuery{err: ErrDatabase}}

	req := httptest.NewRequest(http.MethodPost, "/get-all-tables", nil)
	rec := httptest.NewRecorder()

	handler.GetAllTablesHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}
}
