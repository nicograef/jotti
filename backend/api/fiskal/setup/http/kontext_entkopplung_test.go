//go:build unit

package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nicograef/jotti/backend/api/fiskal/setup/application"
	"github.com/nicograef/jotti/backend/api/middleware"
	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/rs/zerolog"
)

// abbruchProbe spielt den Client, der waehrend der Arbeit des Handlers
// abbricht: Sie storniert den Request-Kontext und haelt fest, was der an die
// Application-Schicht uebergebene Kontext danach meldet.
type abbruchProbe struct {
	abbrechen        context.CancelFunc
	kontextStorniert bool
	korrelation      string
	loggerAktiv      bool
}

// beobachte laeuft anstelle der fiskaly-Sequenz. context.WithCancel schliesst
// den Done-Kanal aller Kinder synchron im cancel-Aufruf, die Pruefung direkt
// danach ist damit deterministisch.
func (p *abbruchProbe) beobachte(ctx context.Context) {
	p.abbrechen()
	p.kontextStorniert = ctx.Err() != nil
	p.korrelation, _ = ctx.Value(middleware.CorrelationIDKey).(string)
	p.loggerAktiv = zerolog.Ctx(ctx).GetLevel() != zerolog.Disabled
}

type abbruchCommand struct {
	*mockSettingsCommand
	probe *abbruchProbe
}

func (c *abbruchCommand) RichteTSEEin(ctx context.Context, credentials tse.SetupCredentials, umgebung tse.Umgebung, neuAnlegen bool) (application.TSESetupErgebnis, error) {
	c.probe.beobachte(ctx)
	return c.mockSettingsCommand.RichteTSEEin(ctx, credentials, umgebung, neuAnlegen)
}

func (c *abbruchCommand) UebernimmTSE(ctx context.Context, credentials tse.SetupCredentials, umgebung tse.Umgebung, tssID, pin, puk string) (application.TSESetupErgebnis, error) {
	c.probe.beobachte(ctx)
	return c.mockSettingsCommand.UebernimmTSE(ctx, credentials, umgebung, tssID, pin, puk)
}

type abbruchQuery struct {
	*mockSettingsQuery
	probe *abbruchProbe
}

func (q *abbruchQuery) TestTSEVerbindung(ctx context.Context) (tse.VerbindungStatus, error) {
	q.probe.beobachte(ctx)
	return q.mockSettingsQuery.TestTSEVerbindung(ctx)
}

func (q *abbruchQuery) CheckTSESetup(ctx context.Context, credentials tse.SetupCredentials) (application.TSESetupBefund, error) {
	q.probe.beobachte(ctx)
	return q.mockSettingsQuery.CheckTSESetup(ctx, credentials)
}

// Die beiden schreibenden Endpunkte fahren einen fiskaly-Lebenszyklus, der
// nicht mittendrin abbrechen darf: Zurueck bliebe eine bezahlte, halbfertige
// TSS, deren PUK und Admin-PIN es nur in der verlorenen Antwort gab. Ein
// Client-Abbruch storniert r.Context() — deshalb laufen sie unter einem davon
// abgekoppelten Kontext (lebenszyklusKontext). Die beiden lesenden Endpunkte
// sind idempotent und wiederholbar; sie behalten r.Context() und sollen mit dem
// Client abbrechen, statt fiskaly ohne Zuhoerer weiter zu befragen.
func TestTSESetupHandler_EntkoppeltNurDieSchreibendenVomClientAbbruch(t *testing.T) {
	faelle := []struct {
		route            string
		body             string
		handler          func(*abbruchProbe) http.HandlerFunc
		erwarteStorniert bool
	}{
		{
			route: "/admin/tse-einrichten",
			body:  `{"apiKey":"key","apiSecret":"secret","umgebung":"TEST"}`,
			handler: func(p *abbruchProbe) http.HandlerFunc {
				h := &CommandHandler{Command: &abbruchCommand{mockSettingsCommand: &mockSettingsCommand{}, probe: p}}
				return h.RichteTSEEinHandler()
			},
			erwarteStorniert: false,
		},
		{
			route: "/admin/tse-uebernehmen",
			body:  `{"apiKey":"key","apiSecret":"secret","umgebung":"TEST","tssId":"tss-123"}`,
			handler: func(p *abbruchProbe) http.HandlerFunc {
				h := &CommandHandler{Command: &abbruchCommand{mockSettingsCommand: &mockSettingsCommand{}, probe: p}}
				return h.UebernimmTSEHandler()
			},
			erwarteStorniert: false,
		},
		{
			route: "/admin/test-tse-verbindung",
			body:  `{}`,
			handler: func(p *abbruchProbe) http.HandlerFunc {
				h := &QueryHandler{Query: &abbruchQuery{mockSettingsQuery: &mockSettingsQuery{}, probe: p}}
				return h.TestTSEVerbindungHandler()
			},
			erwarteStorniert: true,
		},
		{
			route: "/admin/tse-setup-pruefen",
			body:  `{"apiKey":"key","apiSecret":"secret"}`,
			handler: func(p *abbruchProbe) http.HandlerFunc {
				h := &QueryHandler{Query: &abbruchQuery{mockSettingsQuery: &mockSettingsQuery{}, probe: p}}
				return h.CheckTSESetupHandler()
			},
			erwarteStorniert: true,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.route, func(t *testing.T) {
			probe := &abbruchProbe{}

			req := httptest.NewRequest(http.MethodPost, fall.route, strings.NewReader(fall.body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Correlation-ID", "korr-1")
			ctx, cancel := context.WithCancel(req.Context())
			defer cancel()
			probe.abbrechen = cancel
			req = req.WithContext(ctx)

			rec := httptest.NewRecorder()
			kette := middleware.CorrelationIDMiddleware(middleware.LoggingMiddleware(fall.handler(probe)))
			kette.ServeHTTP(rec, req)

			if probe.kontextStorniert != fall.erwarteStorniert {
				t.Fatalf("expected the context handed to the application layer to be cancelled=%v after the client aborted, got %v", fall.erwarteStorniert, probe.kontextStorniert)
			}
			if probe.korrelation != "korr-1" {
				t.Errorf("expected the correlation id to survive in the context, got %q", probe.korrelation)
			}
			if !probe.loggerAktiv {
				t.Error("expected the zerolog logger to survive in the context")
			}
			if rec.Code != http.StatusOK {
				t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
			}
		})
	}
}

// Der abgekoppelte Kontext endet spaetestens, wenn der Handler zurueckkehrt
// (defer cancel) — abgekoppelt heisst nicht unsterblich.
func TestTSESetupHandler_LebenszyklusKontextEndetMitDemHandler(t *testing.T) {
	var erfasst context.Context
	command := &kontextErfassendesCommand{mockSettingsCommand: &mockSettingsCommand{}, erfasst: &erfasst}
	h := &CommandHandler{Command: command}

	req := httptest.NewRequest(http.MethodPost, "/admin/tse-einrichten", strings.NewReader(`{"apiKey":"key","apiSecret":"secret","umgebung":"TEST"}`))
	req.Header.Set("Content-Type", "application/json")
	h.RichteTSEEinHandler().ServeHTTP(httptest.NewRecorder(), req)

	if erfasst == nil {
		t.Fatal("expected the handler to hand a context to the application layer")
	}
	if erfasst.Err() == nil {
		t.Fatal("expected the decoupled context to be cancelled once the handler returned")
	}
	if _, gesetzt := erfasst.Deadline(); !gesetzt {
		t.Error("expected the decoupled context to carry the leak-guard deadline")
	}
}

type kontextErfassendesCommand struct {
	*mockSettingsCommand
	erfasst *context.Context
}

func (c *kontextErfassendesCommand) RichteTSEEin(ctx context.Context, credentials tse.SetupCredentials, umgebung tse.Umgebung, neuAnlegen bool) (application.TSESetupErgebnis, error) {
	*c.erfasst = ctx
	return c.mockSettingsCommand.RichteTSEEin(ctx, credentials, umgebung, neuAnlegen)
}
