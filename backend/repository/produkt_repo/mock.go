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

// Verschiebung records a single move call. The mock cannot reproduce the real
// ordering because the domain model carries no reihenfolge — that column lives
// in the persistence layer only. Tests therefore assert on the recorded calls;
// the actual swap is covered by the repository integration test.
type Verschiebung struct {
	ID   int
	Hoch bool
}

type mockRepo struct {
	produkte  map[int]produkt.Produkt
	varianten map[int]varianteWithProdukt
	err       error

	// ProduktVerschiebungen and VarianteVerschiebungen record the moves the
	// command layer requested, in order.
	ProduktVerschiebungen  []Verschiebung
	VarianteVerschiebungen []Verschiebung

	// SortierteProdukte records the produkt IDs whose varianten were sorted.
	SortierteProdukte []int
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

func (m *mockRepo) VerschiebeProdukt(ctx context.Context, produktID int, hoch bool) error {
	if m.err != nil {
		return m.err
	}
	m.ProduktVerschiebungen = append(m.ProduktVerschiebungen, Verschiebung{ID: produktID, Hoch: hoch})
	return nil
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

func (m *mockRepo) VerschiebeVariante(ctx context.Context, varianteID int, hoch bool) error {
	if m.err != nil {
		return m.err
	}
	m.VarianteVerschiebungen = append(m.VarianteVerschiebungen, Verschiebung{ID: varianteID, Hoch: hoch})
	return nil
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

func (m *mockRepo) SortiereVariantenAlphabetisch(ctx context.Context, produktID int) error {
	if m.err != nil {
		return m.err
	}
	m.SortierteProdukte = append(m.SortierteProdukte, produktID)
	return nil
}
