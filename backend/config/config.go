package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

type postgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

type Config struct {
	Port           int
	Postgres       postgresConfig
	JWTSecret      string
	RelayToken     string
	FiskalyBaseURL string
	EnableTestApi  bool
}

const MinSecretLength = 16

// placeholderSecrets sind die öffentlich im Repo stehenden Beispielwerte aus
// .env.example plus das erratbare Postgres-Passwort "admin": ein bekanntes Secret
// heißt JWT-Forgery und damit Auth-Bypass, deshalb harte Ablehnung.
var placeholderSecrets = map[string]bool{
	"your-256-bit-secret-replace-this-in-production":   true,
	"your-relay-auth-token-replace-this-in-production": true,
	"your-secure-password-here":                        true,
	"admin":                                            true,
}

// Load bricht den Start hart ab, wenn ein Pflicht-Secret fehlt oder die Regeln
// verletzt.
func Load() Config {
	port := parseEnvInt("PORT", 3000)
	postgres := postgresConfig{
		Host:     parseEnvString("POSTGRES_HOST", "localhost"),
		Port:     parseEnvInt("POSTGRES_PORT", 5432),
		User:     parseEnvString("POSTGRES_USER", "admin"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		DBName:   parseEnvString("POSTGRES_DBNAME", "jotti"),
	}

	cfg := Config{
		Port:           port,
		Postgres:       postgres,
		JWTSecret:      os.Getenv("JWT_SECRET"),
		RelayToken:     os.Getenv("RELAY_AUTH_TOKEN"),
		FiskalyBaseURL: parseEnvString("FISKALY_BASE_URL", "https://kassensichv-middleware.fiskaly.com"),
		EnableTestApi:  os.Getenv("JOTTI_ENABLE_TEST_API") == "1",
	}

	if err := ValidateSecrets(cfg); err != nil {
		log.Fatalf("invalid configuration: %v\n", err)
	}

	return cfg
}

func ValidateSecrets(cfg Config) error {
	if err := validateSecret("JWT_SECRET", cfg.JWTSecret); err != nil {
		return err
	}
	if err := validateSecret("RELAY_AUTH_TOKEN", cfg.RelayToken); err != nil {
		return err
	}
	if err := validateSecret("POSTGRES_PASSWORD", cfg.Postgres.Password); err != nil {
		return err
	}
	return nil
}

func validateSecret(name, value string) error {
	if value == "" {
		return fmt.Errorf("%s is not set", name)
	}
	if placeholderSecrets[value] {
		return fmt.Errorf("%s uses a known placeholder value from .env.example; set a real secret (run 'make init')", name)
	}
	if len(value) < MinSecretLength {
		return fmt.Errorf("%s is too short (%d chars); need at least %d", name, len(value), MinSecretLength)
	}
	return nil
}

func parseEnvString(name, defaultValue string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}

	return defaultValue
}

func parseEnvInt(name string, defaultValue int) int {
	v := os.Getenv(name)
	if v == "" {
		return defaultValue
	}

	n, err := strconv.Atoi(v)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid %s value: %v\n", name, err)
		return defaultValue
	}

	if n < 1 {
		fmt.Fprintf(os.Stderr, "Invalid %s value: must be at least 1\n", name)
		return defaultValue
	}

	return n
}
