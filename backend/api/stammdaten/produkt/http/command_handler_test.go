//go:build unit

package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nicograef/jotti/backend/api/stammdaten/produkt/application"
	dom "github.com/nicograef/jotti/backend/domain/produkt"
	"github.com/nicograef/jotti/backend/domain/steuer"
)

type mockCommand struct {
	err error
}

func (m *mockCommand) CreateProdukt(ctx context.Context, name string, kategorie dom.Kategorie, steuersatz steuer.Steuersatz) (int, error) {
	return 1, m.err
}

func (m *mockCommand) UpdateProdukt(ctx context.Context, id int, name string, kategorie dom.Kategorie, steuersatz steuer.Steuersatz) error {
	return m.err
}

func (m *mockCommand) CreateVariante(ctx context.Context, produktID int, name string, preisCents int) (int, error) {
	return 1, m.err
}

func (m *mockCommand) UpdateVariante(ctx context.Context, varianteID int, name string, preisCents int) error {
	return m.err
}

func (m *mockCommand) ActivateVariante(ctx context.Context, varianteID int) error {
	return m.err
}

func (m *mockCommand) DeactivateVariante(ctx context.Context, varianteID int) error {
	return m.err
}

func (m *mockCommand) DeleteProdukt(ctx context.Context, produktID int) error {
	return m.err
}

func (m *mockCommand) DeleteVariante(ctx context.Context, produktID int, varianteID int) error {
	return m.err
}

func TestCreateProduktHandler_Success(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{}}

	body := `{"name":"French Fries","kategorie":"essen","steuersatz":"ermaessigt"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/create-produkt", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.CreateProduktHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestCreateProduktHandler_Failure(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{err: application.ErrDatabase}}

	body := `{"name":"French Fries","kategorie":"essen","steuersatz":"ermaessigt"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/create-produkt", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.CreateProduktHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}
}

func TestUpdateProduktHandler_Success(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{}}

	body := `{"id":1,"name":"French Fries","kategorie":"essen","steuersatz":"ermaessigt"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/update-produkt", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.UpdateProduktHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestUpdateProduktHandler_Failure(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{err: application.ErrDatabase}}

	body := `{"id":1,"name":"French Fries","kategorie":"essen","steuersatz":"ermaessigt"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/update-produkt", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.UpdateProduktHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}
}

func TestDeleteProduktHandler_Success(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{}}

	body := `{"id":1}`
	req := httptest.NewRequest(http.MethodPost, "/admin/delete-produkt", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.DeleteProduktHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestCreateVarianteHandler_Success(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{}}

	body := `{"produktId":1,"name":"Large","preisCents":999}`
	req := httptest.NewRequest(http.MethodPost, "/admin/create-variante", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.CreateVarianteHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestCreateVarianteHandler_Failure(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{err: application.ErrDatabase}}

	body := `{"produktId":1,"name":"Large","preisCents":999}`
	req := httptest.NewRequest(http.MethodPost, "/admin/create-variante", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.CreateVarianteHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}
}

func TestCreateVarianteHandler_InvalidInput(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{}}

	body := `{"produktId":0,"name":"ab","preisCents":-1}`
	req := httptest.NewRequest(http.MethodPost, "/admin/create-variante", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.CreateVarianteHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestUpdateVarianteHandler_InvalidInput(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{}}

	body := `{"id":0,"name":"ab","preisCents":-1}`
	req := httptest.NewRequest(http.MethodPost, "/admin/update-variante", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.UpdateVarianteHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestActivateVarianteHandler_Success(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{}}

	body := `{"id":1}`
	req := httptest.NewRequest(http.MethodPost, "/admin/activate-variante", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ActivateVarianteHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestDeactivateVarianteHandler_Success(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{}}

	body := `{"id":1}`
	req := httptest.NewRequest(http.MethodPost, "/admin/deactivate-variante", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.DeactivateVarianteHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}
