//go:build unit

package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/nicograef/jotti/backend/api/kasse/application"
	"github.com/nicograef/jotti/backend/api/middleware"
)

type mockCommand struct {
	zNr int
	err error
}

func (m *mockCommand) KassensitzungEroeffnen(_ context.Context, _ int, _ string, _ time.Time, _ string) (int, error) {
	return m.zNr, m.err
}

func (m *mockCommand) AnfangsbestandSetzen(_ context.Context, _ int, _ string, _ int) error {
	return m.err
}

func (m *mockCommand) KassenbewegungBuchen(_ context.Context, _ int, _ string, _ string, _ int, _ string) error {
	return m.err
}

func (m *mockCommand) KassensturzDurchfuehren(_ context.Context, _ int, _ string, _ int) error {
	return m.err
}

func (m *mockCommand) TagesabschlussErstellen(_ context.Context, _ int, _ string) error {
	return m.err
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

	req := requestWithUser(`{"datum":"2026-01-01","bezeichnung":"Maihock"}`)
	rec := httptest.NewRecorder()

	handler.KassensitzungEroeffnenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestKassensitzungEroeffnenHandler_MissingDatum(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{}}

	req := requestWithUser(`{"datum":"","bezeichnung":"Maihock"}`)
	rec := httptest.NewRecorder()

	handler.KassensitzungEroeffnenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestKassensitzungEroeffnenHandler_KasseAlreadyOpen(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{err: application.ErrKasseAlreadyOpen}}

	req := requestWithUser(`{"datum":"2026-01-01","bezeichnung":"Maihock"}`)
	rec := httptest.NewRecorder()

	handler.KassensitzungEroeffnenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

// AnfangsbestandSetzen

func TestAnfangsbestandSetzenHandler_Success(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{}}

	req := requestWithUser(`{"betragCents":10000}`)
	rec := httptest.NewRecorder()

	handler.AnfangsbestandSetzenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestAnfangsbestandSetzenHandler_KasseNichtGeoeffnet(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{err: application.ErrKasseNichtGeoeffnet}}

	req := requestWithUser(`{"betragCents":10000}`)
	rec := httptest.NewRecorder()

	handler.AnfangsbestandSetzenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

// KassenbewegungBuchen

func TestKassenbewegungBuchenHandler_Success(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{}}

	req := requestWithUser(`{"art":"geldtransit","betragCents":500,"kommentar":""}`)
	rec := httptest.NewRecorder()

	handler.KassenbewegungBuchenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestKassenbewegungBuchenHandler_MissingArt(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{}}

	req := requestWithUser(`{"art":"","betragCents":500}`)
	rec := httptest.NewRecorder()

	handler.KassenbewegungBuchenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

// KassensturzDurchfuehren

func TestKassensturzDurchfuehrenHandler_Success(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{}}

	req := requestWithUser(`{"istBestandCents":10000}`)
	rec := httptest.NewRecorder()

	handler.KassensturzDurchfuehrenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestKassensturzDurchfuehrenHandler_KasseNichtGeoeffnet(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{err: application.ErrKasseNichtGeoeffnet}}

	req := requestWithUser(`{"istBestandCents":10000}`)
	rec := httptest.NewRecorder()

	handler.KassensturzDurchfuehrenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

// TagesabschlussErstellen

func TestTagesabschlussErstellenHandler_Success(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{}}

	req := requestWithUser(`{}`)
	rec := httptest.NewRecorder()

	handler.TagesabschlussErstellenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestTagesabschlussErstellenHandler_KassensturzErforderlich(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{err: application.ErrKassensturzErforderlich}}

	req := requestWithUser(`{}`)
	rec := httptest.NewRecorder()

	handler.TagesabschlussErstellenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}
