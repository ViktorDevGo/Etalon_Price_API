package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	pool, err := pgxpool.New(context.Background(), os.Getenv("DB_DSN"))
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	ctx := context.Background()

	targetCAEs := []string{"130553", "2449100", "3151004", "3151013", "3251009", "3151009", "3251011"}

	query := "SELECT t.cae, COALESCE(MIN(p.price), 0) as price_kopeks, COALESCE(SUM(p.stock), 0) as total_stock FROM nomenclature_tyres t INNER JOIN tyres_prices_stock p ON t.cae = p.cae WHERE t.cae = ANY($1) GROUP BY t.cae ORDER BY t.cae"

	rows, err := pool.Query(ctx, query, targetCAEs)
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	defer rows.Close()

	file, err := os.Create("/Users/viktor/Desktop/bitrix_filtered.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{"IE_XML_ID", "CATALOG_QUANTITY", "QUANTITY", "CATALOG_STORE_AMOUNT_1", "CV_PRICE_1", "CV_CURRENCY_1"}

	if err := writer.Write(header); err != nil {
		log.Fatal(err)
	}

	count := 0

	for rows.Next() {
		var (
			cae              string
			priceKopeks, stock int
		)

		if err := rows.Scan(&cae, &priceKopeks, &stock); err != nil {
			log.Fatal(err)
		}

		if stock == 0 || priceKopeks == 0 {
			continue
		}

		priceRub := float64(priceKopeks) / 100.0

		record := []string{
			cae,
			fmt.Sprintf("%d", stock),
			fmt.Sprintf("%d", stock),
			fmt.Sprintf("%d", stock),
			fmt.Sprintf("%.2f", priceRub),
			"RUB",
		}

		if err := writer.Write(record); err != nil {
			log.Fatal(err)
		}

		count++
		fmt.Printf("✓ %s: %d шт, %.2f руб\n", cae, stock, priceRub)
	}

	fmt.Println()
	fmt.Printf("✅ Экспортировано: %d позиций\n", count)
	fmt.Println("📁 Файл: /Users/viktor/Desktop/bitrix_filtered.csv")
}
