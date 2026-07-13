package config

import "os"

// Config holds all runtime configuration, sourced from the environment with
// static fallbacks.
type Config struct {
	// Postgres connection string (pgx/libpq DSN or URL).
	DatabaseURL string
	// JWTSecret signs and verifies session tokens. The fallback below is for
	// local development only — production deployments MUST override it via
	// the JWT_SECRET environment variable.
	JWTSecret string
}

// Load reads configuration from the environment, applying static defaults.
func Load() Config {
	return Config{
		DatabaseURL: env("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/controlplane?sslmode=disable"),
		JWTSecret:   env("JWT_SECRET", "dev-jwt-secret-change-me"),
	}
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
