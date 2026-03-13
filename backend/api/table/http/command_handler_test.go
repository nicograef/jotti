//go:build unit

package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nicograef/jotti/backend/api/middleware"
	"github.com/nicograef/jotti/backend/api/table/application"
	"github.com/nicograef/jotti/backend/domain/table"
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

func (m *mockCommand) BestellungAufgeben(ctx context.Context, userID int, userName string, tischID int, positionen []application.BestellPositionInput, kommentar string) error {
	return m.err
}

func (m *mockCommand) ZahlungRegistrieren(ctx context.Context, userID int, userName string, tischID int, positionen []table.PositionRef, kommentar string) error {
	return m.err
}

func (m *mockCommand) ProdukteStornieren(ctx context.Context, userID int, userName string, tischID int, positionen []table.PositionRef, kommentar string) error {
	return m.err
}

func (m *mockCommand) ProdukteLiefern(ctx context.Context, userID int, userName string, tischID int, positionen []table.PositionRef, kommentar string) error {
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

func TestBestellungAufgebenHandler_Conflict(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{err: application.ErrConflict}}

	body := `{"tischId":1,"positionen":[{"varianteId":1,"menge":2}],"kommentar":""}`
	req := httptest.NewRequest(http.MethodPost, "/bestellung-aufgeben", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, 1)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.BestellungAufgebenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("expected status 409, got %d", rec.Code)
	}
}

func TestBestellungAufgebenHandler_ProduktNotFound(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{err: application.ErrProduktNotFound}}

	body := `{"tischId":1,"positionen":[{"varianteId":1,"menge":2}],"kommentar":""}`
	req := httptest.NewRequest(http.MethodPost, "/bestellung-aufgeben", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, 1)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.BestellungAufgebenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestZahlungRegistrierenHandler_Conflict(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{err: application.ErrConflict}}

	body := `{"tischId":1,"positionen":[{"positionId":"abc","menge":1}],"kommentar":""}`
	req := httptest.NewRequest(http.MethodPost, "/zahlung-registrieren", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, 1)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.ZahlungRegistrierenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("expected status 409, got %d", rec.Code)
	}
}

func TestProdukteLiefernHandler_Conflict(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{err: application.ErrConflict}}

	body := `{"tischId":1,"positionen":[{"positionId":"abc","menge":1}],"kommentar":""}`
	req := httptest.NewRequest(http.MethodPost, "/produkte-liefern", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, 1)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.ProdukteLiefernHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("expected status 409, got %d", rec.Code)
	}
}

func TestProdukteStornierenHandler_Conflict(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{err: application.ErrConflict}}

	body := `{"tischId":1,"positionen":[{"positionId":"abc","menge":1}],"kommentar":""}`
	req := httptest.NewRequest(http.MethodPost, "/produkte-stornieren", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, 1)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.ProdukteStornierenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("expected status 409, got %d", rec.Code)
	}
}
