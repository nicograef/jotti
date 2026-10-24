package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nicograef/jotti/backend/api/kasse/enrichment"
	"github.com/nicograef/jotti/backend/api/kasse/tischgeschaeft/application"
	"github.com/nicograef/jotti/backend/api/middleware"
	"github.com/nicograef/jotti/backend/domain/kasse"
)

type mockCommand struct {
	err error
}

func (m *mockCommand) BestellungAufnehmen(context.Context, int, string, string, int, []enrichment.PositionInput, string) error {
	return m.err
}

func (m *mockCommand) BestellungUmbuchen(context.Context, int, string, int, int, []kasse.PositionRef, string) error {
	return m.err
}

func (m *mockCommand) ZahlungKassieren(context.Context, int, string, int, []kasse.PositionRef, string) error {
	return m.err
}

func (m *mockCommand) StornierungErteilen(context.Context, int, string, int, []kasse.PositionRef, string) error {
	return m.err
}

const (
	bestellungBody = `{"bestellungId":"6f9619ff-8b86-d011-b42d-00cf4fc964ff","tischId":1,"positionen":[{"produktId":1,"varianteId":1,"menge":2}],"kommentar":""}`
	positionRef    = `"positionen":[{"positionId":"a87f1b2c-3d4e-5f6a-7b8c-9d0e1f2a3b4c","menge":1}]`
)

// Each row is a request the handler must refuse: a body the schema rejects, or a
// domain error and the status and wire code it maps to.
func TestCommandHandler_Abweisungen(t *testing.T) {
	aufnehmen := (*CommandHandler).BestellungAufnehmenHandler
	umbuchen := (*CommandHandler).BestellungUmbuchenHandler
	tests := []struct {
		name       string
		handler    func(*CommandHandler) http.HandlerFunc
		err        error
		body       string
		wantStatus int
		wantCode   string
	}{
		{"aufnehmen: Konflikt", aufnehmen, application.ErrConflict, bestellungBody, http.StatusConflict, "conflict"},
		{"aufnehmen: Produkt unbekannt", aufnehmen, enrichment.ErrProduktNotFound, bestellungBody, http.StatusBadRequest, "produkt_not_found"},
		{"aufnehmen: Bestell-ID keine UUID", aufnehmen, nil, `{"bestellungId":"nicht-eine-uuid","tischId":1,"positionen":[{"produktId":1,"varianteId":1,"menge":1}],"kommentar":""}`, http.StatusBadRequest, "validation_error"},
		{"aufnehmen: Menge über 999", aufnehmen, nil, `{"bestellungId":"6f9619ff-8b86-d011-b42d-00cf4fc964ff","tischId":1,"positionen":[{"produktId":1,"varianteId":1,"menge":1000}],"kommentar":""}`, http.StatusBadRequest, "validation_error"},
		{"kassieren: Konflikt", (*CommandHandler).ZahlungKassierenHandler, application.ErrConflict, `{"tischId":1,` + positionRef + `,"kommentar":""}`, http.StatusConflict, "conflict"},
		{"stornieren: Konflikt", (*CommandHandler).StornierungErteilenHandler, application.ErrConflict, `{"tischId":1,` + positionRef + `,"kommentar":"Reklamation"}`, http.StatusConflict, "conflict"},
		{"umbuchen: Konflikt", umbuchen, application.ErrConflict, `{"quellTischId":1,"zielTischId":2,` + positionRef + `}`, http.StatusConflict, "conflict"},
		{"umbuchen: gleicher Tisch", umbuchen, application.ErrUmbuchungGleicherTisch, `{"quellTischId":1,"zielTischId":1,` + positionRef + `}`, http.StatusBadRequest, "umbuchung_gleicher_tisch"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))
			req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, 1))
			rec := httptest.NewRecorder()

			tt.handler(&CommandHandler{Command: &mockCommand{err: tt.err}}).ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d: %s", tt.wantStatus, rec.Code, rec.Body.String())
			}
			if !strings.Contains(rec.Body.String(), `"code":"`+tt.wantCode+`"`) {
				t.Errorf("expected code %s, got %s", tt.wantCode, rec.Body.String())
			}
		})
	}
}
