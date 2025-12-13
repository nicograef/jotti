package db

import (
	"database/sql"
	"errors"
	"io"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rs/zerolog/log"
)

// ErrorCode represents a PostgreSQL error code.
// https://www.postgresql.org/docs/17/errcodes-appendix.html
type ErrorCode string

const (
	// UniqueViolation indicates a violation of a unique constraint.
	ErrorCodeUniqueViolation ErrorCode = "23505"
)

// ErrNotFound is returned when a record is not found.
var ErrNotFound = errors.New("not found")

// ErrAlreadyExists is returned when a record already exists.
var ErrAlreadyExists = errors.New("already exists")

// ErrDatabase is returned when there is a database error.
var ErrDatabase = errors.New("database error")

// Error maps a database error to a more specific error.
func Error(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case string(ErrorCodeUniqueViolation):
			return ErrAlreadyExists
		}
	}

	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}

	return ErrDatabase
}

// ResultError checks the result of a SQL operation and returns an appropriate error.
func ResultError(res sql.Result) error {
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return ErrDatabase
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// Close safely closes an io.Closer and logs any error that occurs.
func Close(c io.Closer, name string) {
	if err := c.Close(); err != nil {
		log.Error().Err(err).Str("resource", name).Msg("Error while closing resource")
	}
}
