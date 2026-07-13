package main

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"

	"controlplane/internal/config"
	"controlplane/internal/dag"
	"controlplane/internal/server"
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

	log.Println("Control plane running on :9000")
	if err := server.New(store).Run(":9000"); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
