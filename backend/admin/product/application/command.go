package application

import (
	"context"

	"github.com/nicograef/jotti/backend/admin/product/domain"
	"github.com/rs/zerolog"
)

type productRepository interface {
	GetProduct(ctx context.Context, id int) (domain.Product, error)
	CreateProduct(ctx context.Context, product domain.Product) (int, error)
	UpdateProduct(ctx context.Context, product domain.Product) error
}

type Command struct {
	ProductRepo productRepository
}

// CreateProduct creates a new product in the database.
func (c Command) CreateProduct(ctx context.Context, name, description string, netPriceCents int, category domain.Category) (int, error) {
	log := zerolog.Ctx(ctx)

	product, err := domain.NewProduct(name, description, netPriceCents, category)
	if err != nil {
		log.Warn().Err(err).Str("product_name", name).Msg("Invalid product data")
		return 0, ErrInvalidProductData
	}

	id, err := c.ProductRepo.CreateProduct(ctx, product)
	if err != nil {
		return 0, fromRepositoryError(err, log, 0)
	}

	log.Info().Int("product_id", id).Msg("Product created")
	return id, nil
}

// UpdateProduct updates an existing product in the database.
func (c Command) UpdateProduct(ctx context.Context, id int, name, description string, netPriceCents int, category domain.Category) error {
	log := zerolog.Ctx(ctx)

	product, err := c.ProductRepo.GetProduct(ctx, id)
	if err != nil {
		return fromRepositoryError(err, log, id)

	}

	err = product.Update(name, description, netPriceCents, category)
	if err != nil {
		log.Warn().Err(err).Int("product_id", id).Msg("Invalid product data for update")
		return ErrInvalidProductData
	}

	err = c.ProductRepo.UpdateProduct(ctx, product)
	if err != nil {
		return fromRepositoryError(err, log, id)
	}

	log.Info().Int("product_id", id).Msg("Product updated")
	return nil
}

// ActivateProduct sets the status of a product to active.
func (c Command) ActivateProduct(ctx context.Context, id int) error {
	log := zerolog.Ctx(ctx)

	product, err := c.ProductRepo.GetProduct(ctx, id)
	if err != nil {
		return fromRepositoryError(err, log, id)
	}

	product.Activate()

	err = c.ProductRepo.UpdateProduct(ctx, product)
	if err != nil {
		return fromRepositoryError(err, log, id)
	}

	log.Info().Int("product_id", id).Msg("Product activated")
	return nil
}

// DeactivateProduct sets the status of a product to inactive.
func (c Command) DeactivateProduct(ctx context.Context, id int) error {
	log := zerolog.Ctx(ctx)

	product, err := c.ProductRepo.GetProduct(ctx, id)
	if err != nil {
		return fromRepositoryError(err, log, id)
	}

	product.Deactivate()

	err = c.ProductRepo.UpdateProduct(ctx, product)
	if err != nil {
		return fromRepositoryError(err, log, id)
	}

	log.Info().Int("product_id", id).Msg("Product deactivated")
	return nil
}
