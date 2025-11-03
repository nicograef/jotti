//go:build unit

package config

import (
	"os"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	os.Clearenv()

	cfg := Load()

	if cfg.Port != 3000 {
		t.Errorf("expected default port 3000, got %d", cfg.Port)
	}
	// Postgres defaults
	if cfg.Postgres.Host != "localhost" {
		t.Errorf("expected default Postgres host 'localhost', got %s", cfg.Postgres.Host)
	}
	if cfg.Postgres.Port != 5432 {
		t.Errorf("expected default Postgres port 5432, got %d", cfg.Postgres.Port)
	}
	if cfg.Postgres.User != "admin" {
		t.Errorf("expected default Postgres user 'admin', got %s", cfg.Postgres.User)
	}
	if cfg.Postgres.Password != "admin" {
		t.Errorf("expected default Postgres password 'admin', got %s", cfg.Postgres.Password)
	}
	if cfg.Postgres.DBName != "jotti" {
		t.Errorf("expected default Postgres DBName 'jotti', got %s", cfg.Postgres.DBName)
	}
}

func TestLoad_EnvValues(t *testing.T) {
	os.Clearenv()

	if err := os.Setenv("PORT", "8080"); err != nil {
		t.Fatalf("Failed to set PORT: %v", err)
	}
	if err := os.Setenv("POSTGRES_USER", "testuser"); err != nil {
		t.Fatalf("Failed to set POSTGRES_USER: %v", err)
	}
	if err := os.Setenv("POSTGRES_PASSWORD", "testpassword"); err != nil {
		t.Fatalf("Failed to set POSTGRES_PASSWORD: %v", err)
	}
	if err := os.Setenv("POSTGRES_HOST", "db"); err != nil {
		t.Fatalf("Failed to set POSTGRES_HOST: %v", err)
	}
	if err := os.Setenv("POSTGRES_PORT", "5433"); err != nil {
		t.Fatalf("Failed to set POSTGRES_PORT: %v", err)
	}
	if err := os.Setenv("POSTGRES_DBNAME", "testdb"); err != nil {
		t.Fatalf("Failed to set POSTGRES_DBNAME: %v", err)
	}

	cfg := Load()

	if cfg.Port != 8080 {
		t.Errorf("expected port 8080, got %d", cfg.Port)
	}
	if cfg.Postgres.User != "testuser" {
		t.Errorf("expected Postgres user 'testuser', got %s", cfg.Postgres.User)
	}
	if cfg.Postgres.Password != "testpassword" {
		t.Errorf("expected Postgres password 'testpassword', got %s", cfg.Postgres.Password)
	}
	if cfg.Postgres.Host != "db" {
		t.Errorf("expected Postgres host 'db', got %s", cfg.Postgres.Host)
	}
	if cfg.Postgres.Port != 5433 {
		t.Errorf("expected Postgres port 5433, got %d", cfg.Postgres.Port)
	}
	if cfg.Postgres.DBName != "testdb" {
		t.Errorf("expected Postgres DBName 'testdb', got %s", cfg.Postgres.DBName)
	}
}

func TestLoad_InvalidIntAndLowValues(t *testing.T) {
	os.Clearenv()

	if err := os.Setenv("PORT", "notanint"); err != nil {
		t.Fatalf("Failed to set PORT: %v", err)
	}
	if err := os.Setenv("POSTGRES_PORT", "invalid"); err != nil {
		t.Fatalf("Failed to set POSTGRES_PORT: %v", err)
	}

	cfg := Load()

	if cfg.Port != 3000 {
		t.Errorf("expected fallback port 3000, got %d", cfg.Port)
	}
	if cfg.Postgres.Port != 5432 {
		t.Errorf("expected fallback Postgres port 5432, got %d", cfg.Postgres.Port)
	}
}

func TestLoad_NegativeValues(t *testing.T) {
	os.Clearenv()

	if err := os.Setenv("PORT", "-1"); err != nil {
		t.Fatalf("Failed to set PORT: %v", err)
	}

	cfg := Load()

	// Should fallback to defaults due to validation (must be at least 1)
	if cfg.Port != 3000 {
		t.Errorf("expected fallback port 3000 for negative value, got %d", cfg.Port)
	}
}
