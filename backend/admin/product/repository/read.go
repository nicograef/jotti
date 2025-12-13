package repository

import (
	"context"

	"github.com/nicograef/jotti/backend/admin/product/domain"
	"github.com/nicograef/jotti/backend/db"
	"github.com/rs/zerolog"
)

// GetProduct retrieves a product by its ID from the database.
func (r Repository) GetProduct(ctx context.Context, id int) (domain.Product, error) {
	log := zerolog.Ctx(ctx)

	var dbProduct dbproduct
	err := r.DB.QueryRowContext(ctx,
		"SELECT id, name, description, net_price_cents, status, category, created_at FROM products WHERE id = $1 AND status != 'deleted'",
		id,
	).Scan(&dbProduct.ID, &dbProduct.Name, &dbProduct.Description, &dbProduct.NetPriceCents, &dbProduct.Status, &dbProduct.Category, &dbProduct.CreatedAt)

	if err != nil {
		log.Error().Err(err).Int("product_id", id).Msg("DB Error retrieving product")
		return domain.Product{}, db.Error(err)
	}

	return dbProduct.toDomain(), nil
}

// GetAllProducts retrieves all products from the database.
func (r Repository) GetAllProducts(ctx context.Context) ([]domain.Product, error) {
	log := zerolog.Ctx(ctx)

	rows, err := r.DB.QueryContext(ctx, "SELECT id, name, description, net_price_cents, status, category, created_at FROM products WHERE status != 'deleted' ORDER BY id ASC")
	if err != nil {
		log.Error().Err(err).Msg("DB Error querying all products")
		return nil, db.Error(err)
	}
	defer db.Close(rows, "products", log)

	products := []domain.Product{}
	for rows.Next() {
		var dbProduct dbproduct
		if err := rows.Scan(&dbProduct.ID, &dbProduct.Name, &dbProduct.Description, &dbProduct.NetPriceCents, &dbProduct.Status, &dbProduct.Category, &dbProduct.CreatedAt); err != nil {
			log.Error().Err(err).Msg("DB Error scanning product row")
			return nil, db.Error(err)
		}

		products = append(products, dbProduct.toDomain())
	}

	if err := rows.Err(); err != nil {
		log.Error().Err(err).Msg("DB Error iterating over product rows")
		return nil, db.Error(err)
	}

	return products, nil
}
