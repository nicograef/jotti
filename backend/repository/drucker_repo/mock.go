//go:build unit

package drucker_repo

import "context"

// NewMock creates a new mock repository.
func NewMock(konfigs []DruckerKonfig, err error) *mockRepo {
	return &mockRepo{konfigs: konfigs, err: err}
}

type mockRepo struct {
	konfigs  []DruckerKonfig
	upserted []DruckerKonfig
	err      error
}

func (m *mockRepo) GetAlleKategorieDrucker(ctx context.Context) ([]DruckerKonfig, error) {
	return m.konfigs, m.err
}

func (m *mockRepo) GetKonfigurierteKategorieDrucker(ctx context.Context) (map[string]DruckerKonfig, error) {
	if m.err != nil {
		return nil, m.err
	}
	result := make(map[string]DruckerKonfig)
	for _, k := range m.konfigs {
		if k.DruckerIP != "" {
			result[k.Kategorie] = k
		}
	}
	return result, nil
}

func (m *mockRepo) UpsertKategorieDrucker(ctx context.Context, kategorie, druckerIP, bonmodus string) error {
	if m.err != nil {
		return m.err
	}
	m.upserted = append(m.upserted, DruckerKonfig{
		Kategorie: kategorie,
		DruckerIP: druckerIP,
		Bonmodus:  bonmodus,
	})
	return nil
}
