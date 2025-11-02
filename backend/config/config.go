package config

import (
	"fmt"
	"os"
	"strconv"
)

type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

// Config holds application configuration values loaded from environment variables.
type Config struct {
	Port      int // Port for the HTTP server
	Postgres  PostgresConfig
	JWTSecret string // Secret key for JWT signing
}

// Load reads configuration from environment variables and returns a Config struct.
// Defaults: PORT=3000 CAPACITY=1000, CONSUMER_URL="http://localhost:4000" DELIVERY_ATTEMPTS=3
func Load() Config {
	port := parseEnvInt("PORT", 3000)
	postgres := PostgresConfig{
		Host:     parseEnvString("POSTGRES_HOST", "localhost"),
		Port:     parseEnvInt("POSTGRES_PORT", 5432),
		User:     parseEnvString("POSTGRES_USER", "admin"),
		Password: parseEnvString("POSTGRES_PASSWORD", "admin"),
		DBName:   parseEnvString("POSTGRES_DBNAME", "jotti"),
	}
	jwtSecret := parseEnvString("JWT_SECRET", "your-256-bit-secret")

	return Config{
		Port:      port,
		Postgres:  postgres,
		JWTSecret: jwtSecret,
	}
}

// parseEnvString reads an environment variable by name and returns its value, or the provided default if unset.
func parseEnvString(name, defaultValue string) string {
	v := os.Getenv(name)
	if v == "" && defaultValue != "" {
		return defaultValue
	} else if v == "" {
		fmt.Fprintf(os.Stderr, "Warning: %s is not set and has no default value\n", name)
	}

	return v
}

// parseEnvInt reads an environment variable by name and converts it to int.
// If conversion fails, logs an error and returns the provided default value.
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
