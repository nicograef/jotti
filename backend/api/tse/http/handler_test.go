//go:build unit

package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/nicograef/jotti/backend/repository/tse_repo"
)

type mockSignaturauftragQuery struct {
	queue      tse_repo.SignaturQueueZustand
	stoerungen []tse_repo.Stoerungszeitraum
	err        error
}

func (m *mockSignaturauftragQuery) GetTSESignaturQueueZustand(context.Context) (tse_repo.SignaturQueueZustand, error) {
	return m.queue, m.err
}

func (m *mockSignaturauftragQuery) GetTSEStoerungen(context.Context) ([]tse_repo.Stoerungszeitraum, error) {
	return m.stoerungen, m.err
}

func postJSON(t *testing.T, handler http.HandlerFunc, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func TestGetTSESignaturQueueHandler_Success(t *testing.T) {
	query := &mockSignaturauftragQuery{queue: tse_repo.SignaturQueueZustand{
		OffeneAuftraege:          4,
		FehlgeschlageneAuftraege: 1,
		LetzterFehler:            "fiskaly api error 400",
		RueckstandSekunden:       125,
		SignaturenProMinute:      2.5,
		SignierdauerP95Sekunden:  3.2,
	}}
	handler := &QueryHandler{Query: query}

	rec := postJSON(t, handler.GetTSESignaturQueueHandler(), "/admin/get-tse-signatur-queue", "{}")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		OffeneAuftraege          int     `json:"offeneAuftraege"`
		FehlgeschlageneAuftraege int     `json:"fehlgeschlageneAuftraege"`
		LetzterFehler            string  `json:"letzterFehler"`
		RueckstandSekunden       int     `json:"rueckstandSekunden"`
		SignaturenProMinute      float64 `json:"signaturenProMinute"`
		SignierdauerP95Sekunden  float64 `json:"signierdauerP95Sekunden"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body.OffeneAuftraege != 4 || body.FehlgeschlageneAuftraege != 1 || body.RueckstandSekunden != 125 {
		t.Fatalf("unexpected queue counts: %+v", body)
	}
	if body.LetzterFehler != "fiskaly api error 400" {
		t.Fatalf("unexpected letzter fehler: %q", body.LetzterFehler)
	}
	if body.SignaturenProMinute != 2.5 || body.SignierdauerP95Sekunden != 3.2 {
		t.Fatalf("unexpected queue metrics: %+v", body)
	}
}

func TestGetTSEStoerungenHandler_Success(t *testing.T) {
	ende := time.Date(2026, 6, 11, 12, 0, 0, 0, time.UTC)
	query := &mockSignaturauftragQuery{stoerungen: []tse_repo.Stoerungszeitraum{
		{ID: 2, Beginn: time.Date(2026, 6, 11, 11, 0, 0, 0, time.UTC), GrundArt: "rueckstand", Fehlertext: "Rückstand"},
		{ID: 1, Beginn: time.Date(2026, 6, 11, 10, 0, 0, 0, time.UTC), Ende: &ende, GrundArt: "tse_fehler", Fehlertext: "TSE nicht erreichbar"},
	}}
	handler := &QueryHandler{Query: query}

	rec := postJSON(t, handler.GetTSEStoerungenHandler(), "/admin/get-tse-stoerungen", "{}")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Stoerungen []struct {
			ID       int     `json:"id"`
			Ende     *string `json:"ende"`
			GrundArt string  `json:"grundArt"`
		} `json:"stoerungen"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(body.Stoerungen) != 2 {
		t.Fatalf("expected 2 stoerungen, got %d", len(body.Stoerungen))
	}
	if body.Stoerungen[0].Ende != nil {
		t.Fatalf("expected active stoerung to have null ende, got %v", *body.Stoerungen[0].Ende)
	}
	if body.Stoerungen[1].Ende == nil || body.Stoerungen[1].GrundArt != "tse_fehler" {
		t.Fatalf("unexpected closed stoerung DTO: %+v", body.Stoerungen[1])
	}
}
