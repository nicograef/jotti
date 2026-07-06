package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	_ "time/tzdata"

	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/nicograef/jotti/backend/app"
	"github.com/nicograef/jotti/backend/bootstrap"
	"github.com/nicograef/jotti/backend/config"
	dbpkg "github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/repository/kassenjournal_repo"
	"github.com/nicograef/jotti/backend/repository/user_repo"
	"github.com/nicograef/jotti/backend/seed"
)

// version wird per ldflags einkompiliert (-X main.version=<tag>); der
// Release-Workflow befuellt sie ueber das Docker-Build-Argument VERSION.
var version = "dev"

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})

	cfg := config.Load()

	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.DBName)

	db, err := sql.Open("pgx", psqlconn)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Postgres")
	}

	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(10)

	// Beim Start begrenzt auf die Datenbank warten (Boot-Reihenfolge nach
	// Stromausfall), statt sofort zu sterben. Ohne Datenbank kein Start.
	if err := dbpkg.PingWithRetry(db.Ping, 30*time.Second, time.Second, time.Sleep); err != nil {
		log.Fatal().Err(err).Msg("Failed to ping Postgres")
	}

	log.Info().Msg("Connected to database")

	if len(os.Args) > 1 && os.Args[1] == "rebuild-projections" {
		if err := rebuildProjections(db); err != nil {
			log.Fatal().Err(err).Msg("Failed to rebuild projections")
		}
		return
	}

	if len(os.Args) > 1 && os.Args[1] == "seed" {
		if err := seedDemodaten(db); err != nil {
			log.Fatal().Err(err).Msg("Failed to seed demo data")
		}
		return
	}

	if err := run(cfg, db); err != nil {
		log.Fatal().Err(err).Msg("Application error")
	}
}

func run(cfg config.Config, db *sql.DB) error {
	defer func() {
		if err := db.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close database connection")
		}
	}()

	// Initial-Admin anlegen bzw. dessen Einmalpasswort rotieren, solange die
	// Ersteinrichtung offen ist; der Klartext-Code landet im Log-Strom. Ein Fehler
	// hier ist fatal (run() → main() → log.Fatal), der Container-Restart wiederholt.
	repo := user_repo.NewRepository(db)
	res, err := bootstrap.EnsureInitialAdmin(context.Background(), repo)
	if err != nil {
		return fmt.Errorf("bootstrap initial admin: %w", err)
	}
	res.Log(log.Logger)

	a, err := app.NewApp(cfg, db, version)
	if err != nil {
		return fmt.Errorf("failed to create app: %w", err)
	}

	// Set up signal handling
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
			log.Error().Err(err).Msg("Failed to close database connection")
		}
	}()

	repo := kassenjournal_repo.NewRepository(db)

	log.Info().Msg("Rebuilding tisch-session projections from kassenjournal...")

	count, err := repo.RebuildAllProjections(context.Background())
	if err != nil {
		return fmt.Errorf("rebuild projections: %w", err)
	}

	log.Info().Int("subjects", count).Msg("Projections rebuilt successfully")
	return nil
}

func seedDemodaten(db *sql.DB) error {
	defer func() {
		if err := db.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close database connection")
		}
	}()

	log.Info().Msg("Seeding demo data...")

	if err := seed.Run(context.Background(), db); err != nil {
		return fmt.Errorf("seed demo data: %w", err)
	}

	log.Info().Msg("Demo data seeded successfully")
	return nil
}
