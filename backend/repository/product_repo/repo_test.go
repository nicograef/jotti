//go:build integration

package product_repo

import (
	"context"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	dbpkg "github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/product"
)

func setup(t *testing.T) (Repository, func(t *testing.T)) {
	db := dbpkg.OpenTestDatabase()

	// Clean up in correct order due to foreign key constraints
	_, err := db.Exec("DELETE FROM product_variants")
	if err != nil {
		t.Fatalf("Failed to clean product_variants table: %v", err)
	}
	_, err = db.Exec("DELETE FROM products")
	if err != nil {
		t.Fatalf("Failed to clean products table: %v", err)
	}

	return NewRepository(db), func(t *testing.T) {
		_, err = db.Exec("DELETE FROM product_variants")
		if err != nil {
			t.Fatalf("Failed to clean product_variants table: %v", err)
		}
		_, err = db.Exec("DELETE FROM products")
		if err != nil {
			t.Fatalf("Failed to clean products table: %v", err)
		}

		db.Close()
	}
}

func newProduct(name string, category product.Category) product.Product {
	return product.Product{
		Name:      name,
		Category:  category,
		Variants:  []product.Variant{},
		CreatedAt: time.Now().UTC(),
	}
}

func newVariant(name string, priceCents int, status product.Status) product.Variant {
	return product.Variant{
		Name:       name,
		PriceCents: priceCents,
		Status:     status,
		CreatedAt:  time.Now().UTC(),
	}
}

func TestGetAllProducts(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	_, _ = repo.CreateProduct(ctx, newProduct("Product 1", product.FoodCategory))
	_, _ = repo.CreateProduct(ctx, newProduct("Product 2", product.BeverageCategory))

	products, err := repo.GetAllProducts(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(products) != 2 {
		t.Fatalf("Expected 2 products, got %d", len(products))
	}
}

func TestGetAllProducts_WithVariants(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	productID, _ := repo.CreateProduct(ctx, newProduct("Pizza", product.FoodCategory))
	_, _ = repo.CreateVariant(ctx, productID, newVariant("Small", 899, product.ActiveStatus))
	_, _ = repo.CreateVariant(ctx, productID, newVariant("Large", 1299, product.ActiveStatus))

	products, err := repo.GetAllProducts(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(products) != 1 {
		t.Fatalf("Expected 1 product, got %d", len(products))
	}
	if len(products[0].Variants) != 2 {
		t.Fatalf("Expected 2 variants, got %d", len(products[0].Variants))
	}
}

func TestGetActiveProducts(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()

	// Product with active variant
	product1ID, _ := repo.CreateProduct(ctx, newProduct("Product 1", product.FoodCategory))
	_, _ = repo.CreateVariant(ctx, product1ID, newVariant("Regular", 999, product.ActiveStatus))

	// Product with only inactive variant
	product2ID, _ := repo.CreateProduct(ctx, newProduct("Product 2", product.FoodCategory))
	_, _ = repo.CreateVariant(ctx, product2ID, newVariant("Regular", 999, product.InactiveStatus))

	// Product with no variants
	_, _ = repo.CreateProduct(ctx, newProduct("Product 3", product.FoodCategory))

	products, err := repo.GetActiveProducts(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(products) != 1 {
		t.Fatalf("Expected 1 product with active variants, got %d", len(products))
	}
	if products[0].Name != "Product 1" {
		t.Fatalf("Expected 'Product 1', got %s", products[0].Name)
	}
}

func TestGetProduct(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	productID, _ := repo.CreateProduct(ctx, newProduct("Burger", product.FoodCategory))
	_, _ = repo.CreateVariant(ctx, productID, newVariant("Single", 599, product.ActiveStatus))
	_, _ = repo.CreateVariant(ctx, productID, newVariant("Double", 899, product.ActiveStatus))

	p, err := repo.GetProduct(ctx, productID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if p.Name != "Burger" {
		t.Fatalf("Expected 'Burger', got %s", p.Name)
	}
	if len(p.Variants) != 2 {
		t.Fatalf("Expected 2 variants, got %d", len(p.Variants))
	}
}

func TestGetProduct_NotFound(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	_, err := repo.GetProduct(ctx, 999999)

	if err != dbpkg.ErrNotFound {
		t.Fatalf("Expected ErrNotFound, got %v", err)
	}
}

func TestCreateProduct(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	productID, err := repo.CreateProduct(ctx, newProduct("French Fries", product.FoodCategory))
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
	p := newProduct("Original Product", product.FoodCategory)
	productID, _ := repo.CreateProduct(ctx, p)

	p.ID = productID
	p.Name = "Updated Name"
	p.Category = product.BeverageCategory
	err := repo.UpdateProduct(ctx, p)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	updated, err := repo.GetProduct(ctx, productID)
	if err != nil {
		t.Fatalf("Expected no error retrieving product, got %v", err)
	}
	if updated.Name != "Updated Name" {
		t.Fatalf("Expected product name 'Updated Name', got %s", updated.Name)
	}
	if updated.Category != product.BeverageCategory {
		t.Fatalf("Expected product category 'beverage', got %s", updated.Category)
	}
}

func TestUpdateProduct_NotFound(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	err := repo.UpdateProduct(ctx, product.Product{ID: 999999, Name: "Updated Name", Category: product.BeverageCategory})

	if err != dbpkg.ErrNotFound {
		t.Fatalf("Expected ErrNotFound, got %v", err)
	}
}

// Variant tests

func TestCreateVariant(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	productID, _ := repo.CreateProduct(ctx, newProduct("Cola", product.BeverageCategory))

	variantID, err := repo.CreateVariant(ctx, productID, newVariant("0.5L", 299, product.ActiveStatus))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if variantID < 1 {
		t.Fatalf("Expected valid variant ID, got %d", variantID)
	}
}

func TestGetVariant(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	productID, _ := repo.CreateProduct(ctx, newProduct("Cola", product.BeverageCategory))
	variantID, _ := repo.CreateVariant(ctx, productID, newVariant("0.5L", 299, product.ActiveStatus))

	v, err := repo.GetVariant(ctx, variantID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if v.Name != "0.5L" {
		t.Fatalf("Expected '0.5L', got %s", v.Name)
	}
	if v.PriceCents != 299 {
		t.Fatalf("Expected price 299, got %d", v.PriceCents)
	}
	if v.Status != product.ActiveStatus {
		t.Fatalf("Expected status 'active', got %s", v.Status)
	}
}

func TestGetVariant_NotFound(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	_, err := repo.GetVariant(ctx, 999999)

	if err != dbpkg.ErrNotFound {
		t.Fatalf("Expected ErrNotFound, got %v", err)
	}
}

func TestUpdateVariant(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	productID, _ := repo.CreateProduct(ctx, newProduct("Cola", product.BeverageCategory))
	variantID, _ := repo.CreateVariant(ctx, productID, newVariant("0.5L", 299, product.ActiveStatus))

	v := product.Variant{
		ID:         variantID,
		Name:       "1.0L",
		PriceCents: 499,
		Status:     product.InactiveStatus,
	}
	err := repo.UpdateVariant(ctx, v)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	updated, _ := repo.GetVariant(ctx, variantID)
	if updated.Name != "1.0L" {
		t.Fatalf("Expected '1.0L', got %s", updated.Name)
	}
	if updated.PriceCents != 499 {
		t.Fatalf("Expected price 499, got %d", updated.PriceCents)
	}
	if updated.Status != product.InactiveStatus {
		t.Fatalf("Expected status 'inactive', got %s", updated.Status)
	}
}

func TestUpdateVariant_NotFound(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	err := repo.UpdateVariant(ctx, product.Variant{ID: 999999, Name: "Test", PriceCents: 100, Status: product.ActiveStatus})

	if err != dbpkg.ErrNotFound {
		t.Fatalf("Expected ErrNotFound, got %v", err)
	}
}

func TestDeletedVariantsNotReturned(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	productID, _ := repo.CreateProduct(ctx, newProduct("Pizza", product.FoodCategory))
	_, _ = repo.CreateVariant(ctx, productID, newVariant("Small", 899, product.ActiveStatus))
	deletedVariantID, _ := repo.CreateVariant(ctx, productID, newVariant("Large", 1299, product.ActiveStatus))

	// Simulate soft delete by updating status to 'deleted'
	_, _ = repo.DB.ExecContext(ctx, "UPDATE product_variants SET status = 'deleted' WHERE id = $1", deletedVariantID)

	p, err := repo.GetProduct(ctx, productID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(p.Variants) != 1 {
		t.Fatalf("Expected 1 variant (deleted should be excluded), got %d", len(p.Variants))
	}
	if p.Variants[0].Name != "Small" {
		t.Fatalf("Expected 'Small' variant, got %s", p.Variants[0].Name)
	}
}
