package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/nicograef/jotti/backend/app"
	"github.com/nicograef/jotti/backend/config"
)

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

	err = db.Ping()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to ping Postgres")
	}

	log.Info().Msg("Connected to database")

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

	a, err := app.NewApp(cfg, db)
	if err != nil {
		return fmt.Errorf("failed to create app: %w", err)
	}

	// Set up signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	// Run application
	return a.Run(ctx)
}
