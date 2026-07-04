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

// ErrTischNotActive is returned when an operation is attempted on an inactive or deleted tisch.
var ErrTischNotActive = errors.New("tisch not active")

// fromRepositoryError maps repository errors to application-layer errors with structured logging.
// It consolidates tisch-ID context and error mapping at a single location, avoiding duplicated
// error-handling code across the many command methods that load or update tisch records.
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
