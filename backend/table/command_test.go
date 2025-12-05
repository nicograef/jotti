//go:build unit

package table

import (
	"context"
	"testing"

	"github.com/nicograef/jotti/backend/db"
)

type mockCommandPersistence struct {
	shouldFail bool
	table      *Table
}

func (m *mockCommandPersistence) CreateTable(ctx context.Context, name string) (int, error) {
	if m.shouldFail {
		return 0, db.ErrNotFound
	}
	m.table = &Table{ID: 1, Name: name, Status: InactiveStatus}
	return 1, nil
}

func (m *mockCommandPersistence) UpdateTable(ctx context.Context, id int, name string) error {
	if m.shouldFail {
		return db.ErrNotFound
	}
	return nil
}

func (m *mockCommandPersistence) ActivateTable(ctx context.Context, id int) error {
	if m.shouldFail {
		return db.ErrNotFound
	}
	return nil
}

func (m *mockCommandPersistence) DeactivateTable(ctx context.Context, id int) error {
	if m.shouldFail {
		return db.ErrNotFound
	}
	return nil
}

func TestCreateTable(t *testing.T) {
	command := Command{Persistence: &mockCommandPersistence{table: &Table{ID: 1}}}

	tableId, err := command.CreateTable(context.Background(), "Table 1")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tableId != 1 {
		t.Errorf("expected table ID 1, got %d", tableId)
	}
}

func TestCreateTable_Error(t *testing.T) {
	command := Command{Persistence: &mockCommandPersistence{shouldFail: true}}

	_, err := command.CreateTable(context.Background(), "Table 1")

	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestUpdateTable(t *testing.T) {
	command := Command{Persistence: &mockCommandPersistence{
		table: &Table{ID: 1, Name: "Old Name", Status: ActiveStatus},
	}}

	err := command.UpdateTable(context.Background(), 1, "New Name")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestUpdateTable_NotFound(t *testing.T) {
	command := Command{Persistence: &mockCommandPersistence{shouldFail: true}}

	err := command.UpdateTable(context.Background(), 999, "New Name")

	if err != ErrTableNotFound {
		t.Fatalf("expected ErrTableNotFound, got %v", err)
	}
}

func TestActivateTable(t *testing.T) {
	command := Command{Persistence: &mockCommandPersistence{
		table: &Table{ID: 1, Name: "Table 1", Status: InactiveStatus},
	}}

	err := command.ActivateTable(context.Background(), 1)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestActivateTable_NotFound(t *testing.T) {
	command := Command{Persistence: &mockCommandPersistence{shouldFail: true}}

	err := command.ActivateTable(context.Background(), 999)

	if err != ErrTableNotFound {
		t.Fatalf("expected ErrTableNotFound, got %v", err)
	}
}

func TestDeactivateTable(t *testing.T) {
	command := Command{Persistence: &mockCommandPersistence{
		table: &Table{ID: 1, Name: "Table 1", Status: ActiveStatus},
	}}

	err := command.DeactivateTable(context.Background(), 1)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestDeactivateTable_NotFound(t *testing.T) {
	command := Command{Persistence: &mockCommandPersistence{shouldFail: true}}

	err := command.DeactivateTable(context.Background(), 999)

	if err != ErrTableNotFound {
		t.Fatalf("expected ErrTableNotFound, got %v", err)
	}
}
