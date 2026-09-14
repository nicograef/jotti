//go:build unit

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
// http.ResponseController sucht, und zählt die Aufrufe. deadlineCountBeforeWrite
// zählt nur bis zum ersten Schreibvorgang: Nur die Aufrufe DAVOR können der
// Antwort ein Budget geben.
type deadlineCapturingWriter struct {
	*httptest.ResponseRecorder
	deadline                 time.Time
	deadlineCount            int
	deadlineCountBeforeWrite int
	wroteAnything            bool
}

func newDeadlineCapturingWriter() *deadlineCapturingWriter {
	return &deadlineCapturingWriter{ResponseRecorder: httptest.NewRecorder()}
}

func (w *deadlineCapturingWriter) SetWriteDeadline(t time.Time) error {
	w.deadline = t
	w.deadlineCount++
	if !w.wroteAnything {
		w.deadlineCountBeforeWrite++
	}
	return nil
}

func (w *deadlineCapturingWriter) WriteHeader(code int) {
	w.wroteAnything = true
	w.ResponseRecorder.WriteHeader(code)
}

func (w *deadlineCapturingWriter) Write(b []byte) (int, error) {
	w.wroteAnything = true
	return w.ResponseRecorder.Write(b)
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
	command := &CommandHandler{Command: &mockTSESetupCommand{}}

	faelle := []struct {
		route   string
		handler http.HandlerFunc
		body    string
	}{
		{"/admin/tse-einrichten", command.RichteTSEEinHandler(), `{"apiKey":"key","apiSecret":"secret","umgebung":"TEST"}`},
		{"/admin/tse-uebernehmen", command.UebernimmTSEHandler(), `{"apiKey":"key","apiSecret":"secret","umgebung":"TEST","tssId":"tss-123"}`},
	}

	for _, fall := range faelle {
		t.Run(fall.route, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, fall.route, strings.NewReader(fall.body))
			req.Header.Set("Content-Type", "application/json")
			w := newDeadlineCapturingWriter()

			before := time.Now()
			middleware.LoggingMiddleware(fall.handler).ServeHTTP(w, req)

			if w.deadlineCount == 0 {
				t.Fatal("expected SetWriteDeadline to reach the real ResponseWriter")
			}
			if w.deadlineCountBeforeWrite != 2 {
				t.Fatalf("expected the write deadline to be set twice before the first write (handler entry and right before writing), got %d of %d calls", w.deadlineCountBeforeWrite, w.deadlineCount)
			}
			if min := 2 * time.Minute; w.deadline.Before(before.Add(min)) {
				t.Errorf("expected a write deadline at least %s in the future, got %s", min, w.deadline.Sub(before))
			}
			if w.Code != http.StatusOK {
				t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
			}
		})
	}
}
