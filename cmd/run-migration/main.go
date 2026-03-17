package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"github.com/prokoleso/etalon-price-api/internal/config"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run cmd/run-migration/main.go <migration_file>")
	}

	migrationFile := os.Args[1]

	// Load .env
	_ = godotenv.Load()

	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, cfg.Database.DSN)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	fmt.Printf("Running migration: %s\n", migrationFile)

	// Read and execute migration
	sqlContent, err := os.ReadFile(migrationFile)
	if err != nil {
		log.Fatalf("Failed to read migration file: %v", err)
	}

	_, err = conn.Exec(ctx, string(sqlContent))
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	fmt.Println("✅ Migration completed successfully")
}
