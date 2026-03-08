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
