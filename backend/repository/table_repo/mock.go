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

func (m mockRepo) GetActiveTables(ctx context.Context) ([]table.Tisch, error) {
	var result []table.Tisch
	for _, t := range m.tables {
		if t.Status == table.ActiveStatus {
			result = append(result, t)
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
