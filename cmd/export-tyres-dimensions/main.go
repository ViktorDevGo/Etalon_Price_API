package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	limitFlag := flag.Int("limit", 0, "Limit number of rows (0 = all)")
	outputFlag := flag.String("output", "/Users/viktor/Desktop/tyres_all_dimensions.csv", "Output file path")
	flag.Parse()

	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		log.Fatal("DATABASE_DSN environment variable not set")
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Открываем CSV файл для записи
	file, err := os.Create(*outputFlag)
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer file.Close()

	w := csv.NewWriter(file)
	defer w.Flush()

	// Заголовок
	w.Write([]string{"IE_XML_ID", "IE_NAME", "CP_WEIGHT"})

	// SQL запрос
	query := "SELECT cae, name FROM nomenclature_tyres ORDER BY id"
	if *limitFlag > 0 {
		query = fmt.Sprintf("%s LIMIT %d", query, *limitFlag)
	}

	rows, err := pool.Query(context.Background(), query)
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	defer rows.Close()

	count := 0
	notFound := 0

	for rows.Next() {
		var cae, name string
		if err := rows.Scan(&cae, &name); err != nil {
			log.Printf("Failed to scan row: %v", err)
			continue
		}

		// Извлекаем первое слово из name (до первого пробела) - это размер шины
		firstWord := strings.Split(name, " ")[0]

		// Получаем вес по полному размеру из таблицы соответствия
		weight, ok := weightBySize[firstWord]
		if !ok {
			// Если размер не найден в таблице, используем средний вес (13 кг)
			weight = 13.0
			notFound++
		}

		// Конвертируем вес в граммы
		weightGrams := int(weight * 1000)

		// Записываем строку в CSV
		w.Write([]string{
			cae,
			name,
			fmt.Sprintf("%d", weightGrams),
		})

		count++

		if count%10000 == 0 {
			fmt.Printf("Processed %d rows...\n", count)
		}
	}

	if err := rows.Err(); err != nil {
		log.Fatalf("Row iteration error: %v", err)
	}

	fmt.Printf("\n✅ Export completed!\n")
	fmt.Printf("   Total rows: %d\n", count)
	fmt.Printf("   Found in weight table: %d\n", count-notFound)
	fmt.Printf("   Not found (default 13kg): %d\n", notFound)
	fmt.Printf("   Output file: %s\n", *outputFlag)
}
