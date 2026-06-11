//go:build unit

package druckstation_repo

import (
	"context"

	"github.com/nicograef/jotti/backend/domain/druckstation"
)

// NewMock creates a new mock repository.
func NewMock(konfigs []druckstation.Druckstation, err error) *mockRepo {
	return &mockRepo{konfigs: konfigs, err: err}
}

type mockRepo struct {
	konfigs  []druckstation.Druckstation
	upserted []druckstation.Druckstation
	err      error
}

func (m *mockRepo) GetAlleDruckstationen(ctx context.Context) ([]druckstation.Druckstation, error) {
	return m.konfigs, m.err
}

func (m *mockRepo) GetKonfigurierteDruckstationen(ctx context.Context) (map[string]druckstation.Druckstation, error) {
	if m.err != nil {
		return nil, m.err
	}
	result := make(map[string]druckstation.Druckstation)
	for _, k := range m.konfigs {
		if k.DruckerIP != "" {
			result[string(k.Kategorie)] = k
		}
	}
	return result, nil
}

func (m *mockRepo) UpsertDruckstation(ctx context.Context, station druckstation.Druckstation) error {
	if m.err != nil {
		return m.err
	}
	m.upserted = append(m.upserted, station)
	return nil
}
