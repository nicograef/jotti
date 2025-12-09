//go:build unit

package product_admin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type mockCommand struct {
	err error
}

func (m *mockCommand) CreateProduct(ctx context.Context, name, description string, netPriceCents int, category Category) (int, error) {
	return 1, m.err
}

func (m *mockCommand) UpdateProduct(ctx context.Context, id int, name, description string, netPriceCents int, category Category) error {
	return m.err
}

func (m *mockCommand) ActivateProduct(ctx context.Context, id int) error {
	return m.err
}

func (m *mockCommand) DeactivateProduct(ctx context.Context, id int) error {
	return m.err
}

func TestCreateProductHandler_Success(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{}}

	body := `{"name":"French Fries","description":"The most delicious fries.","netPriceCents":1999,"category":"food"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/create-product", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.CreateProductHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestCreateProductHandler_Failure(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{err: ErrDatabase}}

	body := `{"name":"French Fries","description":"The most delicious fries.","netPriceCents":1999,"category":"food"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/create-product", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.CreateProductHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}
}

func TestUpdateProductHandler_Success(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{}}

	body := `{"id":1,"name":"French Fries","description":"The most delicious fries.","netPriceCents":1999,"category":"food"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/update-product", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.UpdateProductHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestUpdateProductHandler_Failure(t *testing.T) {
	handler := &CommandHandler{Command: &mockCommand{err: ErrDatabase}}

	body := `{"id":1,"name":"French Fries","description":"The most delicious fries.","netPriceCents":1999,"category":"food"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/update-product", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.UpdateProductHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}
}
