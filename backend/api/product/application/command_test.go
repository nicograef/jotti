//go:build unit

package application

import (
	"context"
	"testing"
	"time"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/product"
	"github.com/nicograef/jotti/backend/domain/steuer"
	"github.com/nicograef/jotti/backend/repository/product_repo"
)

var testProduct = product.Produkt{
	ID:         1,
	Name:       "Cola",
	Kategorie:  product.GetraenkKategorie,
	Steuersatz: steuer.RegelSteuersatz,
	Status:     product.ActiveStatus,
	Varianten:  []product.Variante{},
	CreatedAt:  time.Now().UTC(),
	UpdatedAt:  time.Now().UTC(),
}

func TestCreateProduct(t *testing.T) {
	repo := product_repo.NewMock(nil, nil)
	cmd := Command{ProductRepo: repo}

	id, err := cmd.CreateProduct(context.Background(), "Bier", product.GetraenkKategorie, steuer.RegelSteuersatz)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id != 1 {
		t.Errorf("expected product ID 1, got %d", id)
	}
}

func TestCreateProduct_AlreadyExists(t *testing.T) {
	repo := product_repo.NewMock(nil, db.ErrAlreadyExists)
	cmd := Command{ProductRepo: repo}

	_, err := cmd.CreateProduct(context.Background(), "Bier", product.GetraenkKategorie, steuer.RegelSteuersatz)
	if err != ErrProduktAlreadyExists {
		t.Fatalf("expected ErrProduktAlreadyExists, got %v", err)
	}
}

func TestUpdateProduct(t *testing.T) {
	repo := product_repo.NewMock([]product.Produkt{testProduct}, nil)
	cmd := Command{ProductRepo: repo}

	err := cmd.UpdateProduct(context.Background(), 1, "Fanta", product.GetraenkKategorie, steuer.RegelSteuersatz)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	updated, err := repo.GetProduct(context.Background(), 1)
	if err != nil {
		t.Fatalf("failed to get updated product: %v", err)
	}
	if updated.Name != "Fanta" {
		t.Errorf("expected product name 'Fanta', got '%s'", updated.Name)
	}
}

func TestUpdateProduct_NotFound(t *testing.T) {
	repo := product_repo.NewMock(nil, db.ErrNotFound)
	cmd := Command{ProductRepo: repo}

	err := cmd.UpdateProduct(context.Background(), 999, "Fanta", product.GetraenkKategorie, steuer.RegelSteuersatz)
	if err != ErrProduktNotFound {
		t.Fatalf("expected ErrProduktNotFound, got %v", err)
	}
}
