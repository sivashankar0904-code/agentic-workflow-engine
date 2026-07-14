package config

import "os"

// Config holds all runtime configuration, sourced from the environment with
// static fallbacks.
type Config struct {
	// Base URL of the Control Plane Service, the sole source of DAG definitions.
	ControlPlaneURL string
	// ServiceKey authenticates this engine to the Control Plane (sent as the
	// X-Service-Key header). The fallback matches the dev seed in
	// control_plane/schemas/04_services.sql; production must override it.
	ServiceKey string
}

// Load reads configuration from the environment, applying static defaults.
func Load() Config {
	return Config{
		ControlPlaneURL: env("CONTROL_PLANE_URL", "http://localhost:9000"),
		ServiceKey:      env("CONTROL_PLANE_SERVICE_KEY", "dev-orchestrator-api-key-change-me"),
	}
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
