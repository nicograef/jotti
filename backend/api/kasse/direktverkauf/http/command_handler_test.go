//go:build unit

package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nicograef/jotti/backend/api/kasse/direktverkauf/application"
	"github.com/nicograef/jotti/backend/api/kasse/enrichment"
	"github.com/nicograef/jotti/backend/api/middleware"
	"github.com/nicograef/jotti/backend/domain/kasse"
)

type mockCommand struct {
	err error
}

func (m *mockCommand) DirektverkaufTaetigen(_ context.Context, _ int, _ string, _ string, _ []enrichment.PositionInput, _ string) error {
	return m.err
}

func (m *mockCommand) DirektverkaufStornieren(_ context.Context, _ int, _ string, _ string, _ string, _ []kasse.PositionRef, _ string) error {
	return m.err
}

func requestWithUser(body string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/direktverkauf-taetigen", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, 1)
	ctx = context.WithValue(ctx, middleware.UserNameKey, "Test User")
	return req.WithContext(ctx)
}

const validBody = `{"verkaufId":"6f9619ff-8b86-d011-b42d-00cf4fc964ff","positionen":[{"produktId":1,"varianteId":1,"menge":2}],"kommentar":""}`

func TestDirektverkaufTaetigenHandler_Success(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{}}

	rec := httptest.NewRecorder()
	handler.DirektverkaufTaetigenHandler().ServeHTTP(rec, requestWithUser(validBody))

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestDirektverkaufTaetigenHandler_KasseNichtGeoeffnet(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{err: application.ErrKasseNichtGeoeffnet}}

	rec := httptest.NewRecorder()
	handler.DirektverkaufTaetigenHandler().ServeHTTP(rec, requestWithUser(validBody))

	if rec.Code != http.StatusConflict {
		t.Errorf("expected status 409, got %d", rec.Code)
	}
}

func TestDirektverkaufTaetigenHandler_Conflict(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{err: application.ErrConflict}}

	rec := httptest.NewRecorder()
	handler.DirektverkaufTaetigenHandler().ServeHTTP(rec, requestWithUser(validBody))

	if rec.Code != http.StatusConflict {
		t.Errorf("expected status 409, got %d", rec.Code)
	}
}

func TestDirektverkaufTaetigenHandler_ProduktNotFound(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{err: enrichment.ErrProduktNotFound}}

	rec := httptest.NewRecorder()
	handler.DirektverkaufTaetigenHandler().ServeHTTP(rec, requestWithUser(validBody))

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestDirektverkaufTaetigenHandler_ValidationError(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{}}

	// empty positionen violates the Min(1) schema rule
	body := `{"positionen":[],"kommentar":""}`
	rec := httptest.NewRecorder()
	handler.DirektverkaufTaetigenHandler().ServeHTTP(rec, requestWithUser(body))

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func stornoRequestWithUser(body string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/direktverkauf-stornieren", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, 2)
	ctx = context.WithValue(ctx, middleware.UserNameKey, "Leitung")
	return req.WithContext(ctx)
}

const validStornoBody = `{"vorgangId":"a87f1b2c-3d4e-5f6a-7b8c-9d0e1f2a3b4c","verkaufId":"6f9619ff-8b86-d011-b42d-00cf4fc964ff","positionen":[{"positionId":"6f9619ff-8b86-d011-b42d-00cf4fc964ff","menge":1}],"kommentar":"Rueckgabe"}`

func TestDirektverkaufStornierenHandler_Success(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{}}

	rec := httptest.NewRecorder()
	handler.DirektverkaufStornierenHandler().ServeHTTP(rec, stornoRequestWithUser(validStornoBody))

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestDirektverkaufStornierenHandler_Conflict(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{err: application.ErrConflict}}

	rec := httptest.NewRecorder()
	handler.DirektverkaufStornierenHandler().ServeHTTP(rec, stornoRequestWithUser(validStornoBody))

	if rec.Code != http.StatusConflict {
		t.Errorf("expected status 409, got %d", rec.Code)
	}
}

func TestDirektverkaufStornierenHandler_VerkaufNotFound(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{err: application.ErrVerkaufNichtGefunden}}

	rec := httptest.NewRecorder()
	handler.DirektverkaufStornierenHandler().ServeHTTP(rec, stornoRequestWithUser(validStornoBody))

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestDirektverkaufStornierenHandler_ValidationError(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{}}

	// kommentar shorter than 3 characters violates the schema
	body := `{"vorgangId":"a87f1b2c-3d4e-5f6a-7b8c-9d0e1f2a3b4c","verkaufId":"6f9619ff-8b86-d011-b42d-00cf4fc964ff","positionen":[{"positionId":"6f9619ff-8b86-d011-b42d-00cf4fc964ff","menge":1}],"kommentar":"ab"}`
	rec := httptest.NewRecorder()
	handler.DirektverkaufStornierenHandler().ServeHTTP(rec, stornoRequestWithUser(body))

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

// Fehlendes oder ungültiges vorgangId wird über das zog-Schema mit
// validation_error (400) abgelehnt.
func TestDirektverkaufStornierenHandler_VorgangIdPflicht_ValidationError(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{}}

	bodies := map[string]string{
		"ohne vorgangId":       `{"verkaufId":"6f9619ff-8b86-d011-b42d-00cf4fc964ff","positionen":[{"positionId":"6f9619ff-8b86-d011-b42d-00cf4fc964ff","menge":1}],"kommentar":"Rueckgabe"}`,
		"ungültiges vorgangId": `{"vorgangId":"nicht-eine-uuid","verkaufId":"6f9619ff-8b86-d011-b42d-00cf4fc964ff","positionen":[{"positionId":"6f9619ff-8b86-d011-b42d-00cf4fc964ff","menge":1}],"kommentar":"Rueckgabe"}`,
	}

	for name, body := range bodies {
		t.Run(name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			handler.DirektverkaufStornierenHandler().ServeHTTP(rec, stornoRequestWithUser(body))

			if rec.Code != http.StatusBadRequest {
				t.Errorf("expected status 400, got %d", rec.Code)
			}
			if !strings.Contains(rec.Body.String(), "validation_error") {
				t.Errorf("expected validation_error in body, got %s", rec.Body.String())
			}
		})
	}
}

func TestDirektverkaufTaetigenHandler_UngueltigeVerkaufId_ValidationError(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{}}

	body := `{"verkaufId":"nicht-eine-uuid","positionen":[{"produktId":1,"varianteId":1,"menge":1}],"kommentar":""}`
	rec := httptest.NewRecorder()
	handler.DirektverkaufTaetigenHandler().ServeHTTP(rec, requestWithUser(body))

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for non-UUID verkaufId, got %d", rec.Code)
	}
}

// Eine bekannte verkaufId mit abweichenden Nutzdaten ist ein expliziter Konflikt:
// 409 mit dem Code vorgang_daten_abweichend — weder ein stiller Erfolg noch der
// generische 400 der MapError-Abbildung.
func TestDirektverkaufTaetigenHandler_VorgangDatenAbweichend_Conflict(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{err: application.ErrVorgangDatenAbweichend}}

	body := `{"verkaufId":"6f9619ff-8b86-d011-b42d-00cf4fc964ff","positionen":[{"produktId":1,"varianteId":1,"menge":1}],"kommentar":""}`
	rec := httptest.NewRecorder()
	handler.DirektverkaufTaetigenHandler().ServeHTTP(rec, requestWithUser(body))

	if rec.Code != http.StatusConflict {
		t.Errorf("expected status 409, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "vorgang_daten_abweichend") {
		t.Errorf("expected vorgang_daten_abweichend in body, got %s", rec.Body.String())
	}
}

// Ein bekannter vorgangId mit abweichenden Nutzdaten ist ein expliziter Konflikt:
// 409 mit dem Code vorgang_daten_abweichend — weder ein stiller Erfolg noch der
// generische 400 der MapError-Abbildung.
func TestDirektverkaufStornierenHandler_VorgangDatenAbweichend_Conflict(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{err: application.ErrVorgangDatenAbweichend}}

	rec := httptest.NewRecorder()
	handler.DirektverkaufStornierenHandler().ServeHTTP(rec, stornoRequestWithUser(validStornoBody))

	if rec.Code != http.StatusConflict {
		t.Errorf("expected status 409, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "vorgang_daten_abweichend") {
		t.Errorf("expected vorgang_daten_abweichend in body, got %s", rec.Body.String())
	}
}
