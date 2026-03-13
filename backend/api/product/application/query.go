package application

import (
	"context"

	"github.com/nicograef/jotti/backend/domain/product"
	"github.com/rs/zerolog"
)

type productQueryRepo interface {
	GetAllProducts(ctx context.Context) ([]product.Produkt, error)
	GetActiveProducts(ctx context.Context) ([]product.Produkt, error)
}

func GetAllProducts(ctx context.Context, repo productQueryRepo) ([]product.Produkt, error) {
	log := zerolog.Ctx(ctx)

	products, err := repo.GetAllProducts(ctx)
	if err != nil {
		log.Error().Msg("Failed to retrieve all products")
		return nil, ErrDatabase
	}

	log.Info().Int("count", len(products)).Msg("Retrieved all products")
	return products, nil
}

func GetActiveProducts(ctx context.Context, repo productQueryRepo) ([]product.Produkt, error) {
	log := zerolog.Ctx(ctx)

	products, err := repo.GetActiveProducts(ctx)
	if err != nil {
		log.Error().Msg("Failed to retrieve active products")
		return nil, ErrDatabase
	}

	log.Info().Int("count", len(products)).Msg("Retrieved active products")
	return products, nil
}
