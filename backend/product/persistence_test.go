//go:build integration

package product

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

func TestCreateProductInDB(t *testing.T) {
	db := database()
	defer db.Close()
	ctx := context.Background()

	persistence := &Persistence{DB: db}
	productID, err := persistence.CreateProduct(ctx, "French Fries", "The best fries in town", 499, FoodCategory)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if productID < 1 {
		t.Fatalf("Expected valid product ID, got %d", productID)
	}

	// Cleanup
	_, _ = db.ExecContext(ctx, "DELETE FROM products WHERE id = $1", productID)
}

func TestGetProduct(t *testing.T) {
	db := database()
	defer db.Close()
	ctx := context.Background()

	persistence := &Persistence{DB: db}
	productID, _ := persistence.CreateProduct(ctx, "Test Product", "Test Description", 599, FoodCategory)

	product, err := persistence.GetProduct(ctx, productID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if product.ID != productID {
		t.Fatalf("Expected product ID %d, got %d", productID, product.ID)
	}
	if product.Name != "Test Product" {
		t.Fatalf("Expected product name 'Test Product', got %s", product.Name)
	}
	if product.Description != "Test Description" {
		t.Fatalf("Expected description 'Test Description', got %s", product.Description)
	}
	if product.NetPriceCents != 599 {
		t.Fatalf("Expected net price 599, got %d", product.NetPriceCents)
	}
	if product.Status != InactiveStatus {
		t.Fatalf("Expected product to be inactive, got %s", product.Status)
	}
	if product.Category != FoodCategory {
		t.Fatalf("Expected product category 'food', got %s", product.Category)
	}
	if product.CreatedAt.IsZero() {
		t.Fatalf("Expected non-zero created_at, got %v", product.CreatedAt)
	}

	// Cleanup
	_, _ = db.ExecContext(ctx, "DELETE FROM products WHERE id = $1", productID)
}

func TestGetProduct_NotFound(t *testing.T) {
	db := database()
	defer db.Close()
	ctx := context.Background()

	persistence := &Persistence{DB: db}
	_, err := persistence.GetProduct(ctx, 999999)

	if err != dbpkg.ErrNotFound {
		t.Fatalf("Expected product not found error, got %v", err)
	}
}

func TestGetAllProducts(t *testing.T) {
	db := database()
	defer db.Close()
	ctx := context.Background()

	persistence := &Persistence{DB: db}
	productID1, _ := persistence.CreateProduct(ctx, "Product 1", "Description 1", 399, FoodCategory)
	productID2, _ := persistence.CreateProduct(ctx, "Product 2", "Description 2", 499, BeverageCategory)

	products, err := persistence.GetAllProducts(ctx)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(products) < 2 {
		t.Fatalf("Expected at least 2 products, got %d", len(products))
	}

	// Cleanup
	_, _ = db.ExecContext(ctx, "DELETE FROM products WHERE id = $1", productID1)
	_, _ = db.ExecContext(ctx, "DELETE FROM products WHERE id = $1", productID2)
}

func TestGetActiveProducts(t *testing.T) {
	db := database()
	defer db.Close()
	ctx := context.Background()

	persistence := &Persistence{DB: db}
	productID, _ := persistence.CreateProduct(ctx, "Active Product", "Active Description", 699, FoodCategory)
	_ = persistence.ActivateProduct(ctx, productID)

	products, err := persistence.GetActiveProducts(ctx)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(products) < 1 {
		t.Fatalf("Expected at least 1 active product, got %d", len(products))
	}

	// Cleanup
	_, _ = db.ExecContext(ctx, "DELETE FROM products WHERE id = $1", productID)
}

func TestUpdateProduct(t *testing.T) {
	db := database()
	defer db.Close()
	ctx := context.Background()

	persistence := &Persistence{DB: db}
	productID, _ := persistence.CreateProduct(ctx, "Original Product", "Original Description", 799, FoodCategory)

	err := persistence.UpdateProduct(ctx, productID, "Updated Name", "Updated Description", 999, BeverageCategory)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	updatedProduct, err := persistence.GetProduct(ctx, productID)
	if err != nil {
		t.Fatalf("Expected no error retrieving product, got %v", err)
	}
	if updatedProduct.Name != "Updated Name" {
		t.Fatalf("Expected product name 'Updated Name', got %s", updatedProduct.Name)
	}
	if updatedProduct.Description != "Updated Description" {
		t.Fatalf("Expected description 'Updated Description', got %s", updatedProduct.Description)
	}
	if updatedProduct.NetPriceCents != 999 {
		t.Fatalf("Expected net price 999, got %d", updatedProduct.NetPriceCents)
	}
	if updatedProduct.Category != BeverageCategory {
		t.Fatalf("Expected product category 'beverage', got %s", updatedProduct.Category)
	}

	// Cleanup
	_, _ = db.ExecContext(ctx, "DELETE FROM products WHERE id = $1", productID)
}

func TestUpdateProduct_NotFound(t *testing.T) {
	db := database()
	defer db.Close()
	ctx := context.Background()

	persistence := &Persistence{DB: db}
	err := persistence.UpdateProduct(ctx, 999999, "Updated Name", "Updated Description", 999, BeverageCategory)

	if err != dbpkg.ErrNotFound {
		t.Fatalf("Expected product not found error, got %v", err)
	}
}

func TestActivateProduct(t *testing.T) {
	db := database()
	defer db.Close()
	ctx := context.Background()

	persistence := &Persistence{DB: db}
	productID, _ := persistence.CreateProduct(ctx, "Inactive Product", "To be activated", 899, FoodCategory)

	err := persistence.ActivateProduct(ctx, productID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	product, _ := persistence.GetProduct(ctx, productID)
	if product.Status != ActiveStatus {
		t.Fatalf("Expected product status to be active, got %s", product.Status)
	}

	// Cleanup
	_, _ = db.ExecContext(ctx, "DELETE FROM products WHERE id = $1", productID)
}

func TestActivateProduct_NotFound(t *testing.T) {
	db := database()
	defer db.Close()
	ctx := context.Background()

	persistence := &Persistence{DB: db}
	err := persistence.ActivateProduct(ctx, 999999)

	if err != dbpkg.ErrNotFound {
		t.Fatalf("Expected product not found error, got %v", err)
	}
}

func TestDeactivateProduct(t *testing.T) {
	db := database()
	defer db.Close()
	ctx := context.Background()

	persistence := &Persistence{DB: db}
	productID, _ := persistence.CreateProduct(ctx, "Active Product", "To be deactivated", 1099, FoodCategory)
	_ = persistence.ActivateProduct(ctx, productID)

	err := persistence.DeactivateProduct(ctx, productID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	product, _ := persistence.GetProduct(ctx, productID)
	if product.Status != InactiveStatus {
		t.Fatalf("Expected product status to be inactive, got %s", product.Status)
	}

	// Cleanup
	_, _ = db.ExecContext(ctx, "DELETE FROM products WHERE id = $1", productID)
}

func TestDeactivateProduct_NotFound(t *testing.T) {
	db := database()
	defer db.Close()
	ctx := context.Background()

	persistence := &Persistence{DB: db}
	err := persistence.DeactivateProduct(ctx, 999999)

	if err != dbpkg.ErrNotFound {
		t.Fatalf("Expected product not found error, got %v", err)
	}
}
