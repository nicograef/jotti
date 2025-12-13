package application

import (
	"context"

	"github.com/nicograef/jotti/backend/domain/product"
	"github.com/rs/zerolog"
)

type productRepoQuery interface {
	GetAllProducts(ctx context.Context) ([]product.Product, error)
	GetActiveProducts(ctx context.Context) ([]product.Product, error)
}

type Query struct {
	ProductRepo productRepoQuery
}

func (q Query) GetAllProducts(ctx context.Context) ([]product.Product, error) {
	log := zerolog.Ctx(ctx)

	products, err := q.ProductRepo.GetAllProducts(ctx)
	if err != nil {
		log.Error().Msg("Failed to retrieve all products")
		return nil, ErrDatabase
	}

	log.Info().Int("count", len(products)).Msg("Retrieved all products")
	return products, nil
}

func (q Query) GetActiveProducts(ctx context.Context) ([]product.Product, error) {
	log := zerolog.Ctx(ctx)

	products, err := q.ProductRepo.GetActiveProducts(ctx)
	if err != nil {
		log.Error().Msg("Failed to retrieve active products")
		return nil, ErrDatabase
	}

	log.Info().Int("count", len(products)).Msg("Retrieved active products")
	return products, nil
}
