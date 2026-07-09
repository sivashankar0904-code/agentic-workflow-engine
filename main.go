package main

import (
	"context"
	"log"

	"orchestrator/internals/config"
	"orchestrator/internals/kafka"
	"orchestrator/internals/server"
)

const dagFile = "dag.yaml"

func main() {
	store, err := config.NewStore(dagFile)
	if err != nil {
		log.Fatalf("failed to load DAG: %v", err)
	}

	orch, err := kafka.New(store)
	if err != nil {
		log.Fatalf("failed to start kafka clients: %v", err)
	}
	defer orch.Close()

	go orch.Run(context.Background())

	log.Println("Orchestrator running on :8000")
	if err := server.New(store).Run(":8000"); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
