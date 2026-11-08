package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/nicograef/jotti/backend/app"
	"github.com/nicograef/jotti/backend/config"
	dbpkg "github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/internal/logging"
	"github.com/nicograef/jotti/backend/seed"
)

// version wird per ldflags einkompiliert (-X main.version=<tag>).
var version = "dev"

func main() {
	slog.SetDefault(logging.New(os.Stdout))

	cfg := config.Load()

	psqlconn := dbpkg.ConnString(cfg.Postgres.Host, strconv.Itoa(cfg.Postgres.Port), cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.DBName)

	db, err := dbpkg.Open(context.Background(), psqlconn)
	if err != nil {
		fatal("Failed to connect to Postgres", "error", err)
	}

	slog.Info("Connected to database")

	subcommand := ""
	if len(os.Args) > 1 {
		subcommand = os.Args[1]
	}

	if subcommand == "seed" {
		if !seed.AllowedByEnv(os.Getenv) {
			fatal(fmt.Sprintf("seed refused: set %s=1 to allow demo seeding (never do this in production)", seed.AllowSeedEnv))
		}
		if err := seedDemodaten(db); err != nil {
			fatal("Failed to seed demo data", "error", err)
		}
		return
	}

	if err := run(cfg, db); err != nil {
		fatal("Application error", "error", err)
	}
}

func run(cfg config.Config, db *sql.DB) error {
	defer func() {
		if err := db.Close(); err != nil {
			slog.Error("Failed to close database connection", "error", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	return app.Serve(ctx, cfg, db, version)
}

func seedDemodaten(db *sql.DB) error {
	defer func() {
		if err := db.Close(); err != nil {
			slog.Error("Failed to close database connection", "error", err)
		}
	}()

	slog.Info("Seeding demo data...")

	if err := seed.Run(context.Background(), db); err != nil {
		return fmt.Errorf("seed demo data: %w", err)
	}

	slog.Info("Demo data seeded successfully")
	return nil
}

func fatal(msg string, args ...any) {
	slog.Error(msg, args...)
	os.Exit(1)
}
