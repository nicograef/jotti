package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nicograef/jotti/backend/api/kasse/kassenfuehrung/application"
	"github.com/nicograef/jotti/backend/api/middleware"
)

type mockCommand struct {
	err      error
	ergebnis application.KassenabschlussErgebnis
	aufrufe  int
}

func (m *mockCommand) KassensitzungEroeffnen(context.Context, int, string, string, int) (int, error) {
	m.aufrufe++
	return 1, m.err
}

func (m *mockCommand) GeldtransitBuchen(context.Context, int, string, string, string, int, string) error {
	m.aufrufe++
	return m.err
}

func (m *mockCommand) KasseAbschliessen(context.Context, int, string, int) (application.KassenabschlussErgebnis, error) {
	m.aufrufe++
	return m.ergebnis, m.err
}

func requestWithUser(body string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, 1)
	ctx = context.WithValue(ctx, middleware.UserNameKey, "Tester")
	return req.WithContext(ctx)
}

const geldtransitID = `"geldtransitId":"6f9619ff-8b86-d011-b42d-00cf4fc964ff"`

// Each row pins one decision of a handler: a boundary the schema accepts or
// rejects, or a domain error and the status and wire code it maps to. An accepted
// row must also reach the command.
func TestCommandHandler_Entscheidungen(t *testing.T) {
	eroeffnen := (*CommandHandler).KassensitzungEroeffnenHandler
	geldtransit := (*CommandHandler).GeldtransitBuchenHandler
	abschliessen := (*CommandHandler).KasseAbschliessenHandler
	tests := []struct {
		name       string
		handler    func(*CommandHandler) http.HandlerFunc
		cmd        mockCommand
		body       string
		wantStatus int
		wantInBody []string
	}{
		{"eröffnen: Anfangsbestand 0", eroeffnen, mockCommand{}, `{"bezeichnung":"Maihock","betragCents":0}`, http.StatusOK, nil},
		{"eröffnen: Anfangsbestand fehlt", eroeffnen, mockCommand{}, `{"bezeichnung":"Maihock"}`, http.StatusBadRequest, []string{"Anfangsbestand ist erforderlich"}},
		{"eröffnen: Anfangsbestand negativ", eroeffnen, mockCommand{}, `{"bezeichnung":"Maihock","betragCents":-1}`, http.StatusBadRequest, []string{"Anfangsbestand darf nicht negativ sein"}},
		{"eröffnen: Kasse schon offen", eroeffnen, mockCommand{err: application.ErrKasseAlreadyOpen}, `{"bezeichnung":"Maihock","betragCents":10000}`, http.StatusBadRequest, []string{"kasse_bereits_geoeffnet"}},
		{"eröffnen: Betreiber fehlt", eroeffnen, mockCommand{err: application.ErrBetreiberNichtKonfiguriert}, `{"bezeichnung":"Maihock","betragCents":10000}`, http.StatusBadRequest, []string{"betreiber_nicht_konfiguriert"}},
		{"geldtransit: Richtung fehlt", geldtransit, mockCommand{}, `{` + geldtransitID + `,"richtung":"","betragCents":500,"kommentar":"Initialbestand"}`, http.StatusBadRequest, []string{"validation_error"}},
		{"geldtransit: Kommentar zu kurz", geldtransit, mockCommand{}, `{` + geldtransitID + `,"richtung":"einlage","betragCents":500,"kommentar":"ab"}`, http.StatusBadRequest, []string{"validation_error"}},
		{"geldtransit: Betrag 0", geldtransit, mockCommand{}, `{` + geldtransitID + `,"richtung":"einlage","betragCents":0,"kommentar":"Initialbestand"}`, http.StatusBadRequest, []string{"validation_error"}},
		{"geldtransit: ID keine UUID", geldtransit, mockCommand{}, `{"geldtransitId":"nicht-eine-uuid","richtung":"einlage","betragCents":500,"kommentar":"Test"}`, http.StatusBadRequest, []string{"validation_error"}},
		{"abschließen: Ist-Bestand 0", abschliessen, mockCommand{}, `{"istBestandCents":0}`, http.StatusOK, nil},
		{"abschließen: Ist-Bestand fehlt", abschliessen, mockCommand{}, `{}`, http.StatusBadRequest, []string{"Ist-Bestand ist erforderlich"}},
		{"abschließen: Ist-Bestand negativ", abschliessen, mockCommand{}, `{"istBestandCents":-1}`, http.StatusBadRequest, []string{"Ist-Bestand darf nicht negativ sein"}},
		{"abschließen: Kasse nicht geöffnet", abschliessen, mockCommand{err: application.ErrKasseNichtGeoeffnet}, `{"istBestandCents":10000}`, http.StatusConflict, []string{"kasse_nicht_geoeffnet"}},
		{"abschließen: Signaturen ausstehend", abschliessen, mockCommand{err: &application.SignaturenAusstehendError{Anzahl: 2}}, `{"istBestandCents":10000}`, http.StatusConflict, []string{"signaturen_ausstehend", `"anzahl":2`}},
		{"abschließen: Tische mit Saldo", abschliessen, mockCommand{err: application.ErrTischeSaldoOffen}, `{"istBestandCents":10000}`, http.StatusBadRequest, []string{"tische_saldo_offen"}},
		{"abschließen: Buchungen nach Kassensturz", abschliessen, mockCommand{err: application.ErrBuchungenNachKassensturz}, `{"istBestandCents":10000}`, http.StatusBadRequest, []string{"buchungen_nach_kassensturz"}},
		{"abschließen: Ausfall-Reste im Body", abschliessen, mockCommand{ergebnis: application.KassenabschlussErgebnis{AusfallResteAnzahl: 3, OhneKonfigurationAnzahl: 1}}, `{"istBestandCents":10000}`, http.StatusOK, []string{`"ausfallResteAnzahl":3`, `"ohneKonfigurationAnzahl":1`}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.cmd
			rec := httptest.NewRecorder()

			tt.handler(&CommandHandler{Command: &cmd}).ServeHTTP(rec, requestWithUser(tt.body))

			if rec.Code != tt.wantStatus {
				t.Fatalf("expected %d, got %d: %s", tt.wantStatus, rec.Code, rec.Body.String())
			}
			for _, want := range tt.wantInBody {
				if !strings.Contains(rec.Body.String(), want) {
					t.Errorf("expected %s in body, got %s", want, rec.Body.String())
				}
			}
			if tt.wantStatus == http.StatusOK && cmd.aufrufe != 1 {
				t.Errorf("expected one command call, got %d", cmd.aufrufe)
			}
		})
	}
}
