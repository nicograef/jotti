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

// App represents the application with its configuration, router, server, and database connection.
type App struct {
	Server *http.Server
	Config config.Config
	DB     *sql.DB
}

// NewApp creates a new application instance
func NewApp(cfg config.Config, db *sql.DB, version string) (*App, error) {
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
	}, nil
}

// SetupRoutes configures HTTP routes
func SetupRoutes(cfg config.Config, db *sql.DB, version string) http.Handler {
	r := http.NewServeMux()

	healthCheck := health.HealthCheck{DB: db, Version: version}
	r.HandleFunc("/health", healthCheck.Handler())

	deps := api.NewDeps(cfg, db)
	deps.Version = version

	authApi := api.NewAuthApi(cfg, deps)
	r.Handle("/auth/", middleware.RateLimitMiddleware(5)(http.StripPrefix("/auth", authApi)))

	// Der Benutzer-Lookup pro Request stellt sicher, dass deaktivierte Benutzer
	// sofort ausgesperrt sind, nicht erst beim Token-Ablauf.
	admin := middleware.NewJwtMiddleware(cfg.JWTSecret, []string{"admin"}, deps.UserRepo)
	adminApi := api.NewAdminApi(deps)
	r.Handle("/admin/", admin(http.StripPrefix("/admin", adminApi)))

	servicesApi := api.NewServiceApi(deps)
	service := middleware.NewJwtMiddleware(cfg.JWTSecret, []string{"admin", "serviceleitung", "service"}, deps.UserRepo)
	r.Handle("/service/", service(http.StripPrefix("/service", servicesApi)))

	serviceleitungApi := api.NewServiceleitungApi(deps)
	serviceleitung := middleware.NewJwtMiddleware(cfg.JWTSecret, []string{"admin", "serviceleitung"}, deps.UserRepo)
	r.Handle("/serviceleitung/", serviceleitung(http.StripPrefix("/serviceleitung", serviceleitungApi)))

	// Relay — kein JWT, Token-Prüfung im Handler; Rate-Limit gegen Token-Brute-Force
	relayApi := api.NewRelayApi(deps, cfg.RelayToken)
	r.Handle("/relay/", middleware.RateLimitMiddleware(5)(http.StripPrefix("/relay", relayApi)))

	// Wrap the entire router with middleware chain
	// Note: Security headers (HSTS, CSP, X-Frame-Options, etc.) are set by the reverse proxy (Caddy)
	var handler http.Handler = r
	handler = middleware.PostMethodOnlyMiddleware(handler)
	handler = middleware.LoggingMiddleware(handler)
	handler = middleware.CorrelationIDMiddleware(handler)

	return handler
}

// Run starts the application with graceful shutdown
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
		return app.Shutdown()
	case err := <-errChan:
		return fmt.Errorf("server error: %w", err)
	}
}

// Shutdown gracefully stops the application
func (app *App) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := app.Server.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Error shutting down server")
	}

	log.Info().Msg("Shutdown complete")
	return nil
}
