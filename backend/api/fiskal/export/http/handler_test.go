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
	// beiErstellen läuft, während der Archivbau simuliert wird — der Test
	// kann damit den Zustand des ResponseWriters zu diesem Zeitpunkt prüfen.
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

// langerArchivbau hält die Frist fest, die während des Archivbaus gilt, und
// lässt Zeit verstreichen, damit eine danach neu gesetzte Frist später liegt.
func langerArchivbau(w *deadlineCapturingWriter, fristBeimArchivbau *time.Time) func() {
	return func() {
		*fristBeimArchivbau = w.frist
		time.Sleep(time.Millisecond)
	}
}

// Der Export läuft gegen die eigene, verlängerte Schreibfrist statt gegen
// die globale 10-Sekunden-Frist des Servers: Sonst wird ein länger als zehn
// Sekunden dauernder Export stillschweigend abgeschnitten.
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

// In Produktion läuft der Handler nie nackt: LoggingMiddleware umschließt die
// gesamte Routenkette (backend/app/app.go) und reicht den Handlern ihren
// eigenen ResponseWriter-Wrapper. Die Frist muss auch durch diesen Wrapper
// hindurch beim echten ResponseWriter ankommen — sonst bleibt die globale
// 10-Sekunden-Frist wirksam und ein grosses DSFinV-K-Archiv reisst mitten im
// ZIP ab. Der Test oben trifft den Handler direkt und würde das übersehen.
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

// Die erste Setzung sitzt am Handler-Eingang und gilt den frühen Fehlerpfaden
// (unlesbarer Body, invalid_kassensitzung); die Antwort nach einem langen
// Archivbau deckt erst die zweite ab, die auch dieser Fehlerzweig durchläuft.
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
	// Auch die Fehlerantwort geht durch beide Fristen — die zweite gibt ihr ein
	// eigenes Budget, nachdem der Archivbau lange gelaufen ist.
	if !w.fristBeimSchreiben.After(fristBeimArchivbau) {
		t.Errorf("expected the write deadline to be set twice before the first write, got %v during the archive build and %v at the first write", fristBeimArchivbau, w.fristBeimSchreiben)
	}
	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

// Die Frist ist eine absolute Zeit ab Request-Start, kein Budget für den
// Schreibvorgang: Nach fünf Minuten Archivbau wäre die am Handler-Eingang gesetzte
// Frist abgelaufen, wenn die Übertragung beginnt. Erst das zweite Setzen gibt ihr
// ein eigenes Budget.
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
	// Maßgeblich ist die Frist beim ersten Schreibvorgang: Eine Setzung hinter
	// dem Schreiben käme für diese Antwort zu spät.
	if !w.fristBeimSchreiben.After(fristBeimArchivbau) {
		t.Errorf("expected the write deadline to be set again after the archive is built and before writing, got %v during the archive build and %v at the first write", fristBeimArchivbau, w.fristBeimSchreiben)
	}
}

// Unterstützt der ResponseWriter SetWriteDeadline nicht (wie
// httptest.ResponseRecorder), wird das protokolliert; der Export läuft mit
// der globalen Frist unverändert weiter und liefert eine korrekte Antwort.
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
