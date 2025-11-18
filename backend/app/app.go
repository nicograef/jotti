package app

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/nicograef/jotti/backend/api"
	"github.com/nicograef/jotti/backend/config"
	"github.com/nicograef/jotti/backend/domain/auth"
	"github.com/nicograef/jotti/backend/domain/table"
	"github.com/nicograef/jotti/backend/domain/user"
	"github.com/nicograef/jotti/backend/persistence"
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
	userPersistence := persistence.UserPersistence{DB: app.DB}
	tablePersistence := persistence.TablePersistence{DB: app.DB}
	userService := user.Service{DB: &userPersistence}
	tableService := table.Service{DB: &tablePersistence}
	authService := auth.Service{JWTSecret: app.Config.JWTSecret}
	jwtMiddleware := api.NewJWTMiddleware(&authService)

	app.Router.HandleFunc("/health", api.CorsHandler(api.NewHealthHandler()))

	app.Router.HandleFunc("/login", api.CorsHandler(api.LoginHandler(&userService, &authService)))
	app.Router.HandleFunc("/set-password", api.CorsHandler(api.SetPasswordHandler(&userService, &authService)))

	app.Router.HandleFunc("/admin/create-user", api.CorsHandler(jwtMiddleware(api.AdminMiddleware(api.CreateUserHandler(&userService)))))
	app.Router.HandleFunc("/admin/update-user", api.CorsHandler(jwtMiddleware(api.AdminMiddleware(api.UpdateUserHandler(&userService)))))
	app.Router.HandleFunc("/admin/activate-user", api.CorsHandler(jwtMiddleware(api.AdminMiddleware(api.ActivateUserHandler(&userService)))))
	app.Router.HandleFunc("/admin/deactivate-user", api.CorsHandler(jwtMiddleware(api.AdminMiddleware(api.DeactivateUserHandler(&userService)))))
	app.Router.HandleFunc("/admin/get-users", api.CorsHandler(jwtMiddleware(api.AdminMiddleware(api.GetUsersHandler(&userService)))))
	app.Router.HandleFunc("/admin/reset-password", api.CorsHandler(jwtMiddleware(api.AdminMiddleware(api.ResetPasswordHandler(&userService)))))

	app.Router.HandleFunc("/admin/get-tables", api.CorsHandler(jwtMiddleware(api.AdminMiddleware(api.GetTablesHandler(&tableService)))))
	app.Router.HandleFunc("/admin/update-table", api.CorsHandler(jwtMiddleware(api.AdminMiddleware(api.UpdateTableHandler(&tableService)))))
	app.Router.HandleFunc("/admin/create-table", api.CorsHandler(jwtMiddleware(api.AdminMiddleware(api.CreateTableHandler(&tableService)))))
	app.Router.HandleFunc("/admin/activate-table", api.CorsHandler(jwtMiddleware(api.AdminMiddleware(api.ActivateTableHandler(&tableService)))))
	app.Router.HandleFunc("/admin/deactivate-table", api.CorsHandler(jwtMiddleware(api.AdminMiddleware(api.DeactivateTableHandler(&tableService)))))

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
		log.Printf("Error shutting down server: %v", err)
	}

	fmt.Println("Shutdown complete")
	return nil
}
