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

// Config holds application configuration values loaded from environment variables.
type Config struct {
	Port           int // Port for the HTTP server
	Postgres       postgresConfig
	JWTSecret      string // Secret key for JWT signing
	RelayToken     string // Statischer Token für das Print-Relay
	FiskalyBaseURL string // Basis-URL fuer fiskaly SIGN-DE Middleware API
	EnableTestApi  bool   // Schaltet den HTTP-Test-Reset-Endpoint frei (nur E2E, JOTTI_ENABLE_TEST_API=1)
}

// MinSecretLength ist die geforderte Mindestlänge für JWT_SECRET,
// RELAY_AUTH_TOKEN und POSTGRES_PASSWORD. Kürzere Werte werden abgelehnt,
// damit versehentliche Kurz-Secrets nicht in eine laufende Instanz gelangen.
const MinSecretLength = 16

// placeholderSecrets sind die im Repo öffentlich stehenden Beispielwerte aus
// .env.example sowie der frühere POSTGRES_PASSWORD-Default. Ein solcher Wert in
// einer laufenden Instanz bedeutet ein bekanntes Secret (JWT-Forgery = Auth-Bypass)
// und wird deshalb hart abgelehnt.
var placeholderSecrets = map[string]bool{
	"your-256-bit-secret-replace-this-in-production":   true,
	"your-relay-auth-token-replace-this-in-production": true,
	"your-secure-password-here":                        true,
	"admin":                                            true,
}

// Load reads configuration from environment variables and returns a Config struct.
// Defaults: PORT=3000, POSTGRES_HOST="localhost", POSTGRES_PORT=5432.
// JWT_SECRET, RELAY_AUTH_TOKEN und POSTGRES_PASSWORD sind Pflicht und werden
// validiert; fehlt oder verstößt ein Secret, bricht der Start hart ab.
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

// ValidateSecrets prüft die drei Pflicht-Secrets der Konfiguration. Die Funktion
// ist rein (kein os.Exit), damit die Regeln testbar sind; Load ruft sie und bricht
// bei einem Fehler hart ab.
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

// validateSecret setzt die Regeln für ein einzelnes Secret durch: nicht leer,
// kein bekannter Platzhalter, mindestens MinSecretLength Zeichen. Die Fehlermeldung
// nennt immer die betroffene Variable.
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

// parseEnvString reads an environment variable by name and returns its value, or the provided default if unset.
func parseEnvString(name, defaultValue string) string {
	v := os.Getenv(name)
	if v == "" && defaultValue != "" {
		return defaultValue
	}
	if v == "" {
		log.Fatalf("%s is not set and has no default value\n", name)
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
