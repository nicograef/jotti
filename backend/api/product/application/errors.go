package application

import (
	"errors"

	"github.com/nicograef/jotti/backend/db"
)

// ErrProductNotFound is returned when a product is not found.
var ErrProductNotFound = errors.New("product not found")

// ErrProductAlreadyExists is returned when trying to create a product that already exists.
var ErrProductAlreadyExists = errors.New("product already exists")

// ErrVariantNotFound is returned when a variant is not found.
var ErrVariantNotFound = errors.New("variant not found")

// ErrDatabase is returned when there is a database error.
var ErrDatabase = db.ErrDatabase

// ErrInvalidProductData is returned when the provided product data is invalid.
var ErrInvalidProductData = errors.New("invalid product data")

// ErrInvalidVariantData is returned when the provided variant data is invalid.
var ErrInvalidVariantData = errors.New("invalid variant data")
