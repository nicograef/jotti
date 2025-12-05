//go:build unit

package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	usr "github.com/nicograef/jotti/backend/user"
)

type mockUserService struct {
	user *usr.User
	err  error
}

func (m *mockUserService) VerifyPasswordAndGetUser(ctx context.Context, username, password string) (*usr.User, error) {
	return m.user, m.err
}

func (m *mockUserService) SetNewPassword(ctx context.Context, username, password, onetimePassword string) (*usr.User, error) {
	return m.user, m.err
}

func TestLoginHandler_Success(t *testing.T) {
	mockService := &mockUserService{
		user: &usr.User{
			ID:       1,
			Username: "testuser",
			Status:   usr.ActiveStatus,
			Role:     usr.AdminRole,
		},
		err: nil,
	}

	handler := Handler{
		UserService: mockService,
		JWTSecret:   "test-secret",
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
	mockService := &mockUserService{
		user: nil,
		err:  usr.ErrInvalidPassword,
	}

	handler := Handler{
		UserService: mockService,
		JWTSecret:   "test-secret",
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
	mockService := &mockUserService{
		user: &usr.User{
			ID:       1,
			Username: "testuser",
			Status:   usr.InactiveStatus,
			Role:     usr.ServiceRole,
		},
		err: nil,
	}

	handler := Handler{
		UserService: mockService,
		JWTSecret:   "test-secret",
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
