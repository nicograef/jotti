package product_repo

import (
	"context"

	"github.com/nicograef/jotti/backend/domain/product"
)

// NewMock creates a new mock repository with the given products and error.
func NewMock(products []product.Product, err error) *mockRepo {
	productMap := make(map[int]product.Product)
	for _, t := range products {
		productMap[t.ID] = t
	}

	return &mockRepo{
		products: productMap,
		err:      err,
	}
}

type mockRepo struct {
	products map[int]product.Product
	err      error
}

func (m mockRepo) GetProduct(ctx context.Context, id int) (product.Product, error) {
	t, ok := m.products[id]
	if !ok {
		return product.Product{}, m.err
	}
	return t, m.err
}

func (m mockRepo) CreateProduct(ctx context.Context, t product.Product) (int, error) {
	newID := len(m.products) + 1
	t.ID = newID
	m.products[newID] = t
	return newID, m.err
}

func (m mockRepo) UpdateProduct(ctx context.Context, t product.Product) error {
	m.products[t.ID] = t
	return m.err
}
