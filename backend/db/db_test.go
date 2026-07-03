//go:build unit

package db

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestError(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want error
	}{
		{"nil", nil, nil},
		{"unique violation", &pgconn.PgError{Code: string(ErrorCodeUniqueViolation)}, ErrAlreadyExists},
		{"deadlock detected", &pgconn.PgError{Code: string(ErrorCodeDeadlockDetected)}, ErrConflict},
		{"no rows", sql.ErrNoRows, ErrNotFound},
		{"other error", errors.New("boom"), ErrDatabase},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Error(tc.err); !errors.Is(got, tc.want) {
				t.Errorf("Error(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}
