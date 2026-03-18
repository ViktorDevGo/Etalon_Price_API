package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
)

func main() {
	migrationFile := flag.String("file", "", "Path to migration SQL file")
	flag.Parse()

	if *migrationFile == "" {
		log.Fatal("Usage: go run main.go -file=path/to/migration.sql")
	}

	// Read SQL file
	sqlBytes, err := os.ReadFile(*migrationFile)
	if err != nil {
		log.Fatalf("Failed to read migration file: %v", err)
	}

	// Connect to database
	dsn := "postgresql://gen_user:Poison-79@c37e696087932476c61fd621.twc1.net:5432/default_db?sslmode=require"
	conn, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer conn.Close(context.Background())

	// Execute migration
	_, err = conn.Exec(context.Background(), string(sqlBytes))
	if err != nil {
		log.Fatalf("Failed to execute migration: %v", err)
	}

	fmt.Printf("✅ Migration applied successfully: %s\n", *migrationFile)
}
