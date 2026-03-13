//go:build unit

package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nicograef/jotti/backend/api/middleware"
	"github.com/nicograef/jotti/backend/domain/user"
)

type mockCommand struct {
	err          error
	deletedID    int
	passwordHash string
}

func (m *mockCommand) CreateUser(ctx context.Context, name, username string, role user.Role) (int, string, error) {
	return 1, m.passwordHash, m.err
}

func (m *mockCommand) UpdateUser(ctx context.Context, id int, name, username string, role user.Role) error {
	return m.err
}

func (m *mockCommand) ActivateUser(ctx context.Context, id int) error {
	return m.err
}

func (m *mockCommand) DeactivateUser(ctx context.Context, id int) error {
	return m.err
}

func (m *mockCommand) DeleteUser(ctx context.Context, id int) error {
	m.deletedID = id
	return m.err
}

func (m *mockCommand) ResetPassword(ctx context.Context, userID int) (string, error) {
	return m.passwordHash, m.err
}

func TestDeleteUserHandler_CannotDeleteSelf(t *testing.T) {
	mock := &mockCommand{}
	handler := &CommandHandler{Command: mock}

	body := `{"id":42}`
	req := httptest.NewRequest(http.MethodPost, "/admin/delete-user", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Set the current user ID in context to match the delete target
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, 42)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.DeleteUserHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "cannot_delete_self") {
		t.Errorf("expected error code 'cannot_delete_self', got %s", rec.Body.String())
	}
	if mock.deletedID != 0 {
		t.Errorf("expected DeleteUser not to be called, but was called with id %d", mock.deletedID)
	}
}

func TestDeleteUserHandler_DeleteOtherUser(t *testing.T) {
	mock := &mockCommand{}
	handler := &CommandHandler{Command: mock}

	body := `{"id":42}`
	req := httptest.NewRequest(http.MethodPost, "/admin/delete-user", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Set the current user ID in context to a different user
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, 99)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.DeleteUserHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	if mock.deletedID != 42 {
		t.Errorf("expected DeleteUser to be called with id 42, got %d", mock.deletedID)
	}
}
