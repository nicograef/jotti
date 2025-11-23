package app

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

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
	authService := auth.Service{JWTSecret: app.Config.JWTSecret}
	jwtMiddleware := auth.NewJWTMiddleware(&authService)

	app.Router.HandleFunc("/health", api.NewHealthHandler())

	userPersistence := user.Persistence{DB: app.DB}
	userService := user.Service{Persistence: &userPersistence}
	uh := user.Handler{Service: &userService}
	app.Router.HandleFunc("/login", auth.LoginHandler(&userService, &authService))
	app.Router.HandleFunc("/set-password", auth.SetPasswordHandler(&userService, &authService))
	app.Router.HandleFunc("/admin/create-user", jwtMiddleware(auth.AdminMiddleware(uh.CreateUserHandler())))
	app.Router.HandleFunc("/admin/update-user", jwtMiddleware(auth.AdminMiddleware(uh.UpdateUserHandler())))
	app.Router.HandleFunc("/admin/activate-user", jwtMiddleware(auth.AdminMiddleware(uh.ActivateUserHandler())))
	app.Router.HandleFunc("/admin/deactivate-user", jwtMiddleware(auth.AdminMiddleware(uh.DeactivateUserHandler())))
	app.Router.HandleFunc("/admin/get-users", jwtMiddleware(auth.AdminMiddleware(uh.GetUsersHandler())))
	app.Router.HandleFunc("/admin/reset-password", jwtMiddleware(auth.AdminMiddleware(uh.ResetPasswordHandler())))

	tablePersistence := table.Persistence{DB: app.DB}
	tableService := table.Service{Persistence: &tablePersistence}
	th := table.Handler{Service: &tableService}

	app.Router.HandleFunc("/service/get-tables", jwtMiddleware(auth.ServiceMiddleware(th.GetActiveTablesHandler())))
	app.Router.HandleFunc("/admin/get-tables", jwtMiddleware(auth.AdminMiddleware(th.GetAllTablesHandler())))
	app.Router.HandleFunc("/admin/update-table", jwtMiddleware(auth.AdminMiddleware(th.UpdateTableHandler())))
	app.Router.HandleFunc("/admin/create-table", jwtMiddleware(auth.AdminMiddleware(th.CreateTableHandler())))
	app.Router.HandleFunc("/admin/activate-table", jwtMiddleware(auth.AdminMiddleware(th.ActivateTableHandler())))
	app.Router.HandleFunc("/admin/deactivate-table", jwtMiddleware(auth.AdminMiddleware(th.DeactivateTableHandler())))

	productPersistence := product.Persistence{DB: app.DB}
	productService := product.Service{Persistence: &productPersistence}
	ph := product.Handler{Service: &productService}
	app.Router.HandleFunc("/admin/get-products", jwtMiddleware(auth.AdminMiddleware(ph.GetAllProductsHandler())))
	app.Router.HandleFunc("/admin/create-product", jwtMiddleware(auth.AdminMiddleware(ph.CreateProductHandler())))
	app.Router.HandleFunc("/admin/update-product", jwtMiddleware(auth.AdminMiddleware(ph.UpdateProductHandler())))
	app.Router.HandleFunc("/admin/activate-product", jwtMiddleware(auth.AdminMiddleware(ph.ActivateProductHandler())))
	app.Router.HandleFunc("/admin/deactivate-product", jwtMiddleware(auth.AdminMiddleware(ph.DeactivateProductHandler())))

	app.Server.Handler = app.Router
}

// Run starts the application with graceful shutdown
func (app *App) Run(ctx context.Context) error {
	app.SetupRoutes()

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		fmt.Printf("Starting server on port %d\n", app.Config.Port)
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
