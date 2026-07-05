//go:build unit

package tisch_repo

import (
	"context"

	"github.com/nicograef/jotti/backend/domain/tisch"
)

// NewMock creates a new mock repository with the given tables and error.
func NewMock(tables []tisch.Tisch, err error) *mockRepo {
	tableMap := make(map[int]tisch.Tisch)
	for _, t := range tables {
		tableMap[t.ID] = t
	}

	return &mockRepo{
		tables: tableMap,
		err:    err,
	}
}

type mockRepo struct {
	tables map[int]tisch.Tisch
	err    error
}

func (m mockRepo) GetTable(ctx context.Context, id int) (tisch.Tisch, error) {
	t, ok := m.tables[id]
	if !ok {
		return tisch.Tisch{}, m.err
	}
	return t, m.err
}

func (m mockRepo) GetAllTables(ctx context.Context) ([]tisch.Tisch, error) {
	var result []tisch.Tisch
	for _, t := range m.tables {
		result = append(result, t)
	}
	return result, m.err
}

func (m mockRepo) GetActiveTables(ctx context.Context, kassensitzungNr int) ([]tisch.AktiverTisch, error) {
	var result []tisch.AktiverTisch
	for _, t := range m.tables {
		if t.Status == tisch.ActiveStatus {
			result = append(result, tisch.AktiverTisch{ID: t.ID, Name: t.Name, SaldoCents: 0})
		}
	}
	return result, m.err
}

func (m mockRepo) CreateTable(ctx context.Context, t tisch.Tisch) (int, error) {
	newID := len(m.tables) + 1
	t.ID = newID
	m.tables[newID] = t
	return newID, m.err
}

func (m mockRepo) UpdateTable(ctx context.Context, t tisch.Tisch) error {
	m.tables[t.ID] = t
	return m.err
}

func (m mockRepo) GetActiveTablesWithFavorites(_ context.Context, _ int, _ int) ([]tisch.AktiverTischMitFavorit, error) {
	return nil, m.err
}
