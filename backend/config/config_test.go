//go:build unit

package config

import (
	"os"
	"strings"
	"testing"
)

// valid secrets for the happy-path Load tests (>= MinSecretLength, no placeholder).
const (
	validJWTSecret  = "test-jwt-secret-0123456789"
	validRelayToken = "test-relay-token-0123456789"
	validPGPassword = "test-postgres-password-1234" //nolint:gosec // test placeholder, not a real secret
)

func setValidSecrets(t *testing.T) {
	t.Helper()
	t.Setenv("JWT_SECRET", validJWTSecret)
	t.Setenv("RELAY_AUTH_TOKEN", validRelayToken)
	t.Setenv("POSTGRES_PASSWORD", validPGPassword)
}

func TestLoad_Defaults(t *testing.T) {
	os.Clearenv()
	setValidSecrets(t)

	cfg := Load()

	if cfg.Port != 3000 {
		t.Errorf("expected default port 3000, got %d", cfg.Port)
	}
	if cfg.Postgres.Host != "localhost" {
		t.Errorf("expected default Postgres host 'localhost', got %s", cfg.Postgres.Host)
	}
	if cfg.Postgres.Port != 5432 {
		t.Errorf("expected default Postgres port 5432, got %d", cfg.Postgres.Port)
	}
	if cfg.Postgres.User != "admin" {
		t.Errorf("expected default Postgres user 'admin', got %s", cfg.Postgres.User)
	}
	if cfg.Postgres.Password != validPGPassword {
		t.Errorf("expected Postgres password %q, got %s", validPGPassword, cfg.Postgres.Password)
	}
	if cfg.Postgres.DBName != "jotti" {
		t.Errorf("expected default Postgres DBName 'jotti', got %s", cfg.Postgres.DBName)
	}
	if cfg.FiskalyBaseURL != "https://kassensichv-middleware.fiskaly.com" {
		t.Errorf("expected default Fiskaly base URL, got %s", cfg.FiskalyBaseURL)
	}
	if cfg.RelayToken != validRelayToken {
		t.Errorf("expected relay token %q, got %s", validRelayToken, cfg.RelayToken)
	}
}

func TestLoad_EnvValues(t *testing.T) {
	os.Clearenv()
	setValidSecrets(t)

	t.Setenv("PORT", "8080")
	t.Setenv("POSTGRES_USER", "testuser")
	t.Setenv("POSTGRES_HOST", "db")
	t.Setenv("POSTGRES_PORT", "5433")
	t.Setenv("POSTGRES_DBNAME", "testdb")
	t.Setenv("FISKALY_BASE_URL", "https://example.invalid")

	cfg := Load()

	if cfg.Port != 8080 {
		t.Errorf("expected port 8080, got %d", cfg.Port)
	}
	if cfg.Postgres.User != "testuser" {
		t.Errorf("expected Postgres user 'testuser', got %s", cfg.Postgres.User)
	}
	if cfg.Postgres.Password != validPGPassword {
		t.Errorf("expected Postgres password %q, got %s", validPGPassword, cfg.Postgres.Password)
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
	if cfg.FiskalyBaseURL != "https://example.invalid" {
		t.Errorf("expected Fiskaly base URL 'https://example.invalid', got %s", cfg.FiskalyBaseURL)
	}
}

func TestLoad_InvalidIntAndLowValues(t *testing.T) {
	os.Clearenv()
	setValidSecrets(t)

	t.Setenv("PORT", "notanint")
	t.Setenv("POSTGRES_PORT", "invalid")

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
	setValidSecrets(t)

	t.Setenv("PORT", "-1")

	cfg := Load()

	if cfg.Port != 3000 {
		t.Errorf("expected fallback port 3000 for negative value, got %d", cfg.Port)
	}
}

func baseValidConfig() Config {
	return Config{
		JWTSecret:  validJWTSecret,
		RelayToken: validRelayToken,
		Postgres:   postgresConfig{Password: validPGPassword},
	}
}

func TestValidateSecrets_Valid(t *testing.T) {
	if err := ValidateSecrets(baseValidConfig()); err != nil {
		t.Fatalf("expected valid config to pass, got %v", err)
	}
}

func TestValidateSecrets_Rejects(t *testing.T) {
	cases := []struct {
		name    string
		mutate  func(*Config)
		wantVar string
	}{
		{"empty JWT_SECRET", func(c *Config) { c.JWTSecret = "" }, "JWT_SECRET"},
		{"empty RELAY_AUTH_TOKEN", func(c *Config) { c.RelayToken = "" }, "RELAY_AUTH_TOKEN"},
		{"empty POSTGRES_PASSWORD", func(c *Config) { c.Postgres.Password = "" }, "POSTGRES_PASSWORD"},
		{"placeholder JWT_SECRET", func(c *Config) { c.JWTSecret = "your-256-bit-secret-replace-this-in-production" }, "JWT_SECRET"},
		{"placeholder RELAY_AUTH_TOKEN", func(c *Config) { c.RelayToken = "your-relay-auth-token-replace-this-in-production" }, "RELAY_AUTH_TOKEN"},
		{"placeholder POSTGRES_PASSWORD", func(c *Config) { c.Postgres.Password = "your-secure-password-here" }, "POSTGRES_PASSWORD"},
		{"guessable POSTGRES_PASSWORD", func(c *Config) { c.Postgres.Password = "admin" }, "POSTGRES_PASSWORD"},
		{"short JWT_SECRET", func(c *Config) { c.JWTSecret = "short" }, "JWT_SECRET"},
		{"short RELAY_AUTH_TOKEN", func(c *Config) { c.RelayToken = "short" }, "RELAY_AUTH_TOKEN"},
		{"short POSTGRES_PASSWORD", func(c *Config) { c.Postgres.Password = "short" }, "POSTGRES_PASSWORD"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := baseValidConfig()
			tc.mutate(&cfg)
			err := ValidateSecrets(cfg)
			if err == nil {
				t.Fatalf("expected error for %s, got nil", tc.name)
			}
			if !strings.Contains(err.Error(), tc.wantVar) {
				t.Errorf("expected error to name %q, got %q", tc.wantVar, err.Error())
			}
		})
	}
}

func TestValidateSecrets_LengthBoundary(t *testing.T) {
	atMin := strings.Repeat("x", MinSecretLength)
	cfg := baseValidConfig()
	cfg.JWTSecret = atMin
	if err := ValidateSecrets(cfg); err != nil {
		t.Errorf("expected %d-char secret to pass, got %v", MinSecretLength, err)
	}

	cfg.JWTSecret = strings.Repeat("x", MinSecretLength-1)
	if err := ValidateSecrets(cfg); err == nil {
		t.Errorf("expected %d-char secret to fail", MinSecretLength-1)
	}
}
