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

	log.Println("Connected to database successfully")

	// Read and execute migration
	sqlContent, err := os.ReadFile("internal/migrations/000001_init.up.sql")
	if err != nil {
		log.Fatalf("Failed to read migration file: %v", err)
	}

	log.Println("Applying migrations...")

	_, err = conn.Exec(ctx, string(sqlContent))
	if err != nil {
		log.Printf("Migration error (may be already applied): %v", err)
	} else {
		log.Println("✅ Migrations applied successfully")
	}

	// Verify tables
	var count int
	err = conn.QueryRow(ctx, "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public' AND table_name IN ('tyres_api', 'rims_api', 'product_offers_api', 'sync_runs_api')").Scan(&count)
	if err != nil {
		log.Fatalf("Failed to verify tables: %v", err)
	}

	fmt.Printf("✅ Found %d tables in database\n", count)

	if count >= 4 {
		fmt.Println("✅ All tables created successfully!")
	} else {
		fmt.Printf("⚠️  Expected 4 tables, found %d\n", count)
	}
}
