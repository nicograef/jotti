package app

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/nicograef/jotti/backend/api"
	"github.com/nicograef/jotti/backend/api/fiskal/signatur"
	"github.com/nicograef/jotti/backend/api/health"
	"github.com/nicograef/jotti/backend/api/middleware"
	"github.com/nicograef/jotti/backend/config"
)

type App struct {
	Server *http.Server
	Config config.Config
	DB     *sql.DB
}

func NewApp(cfg config.Config, db *sql.DB, version string) *App {
	router := SetupRoutes(cfg, db, version)
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      router,
	}

	return &App{
		Server: server,
		Config: cfg,
		DB:     db,
	}
}

// SetupRoutes registriert alle bereichsgebundenen Routen aus der Routentabelle
// (Areas) — der einzigen Registrierungsquelle. /health ist der einzige Sonderfall
// außerhalb der Tabelle (GET-probebar, kein Präfix).
func SetupRoutes(cfg config.Config, db *sql.DB, version string) http.Handler {
	r := http.NewServeMux()

	healthCheck := health.HealthCheck{DB: db, Version: version}
	r.HandleFunc("/health", healthCheck.Handler())

	deps := api.NewDeps(cfg, db)
	deps.Version = version

	areas := Areas()

	if cfg.EnableTestApi {
		areas = append(areas, testResetArea(db))
	}

	for _, area := range areas {
		mountArea(r, area, cfg, deps)
	}

	// Security headers (HSTS, CSP, X-Frame-Options, etc.) are set by the reverse proxy (Caddy).
	// Recovery liegt innen (nach Logging): Ein Panic wird zu 500, der Request wird trotzdem geloggt.
	var handler http.Handler = r
	handler = middleware.RecoveryMiddleware(handler)
	handler = middleware.PostMethodOnlyMiddleware(handler)
	handler = middleware.LoggingMiddleware(handler)
	handler = middleware.CorrelationIDMiddleware(handler)

	return handler
}

func (app *App) Run(ctx context.Context) error {
	worker := signatur.NewTSESignaturWorker(app.Config.FiskalyBaseURL, app.DB)
	go worker.Run(ctx)

	watchdog := signatur.NewTSERueckstandWatchdog(app.DB)
	go watchdog.Run(ctx)

	errChan := make(chan error, 1)
	go func() {
		log.Info().Int("port", app.Config.Port).Msg("Starting server")
		if err := app.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Info().Msg("Shutdown signal received, gracefully stopping...")
		app.Shutdown()
		return nil
	case err := <-errChan:
		return fmt.Errorf("server error: %w", err)
	}
}

func (app *App) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := app.Server.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Error shutting down server")
	}

	log.Info().Msg("Shutdown complete")
}
