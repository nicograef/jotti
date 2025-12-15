//go:build integration

package product_repo

import (
	"context"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	dbpkg "github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/product"
)

func setup(t *testing.T) (Repository, func(t *testing.T)) {
	db := dbpkg.OpenTestDatabase()

	_, err := db.Exec("DELETE FROM products")
	if err != nil {
		t.Fatalf("Failed to clean products table: %v", err)
	}

	return Repository{DB: db}, func(t *testing.T) {
		_, err = db.Exec("DELETE FROM products")
		if err != nil {
			t.Fatalf("Failed to clean products table: %v", err)
		}

		db.Close()
	}
}

func TestGetAllProducts(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	_, _ = repo.CreateProduct(ctx, product.Product{Name: "Product 1", Description: "Description 1", NetPriceCents: 399, Category: product.FoodCategory})
	_, _ = repo.CreateProduct(ctx, product.Product{Name: "Product 2", Description: "Description 2", NetPriceCents: 499, Category: product.BeverageCategory})

	products, err := repo.GetAllProducts(ctx)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(products) != 2 {
		t.Fatalf("Expected 2 products, got %d", len(products))
	}
}

func TestGetActiveProducts(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	_, _ = repo.CreateProduct(ctx, product.Product{Name: "Product 1", Description: "Description 1", NetPriceCents: 399, Category: product.FoodCategory, Status: product.ActiveStatus})
	_, _ = repo.CreateProduct(ctx, product.Product{Name: "Product 2", Description: "Description 2", NetPriceCents: 499, Category: product.BeverageCategory, Status: product.InactiveStatus})

	products, err := repo.GetActiveProducts(ctx)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(products) != 1 {
		t.Fatalf("Expected 1 product, got %d", len(products))
	}
}

func TestCreateProductInDB(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	productID, err := repo.CreateProduct(ctx, product.Product{Name: "French Fries", Description: "The best fries in town", NetPriceCents: 499, Category: product.FoodCategory})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if productID < 1 {
		t.Fatalf("Expected valid product ID, got %d", productID)
	}
}

func TestUpdateProduct(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	productID, _ := repo.CreateProduct(ctx, product.Product{Name: "Original Product", Description: "Original Description", NetPriceCents: 799, Category: product.FoodCategory})

	err := repo.UpdateProduct(ctx, product.Product{ID: productID, Name: "Updated Name", Description: "Updated Description", NetPriceCents: 999, Category: product.BeverageCategory})
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
	if products[0].Category != product.BeverageCategory {
		t.Fatalf("Expected product category 'beverage', got %s", products[0].Category)
	}
}

func TestUpdateProduct_NotFound(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	err := repo.UpdateProduct(ctx, product.Product{ID: 999999, Name: "Updated Name", Description: "Updated Description", NetPriceCents: 999, Category: product.BeverageCategory})

	if err != dbpkg.ErrNotFound {
		t.Fatalf("Expected product not found error, got %v", err)
	}
}
