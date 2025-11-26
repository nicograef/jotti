//go:build unit

package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
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
	for i := 0; i < 10; i++ {
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
