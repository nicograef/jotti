//go:build unit

package application

import (
	"context"
	"testing"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/table"
	"github.com/nicograef/jotti/backend/repository/table_repo"
)

func TestCreateTable(t *testing.T) {
	ctx := context.Background()
	repo := table_repo.NewMock([]table.Table{}, nil)
	command := Command{TableRepo: repo}

	tableId, err := command.CreateTable(ctx, "Table 1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tableId != 1 {
		t.Errorf("expected table ID 1, got %d", tableId)
	}

	table, err := command.TableRepo.GetTable(ctx, tableId)
	if err != nil {
		t.Fatalf("expected no error retrieving table, got %v", err)
	}
	if table.Name != "Table 1" {
		t.Errorf("expected table name 'Table 1', got %s", table.Name)
	}
}

func TestCreateTable_Error(t *testing.T) {
	repo := table_repo.NewMock([]table.Table{}, db.ErrAlreadyExists)
	command := Command{TableRepo: repo}

	_, err := command.CreateTable(context.Background(), "Table 1")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestUpdateTable(t *testing.T) {
	repo := table_repo.NewMock([]table.Table{{ID: 1, Name: "Old Name", Status: table.ActiveStatus}}, nil)
	command := Command{TableRepo: repo}

	err := command.UpdateTable(context.Background(), 1, "New Name")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	table, err := command.TableRepo.GetTable(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error retrieving table, got %v", err)
	}
	if table.Name != "New Name" {
		t.Errorf("expected table name to be 'New Name', got %s", table.Name)
	}
}

func TestUpdateTable_NotFound(t *testing.T) {
	repo := table_repo.NewMock([]table.Table{}, db.ErrNotFound)
	command := Command{TableRepo: repo}

	err := command.UpdateTable(context.Background(), 999, "New Name")
	if err != ErrTableNotFound {
		t.Fatalf("expected ErrTableNotFound, got %v", err)
	}
}

func TestActivateTable(t *testing.T) {
	repo := table_repo.NewMock([]table.Table{{ID: 1, Name: "Table 1", Status: table.InactiveStatus}}, nil)
	command := Command{TableRepo: repo}

	err := command.ActivateTable(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	tbl, err := repo.GetTable(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error retrieving table, got %v", err)
	}
	if tbl.Status != table.ActiveStatus {
		t.Errorf("expected table status to be Active, got %v", tbl.Status)
	}
}

func TestActivateTable_NotFound(t *testing.T) {
	repo := table_repo.NewMock([]table.Table{}, db.ErrNotFound)
	command := Command{TableRepo: repo}

	err := command.ActivateTable(context.Background(), 999)
	if err != ErrTableNotFound {
		t.Fatalf("expected ErrTableNotFound, got %v", err)
	}
}

func TestDeactivateTable(t *testing.T) {
	repo := table_repo.NewMock([]table.Table{{ID: 1, Name: "Table 1", Status: table.ActiveStatus}}, nil)
	command := Command{TableRepo: repo}

	err := command.DeactivateTable(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	tbl, err := repo.GetTable(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error retrieving table, got %v", err)
	}
	if tbl.Status != table.InactiveStatus {
		t.Errorf("expected table status to be Inactive, got %v", tbl.Status)
	}
}

func TestDeactivateTable_NotFound(t *testing.T) {
	repo := table_repo.NewMock([]table.Table{}, db.ErrNotFound)
	command := Command{TableRepo: repo}

	err := command.DeactivateTable(context.Background(), 999)
	if err != ErrTableNotFound {
		t.Fatalf("expected ErrTableNotFound, got %v", err)
	}
}
