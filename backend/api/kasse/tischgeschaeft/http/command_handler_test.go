//go:build unit

package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nicograef/jotti/backend/api/kasse/tischgeschaeft/application"
	"github.com/nicograef/jotti/backend/api/middleware"
	"github.com/nicograef/jotti/backend/domain/kasse"
)

type mockCommand struct {
	err error
}

func (m *mockCommand) BestellungAufnehmen(ctx context.Context, userID int, userName string, bestellungID string, tischID int, positionen []application.BestellPositionInput, kommentar string) error {
	return m.err
}

func (m *mockCommand) BestellungUmbuchen(ctx context.Context, userID int, userName string, quellTischID int, zielTischID int, positionen []kasse.PositionRef, benutzerKommentar string) error {
	return m.err
}

func (m *mockCommand) ZahlungKassieren(ctx context.Context, userID int, userName string, tischID int, positionen []kasse.PositionRef, kommentar string) error {
	return m.err
}

func (m *mockCommand) StornierungErteilen(ctx context.Context, userID int, userName string, tischID int, positionen []kasse.PositionRef, kommentar string) error {
	return m.err
}

func TestBestellungAufnehmenHandler_Conflict(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{err: application.ErrConflict}}

	body := `{"bestellungId":"6f9619ff-8b86-d011-b42d-00cf4fc964ff","tischId":1,"positionen":[{"produktId":1,"varianteId":1,"menge":2}],"kommentar":""}`
	req := httptest.NewRequest(http.MethodPost, "/bestellung-aufnehmen", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, 1)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.BestellungAufnehmenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("expected status 409, got %d", rec.Code)
	}
}

func TestBestellungAufnehmenHandler_ProduktNotFound(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{err: application.ErrProduktNotFound}}

	body := `{"bestellungId":"6f9619ff-8b86-d011-b42d-00cf4fc964ff","tischId":1,"positionen":[{"produktId":1,"varianteId":1,"menge":2}],"kommentar":""}`
	req := httptest.NewRequest(http.MethodPost, "/bestellung-aufnehmen", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, 1)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.BestellungAufnehmenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestZahlungKassierenHandler_Conflict(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{err: application.ErrConflict}}

	body := `{"tischId":1,"positionen":[{"positionId":"a87f1b2c-3d4e-5f6a-7b8c-9d0e1f2a3b4c","menge":1}],"kommentar":""}`
	req := httptest.NewRequest(http.MethodPost, "/zahlung-kassieren", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, 1)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.ZahlungKassierenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("expected status 409, got %d", rec.Code)
	}
}

func TestStornierungErteilenHandler_Conflict(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{err: application.ErrConflict}}

	body := `{"tischId":1,"positionen":[{"positionId":"a87f1b2c-3d4e-5f6a-7b8c-9d0e1f2a3b4c","menge":1}],"kommentar":"Reklamation"}`
	req := httptest.NewRequest(http.MethodPost, "/stornierung-erteilen", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, 1)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.StornierungErteilenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("expected status 409, got %d", rec.Code)
	}
}

func TestBestellungUmbuchenHandler_Conflict(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{err: application.ErrConflict}}

	body := `{"quellTischId":1,"zielTischId":2,"positionen":[{"positionId":"a87f1b2c-3d4e-5f6a-7b8c-9d0e1f2a3b4c","menge":1}]}`
	req := httptest.NewRequest(http.MethodPost, "/bestellung-umbuchen", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, 1)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.BestellungUmbuchenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("expected status 409, got %d", rec.Code)
	}
}

func TestBestellungUmbuchenHandler_UmbuchungGleicherTisch(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{err: application.ErrUmbuchungGleicherTisch}}

	body := `{"quellTischId":1,"zielTischId":1,"positionen":[{"positionId":"a87f1b2c-3d4e-5f6a-7b8c-9d0e1f2a3b4c","menge":1}]}`
	req := httptest.NewRequest(http.MethodPost, "/bestellung-umbuchen", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, 1)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.BestellungUmbuchenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestBestellungAufnehmenHandler_UngueltigeBestellungId_ValidationError(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{}}

	body := `{"bestellungId":"nicht-eine-uuid","tischId":1,"positionen":[{"produktId":1,"varianteId":1,"menge":1}],"kommentar":""}`
	req := httptest.NewRequest(http.MethodPost, "/bestellung-aufnehmen", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, 1)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.BestellungAufnehmenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for non-UUID bestellungId, got %d", rec.Code)
	}
}
