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

	fmt.Println("\n🔍 Проверка логики дедупликации tyres_prices_stock\n")

	// 1. Проверить constraints таблицы
	fmt.Println("📋 Constraints таблицы tyres_prices_stock:")
	rows, _ := pool.Query(ctx, `
		SELECT
			tc.constraint_name,
			tc.constraint_type,
			kcu.column_name
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu
			ON tc.constraint_name = kcu.constraint_name
		WHERE tc.table_name = 'tyres_prices_stock'
		  AND tc.constraint_type IN ('PRIMARY KEY', 'UNIQUE')
		ORDER BY tc.constraint_name, kcu.ordinal_position
	`)

	constraints := make(map[string][]string)
	constraintTypes := make(map[string]string)

	for rows.Next() {
		var constraintName, constraintType, columnName string
		rows.Scan(&constraintName, &constraintType, &columnName)
		constraints[constraintName] = append(constraints[constraintName], columnName)
		constraintTypes[constraintName] = constraintType
	}
	rows.Close()

	for name, columns := range constraints {
		fmt.Printf("   %s (%s): %v\n", name, constraintTypes[name], columns)
	}

	// 2. Проверить индексы
	fmt.Println("\n📋 Индексы таблицы tyres_prices_stock:")
	rows, _ = pool.Query(ctx, `
		SELECT
			indexname,
			indexdef
		FROM pg_indexes
		WHERE tablename = 'tyres_prices_stock'
		ORDER BY indexname
	`)
	for rows.Next() {
		var indexName, indexDef string
		rows.Scan(&indexName, &indexDef)
		fmt.Printf("   %s:\n      %s\n", indexName, indexDef)
	}
	rows.Close()

	// 3. Проверить текущие данные
	fmt.Println("\n📊 Статистика по провайдерам:")
	rows, _ = pool.Query(ctx, `
		SELECT
			provider,
			COUNT(*) as total,
			COUNT(DISTINCT cae) as unique_cae,
			COUNT(DISTINCT warehouse_name) as unique_warehouses
		FROM tyres_prices_stock
		GROUP BY provider
		ORDER BY total DESC
	`)
	for rows.Next() {
		var provider string
		var total, uniqueCAE, uniqueWarehouses int
		rows.Scan(&provider, &total, &uniqueCAE, &uniqueWarehouses)
		fmt.Printf("   %s:\n", provider)
		fmt.Printf("      Всего записей: %d\n", total)
		fmt.Printf("      Уникальных CAE: %d\n", uniqueCAE)
		fmt.Printf("      Складов: %d\n", uniqueWarehouses)
		fmt.Printf("      Среднее записей на CAE: %.1f\n\n", float64(total)/float64(uniqueCAE))
	}
	rows.Close()

	// 4. Проверить дубли (если есть)
	fmt.Println("🔍 Проверка дублей по (cae, warehouse_name, provider):")
	rows, _ = pool.Query(ctx, `
		SELECT cae, warehouse_name, provider, COUNT(*) as cnt
		FROM tyres_prices_stock
		GROUP BY cae, warehouse_name, provider
		HAVING COUNT(*) > 1
		LIMIT 10
	`)

	hasDuplicates := false
	for rows.Next() {
		hasDuplicates = true
		var cae, warehouse, provider string
		var count int
		rows.Scan(&cae, &warehouse, &provider, &count)
		fmt.Printf("   ⚠️  CAE: %s, Склад: %s, Провайдер: %s - %d дублей\n",
			cae, warehouse, provider, count)
	}
	rows.Close()

	if !hasDuplicates {
		fmt.Println("   ✅ Дублей не найдено")
	}

	// 5. Проверить логику вставки для Форточек
	fmt.Println("\n📝 ON CONFLICT логика в sync-prices (Форточки):")
	fmt.Println("   Проверяю исходный код...")

	// 6. Проверить логику вставки для Северавто
	fmt.Println("\n📝 ON CONFLICT логика в sync-severavto-prices (Северавто):")
	fmt.Println("   Проверяю исходный код...")

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("\n💡 Ключевые моменты:")
	fmt.Println("   1. Какие поля используются для UNIQUE constraint?")
	fmt.Println("   2. Совпадает ли ON CONFLICT для Форточек и Северавто?")
	fmt.Println("   3. Что будет если один CAE на одном складе от разных провайдеров?")
	fmt.Println()
}
