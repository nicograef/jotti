//go:build unit

package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nicograef/jotti/backend/api/product/application"
	"github.com/nicograef/jotti/backend/domain/product"
)

type mockQuery struct {
	err error
}

func (m *mockQuery) GetAllProducts(ctx context.Context) ([]product.Product, error) {
	return []product.Product{{ID: 1, Name: "French Fries", Description: "The most delicious fries.", NetPriceCents: 1999, Status: product.ActiveStatus, Category: product.FoodCategory}}, m.err
}

func (m *mockQuery) GetActiveProducts(ctx context.Context) ([]product.Product, error) {
	return []product.Product{{ID: 1, Name: "French Fries", Description: "The most delicious fries.", NetPriceCents: 1999, Status: product.ActiveStatus, Category: product.FoodCategory}}, m.err
}

func TestGetAllProductsHandler_Success(t *testing.T) {
	handler := &QueryHandler{Query: &mockQuery{}}

	req := httptest.NewRequest(http.MethodGet, "/get-all-products", nil)
	rec := httptest.NewRecorder()

	handler.GetAllProductsHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

}

func TestGetAllProductsHandler_Failure(t *testing.T) {
	handler := &QueryHandler{Query: &mockQuery{err: application.ErrDatabase}}

	req := httptest.NewRequest(http.MethodGet, "/get-all-products", nil)
	rec := httptest.NewRecorder()

	handler.GetAllProductsHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}
}
