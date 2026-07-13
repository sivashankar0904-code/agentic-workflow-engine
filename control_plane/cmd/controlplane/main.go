package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"controlplane/internal/config"
	"controlplane/internal/dag"
	"controlplane/internal/server"
	"controlplane/internal/users"
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

	if len(os.Args) > 1 && os.Args[1] == "reset-admin-password" {
		resetAdminPassword(ctx, pool, os.Args[2:])
		return
	}

	dagStore := dag.NewStore(pool)
	userStore := users.NewStore(pool)

	log.Println("Control plane running on :9000")
	if err := server.New(dagStore, userStore, cfg.JWTSecret).Run(":9000"); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

// resetAdminPassword is the break-glass path for when admin can't log in at
// all: run via `docker exec control-plane-app /app/controlplane
// reset-admin-password <newpass>`. Bypasses the SetActive/SetRole
// last-admin guard since it touches neither role nor active state.
func resetAdminPassword(ctx context.Context, pool *pgxpool.Pool, args []string) {
	var newPassword string
	if len(args) > 0 {
		newPassword = args[0]
	} else {
		fmt.Print("New password for admin: ")
		reader := bufio.NewReader(os.Stdin)
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("failed to read password: %v", err)
		}
		newPassword = strings.TrimSpace(line)
	}
	if newPassword == "" {
		log.Fatal("password must not be empty")
	}

	hash, err := users.Hash(newPassword)
	if err != nil {
		log.Fatalf("failed to hash password: %v", err)
	}

	store := users.NewStore(pool)
	if err := store.SetPassword(ctx, "admin", hash); err != nil {
		log.Fatalf("failed to reset admin password: %v", err)
	}
	fmt.Println("admin password reset")
}
