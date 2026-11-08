package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib" // registriert den Treiber "pgx" für Open
	"github.com/nicograef/jotti/backend/sqlc/dbgen"
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

// ConnString builds the libpq keyword/value connection string (sslmode=disable) for the pgx driver.
// Every value is single-quoted and escaped, so a password with spaces, quotes or backslashes reaches the server unchanged.
func ConnString(host, port, user, password, dbname string) string {
	quote := func(value string) string {
		return "'" + connStringEscaper.Replace(value) + "'"
	}

	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		quote(host), quote(port), quote(user), quote(password), quote(dbname))
}

// Open opens the pgx pool and waits up to 30 s for the database (boot order after a power cut) instead of failing at once.
func Open(ctx context.Context, connString string) (*sql.DB, error) {
	database, err := sql.Open("pgx", connString)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	database.SetConnMaxLifetime(5 * time.Minute)
	database.SetMaxOpenConns(50)
	database.SetMaxIdleConns(10)

	ping := func() error { return database.PingContext(ctx) }
	sleep := func(d time.Duration) {
		timer := time.NewTimer(d)
		defer timer.Stop()
		select {
		case <-ctx.Done():
		case <-timer.C:
		}
	}
	if err := PingWithRetry(ctx, ping, 30*time.Second, time.Second, sleep); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return database, nil
}

// PingWithRetry calls ping until it succeeds or budget/interval attempts (at least one) are used up.
// It logs every failed attempt, so a delayed database shows in the boot log.
// A done ctx ends the retry at once with ctx.Err(), so a stopping caller does not wait out the budget.
func PingWithRetry(ctx context.Context, ping func() error, budget, interval time.Duration, sleep func(time.Duration)) error {
	attempts := max(int(budget/interval), 1)

	var err error
	for attempt := 1; attempt <= attempts; attempt++ {
		if err = ping(); err == nil {
			if attempt > 1 {
				slog.Info("Connected to database after retry", "attempts", attempt)
			}
			return nil
		}
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}

		slog.Warn("Waiting for database", "error", err, "attempt", attempt, "maxAttempts", attempts)
		if attempt < attempts {
			sleep(interval)
		}
	}

	return err
}

// Int4Array turns ids into an int4[] parameter for ANY($1): pgx encodes []int32 natively. sqlc's
// database/sql code would need lib/pq's pq.Array instead, so these queries are written by hand.
func Int4Array(ids []int) []int32 {
	out := make([]int32, len(ids))
	for i, id := range ids {
		out[i] = int32(id)
	}
	return out
}
