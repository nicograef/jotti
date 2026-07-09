//go:build integration

package produkt_repo

import (
	"context"
	"errors"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	dbpkg "github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/produkt"
	"github.com/nicograef/jotti/backend/domain/steuer"
)

func setup(t *testing.T) (Repository, func(t *testing.T)) {
	db := dbpkg.OpenTestDatabase()

	// Clean up in correct order due to foreign key constraints
	_, err := db.Exec("DELETE FROM produkt_varianten")
	if err != nil {
		t.Fatalf("Failed to clean product_variants table: %v", err)
	}
	_, err = db.Exec("DELETE FROM produkte")
	if err != nil {
		t.Fatalf("Failed to clean products table: %v", err)
	}

	return NewRepository(db), func(t *testing.T) {
		_, err = db.Exec("DELETE FROM produkt_varianten")
		if err != nil {
			t.Fatalf("Failed to clean product_variants table: %v", err)
		}
		_, err = db.Exec("DELETE FROM produkte")
		if err != nil {
			t.Fatalf("Failed to clean products table: %v", err)
		}

		_ = db.Close()
	}
}

func newProduct(name string, kategorie produkt.Kategorie) produkt.Produkt {
	now := time.Now().UTC()
	steuersatz := steuer.RegelSteuersatz
	if kategorie == produkt.EssenKategorie {
		steuersatz = steuer.ErmaessigtSteuersatz
	}
	if kategorie == produkt.SonstigesKategorie {
		steuersatz = steuer.BefreitSteuersatz
	}

	return produkt.Produkt{
		Name:       name,
		Kategorie:  kategorie,
		Steuersatz: steuersatz,
		Status:     produkt.ActiveStatus,
		Varianten:  []produkt.Variante{},
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

func newVariant(name string, preisCents int, status produkt.Status) produkt.Variante {
	now := time.Now().UTC()
	return produkt.Variante{
		Name:       name,
		PreisCents: preisCents,
		Status:     status,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

func TestGetAllProducts(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	_, _ = repo.CreateProduct(ctx, newProduct("Product 1", produkt.EssenKategorie))
	_, _ = repo.CreateProduct(ctx, newProduct("Product 2", produkt.GetraenkKategorie))

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
	productID, _ := repo.CreateProduct(ctx, newProduct("Pizza", produkt.EssenKategorie))
	_, _ = repo.CreateVariant(ctx, productID, newVariant("Small", 899, produkt.ActiveStatus))
	_, _ = repo.CreateVariant(ctx, productID, newVariant("Large", 1299, produkt.ActiveStatus))

	products, err := repo.GetAllProducts(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(products) != 1 {
		t.Fatalf("Expected 1 product, got %d", len(products))
	}
	if len(products[0].Varianten) != 2 {
		t.Fatalf("Expected 2 variants, got %d", len(products[0].Varianten))
	}
}

func TestGetActiveProducts(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()

	// Product with active variant
	product1ID, _ := repo.CreateProduct(ctx, newProduct("Product 1", produkt.EssenKategorie))
	_, _ = repo.CreateVariant(ctx, product1ID, newVariant("Regular", 999, produkt.ActiveStatus))

	// Product with only inactive variant
	product2ID, _ := repo.CreateProduct(ctx, newProduct("Product 2", produkt.EssenKategorie))
	_, _ = repo.CreateVariant(ctx, product2ID, newVariant("Regular", 999, produkt.InactiveStatus))

	// Product with no variants
	_, _ = repo.CreateProduct(ctx, newProduct("Product 3", produkt.EssenKategorie))

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
	productID, _ := repo.CreateProduct(ctx, newProduct("Burger", produkt.EssenKategorie))
	_, _ = repo.CreateVariant(ctx, productID, newVariant("Single", 599, produkt.ActiveStatus))
	_, _ = repo.CreateVariant(ctx, productID, newVariant("Double", 899, produkt.ActiveStatus))

	p, err := repo.GetProduct(ctx, productID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if p.Name != "Burger" {
		t.Fatalf("Expected 'Burger', got %s", p.Name)
	}
	if len(p.Varianten) != 2 {
		t.Fatalf("Expected 2 variants, got %d", len(p.Varianten))
	}
}

func TestGetProduct_NotFound(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	_, err := repo.GetProduct(ctx, 999999)

	if !errors.Is(err, dbpkg.ErrNotFound) {
		t.Fatalf("Expected ErrNotFound, got %v", err)
	}
}

func TestCreateProduct(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	productID, err := repo.CreateProduct(ctx, newProduct("French Fries", produkt.EssenKategorie))
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
	p := newProduct("Original Product", produkt.EssenKategorie)
	productID, _ := repo.CreateProduct(ctx, p)

	p.ID = productID
	p.Name = "Updated Name"
	p.Kategorie = produkt.GetraenkKategorie
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
	if updated.Kategorie != produkt.GetraenkKategorie {
		t.Fatalf("Expected product category 'getraenk', got %s", updated.Kategorie)
	}
}

func TestUpdateProduct_NotFound(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	err := repo.UpdateProduct(ctx, produkt.Produkt{ID: 999999, Name: "Updated Name", Kategorie: produkt.GetraenkKategorie, Steuersatz: steuer.RegelSteuersatz, Status: produkt.ActiveStatus, UpdatedAt: time.Now().UTC()})

	if !errors.Is(err, dbpkg.ErrNotFound) {
		t.Fatalf("Expected ErrNotFound, got %v", err)
	}
}

// Variant tests

func TestCreateVariant(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	productID, _ := repo.CreateProduct(ctx, newProduct("Cola", produkt.GetraenkKategorie))

	variantID, err := repo.CreateVariant(ctx, productID, newVariant("0.5L", 299, produkt.ActiveStatus))
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
	productID, _ := repo.CreateProduct(ctx, newProduct("Cola", produkt.GetraenkKategorie))
	variantID, _ := repo.CreateVariant(ctx, productID, newVariant("0.5L", 299, produkt.ActiveStatus))

	v, err := repo.GetVariant(ctx, variantID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if v.Name != "0.5L" {
		t.Fatalf("Expected '0.5L', got %s", v.Name)
	}
	if v.PreisCents != 299 {
		t.Fatalf("Expected price 299, got %d", v.PreisCents)
	}
	if v.Status != produkt.ActiveStatus {
		t.Fatalf("Expected status 'active', got %s", v.Status)
	}
}

func TestGetVariant_NotFound(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	_, err := repo.GetVariant(ctx, 999999)

	if !errors.Is(err, dbpkg.ErrNotFound) {
		t.Fatalf("Expected ErrNotFound, got %v", err)
	}
}

func TestUpdateVariant(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	productID, _ := repo.CreateProduct(ctx, newProduct("Cola", produkt.GetraenkKategorie))
	variantID, _ := repo.CreateVariant(ctx, productID, newVariant("0.5L", 299, produkt.ActiveStatus))

	v := produkt.Variante{
		ID:         variantID,
		Name:       "1.0L",
		PreisCents: 499,
		Status:     produkt.InactiveStatus,
	}
	err := repo.UpdateVariant(ctx, v)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	updated, _ := repo.GetVariant(ctx, variantID)
	if updated.Name != "1.0L" {
		t.Fatalf("Expected '1.0L', got %s", updated.Name)
	}
	if updated.PreisCents != 499 {
		t.Fatalf("Expected price 499, got %d", updated.PreisCents)
	}
	if updated.Status != produkt.InactiveStatus {
		t.Fatalf("Expected status 'inactive', got %s", updated.Status)
	}
}

func TestUpdateVariant_NotFound(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	err := repo.UpdateVariant(ctx, produkt.Variante{ID: 999999, Name: "Test", PreisCents: 100, Status: produkt.ActiveStatus})

	if !errors.Is(err, dbpkg.ErrNotFound) {
		t.Fatalf("Expected ErrNotFound, got %v", err)
	}
}

func TestDeletedVariantsNotReturned(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	productID, _ := repo.CreateProduct(ctx, newProduct("Pizza", produkt.EssenKategorie))
	_, _ = repo.CreateVariant(ctx, productID, newVariant("Small", 899, produkt.ActiveStatus))
	deletedVariantID, _ := repo.CreateVariant(ctx, productID, newVariant("Large", 1299, produkt.ActiveStatus))

	// Simulate soft delete by updating status to 'deleted'
	_, _ = repo.db.ExecContext(ctx, "UPDATE produkt_varianten SET status = 'deleted' WHERE id = $1", deletedVariantID)

	p, err := repo.GetProduct(ctx, productID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(p.Varianten) != 1 {
		t.Fatalf("Expected 1 variant (deleted should be excluded), got %d", len(p.Varianten))
	}
	if p.Varianten[0].Name != "Small" {
		t.Fatalf("Expected 'Small' variant, got %s", p.Varianten[0].Name)
	}
}

func TestVariantOrderStableAfterUpdate(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	productID, _ := repo.CreateProduct(ctx, newProduct("Pizza", produkt.EssenKategorie))

	firstID, _ := repo.CreateVariant(ctx, productID, newVariant("Klein", 899, produkt.ActiveStatus))
	middleID, _ := repo.CreateVariant(ctx, productID, newVariant("Mittel", 1099, produkt.ActiveStatus))
	lastID, _ := repo.CreateVariant(ctx, productID, newVariant("Groß", 1299, produkt.ActiveStatus))

	// Updating the middle variant must not change its position in the aggregated
	// variant list — without ORDER BY in the json_agg, Postgres may reorder rows.
	err := repo.UpdateVariant(ctx, produkt.Variante{
		ID:         middleID,
		Name:       "Mittel neu",
		PreisCents: 1150,
		Status:     produkt.ActiveStatus,
	})
	if err != nil {
		t.Fatalf("Expected no error updating variant, got %v", err)
	}

	wantOrder := []int{firstID, middleID, lastID}

	p, err := repo.GetProduct(ctx, productID)
	if err != nil {
		t.Fatalf("Expected no error from GetProduct, got %v", err)
	}
	assertVariantOrder(t, "GetProduct", p.Varianten, wantOrder)

	all, err := repo.GetAllProducts(ctx)
	if err != nil {
		t.Fatalf("Expected no error from GetAllProducts, got %v", err)
	}
	assertVariantOrder(t, "GetAllProducts", all[0].Varianten, wantOrder)

	active, err := repo.GetActiveProducts(ctx)
	if err != nil {
		t.Fatalf("Expected no error from GetActiveProducts, got %v", err)
	}
	assertVariantOrder(t, "GetActiveProducts", active[0].Varianten, wantOrder)
}

func assertVariantOrder(t *testing.T, query string, varianten []produkt.Variante, wantIDs []int) {
	t.Helper()
	if len(varianten) != len(wantIDs) {
		t.Fatalf("%s: expected %d variants, got %d", query, len(wantIDs), len(varianten))
	}
	for i, wantID := range wantIDs {
		if varianten[i].ID != wantID {
			t.Fatalf("%s: expected variant ID %d at position %d, got %d", query, wantID, i, varianten[i].ID)
		}
	}
}
