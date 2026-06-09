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
	"github.com/nicograef/jotti/backend/domain/kasse"
)

type mockCommand struct {
	err         error
	lastTischID int
	lastZahlung string
	lastVerkauf string
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

func (m *mockCommand) BestellungAufnehmen(ctx context.Context, userID int, userName string, tischID int, positionen []application.BestellPositionInput, kommentar string) error {
	return m.err
}

func (m *mockCommand) BestellungUmbuchen(ctx context.Context, userID int, userName string, quellTischID int, zielTischID int, positionen []kasse.PositionRef) error {
	return m.err
}

func (m *mockCommand) ZahlungKassieren(ctx context.Context, userID int, userName string, tischID int, positionen []kasse.PositionRef, kommentar string) error {
	return m.err
}

func (m *mockCommand) StornierungErteilen(ctx context.Context, userID int, userName string, tischID int, positionen []kasse.PositionRef, kommentar string) error {
	return m.err
}

func (m *mockCommand) AusgabeBestaetigen(ctx context.Context, userID int, userName string, tischID int, positionen []kasse.PositionRef, kommentar string) error {
	return m.err
}

func (m *mockCommand) AuszahlungLeisten(ctx context.Context, userID int, userName string, tischID int, betragCents int, kommentar string) error {
	return m.err
}

func (m *mockCommand) KassenbelegDrucken(_ context.Context, tischID int, zahlungID string, verkaufID string) error {
	m.lastTischID = tischID
	m.lastZahlung = zahlungID
	m.lastVerkauf = verkaufID
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

func TestBestellungAufnehmenHandler_Conflict(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{err: application.ErrConflict}}

	body := `{"tischId":1,"positionen":[{"produktId":1,"varianteId":1,"menge":2}],"kommentar":""}`
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

	body := `{"tischId":1,"positionen":[{"produktId":1,"varianteId":1,"menge":2}],"kommentar":""}`
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

func TestAusgabeBestaetigenHandler_Conflict(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{err: application.ErrConflict}}

	body := `{"tischId":1,"positionen":[{"positionId":"a87f1b2c-3d4e-5f6a-7b8c-9d0e1f2a3b4c","menge":1}],"kommentar":""}`
	req := httptest.NewRequest(http.MethodPost, "/ausgabe-bestaetigen", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, 1)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.AusgabeBestaetigenHandler().ServeHTTP(rec, req)

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

func TestKassenbelegDruckenHandler_Success(t *testing.T) {
	mock := &mockCommand{}
	handler := &CommandHandler{Command: mock}

	body := `{"tischId":1,"zahlungId":"11111111-1111-1111-1111-111111111111"}`
	req := httptest.NewRequest(http.MethodPost, "/beleg-drucken", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.KassenbelegDruckenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	if mock.lastTischID != 1 {
		t.Errorf("expected tischId 1, got %d", mock.lastTischID)
	}
	if mock.lastZahlung != "11111111-1111-1111-1111-111111111111" {
		t.Errorf("expected zahlungId to be forwarded")
	}
	if mock.lastVerkauf != "" {
		t.Errorf("expected verkaufId empty for zahlung path, got %q", mock.lastVerkauf)
	}
}

func TestKassenbelegDruckenHandler_DirektverkaufSuccess(t *testing.T) {
	mock := &mockCommand{}
	handler := &CommandHandler{Command: mock}

	body := `{"verkaufId":"11111111-1111-1111-1111-111111111111"}`
	req := httptest.NewRequest(http.MethodPost, "/beleg-drucken", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.KassenbelegDruckenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	if mock.lastVerkauf != "11111111-1111-1111-1111-111111111111" {
		t.Errorf("expected verkaufId to be forwarded")
	}
	if mock.lastTischID != 0 || mock.lastZahlung != "" {
		t.Errorf("expected zahlung reference to be empty for direktverkauf path")
	}
}

func TestKassenbelegDruckenHandler_DruckerNichtKonfiguriert(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{err: application.ErrKassenbelegDruckerNichtKonfiguriert}}

	body := `{"tischId":1,"zahlungId":"11111111-1111-1111-1111-111111111111"}`
	req := httptest.NewRequest(http.MethodPost, "/beleg-drucken", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.KassenbelegDruckenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestKassenbelegDruckenHandler_DirektverkaufNichtGefunden(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{err: application.ErrVerkaufNichtGefunden}}

	body := `{"verkaufId":"11111111-1111-1111-1111-111111111111"}`
	req := httptest.NewRequest(http.MethodPost, "/beleg-drucken", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.KassenbelegDruckenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestKassenbelegDruckenHandler_MixedReferenceValidationError(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{}}

	body := `{"tischId":1,"zahlungId":"11111111-1111-1111-1111-111111111111","verkaufId":"22222222-2222-2222-2222-222222222222"}`
	req := httptest.NewRequest(http.MethodPost, "/beleg-drucken", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.KassenbelegDruckenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}
