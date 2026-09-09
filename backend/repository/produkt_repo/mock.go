//go:build unit

package produkt_repo

import (
	"context"

	"github.com/nicograef/jotti/backend/db"
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
	produkte         map[int]produkt.Produkt
	varianten        map[int]varianteWithProdukt
	err              error
	updateProduktErr error
}

// SetUpdateProduktError makes UpdateProdukt fail with err while the reads keep
// succeeding — the shape of a UNIQUE violation on the produkt name.
func (m *mockRepo) SetUpdateProduktError(err error) {
	m.updateProduktErr = err
}

// AddVariante adds a variante to the mock repository, associated with a produkt.
func (m *mockRepo) AddVariante(produktID int, v produkt.Variante) {
	m.varianten[v.ID] = varianteWithProdukt{variante: v, produktID: produktID}
}

func (m *mockRepo) GetProdukt(ctx context.Context, id int) (produkt.Produkt, error) {
	if m.err != nil {
		return produkt.Produkt{}, m.err
	}
	t, ok := m.produkte[id]
	if !ok {
		return produkt.Produkt{}, db.ErrNotFound
	}
	return t, nil
}

func (m *mockRepo) CreateProdukt(ctx context.Context, t produkt.Produkt) (int, error) {
	newID := len(m.produkte) + 1
	t.ID = newID
	m.produkte[newID] = t
	return newID, m.err
}

func (m *mockRepo) UpdateProdukt(ctx context.Context, t produkt.Produkt) error {
	if m.updateProduktErr != nil {
		return m.updateProduktErr
	}
	m.produkte[t.ID] = t
	return m.err
}

// VerschiebeProdukt reicht nur den Fehler durch: Die Reihenfolge liegt allein
// in der Persistenz, das Domain-Modell trägt sie nicht. Den Tausch deckt der
// Integrationstest des Repositories ab.
func (m *mockRepo) VerschiebeProdukt(ctx context.Context, produktID int, hoch bool) error {
	return m.err
}

func (m *mockRepo) GetVariante(ctx context.Context, varianteID int) (produkt.Variante, error) {
	if m.err != nil {
		return produkt.Variante{}, m.err
	}
	vp, ok := m.varianten[varianteID]
	if !ok {
		return produkt.Variante{}, db.ErrNotFound
	}
	return vp.variante, nil
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

// GetActiveProdukte spiegelt den INNER JOIN der Query GetAktiveProdukte
// (sqlc/queries/produkte.sql): ein aktives Produkt ohne aktive Variante ist
// nicht bestellbar und fällt raus.
func (m *mockRepo) GetActiveProdukte(ctx context.Context) ([]produkt.Produkt, error) {
	produkte := make([]produkt.Produkt, 0)
	for i := range m.produkte {
		if m.produkte[i].Status == produkt.ActiveStatus && hatAktiveVariante(m.produkte[i].Varianten) {
			produkte = append(produkte, m.produkte[i])
		}
	}
	return produkte, m.err
}

func hatAktiveVariante(varianten []produkt.Variante) bool {
	for _, v := range varianten {
		if v.Status == produkt.ActiveStatus {
			return true
		}
	}
	return false
}

func (m *mockRepo) GetVariantenByIDs(ctx context.Context, ids []int) (map[int]produkt.VarianteMitProdukt, error) {
	if m.err != nil {
		return nil, m.err
	}
	result := make(map[int]produkt.VarianteMitProdukt, len(ids))
	for _, id := range ids {
		if vp, ok := m.varianten[id]; ok {
			result[id] = produkt.VarianteMitProdukt{Variante: vp.variante, ProduktID: vp.produktID}
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
	return m.err
}
