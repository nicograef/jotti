package app

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/nicograef/jotti/backend/api"
	"github.com/nicograef/jotti/backend/config"
	"github.com/nicograef/jotti/backend/event"
	"github.com/nicograef/jotti/backend/order"
	"github.com/nicograef/jotti/backend/product"
	"github.com/nicograef/jotti/backend/table"
	"github.com/nicograef/jotti/backend/user"
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

	admin := user.NewAdminMiddleware(cfg.JWTSecret)
	service := user.NewServiceMiddleware(cfg.JWTSecret)

	// Health check with database connectivity
	healthCheck := api.HealthCheck{DB: db}
	r.HandleFunc("/health", healthCheck.Handler())

	userPersistence := user.Persistence{DB: db}
	ah := user.AuthHandler{JWTSecret: cfg.JWTSecret, Command: &user.Command{Persistence: &userPersistence}}
	r.HandleFunc("/login", ah.LoginHandler())
	r.HandleFunc("/set-password", ah.SetPasswordHandler())

	uch := user.CommandHandler{Command: &user.Command{Persistence: &userPersistence}}
	r.HandleFunc("/create-user", admin(uch.CreateUserHandler()))
	r.HandleFunc("/update-user", admin(uch.UpdateUserHandler()))
	r.HandleFunc("/activate-user", admin(uch.ActivateUserHandler()))
	r.HandleFunc("/deactivate-user", admin(uch.DeactivateUserHandler()))
	r.HandleFunc("/reset-password", admin(uch.ResetPasswordHandler()))
	uqh := user.QueryHandler{Query: &user.Query{Persistence: &userPersistence}}
	r.HandleFunc("/get-all-users", admin(uqh.GetAllUsersHandler()))

	tablePersistence := table.Persistence{DB: db}
	tch := table.CommandHandler{Command: &table.Command{Persistence: &tablePersistence}}
	r.HandleFunc("/update-table", admin(tch.UpdateTableHandler()))
	r.HandleFunc("/create-table", admin(tch.CreateTableHandler()))
	r.HandleFunc("/activate-table", admin(tch.ActivateTableHandler()))
	r.HandleFunc("/deactivate-table", admin(tch.DeactivateTableHandler()))
	tqh := table.QueryHandler{Query: &table.Query{Persistence: &tablePersistence}}
	r.HandleFunc("/get-table", service(tqh.GetTableHandler()))
	r.HandleFunc("/get-active-tables", service(tqh.GetActiveTablesHandler()))
	r.HandleFunc("/get-all-tables", admin(tqh.GetAllTablesHandler()))

	productPersistence := product.Persistence{DB: db}
	pch := product.CommandHandler{Command: &product.Command{Persistence: &productPersistence}}
	r.HandleFunc("/create-product", admin(pch.CreateProductHandler()))
	r.HandleFunc("/update-product", admin(pch.UpdateProductHandler()))
	r.HandleFunc("/activate-product", admin(pch.ActivateProductHandler()))
	r.HandleFunc("/deactivate-product", admin(pch.DeactivateProductHandler()))
	pqh := product.QueryHandler{Query: &product.Query{Persistence: &productPersistence}}
	r.HandleFunc("/get-active-products", service(pqh.GetActiveProductsHandler()))
	r.HandleFunc("/get-all-products", admin(pqh.GetAllProductsHandler()))

	eventPersistence := event.Persistence{DB: db}
	oh := order.NewHandler(&eventPersistence)
	r.HandleFunc("/place-order", service(oh.PlaceOrderHandler()))
	r.HandleFunc("/get-orders", service(oh.GetOrdersHandler()))

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
