package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/nicograef/jotti/backend/app"
	"github.com/nicograef/jotti/backend/bootstrap"
	"github.com/nicograef/jotti/backend/config"
	dbpkg "github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/internal/logging"
	"github.com/nicograef/jotti/backend/repository/kassenjournal_repo"
	"github.com/nicograef/jotti/backend/repository/user_repo"
	"github.com/nicograef/jotti/backend/seed"
)

// version wird per ldflags einkompiliert (-X main.version=<tag>).
var version = "dev"

func main() {
	slog.SetDefault(logging.New(os.Stdout))

	cfg := config.Load()

	psqlconn := dbpkg.ConnString(cfg.Postgres.Host, strconv.Itoa(cfg.Postgres.Port), cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.DBName)

	db, err := sql.Open("pgx", psqlconn)
	if err != nil {
		fatal("Failed to connect to Postgres", "error", err)
	}

	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(10)

	// Begrenzt auf die Datenbank warten (Boot-Reihenfolge nach Stromausfall), statt sofort zu sterben.
	if err := dbpkg.PingWithRetry(db.Ping, 30*time.Second, time.Second, time.Sleep); err != nil {
		fatal("Failed to ping Postgres", "error", err)
	}

	slog.Info("Connected to database")

	subcommand := ""
	if len(os.Args) > 1 {
		subcommand = os.Args[1]
	}

	if subcommand == "rebuild-projections" {
		if err := rebuildProjections(db); err != nil {
			fatal("Failed to rebuild projections", "error", err)
		}
		return
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

	// Der Klartext-Code des Einmalpassworts landet im Log-Strom.
	repo := user_repo.NewRepository(db)
	res, err := bootstrap.EnsureInitialAdmin(context.Background(), repo)
	if err != nil {
		return fmt.Errorf("bootstrap initial admin: %w", err)
	}
	res.Log(slog.Default())

	a := app.NewApp(cfg, db, version)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	return a.Run(ctx)
}

func rebuildProjections(db *sql.DB) error {
	defer func() {
		if err := db.Close(); err != nil {
			slog.Error("Failed to close database connection", "error", err)
		}
	}()

	repo := kassenjournal_repo.NewRepository(db)

	slog.Info("Rebuilding tisch-session projections from kassenjournal...")

	count, err := repo.RebuildAllProjections(context.Background())
	if err != nil {
		return fmt.Errorf("rebuild projections: %w", err)
	}

	slog.Info("Projections rebuilt successfully", "subjects", count)
	return nil
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
