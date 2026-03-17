package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	output := flag.String("output", "/Users/viktor/Desktop/bitrix_simple.csv", "Путь к выходному CSV")
	limit := flag.Int("limit", 0, "Количество позиций (0 = все)")
	flag.Parse()

	_ = godotenv.Load()

	pool, err := pgxpool.New(context.Background(), os.Getenv("DB_DSN"))
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	ctx := context.Background()

	fmt.Println("================================================================")
	fmt.Println("Экспорт для 1C-Битрикс (упрощенный формат)")
	fmt.Println("================================================================")
	fmt.Println()

	query := buildQuery(*limit)

	rows, err := pool.Query(ctx, query)
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	defer rows.Close()

	file, err := os.Create(*output)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{
		"IE_XML_ID",
		"CATALOG_QUANTITY",
		"QUANTITY",
		"CATALOG_STORE_AMOUNT_1",
		"CV_PRICE_1",
		"CV_CURRENCY_1",
	}

	if err := writer.Write(header); err != nil {
		log.Fatal(err)
	}

	fmt.Println("✅ Формат:")
	fmt.Println("  6 колонок")
	fmt.Println("  Фильтр: stock > 0 и price > 0")
	fmt.Println()
	fmt.Println("Экспорт данных...")

	count := 0
	skipped := 0

	for rows.Next() {
		var (
			cae              string
			priceKopeks, stock int
		)

		if err := rows.Scan(&cae, &priceKopeks, &stock); err != nil {
			log.Fatal(err)
		}

		if stock == 0 || priceKopeks == 0 {
			skipped++
			continue
		}

		priceRub := float64(priceKopeks) / 100.0
		stockStr := strconv.Itoa(stock)

		record := []string{
			cae,
			stockStr,
			stockStr,
			stockStr,
			fmt.Sprintf("%.2f", priceRub),
			"RUB",
		}

		if err := writer.Write(record); err != nil {
			log.Fatal(err)
		}

		count++
	}

	fmt.Println()
	fmt.Println("================================================================")
	fmt.Printf("✅ Экспортировано: %d позиций\n", count)
	fmt.Printf("⚠️  Пропущено: %d позиций\n", skipped)
	fmt.Printf("📁 Файл: %s\n", *output)
	fmt.Println()
	fmt.Println("📋 Пример:")
	fmt.Println("  130553,20,20,20,850.00,RUB")
	fmt.Println("================================================================")
}

func buildQuery(limit int) string {
	query := "SELECT t.cae, COALESCE(MIN(p.price), 0) as price_kopeks, COALESCE(SUM(p.stock), 0) as total_stock FROM nomenclature_tyres t INNER JOIN tyres_prices_stock p ON t.cae = p.cae GROUP BY t.cae ORDER BY t.cae"
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}
	return query
}
