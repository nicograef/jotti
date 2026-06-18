//go:build unit

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nicograef/jotti/backend/domain/jwt"
)

func TestCorrelationIDMiddleware_GeneratesID(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		correlationID, ok := r.Context().Value(CorrelationIDKey).(string)
		if !ok || correlationID == "" {
			t.Error("expected correlation ID in context")
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := CorrelationIDMiddleware(handler)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	correlationID := rec.Header().Get("X-Correlation-ID")
	if correlationID == "" {
		t.Error("expected X-Correlation-ID header to be set")
	}
}

func TestCorrelationIDMiddleware_UsesExisting(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := CorrelationIDMiddleware(handler)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Correlation-ID", "test-correlation-id")
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	correlationID := rec.Header().Get("X-Correlation-ID")
	if correlationID != "test-correlation-id" {
		t.Errorf("expected correlation ID 'test-correlation-id', got '%s'", correlationID)
	}
}

func TestRateLimitMiddleware_AllowsWithinLimit(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RateLimitMiddleware(10)(handler)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestRateLimitMiddleware_BlocksExceedingLimit(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RateLimitMiddleware(1)(handler)

	// Fill the limiter
	for range 10 {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		middleware.ServeHTTP(rec, req)
	}

	// This request should be rate limited
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	middleware.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("expected status 429, got %d", rec.Code)
	}
}

func TestJwtMiddleware_ValidToken(t *testing.T) {
	secret := "test-secret"
	token, err := jwt.GenerateJWTTokenForUser(1, "admin", "admin", secret)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := NewJwtMiddleware(secret, []string{"admin"})(handler)
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestJwtMiddleware_NoToken(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := NewJwtMiddleware("test-secret", []string{"admin"})(handler)
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestJwtMiddleware_InvalidBearerFormat(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	testCases := []struct {
		name  string
		value string
	}{
		{"short header", "Bear"},
		{"wrong prefix", "Basic xyz123"},
		{"just Bearer", "Bearer"},
		{"bearer lowercase", "bearer token123"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			middleware := NewJwtMiddleware("test-secret", []string{"admin"})(handler)
			req := httptest.NewRequest(http.MethodGet, "/admin", nil)
			req.Header.Set("Authorization", tc.value)
			rec := httptest.NewRecorder()

			middleware.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("expected status 400 for %q, got %d", tc.value, rec.Code)
			}
		})
	}
}

func TestJwtMiddleware_ServiceRole(t *testing.T) {
	secret := "test-secret"
	token, err := jwt.GenerateJWTTokenForUser(2, "service", "service", secret)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := NewJwtMiddleware(secret, []string{"admin"})(handler)
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
	token, err := jwt.GenerateJWTTokenForUser(2, "service", "service", secret)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := NewJwtMiddleware(secret, []string{"service"})(handler)
	req := httptest.NewRequest(http.MethodGet, "/service", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestServiceleitungRole_AllowedForServiceEndpoints(t *testing.T) {
	secret := "test-secret"
	token, err := jwt.GenerateJWTTokenForUser(3, "serviceleitung", "serviceleitung", secret)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := NewJwtMiddleware(secret, []string{"admin", "serviceleitung", "service"})(handler)
	req := httptest.NewRequest(http.MethodGet, "/service", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestServiceleitungRole_AllowedForCancelEndpoint(t *testing.T) {
	secret := "test-secret"
	token, err := jwt.GenerateJWTTokenForUser(3, "serviceleitung", "serviceleitung", secret)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := NewJwtMiddleware(secret, []string{"admin", "serviceleitung"})(handler)
	req := httptest.NewRequest(http.MethodGet, "/service/cancel", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestJwtMiddleware_SetsUserNameInContext(t *testing.T) {
	secret := "test-secret"
	token, err := jwt.GenerateJWTTokenForUser(1, "admin", "admin", secret)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, ok := r.Context().Value(UserNameKey).(string)
		if !ok || username != "admin" {
			t.Errorf("expected UserNameKey 'admin' in context, got '%s'", username)
		}
		userID, ok := r.Context().Value(UserIDKey).(int)
		if !ok || userID != 1 {
			t.Errorf("expected UserIDKey 1 in context, got %d", userID)
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := NewJwtMiddleware(secret, []string{"admin"})(handler)
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestServiceRole_DeniedForCancelEndpoint(t *testing.T) {
	secret := "test-secret"
	token, err := jwt.GenerateJWTTokenForUser(2, "service", "service", secret)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := NewJwtMiddleware(secret, []string{"admin", "serviceleitung"})(handler)
	req := httptest.NewRequest(http.MethodGet, "/service/cancel", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}
