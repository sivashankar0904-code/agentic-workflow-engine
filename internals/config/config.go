package config

import "os"

// Config holds all runtime configuration, sourced from the environment with
// static fallbacks.
type Config struct {
	// Kafka
	KafkaBroker string

	// DAG storage (local filesystem)
	DAGDir string // directory holding DAG YAML files
	DAGKey string // active DAG file name; the directory holds only DAG files
}

// Load reads configuration from the environment, applying static defaults.
func Load() Config {
	return Config{
		KafkaBroker: env("KAFKA_BROKER", "localhost:9092"),

		DAGDir: env("DAG_DIR", "./dags"),
		DAGKey: env("DAG_KEY", "dag.yaml"),
	}
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
