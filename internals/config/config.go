package config

import "os"

// Config holds all runtime configuration, sourced from the environment with
// static fallbacks.
type Config struct {
	// DAG storage (local filesystem)
	DAGDir string // directory holding DAG YAML files
}

// Load reads configuration from the environment, applying static defaults.
func Load() Config {
	return Config{
		DAGDir: env("DAG_DIR", "./dags"),
	}
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
