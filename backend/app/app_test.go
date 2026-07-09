//go:build unit

package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/nicograef/jotti/backend/config"
)

func setRequiredConfigEnv(t *testing.T) {
	t.Helper()
	t.Setenv("JWT_SECRET", "test-secret-for-app-tests")
	t.Setenv("RELAY_AUTH_TOKEN", "test-relay-token-for-app-tests")
	t.Setenv("POSTGRES_PASSWORD", "test-postgres-password-1234")
}

func TestNewApp(t *testing.T) {
	setRequiredConfigEnv(t)
	cfg := config.Load()
	app, err := NewApp(cfg, &sql.DB{}, "dev")
	if err != nil {
		t.Fatalf("NewApp() failed: %v", err)
	}

	if app.Server == nil {
		t.Error("Server should not be nil")
	}

	if app.Config.Port != 3000 {
		t.Errorf("Expected port 3000, got %d", app.Config.Port)
	}
}

func TestSetupRoutes(t *testing.T) {
	setRequiredConfigEnv(t)
	cfg := config.Load()

	handler := SetupRoutes(cfg, &sql.DB{}, "dev")

	if handler == nil {
		t.Error("Handler should not be nil")
	}
}

func TestSetupRoutes_HealthAllowsGet(t *testing.T) {
	setRequiredConfigEnv(t)
	cfg := config.Load()
	db, err := sql.Open("pgx", "invalid-connection-string")
	if err != nil {
		t.Fatalf("failed to create test db handle: %v", err)
	}
	defer db.Close()

	handler := SetupRoutes(cfg, db, "v9.9.9")

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 200 or 503 for GET /health, got %d", w.Code)
	}

	var body map[string]any
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("expected JSON response for GET /health: %v", err)
	}

	if code, ok := body["code"].(string); ok && code == "method_not_allowed" {
		t.Fatalf("GET /health must not be blocked by method middleware")
	}

	if v, ok := body["version"].(string); !ok || v != "v9.9.9" {
		t.Fatalf("expected version %q in /health response, got %v", "v9.9.9", body["version"])
	}
}

func TestSetupRoutes_NonHealthRejectsGet(t *testing.T) {
	setRequiredConfigEnv(t)
	cfg := config.Load()
	db, err := sql.Open("pgx", "invalid-connection-string")
	if err != nil {
		t.Fatalf("failed to create test db handle: %v", err)
	}
	defer db.Close()

	handler := SetupRoutes(cfg, db, "dev")

	req := httptest.NewRequest(http.MethodGet, "/auth/login", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d for GET on non-health route, got %d", http.StatusBadRequest, w.Code)
	}

	type errorBody struct {
		Code string `json:"code"`
	}

	var body errorBody
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("expected JSON error response: %v", err)
	}

	if body.Code != "method_not_allowed" {
		t.Fatalf("expected code method_not_allowed, got %q", body.Code)
	}
}

// TestSetupRoutes_ResetSeedRouteGuardedByEnv stellt sicher, dass der
// Test-Reset-Endpoint nur bei JOTTI_ENABLE_TEST_API=1 registriert wird: ohne das
// Flag existiert die Route nicht (404), mit dem Flag ist sie erreichbar (kein
// 404). So bleibt der Endpunkt in Produktion unerreichbar.
func TestSetupRoutes_ResetSeedRouteGuardedByEnv(t *testing.T) {
	setRequiredConfigEnv(t)

	t.Run("ohne Flag keine Route (404)", func(t *testing.T) {
		t.Setenv("JOTTI_ENABLE_TEST_API", "0")
		cfg := config.Load()
		handler := SetupRoutes(cfg, &sql.DB{}, "dev")

		req := httptest.NewRequest(http.MethodPost, "/test/reset-and-seed", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404 for POST /test/reset-and-seed without JOTTI_ENABLE_TEST_API, got %d", w.Code)
		}
	})

	t.Run("mit Flag Route registriert (nicht 404)", func(t *testing.T) {
		t.Setenv("JOTTI_ENABLE_TEST_API", "1")
		cfg := config.Load()
		db, err := sql.Open("pgx", "invalid-connection-string")
		if err != nil {
			t.Fatalf("failed to create test db handle: %v", err)
		}
		defer db.Close()

		handler := SetupRoutes(cfg, db, "dev")

		req := httptest.NewRequest(http.MethodPost, "/test/reset-and-seed", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code == http.StatusNotFound {
			t.Fatalf("expected registered route (not 404) with JOTTI_ENABLE_TEST_API=1, got 404")
		}
	})
}

func TestShutdown(t *testing.T) {
	setRequiredConfigEnv(t)
	cfg := config.Load()
	app, err := NewApp(cfg, &sql.DB{}, "dev")
	if err != nil {
		t.Fatalf("NewApp() failed: %v", err)
	}

	err = app.Shutdown()
	if err != nil {
		t.Errorf("Shutdown() failed: %v", err)
	}

}

func TestRun_ContextCancellation(t *testing.T) {
	setRequiredConfigEnv(t)
	cfg := config.Load()
	app, err := NewApp(cfg, &sql.DB{}, "dev")
	if err != nil {
		t.Fatalf("NewApp() failed: %v", err)
	}

	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	// Run the app in a separate goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- app.Run(ctx)
	}()

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	// Cancel the context to trigger shutdown
	cancel()

	// Wait for Run to return
	err = <-errChan
	if err != nil {
		t.Errorf("Run() returned error: %v", err)
	}
}
