package repository

import (
	"context"

	"github.com/nicograef/jotti/backend/admin/product/domain"
	"github.com/nicograef/jotti/backend/db"
)

// GetProduct retrieves a product by its ID from the database.
func (r Repository) GetProduct(ctx context.Context, id int) (domain.Product, error) {
	var dbProduct dbproduct
	err := r.DB.QueryRowContext(ctx,
		"SELECT id, name, description, net_price_cents, status, category, created_at FROM products WHERE id = $1 AND status != 'deleted'",
		id,
	).Scan(&dbProduct.ID, &dbProduct.Name, &dbProduct.Description, &dbProduct.NetPriceCents, &dbProduct.Status, &dbProduct.Category, &dbProduct.CreatedAt)

	if err != nil {
		return domain.Product{}, db.Error(err)
	}

	return dbProduct.toDomain(), nil
}

// GetAllProducts retrieves all products from the database.
func (r Repository) GetAllProducts(ctx context.Context) ([]domain.Product, error) {
	rows, err := r.DB.QueryContext(ctx, "SELECT id, name, description, net_price_cents, status, category, created_at FROM products WHERE status != 'deleted' ORDER BY id ASC")
	if err != nil {
		return nil, db.Error(err)
	}
	defer db.Close(rows, "products")

	products := []domain.Product{}
	for rows.Next() {
		var dbProduct dbproduct
		if err := rows.Scan(&dbProduct.ID, &dbProduct.Name, &dbProduct.Description, &dbProduct.NetPriceCents, &dbProduct.Status, &dbProduct.Category, &dbProduct.CreatedAt); err != nil {
			return nil, db.Error(err)
		}
		products = append(products, dbProduct.toDomain())
	}

	if err := rows.Err(); err != nil {
		return nil, db.Error(err)
	}

	return products, nil
}

// CreateProduct inserts a new product into the database.
func (r Repository) CreateProduct(ctx context.Context, p domain.Product) (int, error) {
	var id int
	err := r.DB.QueryRowContext(ctx,
		"INSERT INTO products (name, description, net_price_cents, category) VALUES ($1, $2, $3, $4) RETURNING id",
		p.Name, p.Description, p.NetPriceCents, string(p.Category),
	).Scan(&id)

	if err != nil {
		return 0, db.Error(err)
	}

	return id, nil
}

// UpdateProduct updates an existing product in the database.
func (r Repository) UpdateProduct(ctx context.Context, p domain.Product) error {
	result, err := r.DB.ExecContext(ctx,
		"UPDATE products SET name = $1, description = $2, net_price_cents = $3, category = $4, status = $5 WHERE id = $6",
		p.Name, p.Description, p.NetPriceCents, string(p.Category), string(p.Status), p.ID,
	)
	if err != nil {
		return db.Error(err)
	}

	return db.ResultError(result)
}
