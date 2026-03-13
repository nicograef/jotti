package application

import (
	"context"
	"errors"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/product"
	"github.com/rs/zerolog"
)

type productRepo interface {
	GetProduct(ctx context.Context, productID int) (product.Produkt, error)
	CreateProduct(ctx context.Context, product product.Produkt) (int, error)
	UpdateProduct(ctx context.Context, product product.Produkt) error
	GetVariant(ctx context.Context, variantID int) (product.Variante, error)
	CreateVariant(ctx context.Context, productID int, variant product.Variante) (int, error)
	UpdateVariant(ctx context.Context, variant product.Variante) error
	GetAllProducts(ctx context.Context) ([]product.Produkt, error)
	GetActiveProducts(ctx context.Context) ([]product.Produkt, error)
}

type Command struct {
	ProductRepo productRepo
}

// Product commands

func (c Command) CreateProduct(ctx context.Context, name string, kategorie product.Kategorie) (int, error) {
	log := zerolog.Ctx(ctx)

	produkt, err := product.NewProdukt(name, kategorie)
	if err != nil {
		log.Warn().Err(err).Str("product_name", name).Msg("Invalid product data")
		return 0, ErrInvalidProduktData
	}

	productID, err := c.ProductRepo.CreateProduct(ctx, produkt)
	if err != nil {
		if errors.Is(err, db.ErrAlreadyExists) {
			log.Warn().Err(err).Str("name", produkt.Name).Msg("Product name already exists")
			return 0, ErrProduktAlreadyExists
		} else {
			log.Error().Str("name", produkt.Name).Msg("Failed to create product")
			return 0, ErrDatabase
		}
	}

	log.Info().Int("product_id", productID).Msg("Product created")
	return productID, nil
}

func (c Command) UpdateProduct(ctx context.Context, productID int, name string, kategorie product.Kategorie) error {
	log := zerolog.Ctx(ctx)

	produkt, err := c.ProductRepo.GetProduct(ctx, productID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Int("product_id", productID).Msg("Product not found for update")
			return ErrProduktNotFound
		} else {
			log.Error().Int("product_id", productID).Msg("Failed to retrieve product for update")
			return ErrDatabase
		}
	}

	err = produkt.UpdateDetails(name, kategorie)
	if err != nil {
		log.Warn().Err(err).Int("product_id", productID).Msg("Invalid product data for update")
		return ErrInvalidProduktData
	}

	err = c.ProductRepo.UpdateProduct(ctx, produkt)
	if err != nil {
		log.Error().Err(err).Int("product_id", productID).Msg("Failed to update product")
		return ErrDatabase
	}

	log.Info().Int("product_id", productID).Msg("Product updated")
	return nil
}

func (c Command) ActivateProduct(ctx context.Context, productID int) error {
	log := zerolog.Ctx(ctx)

	produkt, err := c.ProductRepo.GetProduct(ctx, productID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Int("product_id", productID).Msg("Product not found for activation")
			return ErrProduktNotFound
		} else {
			log.Error().Int("product_id", productID).Msg("Failed to retrieve product for activation")
			return ErrDatabase
		}
	}

	produkt.Activate()

	err = c.ProductRepo.UpdateProduct(ctx, produkt)
	if err != nil {
		log.Error().Err(err).Int("product_id", productID).Msg("Failed to update product")
		return ErrDatabase
	}

	log.Info().Int("product_id", productID).Msg("Product activated")
	return nil
}

func (c Command) DeactivateProduct(ctx context.Context, productID int) error {
	log := zerolog.Ctx(ctx)

	produkt, err := c.ProductRepo.GetProduct(ctx, productID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Int("product_id", productID).Msg("Product not found for deactivation")
			return ErrProduktNotFound
		} else {
			log.Error().Int("product_id", productID).Msg("Failed to retrieve product for deactivation")
			return ErrDatabase
		}
	}

	produkt.Deactivate()

	err = c.ProductRepo.UpdateProduct(ctx, produkt)
	if err != nil {
		log.Error().Err(err).Int("product_id", productID).Msg("Failed to update product")
		return ErrDatabase
	}

	log.Info().Int("product_id", productID).Msg("Product deactivated")
	return nil
}

// Variant commands

func (c Command) CreateVariant(ctx context.Context, productID int, name string, preisCents int) (int, error) {
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

	variante, err := product.NewVariante(name, preisCents)
	if err != nil {
		log.Warn().Err(err).Str("variant_name", name).Msg("Invalid variant data")
		return 0, ErrInvalidVarianteData
	}

	variantID, err := c.ProductRepo.CreateVariant(ctx, productID, variante)
	if err != nil {
		log.Error().Int("product_id", productID).Str("name", variante.Name).Msg("Failed to create variant")
		return 0, ErrDatabase
	}

	log.Info().Int("variant_id", variantID).Int("product_id", productID).Msg("Variant created")
	return variantID, nil
}

func (c Command) UpdateVariant(ctx context.Context, variantID int, name string, preisCents int) error {
	log := zerolog.Ctx(ctx)

	variante, err := c.ProductRepo.GetVariant(ctx, variantID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Int("variant_id", variantID).Msg("Variant not found for update")
			return ErrVarianteNotFound
		} else {
			log.Error().Int("variant_id", variantID).Msg("Failed to retrieve variant for update")
			return ErrDatabase
		}
	}

	err = variante.UpdateDetails(name, preisCents)
	if err != nil {
		log.Warn().Err(err).Int("variant_id", variantID).Msg("Invalid variant data for update")
		return ErrInvalidVarianteData
	}

	err = c.ProductRepo.UpdateVariant(ctx, variante)
	if err != nil {
		log.Error().Err(err).Int("variant_id", variantID).Msg("Failed to update variant")
		return ErrDatabase
	}

	log.Info().Int("variant_id", variantID).Msg("Variant updated")
	return nil
}

func (c Command) ActivateVariant(ctx context.Context, variantID int) error {
	log := zerolog.Ctx(ctx)

	variante, err := c.ProductRepo.GetVariant(ctx, variantID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Int("variant_id", variantID).Msg("Variant not found for activation")
			return ErrVarianteNotFound
		} else {
			log.Error().Int("variant_id", variantID).Msg("Failed to retrieve variant for activation")
			return ErrDatabase
		}
	}

	variante.Activate()

	err = c.ProductRepo.UpdateVariant(ctx, variante)
	if err != nil {
		log.Error().Err(err).Int("variant_id", variantID).Msg("Failed to update variant")
		return ErrDatabase
	}

	log.Info().Int("variant_id", variantID).Msg("Variant activated")
	return nil
}

func (c Command) DeactivateVariant(ctx context.Context, variantID int) error {
	log := zerolog.Ctx(ctx)

	variante, err := c.ProductRepo.GetVariant(ctx, variantID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Int("variant_id", variantID).Msg("Variant not found for deactivation")
			return ErrVarianteNotFound
		} else {
			log.Error().Int("variant_id", variantID).Msg("Failed to retrieve variant for deactivation")
			return ErrDatabase
		}
	}

	variante.Deactivate()

	err = c.ProductRepo.UpdateVariant(ctx, variante)
	if err != nil {
		log.Error().Err(err).Int("variant_id", variantID).Msg("Failed to update variant")
		return ErrDatabase
	}

	log.Info().Int("variant_id", variantID).Msg("Variant deactivated")
	return nil
}

func (c Command) DeleteProdukt(ctx context.Context, productID int) error {
	log := zerolog.Ctx(ctx)

	produkt, err := c.ProductRepo.GetProduct(ctx, productID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Int("product_id", productID).Msg("Product not found for deletion")
			return ErrProduktNotFound
		} else {
			log.Error().Int("product_id", productID).Msg("Failed to retrieve product for deletion")
			return ErrDatabase
		}
	}

	for i := range produkt.Varianten {
		produkt.Varianten[i].Delete()
		err = c.ProductRepo.UpdateVariant(ctx, produkt.Varianten[i])
		if err != nil {
			log.Error().Err(err).Int("variant_id", produkt.Varianten[i].ID).Msg("Failed to delete variant")
			return ErrDatabase
		}
	}

	produkt.Delete()

	err = c.ProductRepo.UpdateProduct(ctx, produkt)
	if err != nil {
		log.Error().Err(err).Int("product_id", productID).Msg("Failed to delete product")
		return ErrDatabase
	}

	log.Info().Int("product_id", productID).Msg("Product deleted")
	return nil
}

func (c Command) DeleteVariante(ctx context.Context, produktID int, variantID int) error {
	log := zerolog.Ctx(ctx)

	_, err := c.ProductRepo.GetProduct(ctx, produktID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Int("product_id", produktID).Msg("Product not found for variant deletion")
			return ErrProduktNotFound
		} else {
			log.Error().Int("product_id", produktID).Msg("Failed to retrieve product for variant deletion")
			return ErrDatabase
		}
	}

	variante, err := c.ProductRepo.GetVariant(ctx, variantID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Int("variant_id", variantID).Msg("Variant not found for deletion")
			return ErrVarianteNotFound
		} else {
			log.Error().Int("variant_id", variantID).Msg("Failed to retrieve variant for deletion")
			return ErrDatabase
		}
	}

	variante.Delete()

	err = c.ProductRepo.UpdateVariant(ctx, variante)
	if err != nil {
		log.Error().Err(err).Int("variant_id", variantID).Msg("Failed to delete variant")
		return ErrDatabase
	}

	log.Info().Int("variant_id", variantID).Msg("Variant deleted")
	return nil
}
