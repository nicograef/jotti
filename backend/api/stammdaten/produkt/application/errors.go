package application

import (
	"errors"

	"github.com/nicograef/jotti/backend/db"
)

var ErrProduktNotFound = errors.New("produkt not found")

var ErrProduktAlreadyExists = errors.New("produkt already exists")

var ErrVarianteNotFound = errors.New("variante not found")

var ErrDatabase = db.ErrDatabase

var ErrInvalidProduktData = errors.New("invalid produkt data")

var ErrInvalidVarianteData = errors.New("invalid variante data")
