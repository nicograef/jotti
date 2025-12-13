//go:build integration

package repository

import (
	"context"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/nicograef/jotti/backend/admin/product/domain"
	"github.com/nicograef/jotti/backend/db"
	dbpkg "github.com/nicograef/jotti/backend/db"
)

func TestCreateProductInDB(t *testing.T) {
	db := db.OpenTestDatabase()
	defer db.Close()
	ctx := context.Background()

	repo := &Repository{DB: db}
	productID, err := repo.CreateProduct(ctx, domain.Product{Name: "French Fries", Description: "The best fries in town", NetPriceCents: 499, Category: domain.FoodCategory})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if productID < 1 {
		t.Fatalf("Expected valid product ID, got %d", productID)
	}

	// Cleanup
	_, _ = db.ExecContext(ctx, "DELETE FROM products")
}

func TestUpdateProduct(t *testing.T) {
	db := db.OpenTestDatabase()
	defer db.Close()
	ctx := context.Background()

	repo := &Repository{DB: db}
	productID, _ := repo.CreateProduct(ctx, domain.Product{Name: "Original Product", Description: "Original Description", NetPriceCents: 799, Category: domain.FoodCategory})

	err := repo.UpdateProduct(ctx, domain.Product{ID: productID, Name: "Updated Name", Description: "Updated Description", NetPriceCents: 999, Category: domain.BeverageCategory})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	products, err := repo.GetAllProducts(ctx)
	if err != nil {
		t.Fatalf("Expected no error retrieving product, got %v", err)
	}
	if products[0].Name != "Updated Name" {
		t.Fatalf("Expected product name 'Updated Name', got %s", products[0].Name)
	}
	if products[0].Description != "Updated Description" {
		t.Fatalf("Expected description 'Updated Description', got %s", products[0].Description)
	}
	if products[0].NetPriceCents != 999 {
		t.Fatalf("Expected net price 999, got %d", products[0].NetPriceCents)
	}
	if products[0].Category != domain.BeverageCategory {
		t.Fatalf("Expected product category 'beverage', got %s", products[0].Category)
	}

	// Cleanup
	_, _ = db.ExecContext(ctx, "DELETE FROM products")
}

func TestUpdateProduct_NotFound(t *testing.T) {
	db := db.OpenTestDatabase()
	defer db.Close()
	ctx := context.Background()

	repo := &Repository{DB: db}
	err := repo.UpdateProduct(ctx, domain.Product{ID: 999999, Name: "Updated Name", Description: "Updated Description", NetPriceCents: 999, Category: domain.BeverageCategory})

	if err != dbpkg.ErrNotFound {
		t.Fatalf("Expected product not found error, got %v", err)
	}
}
