//go:build unit

package table

import (
	"context"
	"testing"

	"github.com/nicograef/jotti/backend/db"
)

type mockQueryPersistence struct {
	err   error
	table Table
}

func (m *mockQueryPersistence) GetTable(ctx context.Context, id int) (*Table, error) {
	return &m.table, m.err
}

func (m *mockQueryPersistence) GetAllTables(ctx context.Context) ([]Table, error) {
	return []Table{m.table}, m.err
}

func TestGetTable(t *testing.T) {
	query := Query{TablePersistence: &mockQueryPersistence{
		table: Table{ID: 1, Name: "Table Name"},
	}}

	table, err := query.GetTable(context.Background(), 1)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if table.ID != 1 {
		t.Errorf("expected table ID 1, got %d", table.ID)
	}
	if table.Name != "Table Name" {
		t.Errorf("expected name 'Table Name', got %s", table.Name)
	}
}

func TestGetTable_NotFound(t *testing.T) {
	query := Query{TablePersistence: &mockQueryPersistence{err: db.ErrNotFound}}

	_, err := query.GetTable(context.Background(), 999)

	if err != ErrTableNotFound {
		t.Fatalf("expected ErrTableNotFound, got %v", err)
	}
}

func TestGetAllTables(t *testing.T) {
	query := Query{TablePersistence: &mockQueryPersistence{
		table: Table{ID: 1, Name: "Table 1"},
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
