package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/nicograef/jotti/backend/api/fiskal/export/application"
	"github.com/nicograef/jotti/backend/api/middleware"
)

type mockService struct {
	archiv application.Archiv
	err    error
	// beiErstellen runs while the archive build is simulated, so the test can inspect
	// the ResponseWriter state at that moment.
	beiErstellen func()
}

func (m *mockService) Erstellen(context.Context, int) (application.Archiv, error) {
	if m.beiErstellen != nil {
		m.beiErstellen()
	}
	return m.archiv, m.err
}

func performRequest(t *testing.T, handler http.HandlerFunc, w http.ResponseWriter, kassensitzungNr int) {
	t.Helper()

	payload, err := json.Marshal(exportRequest{KassensitzungNr: kassensitzungNr})
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(payload))
	handler(w, req)
}

// deadlineCapturingWriter implements the SetWriteDeadline interface that
// http.ResponseController looks for. fristBeimSchreiben records the deadline at the
// first write, the only one that budgets the response.
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

// langerArchivbau records the deadline during the archive build and lets time pass,
// so a deadline set afterwards is later.
func langerArchivbau(w *deadlineCapturingWriter, fristBeimArchivbau *time.Time) func() {
	return func() {
		*fristBeimArchivbau = w.frist
		time.Sleep(time.Millisecond)
	}
}

// The export runs against its own extended write deadline, not the server's global
// 10-second one; otherwise an export over ten seconds is cut silently.
func TestExportHandler_VerlaengertSchreibfristVorErstemSchreibvorgang(t *testing.T) {
	w := newDeadlineCapturingWriter()
	var fristBeimArchivbau time.Time
	svc := &mockService{
		archiv:       application.Archiv{Dateiname: "dsfinvk_1.zip", Inhalt: []byte("zip-inhalt")},
		beiErstellen: langerArchivbau(w, &fristBeimArchivbau),
	}
	h := &Handler{Service: svc}

	before := time.Now()
	performRequest(t, h.ExportHandler(), w, 0)

	if w.fristBeimSchreiben.IsZero() {
		t.Fatal("expected SetWriteDeadline to be called")
	}
	if fristBeimArchivbau.IsZero() || !w.fristBeimSchreiben.After(fristBeimArchivbau) {
		t.Errorf("expected the write deadline to be set twice before the first write (handler entry and right before writing), got %v during the archive build and %v at the first write", fristBeimArchivbau, w.fristBeimSchreiben)
	}
	if min := 5 * time.Minute; w.fristBeimSchreiben.Before(before.Add(min)) {
		t.Errorf("expected a write deadline at least %s in the future, got %s", min, w.fristBeimSchreiben.Sub(before))
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if body := w.Body.String(); body != "zip-inhalt" {
		t.Errorf("expected body %q, got %q", "zip-inhalt", body)
	}
}

// In production LoggingMiddleware wraps the route chain (backend/app/app.go) in its own
// ResponseWriter, and the deadline must reach the real writer through it or a large
// archive breaks off mid-ZIP. The direct-handler test above would miss that.
func TestExportHandler_VerlaengertSchreibfristHinterLoggingMiddleware(t *testing.T) {
	w := newDeadlineCapturingWriter()
	var fristBeimArchivbau time.Time
	svc := &mockService{
		archiv:       application.Archiv{Dateiname: "dsfinvk_1.zip", Inhalt: []byte("zip-inhalt")},
		beiErstellen: langerArchivbau(w, &fristBeimArchivbau),
	}
	h := &Handler{Service: svc}

	before := time.Now()
	performRequest(t, middleware.LoggingMiddleware(h.ExportHandler()).ServeHTTP, w, 0)

	if w.fristBeimSchreiben.IsZero() {
		t.Fatal("expected SetWriteDeadline to reach the real ResponseWriter through the middleware chain")
	}
	if fristBeimArchivbau.IsZero() || !w.fristBeimSchreiben.After(fristBeimArchivbau) {
		t.Errorf("expected the write deadline to be set twice before the first write (handler entry and right before writing), got %v during the archive build and %v at the first write", fristBeimArchivbau, w.fristBeimSchreiben)
	}
	if min := 5 * time.Minute; w.fristBeimSchreiben.Before(before.Add(min)) {
		t.Errorf("expected a write deadline at least %s in the future, got %s", min, w.fristBeimSchreiben.Sub(before))
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if body := w.Body.String(); body != "zip-inhalt" {
		t.Errorf("expected body %q, got %q", "zip-inhalt", body)
	}
}

// The entry deadline covers the early error paths; a response after a long archive
// build is covered only by the second one, which this error branch also passes.
func TestExportHandler_VerlaengertSchreibfristVorDemArchivbau(t *testing.T) {
	w := newDeadlineCapturingWriter()
	var fristBeimArchivbau time.Time
	svc := &mockService{
		err:          application.ErrKassensitzungNichtGefunden,
		beiErstellen: langerArchivbau(w, &fristBeimArchivbau),
	}
	h := &Handler{Service: svc}

	performRequest(t, h.ExportHandler(), w, 5)

	if fristBeimArchivbau.IsZero() {
		t.Error("expected the write deadline to be extended before Erstellen() runs")
	}
	// The error response also passes both deadlines; the second budgets it after the
	// long archive build.
	if !w.fristBeimSchreiben.After(fristBeimArchivbau) {
		t.Errorf("expected the write deadline to be set twice before the first write, got %v during the archive build and %v at the first write", fristBeimArchivbau, w.fristBeimSchreiben)
	}
	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

// The deadline is absolute from request start: after five minutes of archive build
// the entry deadline has expired when the transfer starts. Only the second setting
// gives the transfer its own budget.
func TestExportHandler_SetztSchreibfristVorDemSchreibenErneut(t *testing.T) {
	w := newDeadlineCapturingWriter()
	var fristBeimArchivbau time.Time
	svc := &mockService{
		archiv:       application.Archiv{Dateiname: "dsfinvk_1.zip", Inhalt: []byte("zip-inhalt")},
		beiErstellen: langerArchivbau(w, &fristBeimArchivbau),
	}
	h := &Handler{Service: svc}

	performRequest(t, h.ExportHandler(), w, 0)

	if fristBeimArchivbau.IsZero() {
		t.Error("expected a write deadline before the archive is built")
	}
	// What counts is the deadline at the first write; one set after writing comes too
	// late for this response.
	if !w.fristBeimSchreiben.After(fristBeimArchivbau) {
		t.Errorf("expected the write deadline to be set again after the archive is built and before writing, got %v during the archive build and %v at the first write", fristBeimArchivbau, w.fristBeimSchreiben)
	}
}

// If the ResponseWriter lacks SetWriteDeadline (like httptest.ResponseRecorder), this
// is logged and the export continues with the global deadline and a correct response.
func TestExportHandler_LaeuftWeiterWennFristNichtSetzbar(t *testing.T) {
	svc := &mockService{archiv: application.Archiv{Dateiname: "dsfinvk_1.zip", Inhalt: []byte("zip-inhalt")}}
	h := &Handler{Service: svc}

	rec := httptest.NewRecorder()
	performRequest(t, h.ExportHandler(), rec, 0)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	if contentType := rec.Header().Get("Content-Type"); contentType != "application/zip" {
		t.Errorf("expected Content-Type application/zip, got %q", contentType)
	}
	if body := rec.Body.String(); body != "zip-inhalt" {
		t.Errorf("expected body %q, got %q", "zip-inhalt", body)
	}
}

func TestExportHandler_KassensitzungNichtGefunden(t *testing.T) {
	svc := &mockService{err: application.ErrKassensitzungNichtGefunden}
	h := &Handler{Service: svc}

	rec := httptest.NewRecorder()
	performRequest(t, h.ExportHandler(), rec, 5)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestExportHandler_InvalidKassensitzungNr(t *testing.T) {
	svc := &mockService{}
	h := &Handler{Service: svc}

	rec := httptest.NewRecorder()
	performRequest(t, h.ExportHandler(), rec, -1)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}
