//go:build unit

package table_repo

import (
	"context"

	"github.com/nicograef/jotti/backend/domain/table"
)

// NewMock creates a new mock repository with the given tables and error.
func NewMock(tables []table.Tisch, err error) *mockRepo {
	tableMap := make(map[int]table.Tisch)
	for _, t := range tables {
		tableMap[t.ID] = t
	}

	return &mockRepo{
		tables: tableMap,
		err:    err,
	}
}

type mockRepo struct {
	tables map[int]table.Tisch
	err    error
}

func (m mockRepo) GetTable(ctx context.Context, id int) (table.Tisch, error) {
	t, ok := m.tables[id]
	if !ok {
		return table.Tisch{}, m.err
	}
	return t, m.err
}

func (m mockRepo) GetAllTables(ctx context.Context) ([]table.Tisch, error) {
	var result []table.Tisch
	for _, t := range m.tables {
		result = append(result, t)
	}
	return result, m.err
}

func (m mockRepo) GetActiveTables(ctx context.Context, kassensitzungNr int) ([]table.AktiverTisch, error) {
	var result []table.AktiverTisch
	for _, t := range m.tables {
		if t.Status == table.ActiveStatus {
			result = append(result, table.AktiverTisch{ID: t.ID, Name: t.Name, SaldoCents: 0})
		}
	}
	return result, m.err
}

func (m mockRepo) CreateTable(ctx context.Context, t table.Tisch) (int, error) {
	newID := len(m.tables) + 1
	t.ID = newID
	m.tables[newID] = t
	return newID, m.err
}

func (m mockRepo) UpdateTable(ctx context.Context, t table.Tisch) error {
	m.tables[t.ID] = t
	return m.err
}

func (m mockRepo) GetActiveTablesWithFavorites(_ context.Context, _ int, _ int) ([]table.AktiverTischMitFavorit, error) {
	return nil, m.err
}
