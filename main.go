package main

import (
	"context"
	"log"

	"orchestrator/internals/config"
	"orchestrator/internals/dagconfig"
	"orchestrator/internals/kafka"
	"orchestrator/internals/s3bucket"
	"orchestrator/internals/server"
)

func main() {
	ctx := context.Background()

	cfg := config.Load()

	bucket, err := s3bucket.New(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to init MinIO client: %v", err)
	}

	store, err := dagconfig.NewStore(ctx, bucket, cfg.DAGKey)
	if err != nil {
		log.Fatalf("failed to load DAG from MinIO: %v", err)
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
