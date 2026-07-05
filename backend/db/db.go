package db

import (
	"database/sql"
	"errors"
	"io"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rs/zerolog/log"
)

// ErrorCode represents a PostgreSQL error code.
// https://www.postgresql.org/docs/17/errcodes-appendix.html
type ErrorCode string

const (
	// UniqueViolation indicates a violation of a unique constraint.
	ErrorCodeUniqueViolation ErrorCode = "23505"
	// DeadlockDetected indicates the transaction was aborted as a deadlock victim.
	ErrorCodeDeadlockDetected ErrorCode = "40P01"
)

// ErrNotFound is returned when a record is not found.
var ErrNotFound = errors.New("not found")

// ErrAlreadyExists is returned when a record already exists.
var ErrAlreadyExists = errors.New("already exists")

// ErrConflict is returned when a transaction was aborted because of a
// concurrent transaction (deadlock victim); the request can be retried.
var ErrConflict = errors.New("transaction conflict")

// ErrDatabase is returned when there is a database error.
var ErrDatabase = errors.New("database error")

// Error maps a database error to a more specific error.
func Error(err error) error {
	if err == nil {
		return nil
	}

	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		switch ErrorCode(pgErr.Code) {
		case ErrorCodeUniqueViolation:
			return ErrAlreadyExists
		case ErrorCodeDeadlockDetected:
			return ErrConflict
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

// PingWithRetry calls ping repeatedly until it succeeds or the time budget is
// exhausted (budget/interval attempts, at least one). Every failed attempt is
// logged so a delayed database is visible in the boot log; the caller decides
// what a returned error means (the backend still refuses to start without a
// database). ping and sleep are injected so the retry decision can be
// unit-tested without a real database or real waiting.
func PingWithRetry(ping func() error, budget, interval time.Duration, sleep func(time.Duration)) error {
	attempts := int(budget / interval)
	if attempts < 1 {
		attempts = 1
	}

	var err error
	for attempt := 1; attempt <= attempts; attempt++ {
		if err = ping(); err == nil {
			if attempt > 1 {
				log.Info().Int("attempts", attempt).Msg("Connected to database after retry")
			}
			return nil
		}

		log.Warn().Err(err).Int("attempt", attempt).Int("maxAttempts", attempts).Msg("Waiting for database")
		if attempt < attempts {
			sleep(interval)
		}
	}

	return err
}
