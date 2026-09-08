//go:build unit

package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nicograef/jotti/backend/api/stammdaten/betreiber/application"
	"github.com/nicograef/jotti/backend/domain/betreiber"
)

type mockCommand struct {
	err                 error
	setzenCalled        bool
	zuruecknehmenCalled bool
}

func (m *mockCommand) UpdateBetreiber(_ context.Context, _ betreiber.Betreiber) error {
	return m.err
}

func (m *mockCommand) SetzeElsterMeldung(_ context.Context) error {
	m.setzenCalled = true
	return m.err
}

func (m *mockCommand) NimmElsterMeldungZurueck(_ context.Context) error {
	m.zuruecknehmenCalled = true
	return m.err
}

func TestSetzeElsterMeldungHandler_Success(t *testing.T) {
	cmd := &mockCommand{}
	handler := &CommandHandler{Command: cmd}

	req := httptest.NewRequest(http.MethodPost, "/elster-meldung-setzen", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.SetzeElsterMeldungHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	if !cmd.setzenCalled {
		t.Error("expected SetzeElsterMeldung to be called")
	}
}

func TestSetzeElsterMeldungHandler_Failure(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{err: application.ErrDatabase}}

	req := httptest.NewRequest(http.MethodPost, "/elster-meldung-setzen", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.SetzeElsterMeldungHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}
}

func TestNimmElsterMeldungZurueckHandler_Success(t *testing.T) {
	cmd := &mockCommand{}
	handler := &CommandHandler{Command: cmd}

	req := httptest.NewRequest(http.MethodPost, "/elster-meldung-zuruecknehmen", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.NimmElsterMeldungZurueckHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	if !cmd.zuruecknehmenCalled {
		t.Error("expected NimmElsterMeldungZurueck to be called")
	}
}

func TestNimmElsterMeldungZurueckHandler_Failure(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{err: application.ErrDatabase}}

	req := httptest.NewRequest(http.MethodPost, "/elster-meldung-zuruecknehmen", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.NimmElsterMeldungZurueckHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}
}

func TestUpdateBetreiberHandler_Success(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{}}
	body := `{"vereinsname":"Sportverein","strasse":"Musterstraße 1","plz":"12345","ort":"Musterstadt","steuernummer":null,"ustId":null}`

	req := httptest.NewRequest(http.MethodPost, "/update-betreiber", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.UpdateBetreiberHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

// Der Handler prüft die amtlichen Maximallängen aus domain/betreiber: Ein zu
// langer Vereinsname ist eine ungültige Eingabe und damit 400, kein Serverfehler.
func TestUpdateBetreiberHandler_ZuLangerVereinsname(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{}}
	body := `{"vereinsname":"` + strings.Repeat("a", 61) + `","strasse":"Musterstraße 1","plz":"12345","ort":"Musterstadt","steuernummer":null,"ustId":null}`

	req := httptest.NewRequest(http.MethodPost, "/update-betreiber", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.UpdateBetreiberHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}
