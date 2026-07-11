package main

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"

	"orchestrator/internal/config"
	"orchestrator/internal/dag"
	"orchestrator/internal/server"
)

func main() {
	ctx := context.Background()

	cfg := config.Load()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to create pg pool: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}

	store := dag.NewStore(pool)

	log.Println("Orchestrator running on :8000")
	if err := server.New(store).Run(":8000"); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
