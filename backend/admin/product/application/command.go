package application

import (
	"context"
	"errors"

	"github.com/nicograef/jotti/backend/admin/product/domain"
	"github.com/nicograef/jotti/backend/db"
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
		return 0, err
	}

	id, err := c.ProductRepo.CreateProduct(ctx, product)
	if err != nil {
		if errors.Is(err, db.ErrAlreadyExists) {
			log.Warn().Str("product_name", name).Msg("Product name already exists")
			return 0, ErrProductAlreadyExists
		} else {
			log.Error().Str("product_name", name).Msg("Failed to create product")
			return 0, ErrDatabase
		}
	}

	log.Info().Int("product_id", id).Msg("Product created")
	return id, nil
}

// UpdateProduct updates an existing product in the database.
func (c Command) UpdateProduct(ctx context.Context, id int, name, description string, netPriceCents int, category domain.Category) error {
	log := zerolog.Ctx(ctx)

	product, err := c.ProductRepo.GetProduct(ctx, id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Int("product_id", id).Msg("Product not found for update")
			return ErrProductNotFound
		} else {
			log.Error().Int("product_id", id).Msg("Failed to retrieve product for update")
			return ErrDatabase
		}
	}

	if err = product.Update(name, description, netPriceCents, category); err != nil {
		log.Warn().Err(err).Int("product_id", id).Msg("Invalid product data for update")
		return ErrInvalidProductData
	}

	err = c.ProductRepo.UpdateProduct(ctx, product)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Int("product_id", id).Msg("Product not found for update")
			return ErrProductNotFound
		} else {
			log.Error().Int("product_id", id).Msg("Failed to update product")
			return ErrDatabase
		}
	}

	log.Info().Int("product_id", id).Msg("Product updated")
	return nil
}

// ActivateProduct sets the status of a product to active.
func (c Command) ActivateProduct(ctx context.Context, id int) error {
	log := zerolog.Ctx(ctx)

	product, err := c.ProductRepo.GetProduct(ctx, id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Int("product_id", id).Msg("Product not found for update")
			return ErrProductNotFound
		} else {
			log.Error().Int("product_id", id).Msg("Failed to retrieve product for update")
			return ErrDatabase
		}
	}

	product.Activate()

	err = c.ProductRepo.UpdateProduct(ctx, product)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Int("product_id", id).Msg("Product not found for activation")
			return ErrProductNotFound
		} else {
			log.Error().Int("product_id", id).Msg("Failed to activate product")
			return ErrDatabase
		}
	}

	log.Info().Int("product_id", id).Msg("Product activated")
	return nil
}

// DeactivateProduct sets the status of a product to inactive.
func (c Command) DeactivateProduct(ctx context.Context, id int) error {
	log := zerolog.Ctx(ctx)

	product, err := c.ProductRepo.GetProduct(ctx, id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Int("product_id", id).Msg("Product not found for update")
			return ErrProductNotFound
		} else {
			log.Error().Int("product_id", id).Msg("Failed to retrieve product for update")
			return ErrDatabase
		}
	}

	product.Deactivate()

	err = c.ProductRepo.UpdateProduct(ctx, product)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Int("product_id", id).Msg("Product not found for deactivation")
			return ErrProductNotFound
		} else {
			log.Error().Int("product_id", id).Msg("Failed to deactivate product")
			return ErrDatabase
		}
	}

	log.Info().Int("product_id", id).Msg("Product deactivated")
	return nil
}
