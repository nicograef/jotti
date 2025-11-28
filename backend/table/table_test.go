//go:build unit

package table

import (
	"context"
	"testing"
)

type mockTablePersistence struct {
	shouldFail bool
	table      *Table
}

func (m *mockTablePersistence) CreateTable(ctx context.Context, name string) (int, error) {
	if m.shouldFail {
		return 0, ErrTableNotFound
	}
	m.table = &Table{ID: 1, Name: name, Status: InactiveStatus}
	return 1, nil
}

func (m *mockTablePersistence) GetTable(ctx context.Context, id int) (*Table, error) {
	if m.shouldFail {
		return nil, ErrTableNotFound
	}
	return m.table, nil
}

func (m *mockTablePersistence) GetAllTables(ctx context.Context) ([]*Table, error) {
	if m.shouldFail {
		return nil, ErrTableNotFound
	}
	return []*Table{m.table}, nil
}

func (m *mockTablePersistence) GetActiveTables(ctx context.Context) ([]*TablePublic, error) {
	if m.shouldFail {
		return nil, ErrTableNotFound
	}
	return []*TablePublic{{ID: m.table.ID, Name: m.table.Name}}, nil
}

func (m *mockTablePersistence) UpdateTable(ctx context.Context, id int, name string) error {
	if m.shouldFail {
		return ErrTableNotFound
	}
	return nil
}

func (m *mockTablePersistence) ActivateTable(ctx context.Context, id int) error {
	if m.shouldFail {
		return ErrTableNotFound
	}
	return nil
}

func (m *mockTablePersistence) DeactivateTable(ctx context.Context, id int) error {
	if m.shouldFail {
		return ErrTableNotFound
	}
	return nil
}

func TestCreateTable(t *testing.T) {
	tableService := Service{Persistence: &mockTablePersistence{table: &Table{ID: 1}}}

	table, err := tableService.CreateTable(context.Background(), "Table 1")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if table.ID != 1 {
		t.Errorf("expected table ID 1, got %d", table.ID)
	}
	if table.Name != "Table 1" {
		t.Errorf("expected name 'Table 1', got %s", table.Name)
	}
	if table.Status != InactiveStatus {
		t.Errorf("expected status 'inactive', got %s", table.Status)
	}
}

func TestCreateTable_Error(t *testing.T) {
	tableService := Service{Persistence: &mockTablePersistence{shouldFail: true}}

	_, err := tableService.CreateTable(context.Background(), "Table 1")

	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestUpdateTable(t *testing.T) {
	tableService := Service{Persistence: &mockTablePersistence{
		table: &Table{ID: 1, Name: "Old Name", Status: ActiveStatus},
	}}

	table, err := tableService.UpdateTable(context.Background(), 1, "New Name")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if table.ID != 1 {
		t.Errorf("expected table ID 1, got %d", table.ID)
	}
	if table.Name != "Old Name" {
		t.Errorf("expected name 'Old Name', got %s", table.Name)
	}
}

func TestUpdateTable_NotFound(t *testing.T) {
	tableService := Service{Persistence: &mockTablePersistence{shouldFail: true}}

	_, err := tableService.UpdateTable(context.Background(), 999, "New Name")

	if err != ErrTableNotFound {
		t.Fatalf("expected ErrTableNotFound, got %v", err)
	}
}

func TestGetTable(t *testing.T) {
	tableService := Service{Persistence: &mockTablePersistence{
		table: &Table{ID: 1, Name: "Table Name", Status: ActiveStatus},
	}}

	table, err := tableService.GetTable(context.Background(), 1)

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
	tableService := Service{Persistence: &mockTablePersistence{shouldFail: true}}

	_, err := tableService.GetTable(context.Background(), 999)

	if err != ErrTableNotFound {
		t.Fatalf("expected ErrTableNotFound, got %v", err)
	}
}

func TestGetAllTables(t *testing.T) {
	tableService := Service{Persistence: &mockTablePersistence{
		table: &Table{ID: 1, Name: "Table 1", Status: ActiveStatus},
	}}

	tables, err := tableService.GetAllTables(context.Background())

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
	tableService := Service{Persistence: &mockTablePersistence{
		table: &Table{ID: 1, Name: "Table 1", Status: ActiveStatus},
	}}

	tables, err := tableService.GetActiveTables(context.Background())

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

func TestActivateTable(t *testing.T) {
	tableService := Service{Persistence: &mockTablePersistence{
		table: &Table{ID: 1, Name: "Table 1", Status: InactiveStatus},
	}}

	err := tableService.ActivateTable(context.Background(), 1)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestActivateTable_NotFound(t *testing.T) {
	tableService := Service{Persistence: &mockTablePersistence{shouldFail: true}}

	err := tableService.ActivateTable(context.Background(), 999)

	if err != ErrTableNotFound {
		t.Fatalf("expected ErrTableNotFound, got %v", err)
	}
}

func TestDeactivateTable(t *testing.T) {
	tableService := Service{Persistence: &mockTablePersistence{
		table: &Table{ID: 1, Name: "Table 1", Status: ActiveStatus},
	}}

	err := tableService.DeactivateTable(context.Background(), 1)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestDeactivateTable_NotFound(t *testing.T) {
	tableService := Service{Persistence: &mockTablePersistence{shouldFail: true}}

	err := tableService.DeactivateTable(context.Background(), 999)

	if err != ErrTableNotFound {
		t.Fatalf("expected ErrTableNotFound, got %v", err)
	}
}
