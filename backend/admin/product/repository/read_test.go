//go:build integration

package repository

import (
	"context"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/nicograef/jotti/backend/admin/product/domain"
	"github.com/nicograef/jotti/backend/db"
)

func TestGetAllProducts(t *testing.T) {
	db := db.OpenTestDatabase()
	defer db.Close()

	ctx := context.Background()
	repo := &Repository{DB: db}
	_, _ = repo.CreateProduct(ctx, domain.Product{Name: "Product 1", Description: "Description 1", NetPriceCents: 399, Category: domain.FoodCategory})
	_, _ = repo.CreateProduct(ctx, domain.Product{Name: "Product 2", Description: "Description 2", NetPriceCents: 499, Category: domain.BeverageCategory})

	products, err := repo.GetAllProducts(ctx)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(products) < 2 {
		t.Fatalf("Expected at least 2 products, got %d", len(products))
	}

	// Cleanup
	_, _ = db.ExecContext(ctx, "DELETE FROM products")

}
