package application

import (
	"context"
	"errors"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/produkt"
	"github.com/nicograef/jotti/backend/domain/steuer"
	"github.com/rs/zerolog"
)

type produktRepo interface {
	GetProduct(ctx context.Context, productID int) (produkt.Produkt, error)
	CreateProduct(ctx context.Context, product produkt.Produkt) (int, error)
	UpdateProduct(ctx context.Context, product produkt.Produkt) error
	GetVariant(ctx context.Context, variantID int) (produkt.Variante, error)
	CreateVariant(ctx context.Context, productID int, variant produkt.Variante) (int, error)
	UpdateVariant(ctx context.Context, variant produkt.Variante) error
	GetAllProducts(ctx context.Context) ([]produkt.Produkt, error)
	GetActiveProducts(ctx context.Context) ([]produkt.Produkt, error)
}

type Command struct {
	ProduktRepo produktRepo
}

// Product commands

func (c Command) CreateProduct(ctx context.Context, name string, kategorie produkt.Kategorie, steuersatz steuer.Steuersatz) (int, error) {
	log := zerolog.Ctx(ctx)

	produkt, err := produkt.NewProdukt(name, kategorie, steuersatz)
	if err != nil {
		log.Warn().Err(err).Str("product_name", name).Msg("Invalid product data")
		return 0, ErrInvalidProduktData
	}

	productID, err := c.ProduktRepo.CreateProduct(ctx, produkt)
	if err != nil {
		if errors.Is(err, db.ErrAlreadyExists) {
			log.Warn().Err(err).Str("name", produkt.Name).Msg("Product name already exists")
			return 0, ErrProduktAlreadyExists
		}
		log.Error().Str("name", produkt.Name).Msg("Failed to create product")
		return 0, ErrDatabase
	}

	log.Info().Int("product_id", productID).Msg("Product created")
	return productID, nil
}

func (c Command) UpdateProduct(ctx context.Context, productID int, name string, kategorie produkt.Kategorie, steuersatz steuer.Steuersatz) error {
	log := zerolog.Ctx(ctx)

	produkt, err := c.ProduktRepo.GetProduct(ctx, productID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Int("product_id", productID).Msg("Product not found for update")
			return ErrProduktNotFound
		}
		log.Error().Int("product_id", productID).Msg("Failed to retrieve product for update")
		return ErrDatabase
	}

	err = produkt.UpdateDetails(name, kategorie, steuersatz)
	if err != nil {
		log.Warn().Err(err).Int("product_id", productID).Msg("Invalid product data for update")
		return ErrInvalidProduktData
	}

	err = c.ProduktRepo.UpdateProduct(ctx, produkt)
	if err != nil {
		log.Error().Err(err).Int("product_id", productID).Msg("Failed to update product")
		return ErrDatabase
	}

	log.Info().Int("product_id", productID).Msg("Product updated")
	return nil
}

// Variant commands

func (c Command) CreateVariant(ctx context.Context, productID int, name string, preisCents int) (int, error) {
	log := zerolog.Ctx(ctx)

	// Verify product exists
	_, err := c.ProduktRepo.GetProduct(ctx, productID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Int("product_id", productID).Msg("Product not found for variant creation")
			return 0, ErrProduktNotFound
		}
		log.Error().Int("product_id", productID).Msg("Failed to retrieve product for variant creation")
		return 0, ErrDatabase
	}

	variante, err := produkt.NewVariante(name, preisCents)
	if err != nil {
		log.Warn().Err(err).Str("variant_name", name).Msg("Invalid variant data")
		return 0, ErrInvalidVarianteData
	}

	variantID, err := c.ProduktRepo.CreateVariant(ctx, productID, variante)
	if err != nil {
		log.Error().Int("product_id", productID).Str("name", variante.Name).Msg("Failed to create variant")
		return 0, ErrDatabase
	}

	log.Info().Int("variant_id", variantID).Int("product_id", productID).Msg("Variant created")
	return variantID, nil
}

func (c Command) UpdateVariant(ctx context.Context, variantID int, name string, preisCents int) error {
	log := zerolog.Ctx(ctx)

	variante, err := c.ProduktRepo.GetVariant(ctx, variantID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Int("variant_id", variantID).Msg("Variant not found for update")
			return ErrVarianteNotFound
		}
		log.Error().Int("variant_id", variantID).Msg("Failed to retrieve variant for update")
		return ErrDatabase
	}

	err = variante.UpdateDetails(name, preisCents)
	if err != nil {
		log.Warn().Err(err).Int("variant_id", variantID).Msg("Invalid variant data for update")
		return ErrInvalidVarianteData
	}

	err = c.ProduktRepo.UpdateVariant(ctx, variante)
	if err != nil {
		log.Error().Err(err).Int("variant_id", variantID).Msg("Failed to update variant")
		return ErrDatabase
	}

	log.Info().Int("variant_id", variantID).Msg("Variant updated")
	return nil
}

func (c Command) ActivateVariant(ctx context.Context, variantID int) error {
	return c.applyVarianteStatusChange(ctx, variantID, "Variant activated", func(v *produkt.Variante) { v.Activate() })
}

func (c Command) DeactivateVariant(ctx context.Context, variantID int) error {
	return c.applyVarianteStatusChange(ctx, variantID, "Variant deactivated", func(v *produkt.Variante) { v.Deactivate() })
}

func (c Command) applyVarianteStatusChange(ctx context.Context, variantID int, successMsg string, action func(*produkt.Variante)) error {
	log := zerolog.Ctx(ctx)

	variante, err := c.ProduktRepo.GetVariant(ctx, variantID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Int("variant_id", variantID).Msg("Variant not found for status change")
			return ErrVarianteNotFound
		}
		log.Error().Int("variant_id", variantID).Msg("Failed to retrieve variant for status change")
		return ErrDatabase
	}

	action(&variante)

	if err := c.ProduktRepo.UpdateVariant(ctx, variante); err != nil {
		log.Error().Err(err).Int("variant_id", variantID).Msg("Failed to update variant")
		return ErrDatabase
	}

	log.Info().Int("variant_id", variantID).Msg(successMsg)
	return nil
}

func (c Command) DeleteProdukt(ctx context.Context, productID int) error {
	log := zerolog.Ctx(ctx)

	produkt, err := c.ProduktRepo.GetProduct(ctx, productID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Int("product_id", productID).Msg("Product not found for deletion")
			return ErrProduktNotFound
		}
		log.Error().Int("product_id", productID).Msg("Failed to retrieve product for deletion")
		return ErrDatabase
	}

	for i := range produkt.Varianten {
		produkt.Varianten[i].Delete()
		err = c.ProduktRepo.UpdateVariant(ctx, produkt.Varianten[i])
		if err != nil {
			log.Error().Err(err).Int("variant_id", produkt.Varianten[i].ID).Msg("Failed to delete variant")
			return ErrDatabase
		}
	}

	produkt.Delete()

	err = c.ProduktRepo.UpdateProduct(ctx, produkt)
	if err != nil {
		log.Error().Err(err).Int("product_id", productID).Msg("Failed to delete product")
		return ErrDatabase
	}

	log.Info().Int("product_id", productID).Msg("Product deleted")
	return nil
}

func (c Command) DeleteVariante(ctx context.Context, produktID int, variantID int) error {
	log := zerolog.Ctx(ctx)

	_, err := c.ProduktRepo.GetProduct(ctx, produktID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Int("product_id", produktID).Msg("Product not found for variant deletion")
			return ErrProduktNotFound
		}
		log.Error().Int("product_id", produktID).Msg("Failed to retrieve product for variant deletion")
		return ErrDatabase
	}

	variante, err := c.ProduktRepo.GetVariant(ctx, variantID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Int("variant_id", variantID).Msg("Variant not found for deletion")
			return ErrVarianteNotFound
		}
		log.Error().Int("variant_id", variantID).Msg("Failed to retrieve variant for deletion")
		return ErrDatabase
	}

	variante.Delete()

	err = c.ProduktRepo.UpdateVariant(ctx, variante)
	if err != nil {
		log.Error().Err(err).Int("variant_id", variantID).Msg("Failed to delete variant")
		return ErrDatabase
	}

	log.Info().Int("variant_id", variantID).Msg("Variant deleted")
	return nil
}
