package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// mockCommand panics through its nil embedded interface on any call, so a rejected
// body that still reaches the command fails the test.
type mockCommand struct {
	command
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
