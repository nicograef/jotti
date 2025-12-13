package application

import (
	"errors"

	"github.com/nicograef/jotti/backend/db"
	"github.com/rs/zerolog"
)

// ErrProductNotFound is returned when a product is not found.
var ErrProductNotFound = errors.New("product not found")

// ErrProductAlreadyExists is returned when trying to create a product that already exists.
var ErrProductAlreadyExists = errors.New("product already exists")

// ErrDatabase is returned when there is a database error.
var ErrDatabase = errors.New("database error")

// ErrInvalidProductData is returned when the provided product data is invalid.
var ErrInvalidProductData = errors.New("invalid product data")

func fromRepositoryError(err error, log *zerolog.Logger, id int) error {
	if errors.Is(err, db.ErrNotFound) {
		log.Warn().Err(err).Int("product_id", id).Msg("Product not found")
		return ErrProductNotFound
	}

	if errors.Is(err, db.ErrAlreadyExists) {
		log.Warn().Err(err).Msg("Product already exists")
		return ErrProductAlreadyExists
	}

	log.Error().Err(err).Int("product_id", id).Msg("Database error")
	return ErrDatabase
}
