//go:build unit

package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nicograef/jotti/backend/api/direktverkauf/application"
	"github.com/nicograef/jotti/backend/api/middleware"
)

type mockCommand struct {
	err error
}

func (m *mockCommand) DirektverkaufTaetigen(_ context.Context, _ int, _ string, _ []application.VerkaufPositionInput, _ string) error {
	return m.err
}

func requestWithUser(body string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/direktverkauf-taetigen", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, 1)
	ctx = context.WithValue(ctx, middleware.UserNameKey, "Test User")
	return req.WithContext(ctx)
}

const validBody = `{"positionen":[{"produktId":1,"varianteId":1,"menge":2}],"kommentar":""}`

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
	handler := &CommandHandler{Command: &mockCommand{err: application.ErrProduktNotFound}}

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
