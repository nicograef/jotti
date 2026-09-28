package db

import (
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
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

// fakeDB becomes reachable once its clock reaches upAt; sleeping advances the clock.
type fakeDB struct {
	now, upAt time.Duration
}

var errDown = errors.New("connection refused")

func (d *fakeDB) ping() error {
	if d.now < d.upAt {
		return errDown
	}
	return nil
}

func (d *fakeDB) sleep(dur time.Duration) { d.now += dur }

func TestPingWithRetry(t *testing.T) {
	t.Run("succeeds on the first attempt without sleeping", func(t *testing.T) {
		d := &fakeDB{}
		err := PingWithRetry(d.ping, 30*time.Second, time.Second, d.sleep)

		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
		if d.now != 0 {
			t.Errorf("expected no waiting, waited %v", d.now)
		}
	})

	t.Run("succeeds after transient failures", func(t *testing.T) {
		d := &fakeDB{upAt: 2 * time.Second}
		err := PingWithRetry(d.ping, 30*time.Second, time.Second, d.sleep)

		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
		if d.now != 2*time.Second {
			t.Errorf("expected to stop waiting once the database is up after 2s, waited %v", d.now)
		}
	})

	t.Run("gives up after the budget is exhausted", func(t *testing.T) {
		d := &fakeDB{upAt: time.Hour}
		err := PingWithRetry(d.ping, 5*time.Second, time.Second, d.sleep)

		if !errors.Is(err, errDown) {
			t.Errorf("expected the last ping error, got %v", err)
		}
		if d.now != 4*time.Second {
			t.Errorf("expected 4s of waiting (5 attempts, no sleep after the final one), waited %v", d.now)
		}
	})

	t.Run("tries at least once when the interval exceeds the budget", func(t *testing.T) {
		d := &fakeDB{upAt: time.Hour}
		err := PingWithRetry(d.ping, time.Second, 30*time.Second, d.sleep)

		if !errors.Is(err, errDown) {
			t.Errorf("expected the ping error, got %v", err)
		}
		if d.now != 0 {
			t.Errorf("expected no waiting, waited %v", d.now)
		}
	})
}

func TestConnString(t *testing.T) {
	// An empty password would otherwise be looked up in the developer's ~/.pgpass.
	t.Setenv("PGPASSFILE", filepath.Join(t.TempDir(), "missing"))

	cases := []struct {
		name     string
		password string
	}{
		{"backslash", `ab\cd`},
		{"single quote", `ab'cd`},
		{"space", "ab cd"},
		{"trailing backslash", `abcd\`},
		{"equals sign", "ab=cd"},
		{"plus sign", "ab+cd"},
		{"hash sign", "ab#cd"},
		{"empty", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := pgx.ParseConfig(ConnString("db", "5433", "jotti_user", tc.password, "jotti_db"))
			if err != nil {
				t.Fatalf("ParseConfig failed: %v", err)
			}

			if cfg.Host != "db" {
				t.Errorf("expected host %q, got %q", "db", cfg.Host)
			}
			if cfg.Port != 5433 {
				t.Errorf("expected port 5433, got %d", cfg.Port)
			}
			if cfg.User != "jotti_user" {
				t.Errorf("expected user %q, got %q", "jotti_user", cfg.User)
			}
			if cfg.Password != tc.password {
				t.Errorf("expected password %q, got %q", tc.password, cfg.Password)
			}
			if cfg.Database != "jotti_db" {
				t.Errorf("expected database %q, got %q", "jotti_db", cfg.Database)
			}
			if cfg.TLSConfig != nil {
				t.Error("expected sslmode=disable (no TLS config)")
			}
		})
	}
}
