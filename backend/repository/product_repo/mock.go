//go:build unit

package product_repo

import (
	"context"

	"github.com/nicograef/jotti/backend/domain/product"
)

// NewMock creates a new mock repository with the given products and error.
func NewMock(products []product.Produkt, err error) *mockRepo {
	productMap := make(map[int]product.Produkt)
	for i := range products {
		productMap[products[i].ID] = products[i]
	}

	return &mockRepo{
		products: productMap,
		variants: make(map[int]variantWithProduct),
		err:      err,
	}
}

type variantWithProduct struct {
	variant   product.Variante
	productID int
}

type mockRepo struct {
	products map[int]product.Produkt
	variants map[int]variantWithProduct
	err      error
}

// AddVariant adds a variant to the mock repository, associated with a product.
func (m *mockRepo) AddVariant(productID int, v product.Variante) {
	m.variants[v.ID] = variantWithProduct{variant: v, productID: productID}
}

func (m *mockRepo) GetProduct(ctx context.Context, id int) (product.Produkt, error) {
	t, ok := m.products[id]
	if !ok {
		return product.Produkt{}, m.err
	}
	return t, m.err
}

func (m *mockRepo) CreateProduct(ctx context.Context, t product.Produkt) (int, error) {
	newID := len(m.products) + 1
	t.ID = newID
	m.products[newID] = t
	return newID, m.err
}

func (m *mockRepo) UpdateProduct(ctx context.Context, t product.Produkt) error {
	m.products[t.ID] = t
	return m.err
}

func (m *mockRepo) GetVariant(ctx context.Context, variantID int) (product.Variante, error) {
	vp, ok := m.variants[variantID]
	if !ok {
		return product.Variante{}, m.err
	}
	return vp.variant, m.err
}

func (m *mockRepo) CreateVariant(ctx context.Context, productID int, v product.Variante) (int, error) {
	newID := len(m.variants) + 1
	v.ID = newID
	m.variants[newID] = variantWithProduct{variant: v, productID: productID}
	return newID, m.err
}

func (m *mockRepo) UpdateVariant(ctx context.Context, v product.Variante) error {
	if vp, ok := m.variants[v.ID]; ok {
		m.variants[v.ID] = variantWithProduct{variant: v, productID: vp.productID}
	}
	return m.err
}

func (m *mockRepo) GetAllProducts(ctx context.Context) ([]product.Produkt, error) {
	products := make([]product.Produkt, 0, len(m.products))
	for i := range m.products {
		products = append(products, m.products[i])
	}
	return products, m.err
}

func (m *mockRepo) GetActiveProducts(ctx context.Context) ([]product.Produkt, error) {
	products := make([]product.Produkt, 0)
	for i := range m.products {
		if m.products[i].Status == product.ActiveStatus {
			products = append(products, m.products[i])
		}
	}
	return products, m.err
}

func (m *mockRepo) GetVariantsByIDs(ctx context.Context, ids []int) (map[int]product.Variante, error) {
	if m.err != nil {
		return nil, m.err
	}
	result := make(map[int]product.Variante, len(ids))
	for _, id := range ids {
		if vp, ok := m.variants[id]; ok {
			result[id] = vp.variant
		}
	}
	return result, nil
}

func (m *mockRepo) GetProductsByIDs(ctx context.Context, ids []int) (map[int]product.Produkt, error) {
	if m.err != nil {
		return nil, m.err
	}
	result := make(map[int]product.Produkt, len(ids))
	for _, id := range ids {
		if p, ok := m.products[id]; ok {
			result[id] = p
		}
	}
	return result, nil
}
