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

func (m *mockCommand) DirektverkaufTaetigen(context.Context, int, string, string, []enrichment.PositionInput, string) error {
	return m.err
}

func (m *mockCommand) DirektverkaufStornieren(context.Context, int, string, string, []kasse.PositionRef, string) error {
	return m.err
}

func requestWithUser(body string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, 1)
	ctx = context.WithValue(ctx, middleware.UserNameKey, "Test User")
	return req.WithContext(ctx)
}

const (
	validBody      = `{"verkaufId":"6f9619ff-8b86-d011-b42d-00cf4fc964ff","positionen":[{"produktId":1,"varianteId":1,"menge":2}],"kommentar":""}`
	validStornoBod = `{"verkaufId":"6f9619ff-8b86-d011-b42d-00cf4fc964ff","positionen":[{"positionId":"6f9619ff-8b86-d011-b42d-00cf4fc964ff","menge":1}],"kommentar":"Rueckgabe"}`
)

// Each row is a request the handler must refuse: a body the schema rejects, or a
// domain error and the status and wire code it maps to.
func TestCommandHandler_Abweisungen(t *testing.T) {
	taetigen := (*CommandHandler).DirektverkaufTaetigenHandler
	stornieren := (*CommandHandler).DirektverkaufStornierenHandler
	tests := []struct {
		name       string
		handler    func(*CommandHandler) http.HandlerFunc
		err        error
		body       string
		wantStatus int
		wantCode   string
	}{
		{"tätigen: Kasse nicht geöffnet", taetigen, application.ErrKasseNichtGeoeffnet, validBody, http.StatusConflict, "kasse_nicht_geoeffnet"},
		{"tätigen: Konflikt", taetigen, application.ErrConflict, validBody, http.StatusConflict, "conflict"},
		{"tätigen: Produkt unbekannt", taetigen, enrichment.ErrProduktNotFound, validBody, http.StatusBadRequest, "produkt_not_found"},
		{"tätigen: ohne Verkaufs-ID und Positionen", taetigen, nil, `{"positionen":[],"kommentar":""}`, http.StatusBadRequest, "validation_error"},
		{"tätigen: Verkaufs-ID keine UUID", taetigen, nil, `{"verkaufId":"nicht-eine-uuid","positionen":[{"produktId":1,"varianteId":1,"menge":1}],"kommentar":""}`, http.StatusBadRequest, "validation_error"},
		{"tätigen: Menge über 999", taetigen, nil, `{"verkaufId":"6f9619ff-8b86-d011-b42d-00cf4fc964ff","positionen":[{"produktId":1,"varianteId":1,"menge":1000}],"kommentar":""}`, http.StatusBadRequest, "validation_error"},
		{"stornieren: Konflikt", stornieren, application.ErrConflict, validStornoBod, http.StatusConflict, "conflict"},
		{"stornieren: Verkauf unbekannt", stornieren, application.ErrVerkaufNichtGefunden, validStornoBod, http.StatusBadRequest, "verkauf_not_found"},
		{"stornieren: Kommentar zu kurz", stornieren, nil, `{"verkaufId":"6f9619ff-8b86-d011-b42d-00cf4fc964ff","positionen":[{"positionId":"6f9619ff-8b86-d011-b42d-00cf4fc964ff","menge":1}],"kommentar":"ab"}`, http.StatusBadRequest, "validation_error"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			tt.handler(&CommandHandler{Command: &mockCommand{err: tt.err}}).ServeHTTP(rec, requestWithUser(tt.body))

			if rec.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d: %s", tt.wantStatus, rec.Code, rec.Body.String())
			}
			if !strings.Contains(rec.Body.String(), `"code":"`+tt.wantCode+`"`) {
				t.Errorf("expected code %s, got %s", tt.wantCode, rec.Body.String())
			}
		})
	}
}
