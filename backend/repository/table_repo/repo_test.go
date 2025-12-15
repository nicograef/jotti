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

func setup(t *testing.T) (Repository, func(t *testing.T)) {
	db := dbpkg.OpenTestDatabase()

	_, err := db.Exec("DELETE FROM tables")
	if err != nil {
		t.Fatalf("Failed to clean tables table: %v", err)
	}

	return Repository{DB: db}, func(t *testing.T) {
		_, err = db.Exec("DELETE FROM tables")
		if err != nil {
			t.Fatalf("Failed to clean tables table: %v", err)
		}

		db.Close()
	}
}

func TestGetAllTablesDB(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	_, _ = repo.CreateTable(ctx, table.Table{Name: "GetAll Test 1", Status: table.ActiveStatus, CreatedAt: time.Now()})
	_, _ = repo.CreateTable(ctx, table.Table{Name: "GetAll Test 2", Status: table.ActiveStatus, CreatedAt: time.Now()})

	tables, err := repo.GetAllTables(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(tables) != 2 {
		t.Fatalf("expected exactly 2 tables, got %d", len(tables))
	}
}

func TestGetActiveTablesDB(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	_, _ = repo.CreateTable(ctx, table.Table{Name: "GetAll Test 1", Status: table.ActiveStatus, CreatedAt: time.Now()})
	_, _ = repo.CreateTable(ctx, table.Table{Name: "GetAll Test 2", Status: table.InactiveStatus, CreatedAt: time.Now()})

	tables, err := repo.GetActiveTables(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(tables) != 1 {
		t.Fatalf("expected exactly 1 active table, got %d", len(tables))
	}
}

func TestCreateTableInDB(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	tableID, err := repo.CreateTable(ctx, table.Table{Name: "Integration Test Table", Status: table.ActiveStatus, CreatedAt: time.Now()})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tableID < 1 {
		t.Fatalf("expected valid table ID, got %d", tableID)
	}
}

func TestUpdateTableDB(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	tableID, _ := repo.CreateTable(ctx, table.Table{Name: "Update Test Table", Status: table.ActiveStatus, CreatedAt: time.Now()})

	err := repo.UpdateTable(ctx, table.Table{ID: tableID, Name: "Updated Table Name", Status: table.ActiveStatus, CreatedAt: time.Now()})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	tables, err := repo.GetAllTables(ctx)
	if err != nil {
		t.Fatalf("expected no error getting table, got %v", err)
	}
	if tables[0].Name != "Updated Table Name" {
		t.Fatalf("expected name 'Updated Table Name', got %s", tables[0].Name)
	}
}

func TestUpdateTableDB_NotFound(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	err := repo.UpdateTable(ctx, table.Table{ID: 999999, Name: "New Name", Status: table.ActiveStatus, CreatedAt: time.Now()})

	if err != dbpkg.ErrNotFound {
		t.Fatalf("expected table not found error, got %v", err)
	}
}
