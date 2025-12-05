package product

import (
	"context"
	"errors"

	"github.com/nicograef/jotti/backend/db"
	"github.com/rs/zerolog"
)

type queryPersistence interface {
	GetProduct(ctx context.Context, id int) (*Product, error)
	GetAllProducts(ctx context.Context) ([]Product, error)
	GetActiveProducts(ctx context.Context) ([]ProductPublic, error)
}

type Query struct {
	Persistence queryPersistence
}

// GetProduct retrieves a product by its ID.
func (q *Query) GetProduct(ctx context.Context, id int) (*Product, error) {
	log := zerolog.Ctx(ctx)

	product, err := q.Persistence.GetProduct(ctx, id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Int("product_id", id).Msg("Product not found")
			return nil, ErrProductNotFound
		} else {
			log.Error().Int("product_id", id).Msg("Failed to retrieve product")
			return nil, ErrDatabase
		}
	}

	log.Debug().Int("product_id", id).Msg("Product retrieved")
	return product, nil
}

// GetAllProducts retrieves all products.
func (q *Query) GetAllProducts(ctx context.Context) ([]Product, error) {
	log := zerolog.Ctx(ctx)

	products, err := q.Persistence.GetAllProducts(ctx)
	if err != nil {
		log.Error().Msg("Failed to retrieve all products")
		return nil, ErrDatabase
	}

	log.Debug().Int("count", len(products)).Msg("Retrieved all products")
	return products, nil
}

// GetActiveProducts retrieves active products.
func (q *Query) GetActiveProducts(ctx context.Context) ([]ProductPublic, error) {
	log := zerolog.Ctx(ctx)

	products, err := q.Persistence.GetActiveProducts(ctx)
	if err != nil {
		log.Error().Msg("Failed to retrieve active products")
		return nil, ErrDatabase
	}

	log.Debug().Int("count", len(products)).Msg("Retrieved active products")
	return products, nil
}
