//go:build integration

package tisch_repo

import (
	"context"
	"errors"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	dbpkg "github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/tisch"
)

func setup(t *testing.T) (Repository, func(t *testing.T)) {
	db := dbpkg.OpenTestDatabase()

	_, err := db.Exec("DELETE FROM tische")
	if err != nil {
		t.Fatalf("Failed to clean tische table: %v", err)
	}

	return NewRepository(db), func(t *testing.T) {
		_, err = db.Exec("DELETE FROM tische")
		if err != nil {
			t.Fatalf("Failed to clean tische table: %v", err)
		}

		_ = db.Close()
	}
}

func TestGetAllTablesDB(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	now := time.Now().UTC()
	_, _ = repo.CreateTable(ctx, tisch.Tisch{Name: "GetAll Test 1", Status: tisch.ActiveStatus, CreatedAt: now, UpdatedAt: now})
	_, _ = repo.CreateTable(ctx, tisch.Tisch{Name: "GetAll Test 2", Status: tisch.ActiveStatus, CreatedAt: now, UpdatedAt: now})

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
	now := time.Now().UTC()
	_, _ = repo.CreateTable(ctx, tisch.Tisch{Name: "GetAll Test 1", Status: tisch.ActiveStatus, CreatedAt: now, UpdatedAt: now})
	_, _ = repo.CreateTable(ctx, tisch.Tisch{Name: "GetAll Test 2", Status: tisch.InactiveStatus, CreatedAt: now, UpdatedAt: now})

	tables, err := repo.GetActiveTables(ctx, 0)
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
	now := time.Now().UTC()
	tableID, err := repo.CreateTable(ctx, tisch.Tisch{Name: "Integration Test Table", Status: tisch.ActiveStatus, CreatedAt: now, UpdatedAt: now})
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
	now := time.Now().UTC()
	tableID, _ := repo.CreateTable(ctx, tisch.Tisch{Name: "Update Test Table", Status: tisch.ActiveStatus, CreatedAt: now, UpdatedAt: now})

	err := repo.UpdateTable(ctx, tisch.Tisch{ID: tableID, Name: "Updated Table Name", Status: tisch.ActiveStatus, CreatedAt: now, UpdatedAt: now})
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
	now := time.Now().UTC()
	err := repo.UpdateTable(ctx, tisch.Tisch{ID: 999999, Name: "New Name", Status: tisch.ActiveStatus, CreatedAt: now, UpdatedAt: now})

	if !errors.Is(err, dbpkg.ErrNotFound) {
		t.Fatalf("expected table not found error, got %v", err)
	}
}
