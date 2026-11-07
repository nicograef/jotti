package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/nicograef/jotti/backend/api/kasse/direktverkauf/application"
	"github.com/nicograef/jotti/backend/api/kasse/enrichment"
	"github.com/nicograef/jotti/backend/api/middleware"
	"github.com/nicograef/jotti/backend/domain/kasse"
)

type mockCommand struct {
	err        error
	positionen []enrichment.PositionInput
}

func (m *mockCommand) DirektverkaufTaetigen(_ context.Context, _ int, _ string, _ string, positionen []enrichment.PositionInput, _ string, _ *enrichment.PfandEingabe) error {
	m.positionen = positionen
	return m.err
}

func (m *mockCommand) DirektverkaufStornieren(context.Context, int, string, string, string, []kasse.PositionRef, string) error {
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
	validStornoBod = `{"stornierungId":"0b6a3f1e-2c4d-4e5f-8a9b-1c2d3e4f5a6b","verkaufId":"6f9619ff-8b86-d011-b42d-00cf4fc964ff","positionen":[{"positionId":"6f9619ff-8b86-d011-b42d-00cf4fc964ff","menge":1}],"kommentar":"Rueckgabe"}`
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
		{"tätigen: Pfandrückgabe abweichend", taetigen, application.ErrPfandrueckgabeAbweichend, validBody, http.StatusConflict, "pfandrueckgabe_abweichend"},
		{"tätigen: Produkt unbekannt", taetigen, enrichment.ErrProduktNotFound, validBody, http.StatusBadRequest, "produkt_not_found"},
		{"tätigen: ohne Verkaufs-ID und Positionen", taetigen, nil, `{"positionen":[],"kommentar":""}`, http.StatusBadRequest, "validation_error"},
		{"tätigen: Verkaufs-ID keine UUID", taetigen, nil, `{"verkaufId":"nicht-eine-uuid","positionen":[{"produktId":1,"varianteId":1,"menge":1}],"kommentar":""}`, http.StatusBadRequest, "validation_error"},
		{"tätigen: Menge über 999", taetigen, nil, `{"verkaufId":"6f9619ff-8b86-d011-b42d-00cf4fc964ff","positionen":[{"produktId":1,"varianteId":1,"menge":1000}],"kommentar":""}`, http.StatusBadRequest, "validation_error"},
		{"tätigen: Pfand ohne Rückgabe-ID", taetigen, nil, `{"verkaufId":"6f9619ff-8b86-d011-b42d-00cf4fc964ff","positionen":[{"produktId":1,"varianteId":1,"menge":2}],"kommentar":"","pfand":{"positionen":[{"produktId":2,"varianteId":2,"menge":1}]}}`, http.StatusBadRequest, "validation_error"},
		{"tätigen: Rückgabe-ID gleich Verkaufs-ID", taetigen, nil, `{"verkaufId":"6f9619ff-8b86-d011-b42d-00cf4fc964ff","positionen":[{"produktId":1,"varianteId":1,"menge":2}],"kommentar":"","pfand":{"rueckgabeId":"6f9619ff-8b86-d011-b42d-00cf4fc964ff","positionen":[{"produktId":2,"varianteId":2,"menge":1}]}}`, http.StatusBadRequest, "validation_error"},
		{"tätigen: Positionskommentar über 100 Zeichen", taetigen, nil, `{"verkaufId":"6f9619ff-8b86-d011-b42d-00cf4fc964ff","positionen":[{"produktId":1,"varianteId":1,"menge":1,"kommentar":"` + strings.Repeat("ä", 101) + `"}],"kommentar":""}`, http.StatusBadRequest, "validation_error"},
		{"stornieren: Konflikt", stornieren, application.ErrConflict, validStornoBod, http.StatusConflict, "conflict"},
		{"stornieren: Verkauf unbekannt", stornieren, application.ErrVerkaufNichtGefunden, validStornoBod, http.StatusBadRequest, "verkauf_not_found"},
		{"stornieren: Storno-ID mit anderem Inhalt", stornieren, application.ErrVorgangAbweichend, validStornoBod, http.StatusConflict, "vorgang_abweichend"},
		{"stornieren: ohne Storno-ID", stornieren, nil, `{"verkaufId":"6f9619ff-8b86-d011-b42d-00cf4fc964ff","positionen":[{"positionId":"6f9619ff-8b86-d011-b42d-00cf4fc964ff","menge":1}],"kommentar":"Rueckgabe"}`, http.StatusBadRequest, "validation_error"},
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

func TestDirektverkaufTaetigenHandler_ReichtPositionskommentarDurch(t *testing.T) {
	body := `{"verkaufId":"6f9619ff-8b86-d011-b42d-00cf4fc964ff","positionen":[` +
		`{"produktId":1,"varianteId":1,"menge":2,"kommentar":"1 davon ohne Senf"},` +
		`{"produktId":2,"varianteId":2,"menge":1}],"kommentar":""}`
	rec := httptest.NewRecorder()
	cmd := &mockCommand{}

	(&CommandHandler{Command: cmd}).DirektverkaufTaetigenHandler().ServeHTTP(rec, requestWithUser(body))

	if rec.Code >= http.StatusBadRequest {
		t.Fatalf("expected success, got %d: %s", rec.Code, rec.Body.String())
	}
	want := []enrichment.PositionInput{
		{ProduktID: 1, VarianteID: 1, Menge: 2, Kommentar: "1 davon ohne Senf"},
		{ProduktID: 2, VarianteID: 2, Menge: 1, Kommentar: ""},
	}
	if !slices.Equal(cmd.positionen, want) {
		t.Errorf("positionen = %+v, want %+v", cmd.positionen, want)
	}
}
