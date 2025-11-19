//go:build integration

package product

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

func database() *sql.DB {
	db, err := sql.Open("postgres", "host=localhost port=5432 user=admin password=admin dbname=jotti sslmode=disable")
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

func TestCreateProduct(t *testing.T) {
	db := database()
	defer db.Close()

	persistence := &Persistence{DB: db}
	productID, err := persistence.CreateProduct("French Fries", "The best fries in town", 4.99, FoodCategory)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if productID != 1 {
		t.Fatalf("Expected valid product ID, got %d", productID)
	}

}

func TestGetProduct(t *testing.T) {
	db := database()
	defer db.Close()

	persistence := &Persistence{DB: db}
	product, err := persistence.GetProduct(1)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if product.ID != 1 {
		t.Fatalf("Expected product ID 1, got %d", product.ID)
	}
	if product.Name != "French Fries" {
		t.Fatalf("Expected productname 'French Fries', got %s", product.Name)
	}
	if product.Description != "The best fries in town" {
		t.Fatalf("Expected description 'The best fries in town', got %s", product.Description)
	}
	if product.NetPrice != 4.99 {
		t.Fatalf("Expected net price 4.99, got %f", product.NetPrice)
	}
	if product.Status != InactiveStatus {
		t.Fatalf("Expected product to be active, got %s", product.Status)
	}
	if product.Category != FoodCategory {
		t.Fatalf("Expected product category 'food', got %s", product.Category)
	}
	if product.CreatedAt.IsZero() {
		t.Fatalf("Expected non-zero created_at, got %v", product.CreatedAt)
	}
}

func TestGetProduct_Error(t *testing.T) {
	db := database()
	defer db.Close()

	persistence := &Persistence{DB: db}
	_, err := persistence.GetProduct(100000)

	if err != ErrProductNotFound {
		t.Fatalf("Expected product not found error, got %v", err)
	}
}

func TestGetAllProducts(t *testing.T) {
	db := database()
	defer db.Close()

	persistence := &Persistence{DB: db}
	products, err := persistence.GetAllProducts()

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(products) == 0 {
		t.Fatalf("Expected at least one product, got %d", len(products))
	}
}

func TestUpdateProduct(t *testing.T) {
	db := database()
	defer db.Close()

	persistence := &Persistence{DB: db}
	err := persistence.UpdateProduct(1, "Updated Name", "updatedproductname", 9.99, BeverageCategory)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	updatedProduct, err := persistence.GetProduct(1)
	if err != nil {
		t.Fatalf("Expected no error retrieving product, got %v", err)
	}
	if updatedProduct.Name != "Updated Name" {
		t.Fatalf("Expected product name 'Updated Name', got %s", updatedProduct.Name)
	}
	if updatedProduct.Description != "updatedproductname" {
		t.Fatalf("Expected description 'updatedproductname', got %s", updatedProduct.Description)
	}
	if updatedProduct.NetPrice != 9.99 {
		t.Fatalf("Expected net price 9.99, got %f", updatedProduct.NetPrice)
	}
	if updatedProduct.Category != BeverageCategory {
		t.Fatalf("Expected product category 'beverage', got %s", updatedProduct.Category)
	}
}

func TestUpdateProduct_Error(t *testing.T) {
	db := database()
	defer db.Close()

	persistence := &Persistence{DB: db}
	err := persistence.UpdateProduct(100000, "Updated Name", "updatedproductname", 9.99, BeverageCategory)

	if err != ErrProductNotFound {
		t.Fatalf("Expected product not found error, got %v", err)
	}
}
