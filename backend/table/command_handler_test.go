//go:build unit

package table

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type mockCommand struct {
	err error
}

func (m *mockCommand) CreateTable(ctx context.Context, name string) (int, error) {
	return 1, m.err
}

func (m *mockCommand) UpdateTable(ctx context.Context, id int, name string) error {
	return m.err
}

func (m *mockCommand) ActivateTable(ctx context.Context, id int) error {
	return m.err
}

func (m *mockCommand) DeactivateTable(ctx context.Context, id int) error {
	return m.err
}

func TestCreateTableHandler_Success(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{}}

	body := `{"name":"Table 1"}`
	req := httptest.NewRequest(http.MethodPost, "/create-table", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.CreateTableHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestCreateTabkeHandler_Failure(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{err: ErrDatabase}}

	body := `{"name":"Table 1"}`
	req := httptest.NewRequest(http.MethodPost, "/create-table", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.CreateTableHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}
}

func TestUpdateTableHandler_Success(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{}}

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
	handler := &CommandHandler{Command: &mockCommand{err: ErrTableNotFound}}

	body := `{"id":999,"name":"Updated Table"}`
	req := httptest.NewRequest(http.MethodPost, "/update-table", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.UpdateTableHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestActivateTableHandler_Success(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{}}

	body := `{"id":1}`
	req := httptest.NewRequest(http.MethodPost, "/activate-table", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ActivateTableHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestActivateTableHandler_NotFound(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{err: ErrTableNotFound}}

	body := `{"id":999}`
	req := httptest.NewRequest(http.MethodPost, "/activate-table", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ActivateTableHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestDeactivateTableHandler_Success(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{}}

	body := `{"id":1}`
	req := httptest.NewRequest(http.MethodPost, "/deactivate-table", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.DeactivateTableHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestDeactivateTableHandler_NotFound(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{err: ErrTableNotFound}}

	body := `{"id":999}`
	req := httptest.NewRequest(http.MethodPost, "/deactivate-table", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.DeactivateTableHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}
