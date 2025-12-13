package repository

import (
	"context"

	"github.com/nicograef/jotti/backend/admin/product/domain"
	"github.com/nicograef/jotti/backend/db"
	"github.com/rs/zerolog"
)

// CreateProduct inserts a new product into the database.
func (r Repository) CreateProduct(ctx context.Context, p domain.Product) (int, error) {
	log := zerolog.Ctx(ctx)

	var id int
	err := r.DB.QueryRowContext(ctx,
		"INSERT INTO products (name, description, net_price_cents, category) VALUES ($1, $2, $3, $4) RETURNING id",
		p.Name, p.Description, p.NetPriceCents, string(p.Category),
	).Scan(&id)

	if err != nil {
		log.Error().Err(err).Str("product_name", p.Name).Msg("DB Error creating product")
		return 0, db.Error(err)
	}

	return id, nil
}

// UpdateProduct updates an existing product in the database.
func (r Repository) UpdateProduct(ctx context.Context, p domain.Product) error {
	log := zerolog.Ctx(ctx)

	result, err := r.DB.ExecContext(ctx,
		"UPDATE products SET name = $1, description = $2, net_price_cents = $3, category = $4 WHERE id = $5",
		p.Name, p.Description, p.NetPriceCents, string(p.Category), p.ID,
	)
	if err != nil {
		log.Error().Err(err).Int("product_id", p.ID).Msg("DB Error updating product")
		return db.Error(err)
	}

	return db.ResultError(result)
}
