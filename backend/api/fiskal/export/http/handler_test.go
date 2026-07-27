//go:build unit

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
)

type mockService struct {
	archiv application.Archiv
	err    error
}

func (m *mockService) Erstellen(context.Context, int) (application.Archiv, error) {
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

// deadlineCapturingWriter wraps httptest.ResponseRecorder and implements the
// SetWriteDeadline interface http.ResponseController looks for, so the test
// can observe whether — and when relative to the first write — the handler
// extends the write deadline.
type deadlineCapturingWriter struct {
	*httptest.ResponseRecorder
	deadline            time.Time
	deadlineSet         bool
	deadlineBeforeWrite bool
	wroteAnything       bool
}

func newDeadlineCapturingWriter() *deadlineCapturingWriter {
	return &deadlineCapturingWriter{ResponseRecorder: httptest.NewRecorder()}
}

func (w *deadlineCapturingWriter) SetWriteDeadline(t time.Time) error {
	w.deadline = t
	w.deadlineSet = true
	if !w.wroteAnything {
		w.deadlineBeforeWrite = true
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

// Der Export laeuft gegen die eigene, verlaengerte Schreibfrist statt gegen
// die globale 10-Sekunden-Frist des Servers: Sonst wird ein laenger als zehn
// Sekunden dauernder Export stillschweigend abgeschnitten (Phase 8).
func TestExportHandler_VerlaengertSchreibfristVorErstemSchreibvorgang(t *testing.T) {
	svc := &mockService{archiv: application.Archiv{Dateiname: "dsfinvk_1.zip", Inhalt: []byte("zip-inhalt")}}
	h := &Handler{Service: svc}

	w := newDeadlineCapturingWriter()
	before := time.Now()
	performRequest(t, h.ExportHandler(), w, 0)

	if !w.deadlineSet {
		t.Fatal("expected SetWriteDeadline to be called")
	}
	if !w.deadlineBeforeWrite {
		t.Fatal("expected the write deadline to be set before the first write")
	}
	if min := 5 * time.Minute; w.deadline.Before(before.Add(min)) {
		t.Errorf("expected a write deadline at least %s in the future, got %s", min, w.deadline.Sub(before))
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if body := w.Body.String(); body != "zip-inhalt" {
		t.Errorf("expected body %q, got %q", "zip-inhalt", body)
	}
}

// Unterstuetzt der ResponseWriter SetWriteDeadline nicht (wie
// httptest.ResponseRecorder), wird das protokolliert; der Export laeuft mit
// der globalen Frist unveraendert weiter und liefert eine korrekte Antwort.
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
