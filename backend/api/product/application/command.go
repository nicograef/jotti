package application

import (
	"context"
	"errors"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/product"
	"github.com/rs/zerolog"
)

type commandProductRepo interface {
	GetProduct(ctx context.Context, productID int) (product.Product, error)
	CreateProduct(ctx context.Context, product product.Product) (int, error)
	UpdateProduct(ctx context.Context, product product.Product) error
}

type Command struct {
	ProductRepo commandProductRepo
}

func (c Command) CreateProduct(ctx context.Context, name, description string, netPriceCents int, category product.Category) (int, error) {
	log := zerolog.Ctx(ctx)

	product, err := product.NewProduct(name, description, netPriceCents, category)
	if err != nil {
		log.Warn().Err(err).Str("product_name", name).Msg("Invalid product data")
		return 0, ErrInvalidProductData
	}

	productID, err := c.ProductRepo.CreateProduct(ctx, product)
	if err != nil {
		if errors.Is(err, db.ErrAlreadyExists) {
			log.Warn().Err(err).Str("name", product.Name).Msg("Product name already exists")
			return 0, ErrProductAlreadyExists
		} else {
			log.Error().Str("name", product.Name).Msg("Failed to create product")
			return 0, ErrDatabase
		}
	}

	log.Info().Int("product_id", productID).Msg("Product created")
	return productID, nil
}

func (c Command) UpdateProduct(ctx context.Context, productID int, name, description string, netPriceCents int, category product.Category) error {
	log := zerolog.Ctx(ctx)

	product, err := c.ProductRepo.GetProduct(ctx, productID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Int("product_id", productID).Msg("Product not found for update")
			return ErrProductNotFound
		} else {
			log.Error().Int("product_id", productID).Msg("Failed to retrieve product for update")
			return ErrDatabase
		}
	}

	err = product.UpdateDetails(name, description, netPriceCents, category)
	if err != nil {
		log.Warn().Err(err).Int("product_id", productID).Msg("Invalid product data for update")
		return ErrInvalidProductData
	}

	err = c.ProductRepo.UpdateProduct(ctx, product)
	if err != nil {
		log.Error().Err(err).Int("product_id", productID).Msg("Failed to update product")
		return ErrDatabase
	}

	log.Info().Int("product_id", productID).Msg("Product updated")
	return nil
}

func (c Command) ActivateProduct(ctx context.Context, productID int) error {
	log := zerolog.Ctx(ctx)

	product, err := c.ProductRepo.GetProduct(ctx, productID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Int("product_id", productID).Msg("Product not found for activation")
			return ErrProductNotFound
		} else {
			log.Error().Int("product_id", productID).Msg("Failed to retrieve product for activation")
			return ErrDatabase
		}
	}

	product.Activate()

	err = c.ProductRepo.UpdateProduct(ctx, product)
	if err != nil {
		log.Error().Err(err).Int("product_id", productID).Msg("Failed to update product")
		return ErrDatabase
	}

	log.Info().Int("product_id", productID).Msg("Product activated")
	return nil
}

func (c Command) DeactivateProduct(ctx context.Context, productID int) error {
	log := zerolog.Ctx(ctx)

	product, err := c.ProductRepo.GetProduct(ctx, productID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Int("product_id", productID).Msg("Product not found for deactivation")
			return ErrProductNotFound
		} else {
			log.Error().Int("product_id", productID).Msg("Failed to retrieve product for activation")
			return ErrDatabase
		}
	}

	product.Deactivate()

	err = c.ProductRepo.UpdateProduct(ctx, product)
	if err != nil {
		log.Error().Err(err).Int("product_id", productID).Msg("Failed to update product")
		return ErrDatabase
	}

	log.Info().Int("product_id", productID).Msg("Product deactivated")
	return nil
}
