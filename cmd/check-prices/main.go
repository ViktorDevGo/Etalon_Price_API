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

	fmt.Println("================================================================")
	fmt.Println("Проверка цен и остатков")
	fmt.Println("================================================================")
	fmt.Println()

	// Tyres stats
	var tyresProducts, tyresWarehouses, tyresInStock int
	pool.QueryRow(ctx, "SELECT COUNT(DISTINCT cae) FROM tyres_prices_stock").Scan(&tyresProducts)
	pool.QueryRow(ctx, "SELECT COUNT(*) FROM tyres_prices_stock").Scan(&tyresWarehouses)
	pool.QueryRow(ctx, "SELECT COUNT(DISTINCT cae) FROM tyres_prices_stock WHERE stock > 0").Scan(&tyresInStock)

	fmt.Println("Шины:")
	fmt.Printf("  - Всего товаров: %d\n", tyresProducts)
	fmt.Printf("  - Складских позиций: %d\n", tyresWarehouses)
	fmt.Printf("  - В наличии (stock > 0): %d\n", tyresInStock)
	fmt.Println()

	// Rims stats
	var rimsProducts, rimsWarehouses, rimsInStock int
	pool.QueryRow(ctx, "SELECT COUNT(DISTINCT cae) FROM rims_prices_stock").Scan(&rimsProducts)
	pool.QueryRow(ctx, "SELECT COUNT(*) FROM rims_prices_stock").Scan(&rimsWarehouses)
	pool.QueryRow(ctx, "SELECT COUNT(DISTINCT cae) FROM rims_prices_stock WHERE stock > 0").Scan(&rimsInStock)

	fmt.Println("Диски:")
	fmt.Printf("  - Всего товаров: %d\n", rimsProducts)
	fmt.Printf("  - Складских позиций: %d\n", rimsWarehouses)
	fmt.Printf("  - В наличии (stock > 0): %d\n", rimsInStock)
	fmt.Println()

	// Warehouse stats
	fmt.Println("Склады (шины):")
	rows, err := pool.Query(ctx, `
		SELECT warehouse_id, COUNT(*) AS products, SUM(stock) AS total_stock
		FROM tyres_prices_stock
		WHERE stock > 0
		GROUP BY warehouse_id
		ORDER BY products DESC
		LIMIT 10
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var whID, products, totalStock int
			rows.Scan(&whID, &products, &totalStock)
			fmt.Printf("  - Склад %d: %d товаров, %d шт в наличии\n", whID, products, totalStock)
		}
	}
	fmt.Println()

	// Latest sync
	var syncType, status string
	var totalItems, totalWarehouses int
	err = pool.QueryRow(ctx, `
		SELECT sync_type, status, total_items, total_warehouses
		FROM prices_stock_sync_runs
		ORDER BY started_at DESC
		LIMIT 1
	`).Scan(&syncType, &status, &totalItems, &totalWarehouses)

	if err == nil {
		fmt.Println("Последняя синхронизация:")
		fmt.Printf("  - Тип: %s\n", syncType)
		fmt.Printf("  - Статус: %s\n", status)
		fmt.Printf("  - Товаров: %d\n", totalItems)
		fmt.Printf("  - Складских позиций: %d\n", totalWarehouses)
	}

	fmt.Println()
	fmt.Println("================================================================")
}
