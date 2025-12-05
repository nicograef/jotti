//go:build unit

package table

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type mockQuery struct {
	err error
}

func (m *mockQuery) GetTable(ctx context.Context, id int) (*Table, error) {
	return &Table{ID: id, Name: "Table Name", Status: ActiveStatus}, m.err
}

func (m *mockQuery) GetAllTables(ctx context.Context) ([]Table, error) {
	return []Table{{ID: 1, Name: "Table 1", Status: ActiveStatus}}, m.err
}

func (m *mockQuery) GetActiveTables(ctx context.Context) ([]TablePublic, error) {
	return []TablePublic{{ID: 1, Name: "Table 1"}}, m.err
}

func TestGetTableHandler_Success(t *testing.T) {
	handler := &QueryHandler{Query: &mockQuery{}}

	body := `{"id":1}`
	req := httptest.NewRequest(http.MethodPost, "/get-table", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.GetTableHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestGetTableHandler_NotFound(t *testing.T) {
	handler := &QueryHandler{Query: &mockQuery{err: ErrTableNotFound}}

	body := `{"id":999}`
	req := httptest.NewRequest(http.MethodPost, "/get-table", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.GetTableHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
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

func TestGetActiveTablesHandler_Success(t *testing.T) {
	handler := &QueryHandler{Query: &mockQuery{}}

	req := httptest.NewRequest(http.MethodPost, "/get-active-tables", nil)
	rec := httptest.NewRecorder()

	handler.GetActiveTablesHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestGetActiveTablesHandler_Failure(t *testing.T) {
	handler := &QueryHandler{Query: &mockQuery{err: ErrDatabase}}

	req := httptest.NewRequest(http.MethodPost, "/get-active-tables", nil)
	rec := httptest.NewRecorder()

	handler.GetActiveTablesHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}
}
