//go:build unit

package app

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/nicograef/jotti/backend/config"
)

func TestNewApp(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-for-app-tests")
	cfg := config.Load()
	app, err := NewApp(cfg, &sql.DB{})
	if err != nil {
		t.Fatalf("NewApp() failed: %v", err)
	}

	if app.Server == nil {
		t.Error("Server should not be nil")
	}

	if app.Config.Port != 3000 {
		t.Errorf("Expected port 3000, got %d", app.Config.Port)
	}
}

func TestSetupRoutes(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-for-app-tests")
	cfg := config.Load()

	handler := SetupRoutes(cfg, &sql.DB{})

	if handler == nil {
		t.Error("Handler should not be nil")
	}
}

func TestShutdown(t *testing.T) {
	cfg := config.Load()
	app, err := NewApp(cfg, &sql.DB{})
	if err != nil {
		t.Fatalf("NewApp() failed: %v", err)
	}

	err = app.Shutdown()
	if err != nil {
		t.Errorf("Shutdown() failed: %v", err)
	}

}

func TestRun_ContextCancellation(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-for-app-tests")
	cfg := config.Load()
	app, err := NewApp(cfg, &sql.DB{})
	if err != nil {
		t.Fatalf("NewApp() failed: %v", err)
	}

	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	// Run the app in a separate goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- app.Run(ctx)
	}()

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	// Cancel the context to trigger shutdown
	cancel()

	// Wait for Run to return
	err = <-errChan
	if err != nil {
		t.Errorf("Run() returned error: %v", err)
	}
}
