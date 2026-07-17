//go:build unit

package produkt_repo

import (
	"context"

	"github.com/nicograef/jotti/backend/domain/produkt"
)

// NewMock creates a new mock repository with the given produkte and error.
func NewMock(produkte []produkt.Produkt, err error) *mockRepo {
	produktMap := make(map[int]produkt.Produkt)
	for i := range produkte {
		produktMap[produkte[i].ID] = produkte[i]
	}

	return &mockRepo{
		produkte:  produktMap,
		varianten: make(map[int]varianteWithProdukt),
		err:       err,
	}
}

type varianteWithProdukt struct {
	variante  produkt.Variante
	produktID int
}

type mockRepo struct {
	produkte  map[int]produkt.Produkt
	varianten map[int]varianteWithProdukt
	err       error
}

// AddVariante adds a variante to the mock repository, associated with a produkt.
func (m *mockRepo) AddVariante(produktID int, v produkt.Variante) {
	m.varianten[v.ID] = varianteWithProdukt{variante: v, produktID: produktID}
}

func (m *mockRepo) GetProdukt(ctx context.Context, id int) (produkt.Produkt, error) {
	t, ok := m.produkte[id]
	if !ok {
		return produkt.Produkt{}, m.err
	}
	return t, m.err
}

func (m *mockRepo) CreateProdukt(ctx context.Context, t produkt.Produkt) (int, error) {
	newID := len(m.produkte) + 1
	t.ID = newID
	m.produkte[newID] = t
	return newID, m.err
}

func (m *mockRepo) UpdateProdukt(ctx context.Context, t produkt.Produkt) error {
	m.produkte[t.ID] = t
	return m.err
}

func (m *mockRepo) GetVariante(ctx context.Context, varianteID int) (produkt.Variante, error) {
	vp, ok := m.varianten[varianteID]
	if !ok {
		return produkt.Variante{}, m.err
	}
	return vp.variante, m.err
}

func (m *mockRepo) CreateVariante(ctx context.Context, produktID int, v produkt.Variante) (int, error) {
	newID := len(m.varianten) + 1
	v.ID = newID
	m.varianten[newID] = varianteWithProdukt{variante: v, produktID: produktID}
	return newID, m.err
}

func (m *mockRepo) UpdateVariante(ctx context.Context, v produkt.Variante) error {
	if vp, ok := m.varianten[v.ID]; ok {
		m.varianten[v.ID] = varianteWithProdukt{variante: v, produktID: vp.produktID}
	}
	return m.err
}

func (m *mockRepo) DeleteProduktMitVarianten(ctx context.Context, p produkt.Produkt) error {
	if m.err != nil {
		return m.err
	}
	m.produkte[p.ID] = p
	for i := range p.Varianten {
		v := p.Varianten[i]
		if vp, ok := m.varianten[v.ID]; ok {
			m.varianten[v.ID] = varianteWithProdukt{variante: v, produktID: vp.produktID}
		}
	}
	return nil
}

func (m *mockRepo) GetAllProdukte(ctx context.Context) ([]produkt.Produkt, error) {
	produkte := make([]produkt.Produkt, 0, len(m.produkte))
	for i := range m.produkte {
		produkte = append(produkte, m.produkte[i])
	}
	return produkte, m.err
}

func (m *mockRepo) GetActiveProdukte(ctx context.Context) ([]produkt.Produkt, error) {
	produkte := make([]produkt.Produkt, 0)
	for i := range m.produkte {
		if m.produkte[i].Status == produkt.ActiveStatus {
			produkte = append(produkte, m.produkte[i])
		}
	}
	return produkte, m.err
}

func (m *mockRepo) GetVariantenByIDs(ctx context.Context, ids []int) (map[int]produkt.Variante, error) {
	if m.err != nil {
		return nil, m.err
	}
	result := make(map[int]produkt.Variante, len(ids))
	for _, id := range ids {
		if vp, ok := m.varianten[id]; ok {
			result[id] = vp.variante
		}
	}
	return result, nil
}

func (m *mockRepo) GetProdukteByIDs(ctx context.Context, ids []int) (map[int]produkt.Produkt, error) {
	if m.err != nil {
		return nil, m.err
	}
	result := make(map[int]produkt.Produkt, len(ids))
	for _, id := range ids {
		if p, ok := m.produkte[id]; ok {
			result[id] = p
		}
	}
	return result, nil
}
