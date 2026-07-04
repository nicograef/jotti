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

	"github.com/nicograef/jotti/backend/api/middleware"
	"github.com/nicograef/jotti/backend/repository/tse_repo"
)

type mockSignaturauftragQuery struct {
	auftraege  []tse_repo.Signaturauftrag
	queue      tse_repo.SignaturQueueZustand
	stoerungen []tse_repo.Stoerungszeitraum
	err        error
}

func (m *mockSignaturauftragQuery) GetTSESignaturauftraege(context.Context) ([]tse_repo.Signaturauftrag, error) {
	return m.auftraege, m.err
}

func (m *mockSignaturauftragQuery) GetTSESignaturQueueZustand(context.Context) (tse_repo.SignaturQueueZustand, error) {
	return m.queue, m.err
}

func (m *mockSignaturauftragQuery) GetTSEStoerungen(context.Context) ([]tse_repo.Stoerungszeitraum, error) {
	return m.stoerungen, m.err
}

type mockSignaturauftragCommand struct {
	zurueckgesetzt int
	gesamtCalls    int
	gesamtAnzahl   int
	verworfenID    int
	verworfenGrund string
	verworfenVon   string
	err            error
}

func (m *mockSignaturauftragCommand) TSESignaturauftragZuruecksetzen(_ context.Context, id int) error {
	m.zurueckgesetzt = id
	return m.err
}

func (m *mockSignaturauftragCommand) TSESignaturauftraegeZuruecksetzenGesamt(context.Context) (int, error) {
	m.gesamtCalls++
	return m.gesamtAnzahl, m.err
}

func (m *mockSignaturauftragCommand) TSESignaturauftragVerwerfen(_ context.Context, id int, grund string, benutzer string) error {
	m.verworfenID = id
	m.verworfenGrund = grund
	m.verworfenVon = benutzer
	return m.err
}

func postJSON(t *testing.T, handler http.HandlerFunc, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

// postJSONAlsBenutzer stellt den authentifizierten Benutzer im Request-Context
// bereit, wie es die JWT-Middleware im Betrieb tut.
func postJSONAlsBenutzer(t *testing.T, handler http.HandlerFunc, path, body, benutzer string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, 7)
	ctx = context.WithValue(ctx, middleware.UserNameKey, benutzer)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req.WithContext(ctx))
	return rec
}

func TestGetTSESignaturauftraegeHandler_Success(t *testing.T) {
	erledigtAm := time.Date(2026, 6, 11, 12, 30, 0, 0, time.UTC)
	verworfenAm := time.Date(2026, 6, 11, 13, 0, 0, 0, time.UTC)
	query := &mockSignaturauftragQuery{auftraege: []tse_repo.Signaturauftrag{
		{
			ID:             3,
			TxID:           "8c0f9c4e-3a52-4f5d-9e6b-2d1c7a8b4f01",
			ProcessType:    "Kassenbeleg-V1",
			Status:         "verworfen",
			Versuche:       3,
			LetzterFehler:  "fiskaly timeout",
			ErstelltAm:     time.Date(2026, 6, 11, 10, 0, 0, 0, time.UTC),
			VerworfenGrund: "TSE dauerhaft ausgefallen",
			VerworfenVon:   "admin",
			VerworfenAm:    &verworfenAm,
		},
		{
			ID:          2,
			TxID:        "11111111-2222-4333-8444-555555555555",
			ProcessType: "Bestellung-V1",
			Status:      "erledigt",
			ErstelltAm:  time.Date(2026, 6, 11, 9, 0, 0, 0, time.UTC),
			ErledigtAm:  &erledigtAm,
		},
	}}
	handler := &QueryHandler{Query: query}

	rec := postJSON(t, handler.GetTSESignaturauftraegeHandler(), "/admin/get-tse-signaturauftraege", "{}")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Auftraege []struct {
			ID             int     `json:"id"`
			Status         string  `json:"status"`
			Versuche       int     `json:"versuche"`
			LetzterFehler  string  `json:"letzterFehler"`
			ErledigtAm     *string `json:"erledigtAm"`
			VerworfenGrund string  `json:"verworfenGrund"`
			VerworfenVon   string  `json:"verworfenVon"`
			VerworfenAm    *string `json:"verworfenAm"`
		} `json:"auftraege"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp.Auftraege) != 2 {
		t.Fatalf("expected 2 auftraege, got %d", len(resp.Auftraege))
	}
	a := resp.Auftraege[0]
	if a.ID != 3 || a.Status != "verworfen" || a.Versuche != 3 || a.LetzterFehler != "fiskaly timeout" {
		t.Fatalf("unexpected auftrag DTO: %+v", a)
	}
	if a.VerworfenGrund != "TSE dauerhaft ausgefallen" || a.VerworfenVon != "admin" || a.VerworfenAm == nil {
		t.Fatalf("expected verwerfen protocol in DTO, got %+v", a)
	}
	if a.ErledigtAm != nil {
		t.Fatalf("expected null erledigtAm for verworfenen auftrag, got %v", *a.ErledigtAm)
	}
	if resp.Auftraege[1].ErledigtAm == nil {
		t.Fatal("expected erledigtAm for erledigten auftrag")
	}
}

func TestGetTSESignaturauftraegeHandler_EmptyList(t *testing.T) {
	handler := &QueryHandler{Query: &mockSignaturauftragQuery{}}

	rec := postJSON(t, handler.GetTSESignaturauftraegeHandler(), "/admin/get-tse-signaturauftraege", "{}")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"auftraege":[]`) {
		t.Fatalf("expected empty array, got %s", rec.Body.String())
	}
}

func TestGetTSESignaturQueueHandler_Success(t *testing.T) {
	query := &mockSignaturauftragQuery{queue: tse_repo.SignaturQueueZustand{
		OffeneAuftraege:          4,
		FehlgeschlageneAuftraege: 1,
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

func TestTSESignaturauftragZuruecksetzenHandler_Success(t *testing.T) {
	cmd := &mockSignaturauftragCommand{}
	handler := &CommandHandler{Command: cmd}

	rec := postJSON(t, handler.TSESignaturauftragZuruecksetzenHandler(), "/admin/tse-signaturauftrag-zuruecksetzen", `{"id":42}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if cmd.zurueckgesetzt != 42 {
		t.Fatalf("expected command called with id 42, got %d", cmd.zurueckgesetzt)
	}
}

func TestTSESignaturauftraegeZuruecksetzenHandler_Success(t *testing.T) {
	cmd := &mockSignaturauftragCommand{gesamtAnzahl: 5}
	handler := &CommandHandler{Command: cmd}

	rec := postJSON(t, handler.TSESignaturauftraegeZuruecksetzenHandler(), "/admin/tse-signaturauftraege-zuruecksetzen", `{}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if cmd.gesamtCalls != 1 {
		t.Fatalf("expected gesamt command called once, got %d", cmd.gesamtCalls)
	}
	if !strings.Contains(rec.Body.String(), `"anzahl":5`) {
		t.Fatalf("expected anzahl 5 in response, got %s", rec.Body.String())
	}
}

func TestTSESignaturauftragVerwerfenHandler_Success(t *testing.T) {
	cmd := &mockSignaturauftragCommand{}
	handler := &CommandHandler{Command: cmd}

	rec := postJSONAlsBenutzer(t, handler.TSESignaturauftragVerwerfenHandler(), "/admin/tse-signaturauftrag-verwerfen", `{"id":42,"grund":"TSE dauerhaft defekt"}`, "admin")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if cmd.verworfenID != 42 || cmd.verworfenGrund != "TSE dauerhaft defekt" || cmd.verworfenVon != "admin" {
		t.Fatalf("unexpected verwerfen call: id=%d grund=%q von=%q", cmd.verworfenID, cmd.verworfenGrund, cmd.verworfenVon)
	}
}

func TestTSESignaturauftragVerwerfenHandler_RequiresGrund(t *testing.T) {
	cmd := &mockSignaturauftragCommand{}
	handler := &CommandHandler{Command: cmd}

	for _, body := range []string{`{"id":42}`, `{"id":42,"grund":"   "}`} {
		rec := postJSONAlsBenutzer(t, handler.TSESignaturauftragVerwerfenHandler(), "/admin/tse-signaturauftrag-verwerfen", body, "admin")
		if rec.Code != http.StatusBadRequest {
			t.Errorf("body %s: expected status 400, got %d", body, rec.Code)
		}
	}
	if cmd.verworfenID != 0 {
		t.Fatalf("expected command not called without grund, got id %d", cmd.verworfenID)
	}
}

func TestTSESignaturauftragVerwerfenHandler_RequiresUser(t *testing.T) {
	cmd := &mockSignaturauftragCommand{}
	handler := &CommandHandler{Command: cmd}

	// Ohne authentifizierten Benutzer im Context (keine JWT-Middleware).
	rec := postJSON(t, handler.TSESignaturauftragVerwerfenHandler(), "/admin/tse-signaturauftrag-verwerfen", `{"id":42,"grund":"defekt"}`)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500 without user, got %d", rec.Code)
	}
	if cmd.verworfenID != 0 {
		t.Fatalf("expected command not called without user, got id %d", cmd.verworfenID)
	}
}

func TestTSESignaturauftragCommandHandlers_RejectInvalidID(t *testing.T) {
	cmd := &mockSignaturauftragCommand{}
	zuruecksetzen := (&CommandHandler{Command: cmd}).TSESignaturauftragZuruecksetzenHandler()
	verwerfen := (&CommandHandler{Command: cmd}).TSESignaturauftragVerwerfenHandler()

	for _, body := range []string{`{"id":0}`, `{}`} {
		rec := postJSON(t, zuruecksetzen, "/admin/tse-signaturauftrag-zuruecksetzen", body)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("zuruecksetzen body %s: expected status 400, got %d", body, rec.Code)
		}
		rec = postJSONAlsBenutzer(t, verwerfen, "/admin/tse-signaturauftrag-verwerfen", `{"id":0,"grund":"x"}`, "admin")
		if rec.Code != http.StatusBadRequest {
			t.Errorf("verwerfen body %s: expected status 400, got %d", body, rec.Code)
		}
	}
	if cmd.zurueckgesetzt != 0 || cmd.verworfenID != 0 {
		t.Fatalf("expected commands not called for invalid IDs, got zurueckgesetzt=%d verworfen=%d", cmd.zurueckgesetzt, cmd.verworfenID)
	}
}
