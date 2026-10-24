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

// mockCommand records the guarded calls; the nil embedded interface panics on any
// other, so a rejected body that still reaches the command fails the test.
type mockCommand struct {
	command
	deletedID     int
	deactivatedID int
	updatedID     int
}

func (m *mockCommand) UpdateUser(_ context.Context, id int, _, _ string, _ user.Role) error {
	m.updatedID = id
	return nil
}

func (m *mockCommand) DeactivateUser(_ context.Context, id int) error {
	m.deactivatedID = id
	return nil
}

func (m *mockCommand) DeleteUser(_ context.Context, id int) error {
	m.deletedID = id
	return nil
}

func TestDeleteUserHandler_CannotDeleteSelf(t *testing.T) {
	mock := &mockCommand{}
	handler := &CommandHandler{Command: mock}

	body := `{"id":42}`
	req := httptest.NewRequest(http.MethodPost, "/admin/delete-user", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

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

func TestCreateUserHandler_InvalidInput(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{}}

	body := `{"name":"ab","username":"INVALID","role":"admin"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/create-user", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.CreateUserHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestUpdateUserHandler_InvalidInput(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{}}

	body := `{"id":0,"name":"ab","username":"INVALID","role":"admin"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/update-user", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.UpdateUserHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

// Der letzte Admin darf sich weder deaktivieren noch herabstufen: Beides sperrt
// die Instanz ohne Datenbankzugriff dauerhaft aus.
func TestDeactivateUserHandler_CannotDeactivateSelf(t *testing.T) {
	mock := &mockCommand{}
	handler := &CommandHandler{Command: mock}

	req := httptest.NewRequest(http.MethodPost, "/admin/deactivate-user", strings.NewReader(`{"id":42}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, 42))

	rec := httptest.NewRecorder()
	handler.DeactivateUserHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "cannot_deactivate_self") {
		t.Errorf("expected error code 'cannot_deactivate_self', got %s", rec.Body.String())
	}
	if mock.deactivatedID != 0 {
		t.Errorf("expected DeactivateUser not to be called, but was called with id %d", mock.deactivatedID)
	}
}

func TestDeactivateUserHandler_DeactivateOtherUser(t *testing.T) {
	mock := &mockCommand{}
	handler := &CommandHandler{Command: mock}

	req := httptest.NewRequest(http.MethodPost, "/admin/deactivate-user", strings.NewReader(`{"id":42}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, 99))

	rec := httptest.NewRecorder()
	handler.DeactivateUserHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	if mock.deactivatedID != 42 {
		t.Errorf("expected DeactivateUser to be called with id 42, got %d", mock.deactivatedID)
	}
}

func TestUpdateUserHandler_CannotDemoteSelf(t *testing.T) {
	mock := &mockCommand{}
	handler := &CommandHandler{Command: mock}

	body := `{"id":42,"name":"Anna Admin","username":"anna","role":"service"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/update-user", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, 42))

	rec := httptest.NewRecorder()
	handler.UpdateUserHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "cannot_demote_self") {
		t.Errorf("expected error code 'cannot_demote_self', got %s", rec.Body.String())
	}
	if mock.updatedID != 0 {
		t.Errorf("expected UpdateUser not to be called, but was called with id %d", mock.updatedID)
	}
}

// Name und Benutzername des eigenen Kontos bleiben änderbar, solange die Rolle
// admin bleibt.
func TestUpdateUserHandler_SelfKeepsAdminRole(t *testing.T) {
	mock := &mockCommand{}
	handler := &CommandHandler{Command: mock}

	body := `{"id":42,"name":"Anna Admin","username":"anna","role":"admin"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/update-user", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, 42))

	rec := httptest.NewRecorder()
	handler.UpdateUserHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	if mock.updatedID != 42 {
		t.Errorf("expected UpdateUser to be called with id 42, got %d", mock.updatedID)
	}
}
