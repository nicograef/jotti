//go:build unit

package db

import (
	"database/sql"
	"errors"
	"testing"
	"time"

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

func TestPingWithRetry(t *testing.T) {
	t.Run("succeeds on the first attempt without sleeping", func(t *testing.T) {
		pings, sleeps := 0, 0
		err := PingWithRetry(func() error {
			pings++
			return nil
		}, 30*time.Second, time.Second, func(time.Duration) { sleeps++ })

		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if pings != 1 {
			t.Errorf("expected 1 ping, got %d", pings)
		}
		if sleeps != 0 {
			t.Errorf("expected no sleeps, got %d", sleeps)
		}
	})

	t.Run("succeeds after transient failures", func(t *testing.T) {
		pings, sleeps := 0, 0
		err := PingWithRetry(func() error {
			pings++
			if pings < 3 {
				return errors.New("connection refused")
			}
			return nil
		}, 30*time.Second, time.Second, func(time.Duration) { sleeps++ })

		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if pings != 3 {
			t.Errorf("expected 3 pings, got %d", pings)
		}
		if sleeps != 2 {
			t.Errorf("expected 2 sleeps between attempts, got %d", sleeps)
		}
	})

	t.Run("gives up after the budget is exhausted", func(t *testing.T) {
		down := errors.New("still down")
		pings, sleeps := 0, 0
		err := PingWithRetry(func() error {
			pings++
			return down
		}, 5*time.Second, time.Second, func(time.Duration) { sleeps++ })

		if !errors.Is(err, down) {
			t.Fatalf("expected the last ping error, got %v", err)
		}
		if pings != 5 {
			t.Errorf("expected 5 ping attempts (budget/interval), got %d", pings)
		}
		if sleeps != 4 {
			t.Errorf("expected 4 sleeps (no sleep after the final attempt), got %d", sleeps)
		}
	})

	t.Run("tries at least once when the interval exceeds the budget", func(t *testing.T) {
		pings, sleeps := 0, 0
		err := PingWithRetry(func() error {
			pings++
			return errors.New("down")
		}, time.Second, 30*time.Second, func(time.Duration) { sleeps++ })

		if err == nil {
			t.Fatal("expected an error, got nil")
		}
		if pings != 1 {
			t.Errorf("expected exactly 1 ping, got %d", pings)
		}
		if sleeps != 0 {
			t.Errorf("expected no sleeps, got %d", sleeps)
		}
	})
}
