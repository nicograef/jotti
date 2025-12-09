package product_admin

import (
	"context"
	"errors"

	"github.com/nicograef/jotti/backend/db"
	"github.com/rs/zerolog"
)

type commandPersistence interface {
	CreateProduct(ctx context.Context, name, description string, netPriceCents int, category Category) (int, error)
	UpdateProduct(ctx context.Context, id int, name, description string, netPriceCents int, category Category) error
	ActivateProduct(ctx context.Context, id int) error
	DeactivateProduct(ctx context.Context, id int) error
}

type Command struct {
	Persistence commandPersistence
}

// CreateProduct creates a new product in the database.
func (c *Command) CreateProduct(ctx context.Context, name, description string, netPriceCents int, category Category) (int, error) {
	log := zerolog.Ctx(ctx)

	id, err := c.Persistence.CreateProduct(ctx, name, description, netPriceCents, category)
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
func (c *Command) UpdateProduct(ctx context.Context, id int, name, description string, netPriceCents int, category Category) error {
	log := zerolog.Ctx(ctx)

	err := c.Persistence.UpdateProduct(ctx, id, name, description, netPriceCents, category)
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
func (c *Command) ActivateProduct(ctx context.Context, id int) error {
	log := zerolog.Ctx(ctx)

	err := c.Persistence.ActivateProduct(ctx, id)
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
func (c *Command) DeactivateProduct(ctx context.Context, id int) error {
	log := zerolog.Ctx(ctx)

	err := c.Persistence.DeactivateProduct(ctx, id)
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
