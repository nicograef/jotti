package product_repo

import (
	"context"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/product"
)

func (r Repository) GetVariant(ctx context.Context, id int) (product.Variant, error) {
	row := r.DB.QueryRowContext(ctx,
		"SELECT id, name, price_cents, status, created_at FROM product_variants WHERE id = $1 AND status != 'deleted'",
		id,
	)

	var v dbvariant
	err := row.Scan(&v.ID, &v.Name, &v.PriceCents, &v.Status, &v.CreatedAt)
	if err != nil {
		return product.Variant{}, db.Error(err)
	}

	return v.toDomain(), nil
}

func (r Repository) CreateVariant(ctx context.Context, productID int, v product.Variant) (int, error) {
	var id int
	err := r.DB.QueryRowContext(ctx,
		"INSERT INTO product_variants (product_id, name, price_cents, status, created_at) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		productID, v.Name, v.PriceCents, string(v.Status), v.CreatedAt,
	).Scan(&id)
	if err != nil {
		return 0, db.Error(err)
	}

	return id, nil
}

func (r Repository) UpdateVariant(ctx context.Context, v product.Variant) error {
	result, err := r.DB.ExecContext(ctx,
		"UPDATE product_variants SET name = $1, price_cents = $2, status = $3 WHERE id = $4",
		v.Name, v.PriceCents, string(v.Status), v.ID,
	)
	if err != nil {
		return db.Error(err)
	}

	return db.ResultError(result)
}
