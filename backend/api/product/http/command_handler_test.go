//go:build unit

package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nicograef/jotti/backend/api/product/application"
	"github.com/nicograef/jotti/backend/domain/product"
	"github.com/nicograef/jotti/backend/domain/steuer"
)

type mockCommand struct {
	err error
}

func (m *mockCommand) CreateProduct(ctx context.Context, name string, kategorie product.Kategorie, steuersatz steuer.Steuersatz) (int, error) {
	return 1, m.err
}

func (m *mockCommand) UpdateProduct(ctx context.Context, id int, name string, kategorie product.Kategorie, steuersatz steuer.Steuersatz) error {
	return m.err
}

func (m *mockCommand) CreateVariant(ctx context.Context, productID int, name string, preisCents int) (int, error) {
	return 1, m.err
}

func (m *mockCommand) UpdateVariant(ctx context.Context, variantID int, name string, preisCents int) error {
	return m.err
}

func (m *mockCommand) ActivateVariant(ctx context.Context, variantID int) error {
	return m.err
}

func (m *mockCommand) DeactivateVariant(ctx context.Context, variantID int) error {
	return m.err
}

func (m *mockCommand) DeleteProdukt(ctx context.Context, productID int) error {
	return m.err
}

func (m *mockCommand) DeleteVariante(ctx context.Context, produktID int, variantID int) error {
	return m.err
}

func TestCreateProductHandler_Success(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{}}

	body := `{"name":"French Fries","kategorie":"essen","steuersatz":"ermaessigt"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/create-produkt", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.CreateProductHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestCreateProductHandler_Failure(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{err: application.ErrDatabase}}

	body := `{"name":"French Fries","kategorie":"essen","steuersatz":"ermaessigt"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/create-produkt", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.CreateProductHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}
}

func TestUpdateProductHandler_Success(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{}}

	body := `{"id":1,"name":"French Fries","kategorie":"essen","steuersatz":"ermaessigt"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/update-produkt", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.UpdateProductHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestUpdateProductHandler_Failure(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{err: application.ErrDatabase}}

	body := `{"id":1,"name":"French Fries","kategorie":"essen","steuersatz":"ermaessigt"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/update-produkt", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.UpdateProductHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}
}

func TestCreateVariantHandler_Success(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{}}

	body := `{"produktId":1,"name":"Large","preisCents":999}`
	req := httptest.NewRequest(http.MethodPost, "/admin/create-variante", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.CreateVariantHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestCreateVariantHandler_Failure(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{err: application.ErrDatabase}}

	body := `{"produktId":1,"name":"Large","preisCents":999}`
	req := httptest.NewRequest(http.MethodPost, "/admin/create-variante", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.CreateVariantHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}
}

func TestCreateVariantHandler_InvalidInput(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{}}

	body := `{"produktId":0,"name":"ab","preisCents":-1}`
	req := httptest.NewRequest(http.MethodPost, "/admin/create-variante", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.CreateVariantHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestUpdateVariantHandler_InvalidInput(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{}}

	body := `{"id":0,"name":"ab","preisCents":-1}`
	req := httptest.NewRequest(http.MethodPost, "/admin/update-variante", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.UpdateVariantHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestActivateVariantHandler_Success(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{}}

	body := `{"id":1}`
	req := httptest.NewRequest(http.MethodPost, "/admin/activate-variante", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ActivateVariantHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestDeactivateVariantHandler_Success(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{}}

	body := `{"id":1}`
	req := httptest.NewRequest(http.MethodPost, "/admin/deactivate-variante", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.DeactivateVariantHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}
