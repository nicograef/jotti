//go:build integration

package table_admin

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	dbpkg "github.com/nicograef/jotti/backend/db"
)

func database() *sql.DB {
	db, err := sql.Open("pgx", "host=localhost port=5432 user=admin password=admin dbname=jotti sslmode=disable")
	if err != nil {
		fmt.Printf("failed to connect to Postgres: %v\n", err)
		os.Exit(1)
	}

	err = db.Ping()
	if err != nil {
		fmt.Printf("failed to ping Postgres: %v\n", err)
		os.Exit(1)
	}

	return db
}

func TestCreateTableInDB(t *testing.T) {
	db := database()
	defer db.Close()

	ctx := context.Background()
	persistence := &Persistence{DB: db}
	tableID, err := persistence.CreateTable(ctx, "Integration Test Table")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tableID < 1 {
		t.Fatalf("expected valid table ID, got %d", tableID)
	}

	// Clean up
	_, _ = db.ExecContext(ctx, "DELETE FROM tables")
}

func TestGetAllTablesDB(t *testing.T) {
	db := database()
	defer db.Close()

	ctx := context.Background()
	persistence := &Persistence{DB: db}

	// Create test tables
	_, _ = persistence.CreateTable(ctx, "GetAll Test 1")
	_, _ = persistence.CreateTable(ctx, "GetAll Test 2")

	tables, err := persistence.GetAllTables(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(tables) < 2 {
		t.Fatalf("expected at least 2 tables, got %d", len(tables))
	}

	// Clean up
	_, _ = db.ExecContext(ctx, "DELETE FROM tables")
}

func TestUpdateTableDB(t *testing.T) {
	db := database()
	defer db.Close()

	ctx := context.Background()
	persistence := &Persistence{DB: db}

	// Create a table
	tableID, _ := persistence.CreateTable(ctx, "Update Test Table")

	// Update it
	err := persistence.UpdateTable(ctx, tableID, "Updated Table Name")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify update
	tables, err := persistence.GetAllTables(ctx)
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
	db := database()
	defer db.Close()

	ctx := context.Background()
	persistence := &Persistence{DB: db}
	err := persistence.UpdateTable(ctx, 999999, "New Name")

	if err != dbpkg.ErrNotFound {
		t.Fatalf("expected table not found error, got %v", err)
	}
}

func TestActivateTableDB(t *testing.T) {
	db := database()
	defer db.Close()

	ctx := context.Background()
	persistence := &Persistence{DB: db}

	// Create a table (starts as inactive)
	tableID, _ := persistence.CreateTable(ctx, "Activate Test Table")

	// Activate it
	err := persistence.ActivateTable(ctx, tableID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify status
	tables, err := persistence.GetAllTables(ctx)
	if err != nil {
		t.Fatalf("expected no error getting table, got %v", err)
	}
	if tables[0].Status != ActiveStatus {
		t.Fatalf("expected status 'active', got %s", tables[0].Status)
	}

	// Clean up
	_, _ = db.ExecContext(ctx, "DELETE FROM tables")
}

func TestActivateTableDB_NotFound(t *testing.T) {
	db := database()
	defer db.Close()

	ctx := context.Background()
	persistence := &Persistence{DB: db}
	err := persistence.ActivateTable(ctx, 999999)

	if err != dbpkg.ErrNotFound {
		t.Fatalf("expected table not found error, got %v", err)
	}
}

func TestDeactivateTableDB(t *testing.T) {
	db := database()
	defer db.Close()

	ctx := context.Background()
	persistence := &Persistence{DB: db}

	// Create and activate a table
	tableID, _ := persistence.CreateTable(ctx, "Deactivate Test Table")
	_ = persistence.ActivateTable(ctx, tableID)

	// Deactivate it
	err := persistence.DeactivateTable(ctx, tableID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify status
	tables, err := persistence.GetAllTables(ctx)
	if err != nil {
		t.Fatalf("expected no error getting table, got %v", err)
	}
	if tables[0].Status != InactiveStatus {
		t.Fatalf("expected status 'inactive', got %s", tables[0].Status)
	}

	// Clean up
	_, _ = db.ExecContext(ctx, "DELETE FROM tables")
}

func TestDeactivateTableDB_NotFound(t *testing.T) {
	db := database()
	defer db.Close()

	ctx := context.Background()
	persistence := &Persistence{DB: db}
	err := persistence.DeactivateTable(ctx, 999999)

	if err != dbpkg.ErrNotFound {
		t.Fatalf("expected table not found error, got %v", err)
	}
}
