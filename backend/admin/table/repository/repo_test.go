//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/nicograef/jotti/backend/admin/table/domain"
	dbpkg "github.com/nicograef/jotti/backend/db"
)

func TestGetAllTablesDB(t *testing.T) {
	db := dbpkg.OpenTestDatabase()
	defer db.Close()

	ctx := context.Background()
	repo := &Repository{DB: db}

	// Create test tables
	_, _ = repo.CreateTable(ctx, domain.Table{Name: "GetAll Test 1", Status: domain.ActiveStatus, CreatedAt: time.Now()})
	_, _ = repo.CreateTable(ctx, domain.Table{Name: "GetAll Test 2", Status: domain.ActiveStatus, CreatedAt: time.Now()})

	tables, err := repo.GetAllTables(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(tables) < 2 {
		t.Fatalf("expected at least 2 tables, got %d", len(tables))
	}

	// Clean up
	_, _ = db.ExecContext(ctx, "DELETE FROM tables")
}

func TestCreateTableInDB(t *testing.T) {
	db := dbpkg.OpenTestDatabase()
	defer db.Close()

	ctx := context.Background()
	repo := &Repository{DB: db}

	tableID, err := repo.CreateTable(ctx, domain.Table{Name: "Integration Test Table", Status: domain.ActiveStatus, CreatedAt: time.Now()})
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
	tableID, _ := repo.CreateTable(ctx, domain.Table{Name: "Update Test Table", Status: domain.ActiveStatus, CreatedAt: time.Now()})

	// Update it
	err := repo.UpdateTable(ctx, domain.Table{ID: tableID, Name: "Updated Table Name", Status: domain.ActiveStatus, CreatedAt: time.Now()})
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
	err := persistence.UpdateTable(ctx, domain.Table{ID: 999999, Name: "New Name", Status: domain.ActiveStatus, CreatedAt: time.Now()})

	if err != dbpkg.ErrNotFound {
		t.Fatalf("expected table not found error, got %v", err)
	}
}
