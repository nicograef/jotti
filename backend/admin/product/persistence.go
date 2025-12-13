package product

import (
	"context"
	"database/sql"

	"github.com/nicograef/jotti/backend/db"
	"github.com/rs/zerolog"
)

// Persistence implements product persistence layer using a SQL database.
type Persistence struct {
	DB *sql.DB
}

type dbproduct struct {
	ID            int          `db:"id"`
	Name          string       `db:"name"`
	Description   string       `db:"description"`
	NetPriceCents int          `db:"net_price_cents"`
	Status        string       `db:"status"`
	Category      string       `db:"category"`
	CreatedAt     sql.NullTime `db:"created_at"`
}

// GetAllProducts retrieves all products from the database.
func (p *Persistence) GetAllProducts(ctx context.Context) ([]Product, error) {
	log := zerolog.Ctx(ctx)

	rows, err := p.DB.QueryContext(ctx, "SELECT id, name, description, net_price_cents, status, category, created_at FROM products WHERE status != 'deleted' ORDER BY id ASC")
	if err != nil {
		log.Error().Err(err).Msg("DB Error querying all products")
		return nil, db.Error(err)
	}
	defer db.Close(rows, "products", log)

	products := []Product{}
	for rows.Next() {
		var dbProduct dbproduct
		if err := rows.Scan(&dbProduct.ID, &dbProduct.Name, &dbProduct.Description, &dbProduct.NetPriceCents, &dbProduct.Status, &dbProduct.Category, &dbProduct.CreatedAt); err != nil {
			log.Error().Err(err).Msg("DB Error scanning product row")
			return nil, db.Error(err)
		}

		products = append(products, Product{
			ID:            dbProduct.ID,
			Name:          dbProduct.Name,
			Description:   dbProduct.Description,
			NetPriceCents: dbProduct.NetPriceCents,
			Status:        Status(dbProduct.Status),
			Category:      Category(dbProduct.Category),
			CreatedAt:     dbProduct.CreatedAt.Time,
		})
	}

	if err := rows.Err(); err != nil {
		log.Error().Err(err).Msg("DB Error iterating over product rows")
		return nil, db.Error(err)
	}

	return products, nil
}

// CreateProduct inserts a new product into the database.
func (p *Persistence) CreateProduct(ctx context.Context, name, description string, netPriceCents int, category Category) (int, error) {
	log := zerolog.Ctx(ctx)

	var id int
	err := p.DB.QueryRowContext(ctx,
		"INSERT INTO products (name, description, net_price_cents, category) VALUES ($1, $2, $3, $4) RETURNING id",
		name, description, netPriceCents, string(category),
	).Scan(&id)

	if err != nil {
		log.Error().Err(err).Str("product_name", name).Msg("DB Error creating product")
		return 0, db.Error(err)
	}

	return id, nil
}

// UpdateProduct updates an existing product in the database.
func (p *Persistence) UpdateProduct(ctx context.Context, id int, name, description string, netPriceCents int, category Category) error {
	log := zerolog.Ctx(ctx)

	result, err := p.DB.ExecContext(ctx,
		"UPDATE products SET name = $1, description = $2, net_price_cents = $3, category = $4 WHERE id = $5 AND status != 'deleted'",
		name, description, netPriceCents, string(category), id,
	)
	if err != nil {
		log.Error().Err(err).Int("product_id", id).Msg("DB Error updating product")
		return db.Error(err)
	}

	return db.ResultError(result)
}

// ActivateProduct sets the status of a product to active.
func (p *Persistence) ActivateProduct(ctx context.Context, id int) error {
	log := zerolog.Ctx(ctx)

	result, err := p.DB.ExecContext(ctx, "UPDATE products SET status = 'active' WHERE id = $1 AND status != 'deleted'", id)
	if err != nil {
		log.Error().Err(err).Int("product_id", id).Msg("DB Error activating product")
		return db.Error(err)
	}

	return db.ResultError(result)
}

// DeactivateProduct sets the status of a product to inactive.
func (p *Persistence) DeactivateProduct(ctx context.Context, id int) error {
	log := zerolog.Ctx(ctx)

	result, err := p.DB.ExecContext(ctx, "UPDATE products SET status = 'inactive' WHERE id = $1 AND status != 'deleted'", id)
	if err != nil {
		log.Error().Err(err).Int("product_id", id).Msg("DB Error deactivating product")
		return db.Error(err)
	}

	return db.ResultError(result)
}
