//go:build integration

package table_repo

import (
	"context"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	dbpkg "github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/table"
)

func TestGetAllTablesDB(t *testing.T) {
	db := dbpkg.OpenTestDatabase()
	defer db.Close()

	ctx := context.Background()
	repo := &Repository{DB: db}

	// Create test tables
	_, _ = repo.CreateTable(ctx, table.Table{Name: "GetAll Test 1", Status: table.ActiveStatus, CreatedAt: time.Now()})
	_, _ = repo.CreateTable(ctx, table.Table{Name: "GetAll Test 2", Status: table.ActiveStatus, CreatedAt: time.Now()})

	tables, err := repo.GetAllTables(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(tables) != 2 {
		t.Fatalf("expected exactly 2 tables, got %d", len(tables))
	}

	// Clean up
	_, _ = db.ExecContext(ctx, "DELETE FROM tables")
}

func TestGetActiveTablesDB(t *testing.T) {
	db := dbpkg.OpenTestDatabase()
	defer db.Close()

	ctx := context.Background()
	repo := &Repository{DB: db}

	// Create test tables
	_, _ = repo.CreateTable(ctx, table.Table{Name: "GetAll Test 1", Status: table.ActiveStatus, CreatedAt: time.Now()})
	_, _ = repo.CreateTable(ctx, table.Table{Name: "GetAll Test 2", Status: table.InactiveStatus, CreatedAt: time.Now()})

	tables, err := repo.GetActiveTables(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(tables) != 1 {
		t.Fatalf("expected exactly 1 active table, got %d", len(tables))
	}

	// Clean up
	_, _ = db.ExecContext(ctx, "DELETE FROM tables")
}

func TestCreateTableInDB(t *testing.T) {
	db := dbpkg.OpenTestDatabase()
	defer db.Close()

	ctx := context.Background()
	repo := &Repository{DB: db}

	tableID, err := repo.CreateTable(ctx, table.Table{Name: "Integration Test Table", Status: table.ActiveStatus, CreatedAt: time.Now()})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tableID < 1 {
		t.Fatalf("expected valid table ID, got %d", tableID)
	}

	// Clean up
	_, _ = db.ExecContext(ctx, "DELETE FROM tables")
}

func TestUpdateTableDB(t *testing.T) {
	db := dbpkg.OpenTestDatabase()
	defer db.Close()

	ctx := context.Background()
	repo := &Repository{DB: db}

	// Create a table
	tableID, _ := repo.CreateTable(ctx, table.Table{Name: "Update Test Table", Status: table.ActiveStatus, CreatedAt: time.Now()})

	// Update it
	err := repo.UpdateTable(ctx, table.Table{ID: tableID, Name: "Updated Table Name", Status: table.ActiveStatus, CreatedAt: time.Now()})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify update
	tables, err := repo.GetAllTables(ctx)
	if err != nil {
		t.Fatalf("expected no error getting table, got %v", err)
	}
	if tables[0].Name != "Updated Table Name" {
		t.Fatalf("expected name 'Updated Table Name', got %s", tables[0].Name)
	}

	// Clean up
	_, _ = db.ExecContext(ctx, "DELETE FROM tables")
}

func TestUpdateTableDB_NotFound(t *testing.T) {
	db := dbpkg.OpenTestDatabase()
	defer db.Close()

	ctx := context.Background()
	persistence := &Repository{DB: db}
	err := persistence.UpdateTable(ctx, table.Table{ID: 999999, Name: "New Name", Status: table.ActiveStatus, CreatedAt: time.Now()})

	if err != dbpkg.ErrNotFound {
		t.Fatalf("expected table not found error, got %v", err)
	}
}
