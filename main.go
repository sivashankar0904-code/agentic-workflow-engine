package main

import (
	"log"

	"orchestrator/internals/config"
	"orchestrator/internals/dagconfig"
	"orchestrator/internals/server"
)

func main() {
	cfg := config.Load()

	store, err := dagconfig.NewStore(cfg.DAGDir)
	if err != nil {
		log.Fatalf("failed to init DAG store: %v", err)
	}

	log.Println("Orchestrator running on :8000")
	if err := server.New(store).Run(":8000"); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
