package application

import (
	"errors"

	"github.com/nicograef/jotti/backend/db"
	"github.com/rs/zerolog"
)

// ErrTischNotFound is returned when a tisch is not found.
var ErrTischNotFound = errors.New("tisch not found")

// ErrTischAlreadyExists is returned when a tisch already exists.
var ErrTischAlreadyExists = errors.New("tisch already exists")

// ErrDatabase is returned when there is a database error.
var ErrDatabase = db.ErrDatabase

// ErrInvalidTischData is returned when the provided tisch data is invalid.
var ErrInvalidTischData = errors.New("invalid tisch data")

// ErrConflict is returned when an optimistic concurrency conflict cannot be resolved after retries.
var ErrConflict = errors.New("conflict")

// ErrProduktNotFound is returned when a product or variant is not found during enrichment.
var ErrProduktNotFound = errors.New("produkt not found")

// ErrUserNotFound is returned when the user cannot be found for event creation.
var ErrUserNotFound = errors.New("user not found")

// ErrTischNotActive is returned when an operation is attempted on an inactive or deleted tisch.
var ErrTischNotActive = errors.New("tisch not active")

// ErrInvalidPositionen is returned when the provided positions are not valid.
var ErrInvalidPositionen = errors.New("invalid positionen")

// ErrPositionNichtBezahlbar is returned when a position cannot be paid (not in unbezahlt list).
var ErrPositionNichtBezahlbar = errors.New("position nicht bezahlbar")

// ErrPositionNichtLieferbar is returned when a position cannot be delivered (not in ungeliefert list).
var ErrPositionNichtLieferbar = errors.New("position nicht lieferbar")

// ErrPositionNichtStornierbar is returned when a position cannot be cancelled (not in unbezahlt list).
var ErrPositionNichtStornierbar = errors.New("position nicht stornierbar")

func fromRepositoryError(err error, log *zerolog.Logger, id int) error {
	if errors.Is(err, db.ErrNotFound) {
		log.Warn().Err(err).Int("tisch_id", id).Msg("Tisch not found")
		return ErrTischNotFound
	}

	if errors.Is(err, db.ErrAlreadyExists) {
		log.Warn().Err(err).Msg("Tisch already exists")
		return ErrTischAlreadyExists
	}

	log.Error().Err(err).Int("tisch_id", id).Msg("Database error")
	return ErrDatabase
}
