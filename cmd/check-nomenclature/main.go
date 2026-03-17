package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/prokoleso/etalon-price-api/internal/config"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	pool, err := pgxpool.New(context.Background(), cfg.Database.DSN)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	ctx := context.Background()

	// Check nomenclature tables
	tables := []string{"nomenclature_tyres", "nomenclature_rims", "nomenclature_sync_runs"}

	fmt.Println("Проверка таблиц номенклатуры:")
	fmt.Println("================================================================")

	for _, table := range tables {
		var count int
		err := pool.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
		if err != nil {
			fmt.Printf("❌ Таблица %s: ОШИБКА - %v\n", table, err)
		} else {
			fmt.Printf("✅ Таблица %s: %d записей\n", table, count)
		}
	}

	fmt.Println("================================================================")
}
