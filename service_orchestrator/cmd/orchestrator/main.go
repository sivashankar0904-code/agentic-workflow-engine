package main

import (
	"log"

	"orchestrator/internal/config"
	"orchestrator/internal/controlplane"
	"orchestrator/internal/engine"
	"orchestrator/internal/server"
)

func main() {
	cfg := config.Load()

	cp := controlplane.New(cfg.ControlPlaneURL, cfg.ServiceKey)
	registry := engine.NewRegistry(cp)

	if err := registry.Refresh(); err != nil {
		log.Fatalf("failed to build flows from control plane: %v", err)
	}

	log.Println("Orchestrator running on :8000")
	if err := server.New(registry).Run(":8000"); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
