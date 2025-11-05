package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"database/sql"

	_ "github.com/lib/pq"

	"github.com/nicograef/jotti/backend/app"
	"github.com/nicograef/jotti/backend/config"
)

func main() {
	cfg := config.Load()

	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.DBName)

	db, err := sql.Open("postgres", psqlconn)
	if err != nil {
		fmt.Printf("Failed to connect to Postgres: %v\n", err)
		os.Exit(1)
	}

	defer func() {
		if err := db.Close(); err != nil {
			fmt.Printf("Failed to close database connection: %v\n", err)
		}
	}()

	err = db.Ping()
	if err != nil {
		fmt.Printf("Failed to ping Postgres: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Connected to database.")

	app, err := app.NewApp(cfg, db)
	if err != nil {
		fmt.Printf("Failed to create app: %v\n", err)
		os.Exit(1)
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
	if err := app.Run(ctx); err != nil {
		fmt.Printf("Application error: %v\n", err)
		os.Exit(1)
	}
}
