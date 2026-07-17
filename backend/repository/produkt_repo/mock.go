//go:build unit

package produkt_repo

import (
	"context"

	"github.com/nicograef/jotti/backend/domain/produkt"
)

// NewMock creates a new mock repository with the given products and error.
func NewMock(products []produkt.Produkt, err error) *mockRepo {
	productMap := make(map[int]produkt.Produkt)
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
	variant   produkt.Variante
	productID int
}

type mockRepo struct {
	products map[int]produkt.Produkt
	variants map[int]variantWithProduct
	err      error
}

// AddVariant adds a variant to the mock repository, associated with a produkt.
func (m *mockRepo) AddVariant(productID int, v produkt.Variante) {
	m.variants[v.ID] = variantWithProduct{variant: v, productID: productID}
}

func (m *mockRepo) GetProduct(ctx context.Context, id int) (produkt.Produkt, error) {
	t, ok := m.products[id]
	if !ok {
		return produkt.Produkt{}, m.err
	}
	return t, m.err
}

func (m *mockRepo) CreateProduct(ctx context.Context, t produkt.Produkt) (int, error) {
	newID := len(m.products) + 1
	t.ID = newID
	m.products[newID] = t
	return newID, m.err
}

func (m *mockRepo) UpdateProduct(ctx context.Context, t produkt.Produkt) error {
	m.products[t.ID] = t
	return m.err
}

func (m *mockRepo) GetVariant(ctx context.Context, variantID int) (produkt.Variante, error) {
	vp, ok := m.variants[variantID]
	if !ok {
		return produkt.Variante{}, m.err
	}
	return vp.variant, m.err
}

func (m *mockRepo) CreateVariant(ctx context.Context, productID int, v produkt.Variante) (int, error) {
	newID := len(m.variants) + 1
	v.ID = newID
	m.variants[newID] = variantWithProduct{variant: v, productID: productID}
	return newID, m.err
}

func (m *mockRepo) UpdateVariant(ctx context.Context, v produkt.Variante) error {
	if vp, ok := m.variants[v.ID]; ok {
		m.variants[v.ID] = variantWithProduct{variant: v, productID: vp.productID}
	}
	return m.err
}

func (m *mockRepo) DeleteProduktMitVarianten(ctx context.Context, p produkt.Produkt) error {
	if m.err != nil {
		return m.err
	}
	m.products[p.ID] = p
	for i := range p.Varianten {
		v := p.Varianten[i]
		if vp, ok := m.variants[v.ID]; ok {
			m.variants[v.ID] = variantWithProduct{variant: v, productID: vp.productID}
		}
	}
	return nil
}

func (m *mockRepo) GetAllProducts(ctx context.Context) ([]produkt.Produkt, error) {
	products := make([]produkt.Produkt, 0, len(m.products))
	for i := range m.products {
		products = append(products, m.products[i])
	}
	return products, m.err
}

func (m *mockRepo) GetActiveProducts(ctx context.Context) ([]produkt.Produkt, error) {
	products := make([]produkt.Produkt, 0)
	for i := range m.products {
		if m.products[i].Status == produkt.ActiveStatus {
			products = append(products, m.products[i])
		}
	}
	return products, m.err
}

func (m *mockRepo) GetVariantsByIDs(ctx context.Context, ids []int) (map[int]produkt.Variante, error) {
	if m.err != nil {
		return nil, m.err
	}
	result := make(map[int]produkt.Variante, len(ids))
	for _, id := range ids {
		if vp, ok := m.variants[id]; ok {
			result[id] = vp.variant
		}
	}
	return result, nil
}

func (m *mockRepo) GetProductsByIDs(ctx context.Context, ids []int) (map[int]produkt.Produkt, error) {
	if m.err != nil {
		return nil, m.err
	}
	result := make(map[int]produkt.Produkt, len(ids))
	for _, id := range ids {
		if p, ok := m.products[id]; ok {
			result[id] = p
		}
	}
	return result, nil
}
