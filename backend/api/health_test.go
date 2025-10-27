package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewHealthHandler(t *testing.T) {
	handler := NewHealthHandler()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if w.Body.String() != "OK" {
		t.Errorf("expected body %q, got %q", "OK", w.Body.String())
	}
}

func TestNewHealthHandler_MethodNotAllowed(t *testing.T) {
	handler := NewHealthHandler()

	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}
