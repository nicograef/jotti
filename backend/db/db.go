package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nicograef/jotti/backend/sqlc/dbgen"
	"github.com/rs/zerolog/log"
)

// https://www.postgresql.org/docs/17/errcodes-appendix.html
type ErrorCode string

const (
	ErrorCodeUniqueViolation  ErrorCode = "23505"
	ErrorCodeDeadlockDetected ErrorCode = "40P01"
)

var ErrNotFound = errors.New("not found")

var ErrAlreadyExists = errors.New("already exists")

// ErrConflict is returned when a transaction was aborted because of a
// concurrent transaction (deadlock victim); the request can be retried.
var ErrConflict = errors.New("transaction conflict")

var ErrDatabase = errors.New("database error")

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

// WithTx runs fn in one transaction. fn owns its own error wrapping; only
// begin/commit failures are normalized via Error.
func WithTx(ctx context.Context, database *sql.DB, fn func(*dbgen.Queries) error) error {
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return Error(err)
	}
	defer tx.Rollback() //nolint:errcheck // rollback after commit is a no-op

	if err := fn(dbgen.New(tx)); err != nil {
		return err
	}

	return Error(tx.Commit())
}

// connStringEscaper escapes the two characters with a meaning inside a quoted
// keyword/value connection string value.
var connStringEscaper = strings.NewReplacer(`\`, `\\`, `'`, `\'`)

// ConnString builds the keyword/value connection string (libpq syntax,
// sslmode=disable) for the pgx driver. An unquoted value ends at whitespace and
// a backslash escapes the next character, so every value is single-quoted and
// escaped: a password with spaces, quotes or backslashes reaches the server
// unchanged.
func ConnString(host, port, user, password, dbname string) string {
	quote := func(value string) string {
		return "'" + connStringEscaper.Replace(value) + "'"
	}

	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		quote(host), quote(port), quote(user), quote(password), quote(dbname))
}

// PingWithRetry calls ping until it succeeds or the budget is exhausted
// (budget/interval attempts, at least one), logging every failed attempt so a
// delayed database is visible in the boot log. ping and sleep are injected for
// tests.
func PingWithRetry(ping func() error, budget, interval time.Duration, sleep func(time.Duration)) error {
	attempts := max(int(budget/interval), 1)

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
