package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nicograef/jotti/backend/api/kasse/direktverkauf/application"
	"github.com/nicograef/jotti/backend/api/kasse/enrichment"
	"github.com/nicograef/jotti/backend/domain/kasse"
)

type mockPfandCommand struct {
	err error
}

func (m *mockPfandCommand) PfandZurueckgeben(context.Context, int, string, string, int, []enrichment.PositionInput, string) error {
	return m.err
}

func (m *mockPfandCommand) PfandrueckgabeStornieren(context.Context, int, string, string, string, []kasse.PositionRef, string) error {
	return m.err
}

const validPfandStornoBody = `{"stornierungId":"11111111-1111-4111-8111-111111111111","rueckgabeId":"22222222-2222-4222-8222-222222222222","positionen":[{"positionId":"33333333-3333-4333-8333-333333333333","menge":1}],"kommentar":"Falsch gezählt"}`

// Each row is a request the handler must refuse: a body the schema rejects, or a
// domain error and the status and wire code it maps to.
func TestPfandCommandHandler_Abweisungen(t *testing.T) {
	stornieren := (*PfandCommandHandler).PfandrueckgabeStornierenHandler
	tests := []struct {
		name       string
		handler    func(*PfandCommandHandler) http.HandlerFunc
		err        error
		body       string
		wantStatus int
		wantCode   string
	}{
		{"stornieren: Stornierung abweichend", stornieren, application.ErrPfandrueckgabeAbweichend, validPfandStornoBody, http.StatusConflict, "pfandrueckgabe_abweichend"},
		{"stornieren: Stornierungs-ID gleich Rückgabe-ID", stornieren, nil, `{"stornierungId":"22222222-2222-4222-8222-222222222222","rueckgabeId":"22222222-2222-4222-8222-222222222222","positionen":[{"positionId":"33333333-3333-4333-8333-333333333333","menge":1}],"kommentar":"Falsch gezählt"}`, http.StatusBadRequest, "validation_error"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			tt.handler(&PfandCommandHandler{Command: &mockPfandCommand{err: tt.err}}).ServeHTTP(rec, requestWithUser(tt.body))

			if rec.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d: %s", tt.wantStatus, rec.Code, rec.Body.String())
			}
			if !strings.Contains(rec.Body.String(), `"code":"`+tt.wantCode+`"`) {
				t.Errorf("expected code %s, got %s", tt.wantCode, rec.Body.String())
			}
		})
	}
}
