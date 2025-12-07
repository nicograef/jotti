//go:build unit

package user

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type mockAuthCommand struct {
	user *User
	err  error
}

func (m *mockAuthCommand) VerifyPasswordAndGetUser(ctx context.Context, username, password string) (*User, error) {
	return m.user, m.err
}

func (m *mockAuthCommand) SetNewPassword(ctx context.Context, username, password, onetimePassword string) (*User, error) {
	return m.user, m.err
}

func TestLoginHandler_Success(t *testing.T) {
	command := &mockAuthCommand{
		user: &User{
			ID:       1,
			Username: "testuser",
			Status:   ActiveStatus,
			Role:     AdminRole,
		},
		err: nil,
	}

	handler := AuthHandler{
		Command:   command,
		JWTSecret: "test-secret",
	}

	body := `{"username":"testuser","password":"Test123!"}`
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.LoginHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestLoginHandler_InvalidCredentials(t *testing.T) {
	command := &mockAuthCommand{
		user: nil,
		err:  ErrInvalidPassword,
	}

	handler := AuthHandler{
		Command:   command,
		JWTSecret: "test-secret",
	}

	body := `{"username":"testuser","password":"wrongpassword"}`
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.LoginHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestLoginHandler_InactiveUser(t *testing.T) {
	command := &mockAuthCommand{
		user: &User{
			ID:       1,
			Username: "testuser",
			Status:   InactiveStatus,
			Role:     ServiceRole,
		},
		err: nil,
	}

	handler := AuthHandler{
		Command:   command,
		JWTSecret: "test-secret",
	}

	body := `{"username":"testuser","password":"Test123!"}`
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.LoginHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}
