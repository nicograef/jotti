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
	"github.com/nicograef/jotti/backend/product_admin"
	"github.com/nicograef/jotti/backend/product_service"
	"github.com/nicograef/jotti/backend/table_admin"
	"github.com/nicograef/jotti/backend/table_service"
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

	// Admin API //
	uch := user.CommandHandler{Command: &user.Command{Persistence: &userPersistence}}
	r.HandleFunc("/admin/create-user", admin(uch.CreateUserHandler()))
	r.HandleFunc("/admin/update-user", admin(uch.UpdateUserHandler()))
	r.HandleFunc("/admin/activate-user", admin(uch.ActivateUserHandler()))
	r.HandleFunc("/admin/deactivate-user", admin(uch.DeactivateUserHandler()))
	r.HandleFunc("/admin/reset-password", admin(uch.ResetPasswordHandler()))

	uqh := user.QueryHandler{Query: &user.Query{Persistence: &userPersistence}}
	r.HandleFunc("/admin/get-all-users", admin(uqh.GetAllUsersHandler()))

	pac := product_admin.CommandHandler{Command: &product_admin.Command{Persistence: &product_admin.Persistence{DB: db}}}
	r.HandleFunc("/admin/create-product", admin(pac.CreateProductHandler()))
	r.HandleFunc("/admin/update-product", admin(pac.UpdateProductHandler()))
	r.HandleFunc("/admin/activate-product", admin(pac.ActivateProductHandler()))
	r.HandleFunc("/admin/deactivate-product", admin(pac.DeactivateProductHandler()))

	paq := product_admin.QueryHandler{Query: &product_admin.Query{Persistence: &product_admin.Persistence{DB: db}}}
	r.HandleFunc("/admin/get-all-products", admin(paq.GetAllProductsHandler()))

	tac := table_admin.CommandHandler{Command: &table_admin.Command{Persistence: &table_admin.Persistence{DB: db}}}
	r.HandleFunc("/admin/update-table", admin(tac.UpdateTableHandler()))
	r.HandleFunc("/admin/create-table", admin(tac.CreateTableHandler()))
	r.HandleFunc("/admin/activate-table", admin(tac.ActivateTableHandler()))
	r.HandleFunc("/admin/deactivate-table", admin(tac.DeactivateTableHandler()))

	taq := table_admin.QueryHandler{Query: &table_admin.Query{Persistence: &table_admin.Persistence{DB: db}}}
	r.HandleFunc("/admin/get-all-tables", admin(taq.GetAllTablesHandler()))
	// End of Admin API //

	// Service API //
	psq := product_service.QueryHandler{Query: &product_service.Query{Persistence: &product_service.Persistence{DB: db}}}
	r.HandleFunc("/service/get-all-products", service(psq.GetAllProductsHandler()))

	tsc := table_service.CommandHandler{Command: &table_service.Command{Persistence: &event.Persistence{DB: db}}}
	r.HandleFunc("/service/place-order", service(tsc.PlaceOrderHandler()))

	tsq := table_service.QueryHandler{Query: &table_service.Query{EventPersistence: &event.Persistence{DB: db}, TablePersistence: &table_service.Persistence{DB: db}}}
	r.HandleFunc("/service/get-table", service(tsq.GetTableHandler()))
	r.HandleFunc("/service/get-all-tables", service(tsq.GetAllTablesHandler()))
	r.HandleFunc("/service/get-orders", service(tsq.GetOrdersHandler()))
	r.HandleFunc("/service/get-table-balance", service(tsq.GetTableBalanceHandler()))
	r.HandleFunc("/service/get-table-unpaid-products", service(tsq.GetTableUnpaidProductsHandler()))
	// End of Service API //

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
