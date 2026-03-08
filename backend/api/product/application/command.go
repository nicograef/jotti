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
	GetVariant(ctx context.Context, variantID int) (product.Variant, error)
	CreateVariant(ctx context.Context, productID int, variant product.Variant) (int, error)
	UpdateVariant(ctx context.Context, variant product.Variant) error
}

type Command struct {
	ProductRepo commandProductRepo
}

// Product commands

func (c Command) CreateProduct(ctx context.Context, name string, category product.Category) (int, error) {
	log := zerolog.Ctx(ctx)

	product, err := product.NewProduct(name, category)
	if err != nil {
		log.Warn().Err(err).Str("product_name", name).Msg("Invalid product data")
		return 0, ErrInvalidProduktData
	}

	productID, err := c.ProductRepo.CreateProduct(ctx, product)
	if err != nil {
		if errors.Is(err, db.ErrAlreadyExists) {
			log.Warn().Err(err).Str("name", product.Name).Msg("Product name already exists")
			return 0, ErrProduktAlreadyExists
		} else {
			log.Error().Str("name", product.Name).Msg("Failed to create product")
			return 0, ErrDatabase
		}
	}

	log.Info().Int("product_id", productID).Msg("Product created")
	return productID, nil
}

func (c Command) UpdateProduct(ctx context.Context, productID int, name string, category product.Category) error {
	log := zerolog.Ctx(ctx)

	product, err := c.ProductRepo.GetProduct(ctx, productID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Int("product_id", productID).Msg("Product not found for update")
			return ErrProduktNotFound
		} else {
			log.Error().Int("product_id", productID).Msg("Failed to retrieve product for update")
			return ErrDatabase
		}
	}

	err = product.UpdateDetails(name, category)
	if err != nil {
		log.Warn().Err(err).Int("product_id", productID).Msg("Invalid product data for update")
		return ErrInvalidProduktData
	}

	err = c.ProductRepo.UpdateProduct(ctx, product)
	if err != nil {
		log.Error().Err(err).Int("product_id", productID).Msg("Failed to update product")
		return ErrDatabase
	}

	log.Info().Int("product_id", productID).Msg("Product updated")
	return nil
}

// Variant commands

func (c Command) CreateVariant(ctx context.Context, productID int, name string, priceCents int) (int, error) {
	log := zerolog.Ctx(ctx)

	// Verify product exists
	_, err := c.ProductRepo.GetProduct(ctx, productID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Int("product_id", productID).Msg("Product not found for variant creation")
			return 0, ErrProduktNotFound
		} else {
			log.Error().Int("product_id", productID).Msg("Failed to retrieve product for variant creation")
			return 0, ErrDatabase
		}
	}

	variant, err := product.NewVariant(name, priceCents)
	if err != nil {
		log.Warn().Err(err).Str("variant_name", name).Msg("Invalid variant data")
		return 0, ErrInvalidVarianteData
	}

	variantID, err := c.ProductRepo.CreateVariant(ctx, productID, variant)
	if err != nil {
		log.Error().Int("product_id", productID).Str("name", variant.Name).Msg("Failed to create variant")
		return 0, ErrDatabase
	}

	log.Info().Int("variant_id", variantID).Int("product_id", productID).Msg("Variant created")
	return variantID, nil
}

func (c Command) UpdateVariant(ctx context.Context, variantID int, name string, priceCents int) error {
	log := zerolog.Ctx(ctx)

	variant, err := c.ProductRepo.GetVariant(ctx, variantID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Int("variant_id", variantID).Msg("Variant not found for update")
			return ErrVarianteNotFound
		} else {
			log.Error().Int("variant_id", variantID).Msg("Failed to retrieve variant for update")
			return ErrDatabase
		}
	}

	err = variant.UpdateDetails(name, priceCents)
	if err != nil {
		log.Warn().Err(err).Int("variant_id", variantID).Msg("Invalid variant data for update")
		return ErrInvalidVarianteData
	}

	err = c.ProductRepo.UpdateVariant(ctx, variant)
	if err != nil {
		log.Error().Err(err).Int("variant_id", variantID).Msg("Failed to update variant")
		return ErrDatabase
	}

	log.Info().Int("variant_id", variantID).Msg("Variant updated")
	return nil
}

func (c Command) ActivateVariant(ctx context.Context, variantID int) error {
	log := zerolog.Ctx(ctx)

	variant, err := c.ProductRepo.GetVariant(ctx, variantID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Int("variant_id", variantID).Msg("Variant not found for activation")
			return ErrVarianteNotFound
		} else {
			log.Error().Int("variant_id", variantID).Msg("Failed to retrieve variant for activation")
			return ErrDatabase
		}
	}

	variant.Activate()

	err = c.ProductRepo.UpdateVariant(ctx, variant)
	if err != nil {
		log.Error().Err(err).Int("variant_id", variantID).Msg("Failed to update variant")
		return ErrDatabase
	}

	log.Info().Int("variant_id", variantID).Msg("Variant activated")
	return nil
}

func (c Command) DeactivateVariant(ctx context.Context, variantID int) error {
	log := zerolog.Ctx(ctx)

	variant, err := c.ProductRepo.GetVariant(ctx, variantID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Int("variant_id", variantID).Msg("Variant not found for deactivation")
			return ErrVarianteNotFound
		} else {
			log.Error().Int("variant_id", variantID).Msg("Failed to retrieve variant for deactivation")
			return ErrDatabase
		}
	}

	variant.Deactivate()

	err = c.ProductRepo.UpdateVariant(ctx, variant)
	if err != nil {
		log.Error().Err(err).Int("variant_id", variantID).Msg("Failed to update variant")
		return ErrDatabase
	}

	log.Info().Int("variant_id", variantID).Msg("Variant deactivated")
	return nil
}
