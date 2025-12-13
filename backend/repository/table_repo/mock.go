package table_repo

import (
	"context"

	"github.com/nicograef/jotti/backend/domain/table"
)

// NewMock creates a new mock repository with the given tables and error.
func NewMock(tables []table.Table, err error) *mockRepo {
	tableMap := make(map[int]table.Table)
	for _, t := range tables {
		tableMap[t.ID] = t
	}

	return &mockRepo{
		tables: tableMap,
		err:    err,
	}
}

type mockRepo struct {
	tables map[int]table.Table
	err    error
}

func (m mockRepo) GetTable(ctx context.Context, id int) (table.Table, error) {
	t, ok := m.tables[id]
	if !ok {
		return table.Table{}, m.err
	}
	return t, m.err
}

func (m mockRepo) GetAllTables(ctx context.Context) ([]table.Table, error) {
	var result []table.Table
	for _, t := range m.tables {
		result = append(result, t)
	}
	return result, m.err
}

func (m mockRepo) GetActiveTables(ctx context.Context) ([]table.Table, error) {
	var result []table.Table
	for _, t := range m.tables {
		if t.Status == table.ActiveStatus {
			result = append(result, t)
		}
	}
	return result, m.err
}

func (m mockRepo) CreateTable(ctx context.Context, t table.Table) (int, error) {
	newID := len(m.tables) + 1
	t.ID = newID
	m.tables[newID] = t
	return newID, m.err
}

func (m mockRepo) UpdateTable(ctx context.Context, t table.Table) error {
	m.tables[t.ID] = t
	return m.err
}
