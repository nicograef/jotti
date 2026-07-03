package app

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/nicograef/jotti/backend/api"
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
func NewApp(cfg config.Config, db *sql.DB) (*App, error) {
	router := SetupRoutes(cfg, db)
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
func SetupRoutes(cfg config.Config, db *sql.DB) http.Handler {
	r := http.NewServeMux()

	healthCheck := health.HealthCheck{DB: db}
	r.HandleFunc("/health", healthCheck.Handler())

	authApi := api.NewAuthApi(cfg, db)
	r.Handle("/auth/", middleware.RateLimitMiddleware(5)(http.StripPrefix("/auth", authApi)))

	admin := middleware.NewJwtMiddleware(cfg.JWTSecret, []string{"admin"})
	adminApi := api.NewAdminApi(cfg, db)
	r.Handle("/admin/", admin(http.StripPrefix("/admin", adminApi)))

	servicesApi := api.NewServiceApi(cfg, db)
	service := middleware.NewJwtMiddleware(cfg.JWTSecret, []string{"admin", "serviceleitung", "service"})
	r.Handle("/service/", service(http.StripPrefix("/service", servicesApi)))

	serviceleitungApi := api.NewServiceleitungApi(cfg, db)
	serviceleitung := middleware.NewJwtMiddleware(cfg.JWTSecret, []string{"admin", "serviceleitung"})
	r.Handle("/serviceleitung/", serviceleitung(http.StripPrefix("/serviceleitung", serviceleitungApi)))

	// Relay — kein JWT, Token-Prüfung im Handler; Rate-Limit gegen Token-Brute-Force
	relayApi := api.NewRelayApi(db, cfg.RelayToken)
	r.Handle("/relay/", middleware.RateLimitMiddleware(5)(http.StripPrefix("/relay", relayApi)))

	// Wrap the entire router with middleware chain
	// Note: Security headers (HSTS, CSP, X-Frame-Options, etc.) are set by nginx
	var handler http.Handler = r
	handler = middleware.PostMethodOnlyMiddleware(handler)
	handler = middleware.LoggingMiddleware(handler)
	handler = middleware.CorrelationIDMiddleware(handler)

	return handler
}

// Run starts the application with graceful shutdown
func (app *App) Run(ctx context.Context) error {
	worker := newTSENachsignierWorker(app.Config, app.DB)
	go worker.run(ctx)

	errChan := make(chan error, 1)
	go func() {
		log.Info().Int("port", app.Config.Port).Msg("Starting server")
		if err := app.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		fmt.Println("Shutdown signal received, gracefully stopping...")
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
		log.Printf("ERROR shutting down server: %v", err)
	}

	fmt.Println("Shutdown complete")
	return nil
}
