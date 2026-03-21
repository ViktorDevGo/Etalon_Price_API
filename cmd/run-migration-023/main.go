package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	dbDSN := os.Getenv("DATABASE_DSN")
	if dbDSN == "" {
		log.Fatal("DATABASE_DSN must be set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbDSN)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	fmt.Println("🔄 Running migration 023: create severavto_tyres_mapping table...")

	sql := `
-- Migration: Create Severavto mapping table for CAE to commodity_id mapping
CREATE TABLE IF NOT EXISTS severavto_tyres_mapping (
    id BIGSERIAL PRIMARY KEY,
    cae VARCHAR(50) NOT NULL UNIQUE,
    commodity_id VARCHAR(50) NOT NULL,
    match_priority INTEGER NOT NULL,
    match_details JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_severavto_mapping_cae ON severavto_tyres_mapping(cae);
CREATE INDEX IF NOT EXISTS idx_severavto_mapping_commodity_id ON severavto_tyres_mapping(commodity_id);
CREATE INDEX IF NOT EXISTS idx_severavto_mapping_priority ON severavto_tyres_mapping(match_priority);
`

	_, err = pool.Exec(ctx, sql)
	if err != nil {
		log.Fatalf("Failed to run migration: %v", err)
	}

	fmt.Println("✅ Migration completed successfully!")
	fmt.Println("   Table: severavto_tyres_mapping")
	fmt.Println("   Indexes: cae, commodity_id, match_priority")
}
