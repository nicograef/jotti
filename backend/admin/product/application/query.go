package application

import (
	"context"

	"github.com/nicograef/jotti/backend/admin/product/domain"
	"github.com/rs/zerolog"
)

type productRepoQuery interface {
	GetAllProducts(ctx context.Context) ([]domain.Product, error)
}

type Query struct {
	ProductRepo productRepoQuery
}

// GetAllProducts retrieves all products.
func (q Query) GetAllProducts(ctx context.Context) ([]domain.Product, error) {
	log := zerolog.Ctx(ctx)

	products, err := q.ProductRepo.GetAllProducts(ctx)
	if err != nil {
		log.Error().Msg("Failed to retrieve all products")
		return nil, ErrDatabase
	}

	log.Info().Int("count", len(products)).Msg("Retrieved all products")
	return products, nil
}
