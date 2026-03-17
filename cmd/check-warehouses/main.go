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
	fmt.Println("Проверка складов")
	fmt.Println("================================================================")
	fmt.Println()

	// Count warehouses
	var warehousesCount int
	pool.QueryRow(ctx, "SELECT COUNT(*) FROM warehouses").Scan(&warehousesCount)
	fmt.Printf("Всего складов в БД: %d\n\n", warehousesCount)

	// Check missing warehouse_id in tyres_prices_stock
	rows, err := pool.Query(ctx, `
		SELECT DISTINCT warehouse_id
		FROM tyres_prices_stock
		WHERE warehouse_id NOT IN (SELECT id FROM warehouses)
		ORDER BY warehouse_id
	`)
	if err == nil {
		defer rows.Close()
		missingIDs := []int{}
		for rows.Next() {
			var id int
			rows.Scan(&id)
			missingIDs = append(missingIDs, id)
		}
		if len(missingIDs) > 0 {
			fmt.Printf("⚠️  warehouse_id отсутствующие в warehouses (tyres): %v\n", missingIDs)
		} else {
			fmt.Println("✅ Все warehouse_id из tyres_prices_stock есть в warehouses")
		}
	}

	// Check missing warehouse_id in rims_prices_stock
	rows, err = pool.Query(ctx, `
		SELECT DISTINCT warehouse_id
		FROM rims_prices_stock
		WHERE warehouse_id NOT IN (SELECT id FROM warehouses)
		ORDER BY warehouse_id
	`)
	if err == nil {
		defer rows.Close()
		missingIDs := []int{}
		for rows.Next() {
			var id int
			rows.Scan(&id)
			missingIDs = append(missingIDs, id)
		}
		if len(missingIDs) > 0 {
			fmt.Printf("⚠️  warehouse_id отсутствующие в warehouses (rims): %v\n", missingIDs)
		} else {
			fmt.Println("✅ Все warehouse_id из rims_prices_stock есть в warehouses")
		}
	}

	fmt.Println()

	// Show warehouse stats
	fmt.Println("Топ-10 складов по количеству товаров (шины):")
	fmt.Println("-----------------------------------------------------------------")
	rows, err = pool.Query(ctx, `
		SELECT w.id, w.short_name, w.name, COUNT(*) AS products
		FROM tyres_prices_stock p
		JOIN warehouses w ON p.warehouse_id = w.id
		WHERE p.stock > 0
		GROUP BY w.id, w.short_name, w.name
		ORDER BY products DESC
		LIMIT 10
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var id, products int
			var shortName, name string
			rows.Scan(&id, &shortName, &name, &products)
			fmt.Printf("  [%d] %s - %s: %d товаров\n", id, shortName, name, products)
		}
	}

	fmt.Println()
	fmt.Println("================================================================")
}
