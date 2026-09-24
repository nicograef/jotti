// Package dbtest öffnet die Postgres-Testdatenbank der Integrationstests aus
// den POSTGRES_*-Umgebungsvariablen (Vorgaben: localhost:5432, admin/admin, jotti).
package dbtest

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"

	dbpkg "github.com/nicograef/jotti/backend/db"
)

// Open verbindet sich mit der Testdatenbank und beendet den Testprozess, wenn
// keine Verbindung zustande kommt: ohne Datenbank ist kein Integrationstest
// aussagekräftig.
func Open() *sql.DB {
	host := envOrDefault("POSTGRES_HOST", "localhost")
	port := envOrDefault("POSTGRES_PORT", "5432")
	user := envOrDefault("POSTGRES_USER", "admin")
	password := envOrDefault("POSTGRES_PASSWORD", "admin")
	dbName := envAnyOrDefault([]string{"POSTGRES_DBNAME", "POSTGRES_DB"}, "jotti")

	db, err := sql.Open("pgx", dbpkg.ConnString(host, port, user, password, dbName))
	if err != nil {
		fmt.Printf("Failed to connect to Postgres: %v\n", err)
		os.Exit(1)
	}

	err = db.Ping()
	if err != nil {
		fmt.Printf("Failed to ping Postgres: %v\n", err)
		os.Exit(1)
	}

	return db
}

func envOrDefault(name string, fallback string) string {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}

	return value
}

func envAnyOrDefault(names []string, fallback string) string {
	for _, name := range names {
		if value := os.Getenv(name); value != "" {
			return value
		}
	}

	return fallback
}
