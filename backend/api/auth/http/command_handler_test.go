//go:build unit

package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nicograef/jotti/backend/api/auth/application"
)

type mockAuthCommand struct {
	token string
	err   error
}

func (m mockAuthCommand) GenerateJWTToken(ctx context.Context, username, password string) (string, error) {
	return m.token, m.err
}

func (m mockAuthCommand) SetNewPassword(ctx context.Context, username, password, onetimePassword string) error {
	return m.err
}

func TestLoginHandler_Success(t *testing.T) {
	command := mockAuthCommand{token: "test-token", err: nil}
	handler := CommandHandler{Command: command}

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
	command := mockAuthCommand{token: "", err: application.ErrInvalidPassword}
	handler := CommandHandler{Command: command}

	body := `{"username":"testuser","password":"wrongpassword"}`
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.LoginHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestLoginHandler_Throttled(t *testing.T) {
	command := mockAuthCommand{token: "", err: application.ErrLoginThrottled}
	handler := CommandHandler{Command: command}

	body := `{"username":"testuser","password":"Test123!"}`
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.LoginHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("expected status 429 for throttled login, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "login_throttled") {
		t.Errorf("expected body to carry the login_throttled code, got %s", rec.Body.String())
	}
}

func TestLoginHandler_InactiveUser(t *testing.T) {
	command := mockAuthCommand{token: "", err: application.ErrNotActive}
	handler := CommandHandler{Command: command}

	body := `{"username":"testuser","password":"Test123!"}`
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.LoginHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestLoginHandler_InvalidInput(t *testing.T) {
	command := mockAuthCommand{token: "test-token", err: nil}
	handler := CommandHandler{Command: command}

	body := `{"username":"","password":""}`
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.LoginHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestSetPasswordHandler_InvalidInput(t *testing.T) {
	command := mockAuthCommand{err: nil}
	handler := CommandHandler{Command: command}

	body := `{"username":"INVALID","password":"123","onetimePassword":""}`
	req := httptest.NewRequest(http.MethodPost, "/set-password", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.SetPasswordHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}
