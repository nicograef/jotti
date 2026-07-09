//go:build unit

package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/nicograef/jotti/backend/api/kasse/kassenfuehrung/application"
	"github.com/nicograef/jotti/backend/api/middleware"
)

type mockCommand struct {
	zNr      int
	err      error
	ergebnis application.KassenabschlussErgebnis
}

func (m *mockCommand) KassensitzungEroeffnen(_ context.Context, _ int, _ string, _ string, _ int) (int, error) {
	return m.zNr, m.err
}

func (m *mockCommand) GeldtransitBuchen(_ context.Context, _ int, _ string, _ string, _ string, _ int, _ string) error {
	return m.err
}

func (m *mockCommand) KasseAbschliessen(_ context.Context, _ int, _ string, _ int) (application.KassenabschlussErgebnis, error) {
	return m.ergebnis, m.err
}

func requestWithUser(body string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, 1)
	ctx = context.WithValue(ctx, middleware.UserNameKey, "Tester")
	return req.WithContext(ctx)
}

// KassensitzungEroeffnen

func TestKassensitzungEroeffnenHandler_Success(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{zNr: 1}}

	req := requestWithUser(`{"bezeichnung":"Maihock","betragCents":10000}`)
	rec := httptest.NewRecorder()

	handler.KassensitzungEroeffnenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestKassensitzungEroeffnenHandler_KasseAlreadyOpen(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{err: application.ErrKasseAlreadyOpen}}

	req := requestWithUser(`{"bezeichnung":"Maihock","betragCents":10000}`)
	rec := httptest.NewRecorder()

	handler.KassensitzungEroeffnenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestKassensitzungEroeffnenHandler_BetreiberNichtKonfiguriert(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{err: application.ErrBetreiberNichtKonfiguriert}}

	req := requestWithUser(`{"bezeichnung":"Maihock","betragCents":10000}`)
	rec := httptest.NewRecorder()

	handler.KassensitzungEroeffnenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "betreiber_nicht_konfiguriert") {
		t.Errorf("expected code betreiber_nicht_konfiguriert in body, got %s", rec.Body.String())
	}
}

func TestKassensitzungEroeffnenHandler_NullBetrag(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{zNr: 1}}

	req := requestWithUser(`{"bezeichnung":"Maihock","betragCents":0}`)
	rec := httptest.NewRecorder()

	handler.KassensitzungEroeffnenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestKassensitzungEroeffnenHandler_MissingBetrag(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{zNr: 1}}

	req := requestWithUser(`{"bezeichnung":"Maihock"}`)
	rec := httptest.NewRecorder()

	handler.KassensitzungEroeffnenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Anfangsbestand ist erforderlich") {
		t.Errorf("expected field error message in body, got %s", rec.Body.String())
	}
}

func TestKassensitzungEroeffnenHandler_NegativerBetrag(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{zNr: 1}}

	req := requestWithUser(`{"bezeichnung":"Maihock","betragCents":-1}`)
	rec := httptest.NewRecorder()

	handler.KassensitzungEroeffnenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Anfangsbestand darf nicht negativ sein") {
		t.Errorf("expected field error message in body, got %s", rec.Body.String())
	}
}

// GeldtransitBuchen

func TestGeldtransitBuchenHandler_Success(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{}}

	req := requestWithUser(`{"geldtransitId":"6f9619ff-8b86-d011-b42d-00cf4fc964ff","richtung":"einlage","betragCents":500,"kommentar":"Initialbestand"}`)
	rec := httptest.NewRecorder()

	handler.GeldtransitBuchenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestGeldtransitBuchenHandler_MissingRichtung(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{}}

	req := requestWithUser(`{"richtung":"","betragCents":500}`)
	rec := httptest.NewRecorder()

	handler.GeldtransitBuchenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestGeldtransitBuchenHandler_InvalidKommentar(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{}}

	req := requestWithUser(`{"richtung":"einlage","betragCents":500,"kommentar":"ab"}`)
	rec := httptest.NewRecorder()

	handler.GeldtransitBuchenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestGeldtransitBuchenHandler_NullBetrag(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{}}

	req := requestWithUser(`{"richtung":"einlage","betragCents":0,"kommentar":"Initialbestand"}`)
	rec := httptest.NewRecorder()

	handler.GeldtransitBuchenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestGeldtransitBuchenHandler_UngueltigeGeldtransitId_ValidationError(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{}}

	req := requestWithUser(`{"geldtransitId":"nicht-eine-uuid","richtung":"einlage","betragCents":500,"kommentar":"Test"}`)
	rec := httptest.NewRecorder()

	handler.GeldtransitBuchenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for non-UUID geldtransitId, got %d", rec.Code)
	}
}

// KasseAbschliessen

func TestKasseAbschliessenHandler_Success(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{}}

	req := requestWithUser(`{"istBestandCents":10000}`)
	rec := httptest.NewRecorder()

	handler.KasseAbschliessenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestKasseAbschliessenHandler_NullBestand(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{}}

	req := requestWithUser(`{"istBestandCents":0}`)
	rec := httptest.NewRecorder()

	handler.KasseAbschliessenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestKasseAbschliessenHandler_MissingBestand(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{}}

	req := requestWithUser(`{}`)
	rec := httptest.NewRecorder()

	handler.KasseAbschliessenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Ist-Bestand ist erforderlich") {
		t.Errorf("expected field error message in body, got %s", rec.Body.String())
	}
}

func TestKasseAbschliessenHandler_NegativerBestand(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{}}

	req := requestWithUser(`{"istBestandCents":-1}`)
	rec := httptest.NewRecorder()

	handler.KasseAbschliessenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Ist-Bestand darf nicht negativ sein") {
		t.Errorf("expected field error message in body, got %s", rec.Body.String())
	}
}

func TestKasseAbschliessenHandler_KasseNichtGeoeffnet(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{err: application.ErrKasseNichtGeoeffnet}}

	req := requestWithUser(`{"istBestandCents":10000}`)
	rec := httptest.NewRecorder()

	handler.KasseAbschliessenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", rec.Code)
	}
}

func TestKasseAbschliessenHandler_SignaturenAusstehend(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{err: &application.SignaturenAusstehendError{
		Anzahl:              2,
		AeltesterErstelltAm: time.Now().Add(-30 * time.Second),
	}}}

	req := requestWithUser(`{"istBestandCents":10000}`)
	rec := httptest.NewRecorder()

	handler.KasseAbschliessenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "signaturen_ausstehend") {
		t.Errorf("expected code signaturen_ausstehend in body, got %s", body)
	}
	if !strings.Contains(body, `"anzahl":2`) {
		t.Errorf("expected anzahl detail in body, got %s", body)
	}
	if !strings.Contains(body, `"alterSekunden"`) {
		t.Errorf("expected alterSekunden detail in body, got %s", body)
	}
}

func TestKasseAbschliessenHandler_AusfallResteImBody(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{ergebnis: application.KassenabschlussErgebnis{
		AusfallResteAnzahl:      3,
		OhneKonfigurationAnzahl: 1,
	}}}

	req := requestWithUser(`{"istBestandCents":10000}`)
	rec := httptest.NewRecorder()

	handler.KasseAbschliessenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"ausfallResteAnzahl":3`) {
		t.Errorf("expected ausfallResteAnzahl in body, got %s", body)
	}
	if !strings.Contains(body, `"ohneKonfigurationAnzahl":1`) {
		t.Errorf("expected ohneKonfigurationAnzahl in body, got %s", body)
	}
}

func TestKasseAbschliessenHandler_TischeSaldoOffen(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{err: application.ErrTischeSaldoOffen}}

	req := requestWithUser(`{"istBestandCents":10000}`)
	rec := httptest.NewRecorder()

	handler.KasseAbschliessenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "tische_saldo_offen") {
		t.Errorf("expected code tische_saldo_offen in body, got %s", rec.Body.String())
	}
}

func TestKasseAbschliessenHandler_BuchungenNachKassensturz(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{err: application.ErrBuchungenNachKassensturz}}

	req := requestWithUser(`{"istBestandCents":10000}`)
	rec := httptest.NewRecorder()

	handler.KasseAbschliessenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "buchungen_nach_kassensturz") {
		t.Errorf("expected code buchungen_nach_kassensturz in body, got %s", rec.Body.String())
	}
}
