//go:build integration

package product

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func database() *sql.DB {
	db, err := sql.Open("pgx", "host=localhost port=5432 user=admin password=admin dbname=jotti sslmode=disable")
	if err != nil {
		fmt.Printf("Failed to connect to Postgres: %v\n", err)
		os.Exit(1)
	}

	err = db.Ping()
	if err != nil {
		fmt.Printf("Failed to ping Postgres: %v\n", err)
		os.Exit(1)
	}

	return db
}

func createTestProduct(t *testing.T, db *sql.DB, name string) int {
	ctx := context.Background()
	var productID int
	err := db.QueryRowContext(ctx, "INSERT INTO products (name, description, net_price_cents, category, status) VALUES ($1, $2, $3, $4, 'active') RETURNING id", name, "Test Description", 550, "food").Scan(&productID)
	if err != nil {
		t.Fatalf("failed to create test product: %v", err)
	}
	return productID
}

func TestGetAllProducts(t *testing.T) {
	db := database()
	defer db.Close()
	ctx := context.Background()

	persistence := &Persistence{DB: db}
	_ = createTestProduct(t, db, "Product 1")
	_ = createTestProduct(t, db, "Product 2")

	products, err := persistence.GetAllProducts(ctx)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(products) < 2 {
		t.Fatalf("Expected at least 2 products, got %d", len(products))
	}

	// Cleanup
	_, _ = db.ExecContext(ctx, "DELETE FROM products")
}
