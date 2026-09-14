//go:build unit

package health

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestHealthCheck_DegradedWhenPingFails(t *testing.T) {
	db, err := sql.Open("pgx", "invalid-connection-string")
	if err != nil {
		t.Fatalf("open test db handle: %v", err)
	}
	defer func() { _ = db.Close() }()

	hc := HealthCheck{DB: db}
	handler := hc.Handler()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status %d, got %d", http.StatusServiceUnavailable, w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}
}
