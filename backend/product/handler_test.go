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

func (m *mockProductService) CreateProduct(ctx context.Context, name, description string, netPrice float64, category Category) (*Product, error) {
	return &Product{
		ID:          1,
		Name:        name,
		Description: description,
		NetPrice:    netPrice,
		Status:      InactiveStatus,
		Category:    category,
	}, nil
}

func (m *mockProductService) UpdateProduct(ctx context.Context, id int, name, description string, netPrice float64, category Category) (*Product, error) {
	return &Product{
		ID:          id,
		Name:        name,
		Description: description,
		NetPrice:    netPrice,
		Status:      ActiveStatus,
		Category:    category,
	}, nil
}

func (m *mockProductService) GetAllProducts(ctx context.Context) ([]*Product, error) {
	return []*Product{
		{
			ID:          1,
			Name:        "French Fries",
			Description: "The most delicious fries.",
			NetPrice:    19.99,
			Status:      ActiveStatus,
			Category:    FoodCategory,
		},
	}, nil
}

func (m *mockProductService) GetActiveProducts(ctx context.Context) ([]*ProductPublic, error) {
	return []*ProductPublic{
		{
			ID:          1,
			Name:        "French Fries",
			Description: "The most delicious fries.",
			NetPrice:    19.99,
			Category:    FoodCategory,
		},
	}, nil
}

func (m *mockProductService) ActivateProduct(ctx context.Context, id int) error {
	return nil
}

func (m *mockProductService) DeactivateProduct(ctx context.Context, id int) error {
	return nil
}

func TestCreateProductHandler_MethodNotAllowed(t *testing.T) {
	handler := &Handler{Service: &mockProductService{}}

	req := httptest.NewRequest(http.MethodGet, "/create-product", nil)
	rec := httptest.NewRecorder()

	handler.CreateProductHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", rec.Code)
	}
}

func TestCreateProductHandler_Success(t *testing.T) {
	handler := &Handler{Service: &mockProductService{}}

	body := `{"name":"French Fries","description":"The most delicious fries.","netPrice":19.99,"category":"food"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/create-product", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.CreateProductHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}
