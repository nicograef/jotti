//go:build integration

package api

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/lib/pq"
)

func TestHealthCheck_WithDatabase(t *testing.T) {
	// This is a placeholder for integration tests
	// In a real scenario, you'd use testcontainers to spin up a real PostgreSQL instance
	t.Skip("Integration test requires database - run with testcontainers")
}

func TestHealthCheck_WithMockDB(t *testing.T) {
	// Create a mock database connection (will fail on ping)
	db, err := sql.Open("postgres", "invalid-connection-string")
	if err != nil {
		t.Fatalf("Failed to create mock DB: %v", err)
	}
	defer db.Close()

	hc := HealthCheck{DB: db}
	handler := hc.Handler()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Should return 503 because database ping will fail
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status %d, got %d", http.StatusServiceUnavailable, w.Code)
	}

	// Should return JSON
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}
}
