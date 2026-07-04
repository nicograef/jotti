//go:build unit

package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nicograef/jotti/backend/api/middleware"
	"github.com/nicograef/jotti/backend/api/stammdaten/tisch/application"
)

type mockCommand struct {
	err error
}

func (m *mockCommand) TischErstellen(ctx context.Context, name string) (int, error) {
	return 1, m.err
}

func (m *mockCommand) TischAktualisieren(ctx context.Context, id int, name string) error {
	return m.err
}

func (m *mockCommand) TischAktivieren(ctx context.Context, id int) error {
	return m.err
}

func (m *mockCommand) TischDeaktivieren(ctx context.Context, id int) error {
	return m.err
}

func (m *mockCommand) TischLoeschen(ctx context.Context, id int) error {
	return m.err
}

func (m *mockCommand) FavoritHinzufuegen(_ context.Context, _, _ int) error {
	return m.err
}

func (m *mockCommand) FavoritEntfernen(_ context.Context, _, _ int) error {
	return m.err
}

func TestTischErstellenHandler_Success(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{}}

	body := `{"name":"Tisch 1"}`
	req := httptest.NewRequest(http.MethodPost, "/create-tisch", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.TischErstellenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestTischErstellenHandler_Failure(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{err: application.ErrDatabase}}

	body := `{"name":"Tisch 1"}`
	req := httptest.NewRequest(http.MethodPost, "/create-tisch", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.TischErstellenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}
}

func TestTischErstellenHandler_InvalidInput(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{}}

	body := `{"name":"ab"}`
	req := httptest.NewRequest(http.MethodPost, "/create-tisch", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.TischErstellenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestTischAktualisierenHandler_Success(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{}}

	body := `{"id":1,"name":"Aktualisierter Tisch"}`
	req := httptest.NewRequest(http.MethodPost, "/update-tisch", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.TischAktualisierenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestTischAktualisierenHandler_NotFound(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{err: application.ErrTischNotFound}}

	body := `{"id":999,"name":"Aktualisierter Tisch"}`
	req := httptest.NewRequest(http.MethodPost, "/update-tisch", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.TischAktualisierenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestTischAktualisierenHandler_InvalidInput(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{}}

	body := `{"id":0,"name":"ab"}`
	req := httptest.NewRequest(http.MethodPost, "/update-tisch", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.TischAktualisierenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestTischAktivierenHandler_Success(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{}}

	body := `{"id":1}`
	req := httptest.NewRequest(http.MethodPost, "/activate-tisch", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.TischAktivierenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestTischAktivierenHandler_NotFound(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{err: application.ErrTischNotFound}}

	body := `{"id":999}`
	req := httptest.NewRequest(http.MethodPost, "/activate-tisch", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.TischAktivierenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestTischDeaktivierenHandler_Success(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{}}

	body := `{"id":1}`
	req := httptest.NewRequest(http.MethodPost, "/deactivate-tisch", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.TischDeaktivierenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestTischDeaktivierenHandler_NotFound(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{err: application.ErrTischNotFound}}

	body := `{"id":999}`
	req := httptest.NewRequest(http.MethodPost, "/deactivate-tisch", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.TischDeaktivierenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestFavoritHinzufuegenHandler_InvalidInput(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{}}

	body := `{"tischId":0}`
	req := httptest.NewRequest(http.MethodPost, "/favorit-hinzufuegen", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, 1)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.FavoritHinzufuegenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}
