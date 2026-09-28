package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/nicograef/jotti/backend/api/middleware"
)

// deadlineCapturingWriter implementiert das SetWriteDeadline-Interface, das
// http.ResponseController sucht. fristBeimSchreiben hält die Frist fest, die
// beim ersten Schreibvorgang gilt: Nur sie gibt der Antwort ein Budget.
type deadlineCapturingWriter struct {
	*httptest.ResponseRecorder
	frist              time.Time
	fristBeimSchreiben time.Time
	geschrieben        bool
}

func newDeadlineCapturingWriter() *deadlineCapturingWriter {
	return &deadlineCapturingWriter{ResponseRecorder: httptest.NewRecorder()}
}

func (w *deadlineCapturingWriter) SetWriteDeadline(t time.Time) error {
	w.frist = t
	return nil
}

func (w *deadlineCapturingWriter) WriteHeader(code int) {
	w.merkeErstenSchreibvorgang()
	w.ResponseRecorder.WriteHeader(code)
}

func (w *deadlineCapturingWriter) Write(b []byte) (int, error) {
	w.merkeErstenSchreibvorgang()
	return w.ResponseRecorder.Write(b)
}

func (w *deadlineCapturingWriter) merkeErstenSchreibvorgang() {
	if !w.geschrieben {
		w.geschrieben = true
		w.fristBeimSchreiben = w.frist
	}
}

// Die beiden schreibenden TSE-Endpunkte fahren einen fiskaly-Lebenszyklus, der die
// globale 10-Sekunden-Schreibfrist überschreiten kann. Ohne verlängerte Frist
// stirbt die Antwort auf der Verbindung — samt PUK und Admin-PIN, die genau einmal
// ausgeliefert und nirgends persistiert werden. Der Test läuft durch die
// LoggingMiddleware, weil sie in Produktion die gesamte Routenkette umschließt
// (app/app.go) und die Frist auch durch ihren ResponseWriter-Wrapper ankommen muss.
//
// Die Frist muss ZWEIMAL gesetzt werden: Sie ist eine absolute Zeit ab
// Request-Start. Der Aufruf am Handler-Eingang deckt die frühen Fehlerpfade ab,
// der Aufruf vor dem Schreiben gibt der Antwort ein eigenes Budget.
func TestTSESetupHandler_VerlaengertSchreibfristVorErstemSchreibvorgang(t *testing.T) {
	faelle := []struct {
		route   string
		handler func(*CommandHandler) http.HandlerFunc
		body    string
	}{
		{"/admin/tse-einrichten", (*CommandHandler).RichteTSEEinHandler, `{"apiKey":"key","apiSecret":"secret","umgebung":"TEST"}`},
		{"/admin/tse-uebernehmen", (*CommandHandler).UebernimmTSEHandler, `{"apiKey":"key","apiSecret":"secret","umgebung":"TEST","tssId":"tss-123"}`},
	}

	for _, fall := range faelle {
		t.Run(fall.route, func(t *testing.T) {
			w := newDeadlineCapturingWriter()
			// Der simulierte Lebenszyklus hält die Frist fest, die während
			// fiskaly gilt, und lässt Zeit verstreichen, damit eine danach neu
			// gesetzte Frist später liegt.
			var fristImLebenszyklus time.Time
			command := &CommandHandler{Command: &mockTSESetupCommand{waehrendLebenszyklus: func() {
				fristImLebenszyklus = w.frist
				time.Sleep(time.Millisecond)
			}}}
			req := httptest.NewRequest(http.MethodPost, fall.route, strings.NewReader(fall.body))
			req.Header.Set("Content-Type", "application/json")

			before := time.Now()
			middleware.LoggingMiddleware(fall.handler(command)).ServeHTTP(w, req)

			if w.fristBeimSchreiben.IsZero() {
				t.Fatal("expected SetWriteDeadline to reach the real ResponseWriter")
			}
			if fristImLebenszyklus.IsZero() || !w.fristBeimSchreiben.After(fristImLebenszyklus) {
				t.Errorf("expected the write deadline to be set twice before the first write (handler entry and right before writing), got %v during the lifecycle and %v at the first write", fristImLebenszyklus, w.fristBeimSchreiben)
			}
			if min := 2 * time.Minute; w.fristBeimSchreiben.Before(before.Add(min)) {
				t.Errorf("expected a write deadline at least %s in the future, got %s", min, w.fristBeimSchreiben.Sub(before))
			}
			if w.Code != http.StatusOK {
				t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
			}
		})
	}
}
