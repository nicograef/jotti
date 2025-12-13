package repository

import (
	"context"

	"github.com/nicograef/jotti/backend/admin/table/domain"
)

// NewMock creates a new mock repository with the given tables and error.
func NewMock(tables []domain.Table, err error) *mockRepo {
	tableMap := make(map[int]domain.Table)
	for _, t := range tables {
		tableMap[t.ID] = t
	}

	return &mockRepo{
		tables: tableMap,
		err:    err,
	}
}

type mockRepo struct {
	tables map[int]domain.Table
	err    error
}

func (m *mockRepo) GetTable(ctx context.Context, id int) (domain.Table, error) {
	t, ok := m.tables[id]
	if !ok {
		return domain.Table{}, m.err
	}
	return t, m.err
}

func (m *mockRepo) CreateTable(ctx context.Context, t domain.Table) (int, error) {
	newID := len(m.tables) + 1
	t.ID = newID
	m.tables[newID] = t
	return newID, m.err
}

func (m *mockRepo) UpdateTable(ctx context.Context, t domain.Table) error {
	m.tables[t.ID] = t
	return m.err
}
