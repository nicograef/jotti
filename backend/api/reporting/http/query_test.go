//go:build unit

package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetTagesabrechnungHandler_InvalidJSON(t *testing.T) {
	handler := QueryHandler{}

	req := httptest.NewRequest(http.MethodPost, "/get-tagesabrechnung", strings.NewReader("{invalid"))
	rec := httptest.NewRecorder()

	handler.GetTagesabrechnungHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestGetTagesabrechnungHandler_MissingVon(t *testing.T) {
	handler := QueryHandler{}

	body := `{"bis": "2026-03-14T03:00:00Z"}`
	req := httptest.NewRequest(http.MethodPost, "/get-tagesabrechnung", strings.NewReader(body))
	rec := httptest.NewRecorder()

	handler.GetTagesabrechnungHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestGetTagesabrechnungHandler_MissingBis(t *testing.T) {
	handler := QueryHandler{}

	body := `{"von": "2026-03-13T17:00:00Z"}`
	req := httptest.NewRequest(http.MethodPost, "/get-tagesabrechnung", strings.NewReader(body))
	rec := httptest.NewRecorder()

	handler.GetTagesabrechnungHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestGetTagesabrechnungHandler_VonAfterBis(t *testing.T) {
	handler := QueryHandler{}

	body := `{"von": "2026-03-14T03:00:00Z", "bis": "2026-03-13T17:00:00Z"}`
	req := httptest.NewRequest(http.MethodPost, "/get-tagesabrechnung", strings.NewReader(body))
	rec := httptest.NewRecorder()

	handler.GetTagesabrechnungHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestGetTagesabrechnungHandler_VonEqualsBis(t *testing.T) {
	handler := QueryHandler{}

	body := `{"von": "2026-03-13T17:00:00Z", "bis": "2026-03-13T17:00:00Z"}`
	req := httptest.NewRequest(http.MethodPost, "/get-tagesabrechnung", strings.NewReader(body))
	rec := httptest.NewRecorder()

	handler.GetTagesabrechnungHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}
