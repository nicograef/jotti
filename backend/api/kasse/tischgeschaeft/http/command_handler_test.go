package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/nicograef/jotti/backend/api/kasse/enrichment"
	"github.com/nicograef/jotti/backend/api/kasse/tischgeschaeft/application"
	"github.com/nicograef/jotti/backend/api/middleware"
	"github.com/nicograef/jotti/backend/domain/kasse"
)

type mockCommand struct {
	err        error
	positionen []enrichment.PositionInput
}

func (m *mockCommand) BestellungAufnehmen(_ context.Context, _ int, _ string, _ string, _ int, positionen []enrichment.PositionInput, _ string) error {
	m.positionen = positionen
	return m.err
}

func (m *mockCommand) BestellungUmbuchen(context.Context, int, string, string, int, int, []kasse.PositionRef, string) error {
	return m.err
}

func (m *mockCommand) ZahlungKassieren(context.Context, int, string, string, int, []kasse.PositionRef, string, *enrichment.PfandEingabe) error {
	return m.err
}

func (m *mockCommand) StornierungErteilen(context.Context, int, string, uuid.UUID, int, []kasse.PositionRef, string) error {
	return m.err
}

const (
	bestellungBody = `{"bestellungId":"6f9619ff-8b86-d011-b42d-00cf4fc964ff","tischId":1,"positionen":[{"produktId":1,"varianteId":1,"menge":2}],"kommentar":""}`
	positionRef    = `"positionen":[{"positionId":"a87f1b2c-3d4e-5f6a-7b8c-9d0e1f2a3b4c","menge":1}]`
	pfandPosition  = `"positionen":[{"produktId":2,"varianteId":2,"menge":1}]`
	zahlungBody    = `{"zahlungId":"0b6a3f1e-2c4d-4e5f-8a9b-1c2d3e4f5a6b","tischId":1,` + positionRef + `,"kommentar":""}`
	stornoBody     = `{"stornoId":"0b6a3f1e-2c4d-4e5f-8a9b-1c2d3e4f5a6b","tischId":1,` + positionRef + `,"kommentar":"Reklamation"}`
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
		{"aufnehmen: Positionskommentar über 100 Zeichen", aufnehmen, nil, `{"bestellungId":"6f9619ff-8b86-d011-b42d-00cf4fc964ff","tischId":1,"positionen":[{"produktId":1,"varianteId":1,"menge":1,"kommentar":"` + strings.Repeat("ä", 101) + `"}],"kommentar":""}`, http.StatusBadRequest, "validation_error"},
		{"aufnehmen: Menge über 999", aufnehmen, nil, `{"bestellungId":"6f9619ff-8b86-d011-b42d-00cf4fc964ff","tischId":1,"positionen":[{"produktId":1,"varianteId":1,"menge":1000}],"kommentar":""}`, http.StatusBadRequest, "validation_error"},
		{"kassieren: Konflikt", (*CommandHandler).ZahlungKassierenHandler, application.ErrConflict, zahlungBody, http.StatusConflict, "conflict"},
		{"kassieren: Zahlungs-ID mit anderem Inhalt", (*CommandHandler).ZahlungKassierenHandler, application.ErrVorgangAbweichend, zahlungBody, http.StatusConflict, "vorgang_abweichend"},
		{"kassieren: ohne Zahlungs-ID", (*CommandHandler).ZahlungKassierenHandler, nil, `{"tischId":1,` + positionRef + `,"kommentar":""}`, http.StatusBadRequest, "validation_error"},
		{"kassieren: Pfandrückgabe abweichend", (*CommandHandler).ZahlungKassierenHandler, application.ErrPfandrueckgabeAbweichend, zahlungBody, http.StatusConflict, "pfandrueckgabe_abweichend"},
		{"kassieren: Pfand ohne Rückgabe-ID", (*CommandHandler).ZahlungKassierenHandler, nil, `{"zahlungId":"0b6a3f1e-2c4d-4e5f-8a9b-1c2d3e4f5a6b","tischId":1,` + positionRef + `,"kommentar":"","pfand":{` + pfandPosition + `}}`, http.StatusBadRequest, "validation_error"},
		{"stornieren: Konflikt", (*CommandHandler).StornierungErteilenHandler, application.ErrConflict, stornoBody, http.StatusConflict, "conflict"},
		{"stornieren: ohne Storno-ID", (*CommandHandler).StornierungErteilenHandler, nil, `{"tischId":1,` + positionRef + `,"kommentar":"Reklamation"}`, http.StatusBadRequest, "validation_error"},
		{"umbuchen: Konflikt", umbuchen, application.ErrConflict, `{"umbuchungId":"0b6a3f1e-2c4d-4e5f-8a9b-1c2d3e4f5a6b","quellTischId":1,"zielTischId":2,` + positionRef + `}`, http.StatusConflict, "conflict"},
		{"umbuchen: gleicher Tisch", umbuchen, application.ErrUmbuchungGleicherTisch, `{"umbuchungId":"0b6a3f1e-2c4d-4e5f-8a9b-1c2d3e4f5a6b","quellTischId":1,"zielTischId":1,` + positionRef + `}`, http.StatusBadRequest, "umbuchung_gleicher_tisch"},
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

func TestBestellungAufnehmenHandler_ReichtPositionskommentarDurch(t *testing.T) {
	body := `{"bestellungId":"6f9619ff-8b86-d011-b42d-00cf4fc964ff","tischId":1,"positionen":[` +
		`{"produktId":1,"varianteId":1,"menge":1,"kommentar":"ohne Kümmel"},` +
		`{"produktId":2,"varianteId":2,"menge":1}],"kommentar":""}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, 1))
	rec := httptest.NewRecorder()
	cmd := &mockCommand{}

	(&CommandHandler{Command: cmd}).BestellungAufnehmenHandler().ServeHTTP(rec, req)

	if rec.Code >= http.StatusBadRequest {
		t.Fatalf("expected success, got %d: %s", rec.Code, rec.Body.String())
	}
	want := []enrichment.PositionInput{
		{ProduktID: 1, VarianteID: 1, Menge: 1, Kommentar: "ohne Kümmel"},
		{ProduktID: 2, VarianteID: 2, Menge: 1, Kommentar: ""},
	}
	if !slices.Equal(cmd.positionen, want) {
		t.Errorf("positionen = %+v, want %+v", cmd.positionen, want)
	}
}
