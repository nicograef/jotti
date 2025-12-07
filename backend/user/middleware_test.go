//go:build unit

package user

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAdminMiddleware_ValidToken(t *testing.T) {
	secret := "test-secret"
	user := User{
		ID:       1,
		Username: "admin",
		Role:     AdminRole,
		Status:   ActiveStatus,
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

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestAdminMiddleware_ServiceRole(t *testing.T) {
	secret := "test-secret"
	user := User{
		ID:       2,
		Username: "service",
		Role:     ServiceRole,
		Status:   ActiveStatus,
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

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestServiceMiddleware_ValidToken(t *testing.T) {
	secret := "test-secret"
	user := User{
		ID:       2,
		Username: "service",
		Role:     ServiceRole,
		Status:   ActiveStatus,
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
