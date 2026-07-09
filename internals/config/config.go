package config

import "os"

// Config holds all runtime configuration, sourced from the environment with
// static fallbacks.
type Config struct {
	// Kafka
	KafkaBroker string

	// S3 / MinIO
	S3Endpoint      string
	S3Bucket        string
	S3Region        string
	S3AccessKey     string
	S3SecretKey     string
	DAGKey          string // active DAG object key; the bucket holds only DAG files
}

// Load reads configuration from the environment, applying static defaults.
func Load() Config {
	return Config{
		KafkaBroker: env("KAFKA_BROKER", "localhost:9092"),

		S3Endpoint:  env("S3_ENDPOINT", "http://localhost:9000"),
		S3Bucket:    os.Getenv("S3_BUCKET"),
		S3Region:    env("AWS_REGION", "us-east-1"),
		S3AccessKey: os.Getenv("AWS_ACCESS_KEY_ID"),
		S3SecretKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
		DAGKey:      env("S3_DAG_KEY", "dag.yaml"),
	}
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
