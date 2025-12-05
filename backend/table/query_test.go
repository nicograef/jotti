//go:build unit

package table

import (
	"context"
	"testing"

	"github.com/nicograef/jotti/backend/db"
)

type mockQueryPersistence struct {
	shouldFail bool
	table      Table
}

func (m *mockQueryPersistence) GetTable(ctx context.Context, id int) (*Table, error) {
	if m.shouldFail {
		return nil, db.ErrNotFound
	}
	return &m.table, nil
}

func (m *mockQueryPersistence) GetAllTables(ctx context.Context) ([]Table, error) {
	if m.shouldFail {
		return nil, db.ErrNotFound
	}
	return []Table{m.table}, nil
}

func (m *mockQueryPersistence) GetActiveTables(ctx context.Context) ([]TablePublic, error) {
	if m.shouldFail {
		return nil, db.ErrNotFound
	}
	return []TablePublic{{ID: m.table.ID, Name: m.table.Name}}, nil
}

func TestGetTable(t *testing.T) {
	query := Query{Persistence: &mockQueryPersistence{
		table: Table{ID: 1, Name: "Table Name", Status: ActiveStatus},
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
	query := Query{Persistence: &mockQueryPersistence{shouldFail: true}}

	_, err := query.GetTable(context.Background(), 999)

	if err != ErrTableNotFound {
		t.Fatalf("expected ErrTableNotFound, got %v", err)
	}
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

func TestGetActiveTables(t *testing.T) {
	query := Query{Persistence: &mockQueryPersistence{
		table: Table{ID: 1, Name: "Table 1", Status: ActiveStatus},
	}}

	tables, err := query.GetActiveTables(context.Background())

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
