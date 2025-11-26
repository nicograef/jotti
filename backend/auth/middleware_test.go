//go:build unit

package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	usr "github.com/nicograef/jotti/backend/user"
)

func TestAdminMiddleware_ValidToken(t *testing.T) {
	secret := "test-secret"
	user := usr.User{
		ID:       1,
		Username: "admin",
		Role:     usr.AdminRole,
		Status:   usr.ActiveStatus,
	}

	token, err := generateJWTTokenForUser(user, secret)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := NewAdminMiddleware(secret)(handler)
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestAdminMiddleware_NoToken(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := NewAdminMiddleware("test-secret")(handler)
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

func TestAdminMiddleware_ServiceRole(t *testing.T) {
	secret := "test-secret"
	user := usr.User{
		ID:       2,
		Username: "service",
		Role:     usr.ServiceRole,
		Status:   usr.ActiveStatus,
	}

	token, err := generateJWTTokenForUser(user, secret)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := NewAdminMiddleware(secret)(handler)
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rec.Code)
	}
}

func TestServiceMiddleware_ValidToken(t *testing.T) {
	secret := "test-secret"
	user := usr.User{
		ID:       2,
		Username: "service",
		Role:     usr.ServiceRole,
		Status:   usr.ActiveStatus,
	}

	token, err := generateJWTTokenForUser(user, secret)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := NewServiceMiddleware(secret)(handler)
	req := httptest.NewRequest(http.MethodGet, "/service", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}
