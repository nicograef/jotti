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
	err     error
	tischID int
}

func (m *mockPfandCommand) PfandZurueckgeben(_ context.Context, _ int, _ string, _ string, tischID int, _ []enrichment.PositionInput, _ string) error {
	m.tischID = tischID
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

// mockQuery records the tischId of a Pfandrückgabe-Historie call.
type mockQuery struct {
	query
	pfandTischID int
}

func (m *mockQuery) GetPfandrueckgabeHistorie(_ context.Context, tischID int) ([]kasse.PfandrueckgabeHistorieEintrag, error) {
	m.pfandTischID = tischID
	return nil, nil
}

// tischId 0 is the Direktverkauf and reaches the application as 0; a missing tischId is a
// validation error. The pointer schema exists because zog counts a required 0 as missing.
func TestPfandHandler_TischID(t *testing.T) {
	const nichtAufgerufen = -1
	zurueckgeben := func(rec *httptest.ResponseRecorder, body string) int {
		cmd := &mockPfandCommand{tischID: nichtAufgerufen}
		(&PfandCommandHandler{Command: cmd}).PfandZurueckgebenHandler().ServeHTTP(rec, requestWithUser(body))
		return cmd.tischID
	}
	historie := func(rec *httptest.ResponseRecorder, body string) int {
		query := &mockQuery{pfandTischID: nichtAufgerufen}
		(&QueryHandler{Query: query}).GetPfandrueckgabeHistorieHandler().ServeHTTP(rec, requestWithUser(body))
		return query.pfandTischID
	}
	const positionen = `"rueckgabeId":"22222222-2222-4222-8222-222222222222","positionen":[{"produktId":1,"varianteId":1,"menge":2}]`
	tests := []struct {
		name        string
		serve       func(*httptest.ResponseRecorder, string) int
		body        string
		wantStatus  int
		wantTischID int
	}{
		{"zurückgeben: Direktverkauf", zurueckgeben, `{` + positionen + `,"tischId":0}`, http.StatusOK, 0},
		{"zurückgeben: ohne tischId", zurueckgeben, `{` + positionen + `}`, http.StatusBadRequest, nichtAufgerufen},
		{"Historie: Direktverkauf", historie, `{"tischId":0}`, http.StatusOK, 0},
		{"Historie: ohne tischId", historie, `{}`, http.StatusBadRequest, nichtAufgerufen},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			gotTischID := tt.serve(rec, tt.body)

			if rec.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d: %s", tt.wantStatus, rec.Code, rec.Body.String())
			}
			if gotTischID != tt.wantTischID {
				t.Errorf("expected tischId %d at the application, got %d", tt.wantTischID, gotTischID)
			}
			if tt.wantStatus == http.StatusBadRequest && !strings.Contains(rec.Body.String(), `"code":"validation_error"`) {
				t.Errorf("expected validation_error, got %s", rec.Body.String())
			}
		})
	}
}
