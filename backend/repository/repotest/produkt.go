package repotest

import (
	"context"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/produkt"
)

// NewProduktRepo creates a new mock repository with the given produkte and error.
func NewProduktRepo(produkte []produkt.Produkt, err error) *ProduktRepo {
	produktMap := make(map[int]produkt.Produkt)
	for i := range produkte {
		produktMap[produkte[i].ID] = produkte[i]
	}

	return &ProduktRepo{
		produkte:  produktMap,
		varianten: make(map[int]produkt.VarianteMitProdukt),
		err:       err,
	}
}

type ProduktRepo struct {
	produkte         map[int]produkt.Produkt
	varianten        map[int]produkt.VarianteMitProdukt
	err              error
	updateProduktErr error
}

// SetUpdateProduktError makes UpdateProdukt fail with err while the reads keep
// succeeding — the shape of a UNIQUE violation on the produkt name.
func (m *ProduktRepo) SetUpdateProduktError(err error) {
	m.updateProduktErr = err
}

// AddVariante adds a variante to the mock repository, associated with a produkt.
func (m *ProduktRepo) AddVariante(produktID int, v produkt.Variante) {
	m.varianten[v.ID] = produkt.VarianteMitProdukt{Variante: v, ProduktID: produktID}
}

func (m *ProduktRepo) GetProdukt(ctx context.Context, id int) (produkt.Produkt, error) {
	if m.err != nil {
		return produkt.Produkt{}, m.err
	}
	t, ok := m.produkte[id]
	if !ok {
		return produkt.Produkt{}, db.ErrNotFound
	}
	return t, nil
}

func (m *ProduktRepo) CreateProdukt(ctx context.Context, t produkt.Produkt) (int, error) {
	newID := len(m.produkte) + 1
	t.ID = newID
	m.produkte[newID] = t
	return newID, m.err
}

func (m *ProduktRepo) UpdateProdukt(ctx context.Context, t produkt.Produkt) error {
	if m.updateProduktErr != nil {
		return m.updateProduktErr
	}
	m.produkte[t.ID] = t
	return m.err
}

// VerschiebeProdukt reicht nur den Fehler durch: Die Reihenfolge liegt allein
// in der Persistenz, das Domain-Modell trägt sie nicht. Den Tausch deckt der
// Integrationstest des Repositories ab.
func (m *ProduktRepo) VerschiebeProdukt(ctx context.Context, produktID int, hoch bool) error {
	return m.err
}

func (m *ProduktRepo) GetVariante(ctx context.Context, varianteID int) (produkt.Variante, error) {
	if m.err != nil {
		return produkt.Variante{}, m.err
	}
	vp, ok := m.varianten[varianteID]
	if !ok {
		return produkt.Variante{}, db.ErrNotFound
	}
	return vp.Variante, nil
}

func (m *ProduktRepo) CreateVariante(ctx context.Context, produktID int, v produkt.Variante) (int, error) {
	newID := len(m.varianten) + 1
	v.ID = newID
	m.varianten[newID] = produkt.VarianteMitProdukt{Variante: v, ProduktID: produktID}
	return newID, m.err
}

func (m *ProduktRepo) UpdateVariante(ctx context.Context, v produkt.Variante) error {
	if vp, ok := m.varianten[v.ID]; ok {
		m.varianten[v.ID] = produkt.VarianteMitProdukt{Variante: v, ProduktID: vp.ProduktID}
	}
	return m.err
}

func (m *ProduktRepo) VerschiebeVariante(ctx context.Context, varianteID int, hoch bool) error {
	return m.err
}

func (m *ProduktRepo) DeleteProduktMitVarianten(ctx context.Context, p produkt.Produkt) error {
	if m.err != nil {
		return m.err
	}
	m.produkte[p.ID] = p
	for i := range p.Varianten {
		v := p.Varianten[i]
		if vp, ok := m.varianten[v.ID]; ok {
			m.varianten[v.ID] = produkt.VarianteMitProdukt{Variante: v, ProduktID: vp.ProduktID}
		}
	}
	return nil
}

func (m *ProduktRepo) GetAllProdukte(ctx context.Context) ([]produkt.Produkt, error) {
	produkte := make([]produkt.Produkt, 0, len(m.produkte))
	for i := range m.produkte {
		produkte = append(produkte, m.produkte[i])
	}
	return produkte, m.err
}

// GetActiveProdukte mirrors the query GetAktiveProdukte (sqlc/queries/produkte.sql):
// a Produkt without an active Variante drops out, and only active Varianten are returned.
func (m *ProduktRepo) GetActiveProdukte(ctx context.Context) ([]produkt.Produkt, error) {
	produkte := make([]produkt.Produkt, 0)
	for i := range m.produkte {
		p := m.produkte[i]
		if p.Status != produkt.ActiveStatus {
			continue
		}
		aktive := aktiveVarianten(p.Varianten)
		if len(aktive) == 0 {
			continue
		}
		p.Varianten = aktive
		produkte = append(produkte, p)
	}
	return produkte, m.err
}

func aktiveVarianten(varianten []produkt.Variante) []produkt.Variante {
	aktive := make([]produkt.Variante, 0, len(varianten))
	for _, v := range varianten {
		if v.Status == produkt.ActiveStatus {
			aktive = append(aktive, v)
		}
	}
	return aktive
}

func (m *ProduktRepo) GetVariantenByIDs(ctx context.Context, ids []int) (map[int]produkt.VarianteMitProdukt, error) {
	if m.err != nil {
		return nil, m.err
	}
	result := make(map[int]produkt.VarianteMitProdukt, len(ids))
	for _, id := range ids {
		if vp, ok := m.varianten[id]; ok {
			result[id] = vp
		}
	}
	return result, nil
}

func (m *ProduktRepo) GetProdukteByIDs(ctx context.Context, ids []int) (map[int]produkt.Produkt, error) {
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

func (m *ProduktRepo) SortiereVariantenAlphabetisch(ctx context.Context, produktID int) error {
	return m.err
}
