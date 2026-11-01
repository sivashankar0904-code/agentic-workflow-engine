package config

import (
	"controlplane/internal/datastore"
	"controlplane/internal/http"
	"os"
	"strings"
	"time"
)

// Env-backed application settings.
type Config struct {
	// Secret key for session token signing/verification.
	JWTSecret string
	// Postgres connection string, e.g. postgres://user:pass@host:port/db.
	DatabaseURL string
}

// Load builds a Config from env vars.
func Load() *Config {
	cfg := &Config{}
	cfg.JWTSecret = cfg.getEnv("JWT_SECRET", "")
	cfg.DatabaseURL = cfg.getEnv("DATABASE_URL", "")
	return cfg
}

// Fetches an env var, falling back to defaultValue when unset or blank.
func (cfg *Config) getEnv(key string, defaultValue string) string {
	v := strings.Trim(strings.TrimSpace(os.Getenv(key)), "\n")
	if v == "" {
		v = defaultValue
	}
	return v
}

// Builds the HTTP server config from env vars.
func (cfg *Config) HTTP() *http.Config {
	host := cfg.getEnv("HTTP_HOST", "")
	port := cfg.getEnv("HTTP_PORT", "8001")

	hcfg := &http.Config{
		Host:         host,
		Port:         port,
		ReadTimeout:  time.Second * 30,
		WriteTimeout: time.Second * 60,
	}
	return hcfg
}

// Assembles datastore connection settings from the key vault.
func (cfg *Config) DataStore() *datastore.Config {

	return &datastore.Config{
		Driver:    "postgres",
		Host:      cfg.getEnv("DATA_STORE_HOST", ""),
		Port:      cfg.getEnv("DATA_STORE_PORT", ""),
		Username:  cfg.getEnv("DATA_STORE_USER", ""),
		Password:  cfg.getEnv("DATA_STORE_PASSWORD", ""),
		StoreName: cfg.getEnv("DATA_STORE_NAME", ""),
		SSLMode:   cfg.getEnv("DATA_STORE_SSLMODE", ""),

		DialTimeoutSecs:  time.Second * 5,
		ReadTimeoutSecs:  time.Second * 10,
		WriteTimeoutSecs: time.Second * 20,
		IdleTimeoutSecs:  time.Second * 120,
		PoolSize:         64,
	}
}
