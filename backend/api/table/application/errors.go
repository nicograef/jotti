package application

import (
	"errors"

	"github.com/nicograef/jotti/backend/db"
	"github.com/rs/zerolog"
)

// ErrTableNotFound is returned when a table is not found.
var ErrTableNotFound = errors.New("table not found")

// ErrTableAlreadyExists is returned when a table already exists.
var ErrTableAlreadyExists = errors.New("table already exists")

// ErrDatabase is returned when there is a database error.
var ErrDatabase = errors.New("database error")

// ErrInvalidTableData is returned when the provided table data is invalid.
var ErrInvalidTableData = errors.New("invalid table data")

func fromRepositoryError(err error, log *zerolog.Logger, id int) error {
	if errors.Is(err, db.ErrNotFound) {
		log.Warn().Err(err).Int("table_id", id).Msg("Table not found")
		return ErrTableNotFound
	}

	if errors.Is(err, db.ErrAlreadyExists) {
		log.Warn().Err(err).Msg("Table already exists")
		return ErrTableAlreadyExists
	}

	log.Error().Err(err).Int("table_id", id).Msg("Database error")
	return ErrDatabase
}
