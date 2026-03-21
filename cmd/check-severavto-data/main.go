package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

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

	fmt.Println("\n📊 Анализ данных Северавто\n")

	// 1. Номенклатура
	var nomenclatureCount int
	pool.QueryRow(ctx, "SELECT COUNT(*) FROM nomenclature_tyres_severavto").Scan(&nomenclatureCount)
	fmt.Printf("1️⃣  nomenclature_tyres_severavto: %d товаров\n", nomenclatureCount)

	// Структура nomenclature_tyres_severavto
	fmt.Println("\n   Колонки nomenclature_tyres_severavto:")
	rows, _ := pool.Query(ctx, `
		SELECT column_name, data_type
		FROM information_schema.columns
		WHERE table_name = 'nomenclature_tyres_severavto'
		ORDER BY ordinal_position
	`)
	for rows.Next() {
		var col, dtype string
		rows.Scan(&col, &dtype)
		fmt.Printf("      - %s (%s)\n", col, dtype)
	}
	rows.Close()

	// Пример записи
	fmt.Println("\n   Пример записи:")
	rows, _ = pool.Query(ctx, `
		SELECT commodity_id, brand, model, width, height, diameter,
		       load_index, speed_index, season, price, stock, warehouse_name
		FROM nomenclature_tyres_severavto
		LIMIT 1
	`)
	for rows.Next() {
		var commodityID, brand, model, diameter, loadIndex, speedIndex, season, warehouse string
		var width, height float64
		var price, stock int
		rows.Scan(&commodityID, &brand, &model, &width, &height, &diameter,
			&loadIndex, &speedIndex, &season, &price, &stock, &warehouse)
		fmt.Printf("      commodity_id: %s\n", commodityID)
		fmt.Printf("      brand: %s\n", brand)
		fmt.Printf("      model: %s\n", model)
		fmt.Printf("      размеры: %.0f/%.0f R%s\n", width, height, diameter)
		fmt.Printf("      индексы: %s %s\n", loadIndex, speedIndex)
		fmt.Printf("      сезон: %s\n", season)
		fmt.Printf("      цена: %d руб\n", price)
		fmt.Printf("      остаток: %d шт\n", stock)
		fmt.Printf("      склад: %s\n", warehouse)
	}
	rows.Close()

	// 2. Проверяем, есть ли отдельная таблица цен/остатков
	var pricesTableExists bool
	pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_name = 'severavto_prices_stock'
		)
	`).Scan(&pricesTableExists)

	if pricesTableExists {
		var pricesCount int
		pool.QueryRow(ctx, "SELECT COUNT(*) FROM severavto_prices_stock").Scan(&pricesCount)
		fmt.Printf("\n2️⃣  severavto_prices_stock: %d записей\n", pricesCount)
	} else {
		fmt.Println("\n2️⃣  severavto_prices_stock: таблица НЕ существует")
		fmt.Println("      (цены и остатки хранятся в nomenclature_tyres_severavto)")
	}

	// 3. Целевая таблица
	fmt.Println("\n3️⃣  Целевая таблица: tyres_prices_stock")
	fmt.Println("\n   Колонки tyres_prices_stock:")
	rows, _ = pool.Query(ctx, `
		SELECT column_name, data_type
		FROM information_schema.columns
		WHERE table_name = 'tyres_prices_stock'
		ORDER BY ordinal_position
	`)
	for rows.Next() {
		var col, dtype string
		rows.Scan(&col, &dtype)
		fmt.Printf("      - %s (%s)\n", col, dtype)
	}
	rows.Close()

	// Текущие данные в tyres_prices_stock
	var currentCount int
	pool.QueryRow(ctx, "SELECT COUNT(*) FROM tyres_prices_stock").Scan(&currentCount)
	fmt.Printf("\n   Текущих записей: %d\n", currentCount)

	// По провайдерам
	fmt.Println("\n   По провайдерам:")
	rows, _ = pool.Query(ctx, `
		SELECT provider, COUNT(*) as cnt
		FROM tyres_prices_stock
		GROUP BY provider
		ORDER BY cnt DESC
	`)
	for rows.Next() {
		var provider string
		var count int
		rows.Scan(&provider, &count)
		fmt.Printf("      - %s: %d записей\n", provider, count)
	}
	rows.Close()

	// 4. Маппинг
	var mappingCount int
	pool.QueryRow(ctx, "SELECT COUNT(*) FROM severavto_tyres_mapping").Scan(&mappingCount)
	fmt.Printf("\n4️⃣  severavto_tyres_mapping: %d связей\n", mappingCount)

	// 5. Потенциал для импорта
	fmt.Println("\n📈 Потенциал для импорта от Северавто:")

	// Сколько товаров с маппингом имеют цены/остатки в Северавто
	var withPrices int
	pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT sm.cae)
		FROM severavto_tyres_mapping sm
		INNER JOIN nomenclature_tyres_severavto nts ON sm.commodity_id = nts.commodity_id
		WHERE nts.price > 0 AND nts.stock > 0
	`).Scan(&withPrices)
	fmt.Printf("   Товары с маппингом И ценой/остатком: %d\n", withPrices)

	// Общее количество записей цен/остатков (может быть несколько складов)
	var totalPriceRecords int
	pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM severavto_tyres_mapping sm
		INNER JOIN nomenclature_tyres_severavto nts ON sm.commodity_id = nts.commodity_id
		WHERE nts.price > 0 AND nts.stock > 0
	`).Scan(&totalPriceRecords)
	fmt.Printf("   Всего записей цен/остатков: %d\n", totalPriceRecords)

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("\n💡 План импорта:\n")
	fmt.Println("   ИЗ: nomenclature_tyres_severavto")
	fmt.Println("      - commodity_id, price, stock, warehouse_name")
	fmt.Println()
	fmt.Println("   ЧЕРЕЗ: severavto_tyres_mapping")
	fmt.Println("      - commodity_id → cae")
	fmt.Println()
	fmt.Println("   В: tyres_prices_stock")
	fmt.Println("      - cae, price, stock, warehouse_name, provider='Северавто'")
	fmt.Println()
}
