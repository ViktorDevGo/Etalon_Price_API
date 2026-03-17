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
	fmt.Println("Проверка структуры таблиц цен и остатков")
	fmt.Println("================================================================")
	fmt.Println()

	// Check tyres_prices_stock structure
	fmt.Println("Таблица tyres_prices_stock:")
	fmt.Println("-----------------------------------------------------------------")
	rows, err := pool.Query(ctx, `
		SELECT column_name, data_type, is_nullable
		FROM information_schema.columns
		WHERE table_name = 'tyres_prices_stock'
		ORDER BY ordinal_position
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var colName, dataType, nullable string
			rows.Scan(&colName, &dataType, &nullable)
			fmt.Printf("  %-20s %-20s NULL: %s\n", colName, dataType, nullable)
		}
	}
	fmt.Println()

	// Sample data from tyres_prices_stock
	fmt.Println("Примеры данных (tyres_prices_stock):")
	fmt.Println("-----------------------------------------------------------------")
	rows, err = pool.Query(ctx, `
		SELECT cae, warehouse_name, price, stock, isimport, provider
		FROM tyres_prices_stock
		LIMIT 5
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var cae, warehouseName, provider string
			var price, stock, isimport int
			rows.Scan(&cae, &warehouseName, &price, &stock, &isimport, &provider)
			fmt.Printf("  CAE: %s, Склад: %s, Цена: %.2f руб, Остаток: %d, IsImport: %d, Поставщик: %s\n",
				cae, warehouseName, float64(price)/100, stock, isimport, provider)
		}
	}
	fmt.Println()

	// Check indexes
	fmt.Println("Индексы на tyres_prices_stock:")
	fmt.Println("-----------------------------------------------------------------")
	rows, err = pool.Query(ctx, `
		SELECT indexname, indexdef
		FROM pg_indexes
		WHERE tablename = 'tyres_prices_stock'
		ORDER BY indexname
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var indexName, indexDef string
			rows.Scan(&indexName, &indexDef)
			fmt.Printf("  %s\n", indexName)
		}
	}
	fmt.Println()

	// Statistics
	var totalRecords, uniqueCAE, uniqueWarehouses, uniqueProviders int
	pool.QueryRow(ctx, "SELECT COUNT(*) FROM tyres_prices_stock").Scan(&totalRecords)
	pool.QueryRow(ctx, "SELECT COUNT(DISTINCT cae) FROM tyres_prices_stock").Scan(&uniqueCAE)
	pool.QueryRow(ctx, "SELECT COUNT(DISTINCT warehouse_name) FROM tyres_prices_stock").Scan(&uniqueWarehouses)
	pool.QueryRow(ctx, "SELECT COUNT(DISTINCT provider) FROM tyres_prices_stock").Scan(&uniqueProviders)

	fmt.Println("Статистика:")
	fmt.Println("-----------------------------------------------------------------")
	fmt.Printf("  Всего записей: %d\n", totalRecords)
	fmt.Printf("  Уникальных товаров: %d\n", uniqueCAE)
	fmt.Printf("  Уникальных складов: %d\n", uniqueWarehouses)
	fmt.Printf("  Поставщиков: %d\n", uniqueProviders)

	fmt.Println()

	// Check rims_prices_stock structure
	fmt.Println("================================================================")
	fmt.Println("Таблица rims_prices_stock:")
	fmt.Println("-----------------------------------------------------------------")
	rows, err = pool.Query(ctx, `
		SELECT column_name, data_type, is_nullable
		FROM information_schema.columns
		WHERE table_name = 'rims_prices_stock'
		ORDER BY ordinal_position
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var colName, dataType, nullable string
			rows.Scan(&colName, &dataType, &nullable)
			fmt.Printf("  %-20s %-20s NULL: %s\n", colName, dataType, nullable)
		}
	}
	fmt.Println()

	// Sample data from rims_prices_stock
	fmt.Println("Примеры данных (rims_prices_stock):")
	fmt.Println("-----------------------------------------------------------------")
	rows, err = pool.Query(ctx, `
		SELECT cae, warehouse_name, price, stock, isimport, provider
		FROM rims_prices_stock
		LIMIT 5
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var cae, warehouseName, provider string
			var price, stock, isimport int
			rows.Scan(&cae, &warehouseName, &price, &stock, &isimport, &provider)
			fmt.Printf("  CAE: %s, Склад: %s, Цена: %.2f руб, Остаток: %d, IsImport: %d, Поставщик: %s\n",
				cae, warehouseName, float64(price)/100, stock, isimport, provider)
		}
	}
	fmt.Println()

	// Check indexes on rims
	fmt.Println("Индексы на rims_prices_stock:")
	fmt.Println("-----------------------------------------------------------------")
	rows, err = pool.Query(ctx, `
		SELECT indexname
		FROM pg_indexes
		WHERE tablename = 'rims_prices_stock'
		ORDER BY indexname
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var indexName string
			rows.Scan(&indexName)
			fmt.Printf("  %s\n", indexName)
		}
	}
	fmt.Println()

	// Statistics for rims
	var totalRimsRecords, uniqueRimsCAE, uniqueRimsWarehouses, uniqueRimsProviders int
	pool.QueryRow(ctx, "SELECT COUNT(*) FROM rims_prices_stock").Scan(&totalRimsRecords)
	pool.QueryRow(ctx, "SELECT COUNT(DISTINCT cae) FROM rims_prices_stock").Scan(&uniqueRimsCAE)
	pool.QueryRow(ctx, "SELECT COUNT(DISTINCT warehouse_name) FROM rims_prices_stock").Scan(&uniqueRimsWarehouses)
	pool.QueryRow(ctx, "SELECT COUNT(DISTINCT provider) FROM rims_prices_stock").Scan(&uniqueRimsProviders)

	fmt.Println("Статистика (диски):")
	fmt.Println("-----------------------------------------------------------------")
	fmt.Printf("  Всего записей: %d\n", totalRimsRecords)
	fmt.Printf("  Уникальных товаров: %d\n", uniqueRimsCAE)
	fmt.Printf("  Уникальных складов: %d\n", uniqueRimsWarehouses)
	fmt.Printf("  Поставщиков: %d\n", uniqueRimsProviders)

	fmt.Println()
	fmt.Println("================================================================")
}
