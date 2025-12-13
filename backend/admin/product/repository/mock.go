package repository

import (
	"context"

	"github.com/nicograef/jotti/backend/admin/product/domain"
)

// NewMock creates a new mock repository with the given products and error.
func NewMock(products []domain.Product, err error) *mockRepo {
	productMap := make(map[int]domain.Product)
	for _, t := range products {
		productMap[t.ID] = t
	}

	return &mockRepo{
		products: productMap,
		err:      err,
	}
}

type mockRepo struct {
	products map[int]domain.Product
	err      error
}

func (m *mockRepo) GetProduct(ctx context.Context, id int) (domain.Product, error) {
	t, ok := m.products[id]
	if !ok {
		return domain.Product{}, m.err
	}
	return t, m.err
}

func (m *mockRepo) CreateProduct(ctx context.Context, t domain.Product) (int, error) {
	newID := len(m.products) + 1
	t.ID = newID
	m.products[newID] = t
	return newID, m.err
}

func (m *mockRepo) UpdateProduct(ctx context.Context, t domain.Product) error {
	m.products[t.ID] = t
	return m.err
}
