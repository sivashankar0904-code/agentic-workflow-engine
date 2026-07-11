package main

import (
	"context"
	"log"

	"orchestrator/internals/config"
	"orchestrator/internals/dagconfig"
	"orchestrator/internals/kafka"
	"orchestrator/internals/server"
)

func main() {
	ctx := context.Background()

	cfg := config.Load()

	store, err := dagconfig.NewStore(ctx, cfg.DAGDir, cfg.DAGKey)
	if err != nil {
		log.Fatalf("failed to load DAG: %v", err)
	}

	orch, err := kafka.New(cfg.KafkaBroker, store)
	if err != nil {
		log.Fatalf("failed to start kafka clients: %v", err)
	}
	defer orch.Close()

	go orch.Run(ctx)

	log.Println("Orchestrator running on :8000")
	if err := server.New(store).Run(":8000"); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
