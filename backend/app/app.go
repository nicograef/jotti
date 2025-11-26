package app

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/nicograef/jotti/backend/api"
	"github.com/nicograef/jotti/backend/auth"
	"github.com/nicograef/jotti/backend/config"
	"github.com/nicograef/jotti/backend/product"
	"github.com/nicograef/jotti/backend/table"
	"github.com/nicograef/jotti/backend/user"
)

// App represents the application with its configuration, router, server, and database connection.
type App struct {
	Server *http.Server
	Config config.Config
	Router *http.ServeMux
	DB     *sql.DB
}

// NewApp creates a new application instance
func NewApp(cfg config.Config, db *sql.DB) (*App, error) {
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	router := http.NewServeMux()

	return &App{
		Server: server,
		Config: cfg,
		Router: router,
		DB:     db,
	}, nil
}

// SetupRoutes configures HTTP routes
func (app *App) SetupRoutes() {
	// Health check with database connectivity
	healthCheck := api.HealthCheck{DB: app.DB}
	app.Router.HandleFunc("/health", healthCheck.Handler())

	userPersistence := user.Persistence{DB: app.DB}
	userService := user.Service{Persistence: &userPersistence}

	ah := auth.Handler{JWTSecret: app.Config.JWTSecret, UserService: &userService}
	app.Router.HandleFunc("/login", ah.LoginHandler())
	app.Router.HandleFunc("/set-password", ah.SetPasswordHandler())

	admin := auth.NewAdminMiddleware(app.Config.JWTSecret)
	service := auth.NewServiceMiddleware(app.Config.JWTSecret)

	uh := user.Handler{Service: &userService}
	app.Router.HandleFunc("/create-user", admin(uh.CreateUserHandler()))
	app.Router.HandleFunc("/update-user", admin(uh.UpdateUserHandler()))
	app.Router.HandleFunc("/activate-user", admin(uh.ActivateUserHandler()))
	app.Router.HandleFunc("/deactivate-user", admin(uh.DeactivateUserHandler()))
	app.Router.HandleFunc("/get-all-users", admin(uh.GetAllUsersHandler()))
	app.Router.HandleFunc("/reset-password", admin(uh.ResetPasswordHandler()))

	tablePersistence := table.Persistence{DB: app.DB}
	tableService := table.Service{Persistence: &tablePersistence}
	th := table.Handler{Service: &tableService}
	app.Router.HandleFunc("/get-active-tables", service(th.GetActiveTablesHandler()))
	app.Router.HandleFunc("/get-all-tables", admin(th.GetAllTablesHandler()))
	app.Router.HandleFunc("/update-table", admin(th.UpdateTableHandler()))
	app.Router.HandleFunc("/create-table", admin(th.CreateTableHandler()))
	app.Router.HandleFunc("/activate-table", admin(th.ActivateTableHandler()))
	app.Router.HandleFunc("/deactivate-table", admin(th.DeactivateTableHandler()))

	productPersistence := product.Persistence{DB: app.DB}
	productService := product.Service{Persistence: &productPersistence}
	ph := product.Handler{Service: &productService}
	app.Router.HandleFunc("/get-active-products", service(ph.GetActiveProductsHandler()))
	app.Router.HandleFunc("/get-all-products", admin(ph.GetAllProductsHandler()))
	app.Router.HandleFunc("/create-product", admin(ph.CreateProductHandler()))
	app.Router.HandleFunc("/update-product", admin(ph.UpdateProductHandler()))
	app.Router.HandleFunc("/activate-product", admin(ph.ActivateProductHandler()))
	app.Router.HandleFunc("/deactivate-product", admin(ph.DeactivateProductHandler()))

	// Wrap the entire router with middleware chain
	var handler http.Handler = app.Router
	handler = api.RateLimitMiddleware(100)(handler)
	handler = api.LoggingMiddleware(handler)

	app.Server.Handler = handler
}

// Run starts the application with graceful shutdown
func (app *App) Run(ctx context.Context) error {
	app.SetupRoutes()

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
