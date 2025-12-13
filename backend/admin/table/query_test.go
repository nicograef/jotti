//go:build unit

package table

import (
	"context"
	"testing"
)

type mockQueryPersistence struct {
	err   error
	table Table
}

func (m *mockQueryPersistence) GetAllTables(ctx context.Context) ([]Table, error) {
	return []Table{m.table}, m.err
}

func TestGetAllTables(t *testing.T) {
	query := Query{Persistence: &mockQueryPersistence{
		table: Table{ID: 1, Name: "Table 1", Status: ActiveStatus},
	}}

	tables, err := query.GetAllTables(context.Background())

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(tables) != 1 {
		t.Fatalf("expected 1 table, got %d", len(tables))
	}
	if tables[0].Name != "Table 1" {
		t.Errorf("expected name 'Table 1', got %s", tables[0].Name)
	}
}
