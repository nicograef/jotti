package application

import (
	"errors"

	"github.com/nicograef/jotti/backend/db"
	"github.com/rs/zerolog"
)

var ErrTischNotFound = errors.New("tisch not found")

var ErrTischAlreadyExists = errors.New("tisch already exists")

var ErrDatabase = db.ErrDatabase

var ErrInvalidTischData = errors.New("invalid tisch data")

var ErrTischNotActive = errors.New("tisch not active")

// ErrTischSaldoOffen is returned when a tisch cannot be deactivated or deleted
// because it carries an open balance in the currently open Kassensitzung.
var ErrTischSaldoOffen = errors.New("tisch has open saldo")

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
