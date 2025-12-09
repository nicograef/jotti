package product_service

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
	ID            int    `db:"id"`
	Name          string `db:"name"`
	Description   string `db:"description"`
	NetPriceCents int    `db:"net_price_cents"`
	Category      string `db:"category"`
}

// GetAllProducts retrieves all products from the database.
func (p *Persistence) GetAllProducts(ctx context.Context) ([]Product, error) {
	log := zerolog.Ctx(ctx)

	rows, err := p.DB.QueryContext(ctx, "SELECT id, name, description, net_price_cents, category FROM products WHERE status = 'active'")
	if err != nil {
		log.Error().Err(err).Msg("DB Error querying all products")
		return nil, db.Error(err)
	}
	defer db.Close(rows, "products", log)

	products := []Product{}
	for rows.Next() {
		var dbProduct dbproduct
		if err := rows.Scan(&dbProduct.ID, &dbProduct.Name, &dbProduct.Description, &dbProduct.NetPriceCents, &dbProduct.Category); err != nil {
			log.Error().Err(err).Msg("DB Error scanning product row")
			return nil, db.Error(err)
		}

		products = append(products, Product{
			ID:            dbProduct.ID,
			Name:          dbProduct.Name,
			Description:   dbProduct.Description,
			NetPriceCents: dbProduct.NetPriceCents,
			Category:      Category(dbProduct.Category),
		})
	}

	if err := rows.Err(); err != nil {
		log.Error().Err(err).Msg("DB Error iterating over product rows")
		return nil, db.Error(err)
	}

	return products, nil
}
