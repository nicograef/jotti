//go:build integration

package table

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

func createTestTable(t *testing.T, db *sql.DB, name string) int {
	ctx := context.Background()
	var tableID int
	err := db.QueryRowContext(ctx, "INSERT INTO tables (name, status) VALUES ($1, 'active') RETURNING id", name).Scan(&tableID)
	if err != nil {
		t.Fatalf("failed to create test table: %v", err)
	}
	return tableID
}

func TestGetTableDB(t *testing.T) {
	db := database()
	defer db.Close()

	tableID := createTestTable(t, db, "Table 01")

	ctx := context.Background()
	persistence := &Persistence{DB: db}

	table, err := persistence.GetTable(ctx, tableID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if table.ID != tableID {
		t.Fatalf("expected table ID %d, got %d", tableID, table.ID)
	}
	if table.Name != "Table 01" {
		t.Fatalf("expected name 'Table 01', got %s", table.Name)
	}

	// Clean up
	_, _ = db.ExecContext(ctx, "DELETE FROM tables")
}

func TestGetTableDB_NotFound(t *testing.T) {
	db := database()
	defer db.Close()

	ctx := context.Background()
	persistence := &Persistence{DB: db}
	_, err := persistence.GetTable(ctx, 999999)

	if err != dbpkg.ErrNotFound {
		t.Fatalf("expected table not found error, got %v", err)
	}
}

func TestGetAllTablesDB(t *testing.T) {
	db := database()
	defer db.Close()

	ctx := context.Background()
	persistence := &Persistence{DB: db}

	// Create test tables
	_ = createTestTable(t, db, "Table 01")
	_ = createTestTable(t, db, "Table 02")

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
