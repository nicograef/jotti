//go:build unit

package druckstation_repo

import "context"

// NewMock creates a new mock repository.
func NewMock(konfigs []Druckstation, err error) *mockRepo {
	return &mockRepo{konfigs: konfigs, err: err}
}

type mockRepo struct {
	konfigs  []Druckstation
	upserted []Druckstation
	err      error
}

func (m *mockRepo) GetAlleDruckstationen(ctx context.Context) ([]Druckstation, error) {
	return m.konfigs, m.err
}

func (m *mockRepo) GetKonfigurierteDruckstationen(ctx context.Context) (map[string]Druckstation, error) {
	if m.err != nil {
		return nil, m.err
	}
	result := make(map[string]Druckstation)
	for _, k := range m.konfigs {
		if k.DruckerIP != "" {
			result[k.Kategorie] = k
		}
	}
	return result, nil
}

func (m *mockRepo) UpsertDruckstation(ctx context.Context, kategorie, druckerIP, bonmodus string) error {
	if m.err != nil {
		return m.err
	}
	m.upserted = append(m.upserted, Druckstation{
		Kategorie: kategorie,
		DruckerIP: druckerIP,
		Bonmodus:  bonmodus,
	})
	return nil
}
