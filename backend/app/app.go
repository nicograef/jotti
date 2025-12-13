package app

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"

	adm "github.com/nicograef/jotti/backend/admin/api"
	"github.com/nicograef/jotti/backend/api"
	auth "github.com/nicograef/jotti/backend/auth/api"
	"github.com/nicograef/jotti/backend/config"
	service "github.com/nicograef/jotti/backend/service/api"
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
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
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

	// Health check with database connectivity
	healthCheck := api.HealthCheck{DB: db}
	r.HandleFunc("/health", healthCheck.Handler())

	authApi := auth.NewApi(cfg, db)
	r.Handle("/auth/", http.StripPrefix("/auth", authApi))

	admin := api.NewJwtMiddleware(cfg.JWTSecret, []string{"admin"})
	adminApi := adm.NewApi(db)
	r.Handle("/admin/", admin(http.StripPrefix("/admin", adminApi)))

	servicesApi := service.NewApi(db)
	service := api.NewJwtMiddleware(cfg.JWTSecret, []string{"admin", "service"})
	r.Handle("/service/", service(http.StripPrefix("/service", servicesApi)))

	// Wrap the entire router with middleware chain
	// Note: Security headers (HSTS, CSP, X-Frame-Options, etc.) are set by nginx
	var handler http.Handler = r
	handler = api.PostMethodOnlyMiddleware(handler) // Enforce POST method
	handler = api.RateLimitMiddleware(100)(handler) // Rate limiting
	handler = api.LoggingMiddleware(handler)        // Logging
	handler = api.CorrelationIDMiddleware(handler)  // Correlation ID

	return handler
}

// Run starts the application with graceful shutdown
func (app *App) Run(ctx context.Context) error {
	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		log.Info().Int("port", app.Config.Port).Msg("Starting server")
		if err := app.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// Wait for context cancellation or server error
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
	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := app.Server.Shutdown(ctx); err != nil {
		log.Printf("ERROR shutting down server: %v", err)
	}

	fmt.Println("Shutdown complete")
	return nil
}
