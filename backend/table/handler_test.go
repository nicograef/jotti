//go:build unit

package table

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type mockTableService struct {
	shouldFail bool
	table      *Table
}

func (m *mockTableService) CreateTable(ctx context.Context, name string) (*Table, error) {
	if m.shouldFail {
		return nil, ErrDatabase
	}
	return &Table{
		ID:     1,
		Name:   name,
		Status: InactiveStatus,
	}, nil
}

func (m *mockTableService) UpdateTable(ctx context.Context, id int, name string) (*Table, error) {
	if m.shouldFail {
		return nil, ErrTableNotFound
	}
	return &Table{
		ID:     id,
		Name:   name,
		Status: ActiveStatus,
	}, nil
}

func (m *mockTableService) GetTable(ctx context.Context, id int) (*Table, error) {
	if m.shouldFail {
		return nil, ErrTableNotFound
	}
	return &Table{
		ID:     id,
		Name:   "Table Name",
		Status: ActiveStatus,
	}, nil
}

func (m *mockTableService) GetAllTables(ctx context.Context) ([]*Table, error) {
	if m.shouldFail {
		return nil, ErrDatabase
	}
	return []*Table{
		{
			ID:     1,
			Name:   "Table 1",
			Status: ActiveStatus,
		},
	}, nil
}

func (m *mockTableService) GetActiveTables(ctx context.Context) ([]*TablePublic, error) {
	if m.shouldFail {
		return nil, ErrDatabase
	}
	return []*TablePublic{
		{
			ID:   1,
			Name: "Table 1",
		},
	}, nil
}

func (m *mockTableService) ActivateTable(ctx context.Context, id int) error {
	if m.shouldFail {
		return ErrTableNotFound
	}
	return nil
}

func (m *mockTableService) DeactivateTable(ctx context.Context, id int) error {
	if m.shouldFail {
		return ErrTableNotFound
	}
	return nil
}

func TestCreateTableHandler_MethodNotAllowed(t *testing.T) {
	handler := &Handler{Service: &mockTableService{}}

	req := httptest.NewRequest(http.MethodGet, "/create-table", nil)
	rec := httptest.NewRecorder()

	handler.CreateTableHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", rec.Code)
	}
}

func TestCreateTableHandler_Success(t *testing.T) {
	handler := &Handler{Service: &mockTableService{}}

	body := `{"name":"Table 1"}`
	req := httptest.NewRequest(http.MethodPost, "/create-table", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.CreateTableHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestUpdateTableHandler_Success(t *testing.T) {
	handler := &Handler{Service: &mockTableService{}}

	body := `{"id":1,"name":"Updated Table"}`
	req := httptest.NewRequest(http.MethodPost, "/update-table", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.UpdateTableHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestUpdateTableHandler_NotFound(t *testing.T) {
	handler := &Handler{Service: &mockTableService{shouldFail: true}}

	body := `{"id":999,"name":"Updated Table"}`
	req := httptest.NewRequest(http.MethodPost, "/update-table", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.UpdateTableHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestGetTableHandler_Success(t *testing.T) {
	handler := &Handler{Service: &mockTableService{}}

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
	handler := &Handler{Service: &mockTableService{shouldFail: true}}

	body := `{"id":999}`
	req := httptest.NewRequest(http.MethodPost, "/get-table", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.GetTableHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestGetAllTablesHandler_Success(t *testing.T) {
	handler := &Handler{Service: &mockTableService{}}

	req := httptest.NewRequest(http.MethodPost, "/get-all-tables", nil)
	rec := httptest.NewRecorder()

	handler.GetAllTablesHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestGetActiveTablesHandler_Success(t *testing.T) {
	handler := &Handler{Service: &mockTableService{}}

	req := httptest.NewRequest(http.MethodPost, "/get-active-tables", nil)
	rec := httptest.NewRecorder()

	handler.GetActiveTablesHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestActivateTableHandler_Success(t *testing.T) {
	handler := &Handler{Service: &mockTableService{}}

	body := `{"id":1}`
	req := httptest.NewRequest(http.MethodPost, "/activate-table", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ActivateTableHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", rec.Code)
	}
}

func TestActivateTableHandler_NotFound(t *testing.T) {
	handler := &Handler{Service: &mockTableService{shouldFail: true}}

	body := `{"id":999}`
	req := httptest.NewRequest(http.MethodPost, "/activate-table", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ActivateTableHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestDeactivateTableHandler_Success(t *testing.T) {
	handler := &Handler{Service: &mockTableService{}}

	body := `{"id":1}`
	req := httptest.NewRequest(http.MethodPost, "/deactivate-table", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.DeactivateTableHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", rec.Code)
	}
}

func TestDeactivateTableHandler_NotFound(t *testing.T) {
	handler := &Handler{Service: &mockTableService{shouldFail: true}}

	body := `{"id":999}`
	req := httptest.NewRequest(http.MethodPost, "/deactivate-table", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.DeactivateTableHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}
