//go:build unit

package product

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type mockProductService struct{}

func (m *mockProductService) CreateProduct(ctx context.Context, name, description string, netPriceCents int, category Category) (*Product, error) {
	return &Product{
		ID:            1,
		Name:          name,
		Description:   description,
		NetPriceCents: netPriceCents,
		Status:        InactiveStatus,
		Category:      category,
	}, nil
}

func (m *mockProductService) UpdateProduct(ctx context.Context, id int, name, description string, netPriceCents int, category Category) (*Product, error) {
	return &Product{
		ID:            id,
		Name:          name,
		Description:   description,
		NetPriceCents: netPriceCents,
		Status:        ActiveStatus,
		Category:      category,
	}, nil
}

func (m *mockProductService) GetAllProducts(ctx context.Context) ([]*Product, error) {
	return []*Product{
		{
			ID:            1,
			Name:          "French Fries",
			Description:   "The most delicious fries.",
			NetPriceCents: 1999,
			Status:        ActiveStatus,
			Category:      FoodCategory,
		},
	}, nil
}

func (m *mockProductService) GetActiveProducts(ctx context.Context) ([]*ProductPublic, error) {
	return []*ProductPublic{
		{
			ID:            1,
			Name:          "French Fries",
			Description:   "The most delicious fries.",
			NetPriceCents: 1999,
			Category:      FoodCategory,
		},
	}, nil
}

func (m *mockProductService) ActivateProduct(ctx context.Context, id int) error {
	return nil
}

func (m *mockProductService) DeactivateProduct(ctx context.Context, id int) error {
	return nil
}

func TestCreateProductHandler_Success(t *testing.T) {
	handler := &Handler{Service: &mockProductService{}}

	body := `{"name":"French Fries","description":"The most delicious fries.","netPriceCents":1999,"category":"food"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/create-product", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.CreateProductHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}
