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

	fmt.Println("\n🔍 Проверка РЕАЛЬНОЙ структуры Северавто\n")

	// Все таблицы со словом severavto
	fmt.Println("📋 Таблицы Северавто:")
	rows, _ := pool.Query(ctx, `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = 'public'
		  AND table_name LIKE '%severavto%'
		ORDER BY table_name
	`)
	tables := []string{}
	for rows.Next() {
		var table string
		rows.Scan(&table)
		tables = append(tables, table)
		fmt.Printf("   - %s\n", table)
	}
	rows.Close()

	// Для каждой таблицы - структура и примеры
	for _, table := range tables {
		fmt.Printf("\n" + strings.Repeat("=", 60) + "\n")
		fmt.Printf("📊 Таблица: %s\n", table)
		fmt.Println(strings.Repeat("=", 60))

		// Количество записей
		var count int
		pool.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
		fmt.Printf("\nЗаписей: %d\n", count)

		// Структура
		fmt.Println("\nКолонки:")
		rows, _ := pool.Query(ctx, fmt.Sprintf(`
			SELECT column_name, data_type, is_nullable
			FROM information_schema.columns
			WHERE table_name = '%s'
			ORDER BY ordinal_position
		`, table))
		for rows.Next() {
			var col, dtype, nullable string
			rows.Scan(&col, &dtype, &nullable)
			fmt.Printf("   %-25s %-20s %s\n", col, dtype, nullable)
		}
		rows.Close()

		// Первые 2 записи
		fmt.Println("\nПримеры данных:")
		rows, _ = pool.Query(ctx, fmt.Sprintf("SELECT * FROM %s LIMIT 2", table))

		// Получаем имена колонок
		fields := rows.FieldDescriptions()
		colNames := make([]string, len(fields))
		for i, field := range fields {
			colNames[i] = string(field.Name)
		}

		i := 0
		for rows.Next() {
			i++
			fmt.Printf("\n   Запись #%d:\n", i)

			values, _ := rows.Values()
			for j, val := range values {
				fmt.Printf("      %-20s: %v\n", colNames[j], val)
			}
		}
		rows.Close()
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("\n❓ ВОПРОС: Откуда брать price/stock/warehouse для Северавто?")
	fmt.Println("   1. Есть ли отдельная таблица цен/остатков?")
	fmt.Println("   2. Нужно ли загружать через API?")
	fmt.Println()
}
