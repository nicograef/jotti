package product_repo

import (
	"context"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/product"
)

func (r Repository) GetProduct(ctx context.Context, id int) (product.Product, error) {
	row := r.DB.QueryRowContext(ctx,
		"SELECT id, name, description, net_price_cents, status, category, created_at FROM products WHERE id = $1 AND status != 'deleted'",
		id,
	)

	var p dbproduct
	err := row.Scan(&p.ID, &p.Name, &p.Description, &p.NetPriceCents, &p.Status, &p.Category, &p.CreatedAt)

	if err != nil {
		return product.Product{}, db.Error(err)
	}

	return p.toDomain(), nil
}

func (r Repository) GetAllProducts(ctx context.Context) ([]product.Product, error) {
	rows, err := r.DB.QueryContext(ctx, "SELECT id, name, description, net_price_cents, status, category, created_at FROM products WHERE status != 'deleted' ORDER BY id ASC")
	if err != nil {
		return nil, db.Error(err)
	}
	defer db.Close(rows, "products")

	products := []product.Product{}
	for rows.Next() {
		var p dbproduct
		err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.NetPriceCents, &p.Status, &p.Category, &p.CreatedAt)
		if err != nil {
			return nil, db.Error(err)
		}
		products = append(products, p.toDomain())
	}

	if err := rows.Err(); err != nil {
		return nil, db.Error(err)
	}

	return products, nil
}

func (r Repository) GetActiveProducts(ctx context.Context) ([]product.Product, error) {
	rows, err := r.DB.QueryContext(ctx, "SELECT id, name, description, net_price_cents, status, category, created_at FROM products WHERE status = 'active' ORDER BY id ASC")
	if err != nil {
		return nil, db.Error(err)
	}
	defer db.Close(rows, "products")

	products := []product.Product{}
	for rows.Next() {
		var p dbproduct
		err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.NetPriceCents, &p.Status, &p.Category, &p.CreatedAt)
		if err != nil {
			return nil, db.Error(err)
		}
		products = append(products, p.toDomain())
	}

	if err := rows.Err(); err != nil {
		return nil, db.Error(err)
	}

	return products, nil
}

func (r Repository) CreateProduct(ctx context.Context, p product.Product) (int, error) {
	var id int
	err := r.DB.QueryRowContext(ctx,
		"INSERT INTO products (name, description, net_price_cents, category, status, created_at) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id",
		p.Name, p.Description, p.NetPriceCents, string(p.Category), string(p.Status), p.CreatedAt,
	).Scan(&id)

	if err != nil {
		return 0, db.Error(err)
	}

	return id, nil
}

func (r Repository) UpdateProduct(ctx context.Context, p product.Product) error {
	result, err := r.DB.ExecContext(ctx,
		"UPDATE products SET name = $1, description = $2, net_price_cents = $3, category = $4, status = $5 WHERE id = $6",
		p.Name, p.Description, p.NetPriceCents, string(p.Category), string(p.Status), p.ID,
	)
	if err != nil {
		return db.Error(err)
	}

	return db.ResultError(result)
}
