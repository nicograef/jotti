package application

import (
	"errors"

	"github.com/nicograef/jotti/backend/db"
)

// ErrProduktNotFound is returned when a product is not found.
var ErrProduktNotFound = errors.New("produkt not found")

// ErrProduktAlreadyExists is returned when trying to create a product that already exists.
var ErrProduktAlreadyExists = errors.New("produkt already exists")

// ErrVarianteNotFound is returned when a variant is not found.
var ErrVarianteNotFound = errors.New("variante not found")

// ErrDatabase is returned when there is a database error.
var ErrDatabase = db.ErrDatabase

// ErrInvalidProduktData is returned when the provided product data is invalid.
var ErrInvalidProduktData = errors.New("invalid produkt data")

// ErrInvalidVarianteData is returned when the provided variant data is invalid.
var ErrInvalidVarianteData = errors.New("invalid variante data")
