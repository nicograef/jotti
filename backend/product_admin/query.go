package product_admin

import (
	"context"

	"github.com/rs/zerolog"
)

type queryPersistence interface {
	GetAllProducts(ctx context.Context) ([]Product, error)
}

type Query struct {
	Persistence queryPersistence
}

// GetAllProducts retrieves all products.
func (q *Query) GetAllProducts(ctx context.Context) ([]Product, error) {
	log := zerolog.Ctx(ctx)

	products, err := q.Persistence.GetAllProducts(ctx)
	if err != nil {
		log.Error().Msg("Failed to retrieve all products")
		return nil, ErrDatabase
	}

	log.Info().Int("count", len(products)).Msg("Retrieved all products")
	return products, nil
}
