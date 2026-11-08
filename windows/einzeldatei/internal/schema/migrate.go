// Package schema legt die Datenbank an und spielt die Migrationen mit golang-migrate ein,
// in dieselbe Tabelle schema_migrations wie das Migrate-Image.
package schema

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5" // registriert pgx5://
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib" // registriert den Treiber "pgx"
)

// EnsureDatabase legt die Datenbank name an, falls sie fehlt; dsn zeigt auf die Datenbank postgres.
func EnsureDatabase(ctx context.Context, dsn, name string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	defer db.Close() //nolint:errcheck // nur lesend genutzt bis auf CREATE DATABASE

	var found bool
	if err := db.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)", name).Scan(&found); err != nil {
		return fmt.Errorf("bestehende Datenbank suchen: %w", err)
	}
	if found {
		return nil
	}
	if _, err := db.ExecContext(ctx, "CREATE DATABASE "+pgx.Identifier{name}.Sanitize()); err != nil {
		return fmt.Errorf("neue Datenbank anlegen: %w", err)
	}
	return nil
}

// Migrate spielt alle Migrationen aus src (NN_name.up.sql) ein; dsn hat das Schema pgx5://.
func Migrate(ctx context.Context, dsn string, src fs.FS) error {
	source, err := iofs.New(src, ".")
	if err != nil {
		return err
	}
	m, err := migrate.NewWithSourceInstance("iofs", source, dsn)
	if err != nil {
		return err
	}
	defer m.Close() //nolint:errcheck // Fehler beim Schließen ändern das Ergebnis nicht

	done := make(chan error, 1)
	go func() { done <- m.Up() }()
	select {
	case err := <-done:
		if err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return err
		}
		return nil
	case <-ctx.Done():
		m.GracefulStop <- true
		<-done
		return ctx.Err()
	}
}
