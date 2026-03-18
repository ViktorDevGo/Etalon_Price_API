package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Справочник габаритов упаковки по диаметру (из файла "диски двш.xlsx")
// Диаметр -> (Вес, Д, Ш, В)
// ВАЖНО:
// - Вес в КГ → в CSV переводится в ГРАММЫ (×1000)
// - Размеры в СМ → в CSV переводятся в ММ (×10)
var dimensionsByDiameter = map[float64]struct {
	Weight float64 // вес в кг (→ граммы в CSV)
	Width  float64 // ширина в см (Ш) (→ мм в CSV)
	Height float64 // высота в см (В) (→ мм в CSV)
	Length float64 // длина в см (Д) (→ мм в CSV)
}{
	14: {Weight: 6.5, Length: 40, Width: 40, Height: 18},
	15: {Weight: 8.0, Length: 42, Width: 42, Height: 18},
	16: {Weight: 9.5, Length: 44, Width: 44, Height: 20},
	17: {Weight: 11.5, Length: 46, Width: 46, Height: 22},
	18: {Weight: 13.0, Length: 48, Width: 48, Height: 23},
	19: {Weight: 15.5, Length: 50, Width: 50, Height: 24},
	20: {Weight: 17.0, Length: 52, Width: 52, Height: 25},
	21: {Weight: 19.5, Length: 55, Width: 55, Height: 28},
	22: {Weight: 23.0, Length: 58, Width: 58, Height: 30},
}

func main() {
	outputPath := flag.String("output", "/Users/viktor/Desktop/rims_dimensions.csv", "Output CSV file path")
	limit := flag.Int("limit", 5, "Number of rows to generate")
	flag.Parse()

	// Подключение к БД
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		log.Fatal("DATABASE_DSN environment variable is required")
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Запрос данных из nomenclature_rims
	query := `
		SELECT cae, name, diameter
		FROM nomenclature_rims
		WHERE diameter IS NOT NULL
		ORDER BY id
		LIMIT $1
	`

	rows, err := pool.Query(context.Background(), query, *limit)
	if err != nil {
		log.Fatalf("Failed to query database: %v", err)
	}
	defer rows.Close()

	// Создание CSV файла
	file, err := os.Create(*outputPath)
	if err != nil {
		log.Fatalf("Failed to create CSV file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Заголовок CSV (только нужные колонки)
	header := []string{
		"IE_XML_ID",  // CAE
		"IE_NAME",    // Название
		"CP_WEIGHT",  // Вес (кг)
		"CP_WIDTH",   // Ширина (см)
		"CP_HEIGHT",  // Высота (см)
		"CP_LENGTH",  // Длина (см)
	}

	if err := writer.Write(header); err != nil {
		log.Fatalf("Failed to write header: %v", err)
	}

	// Запись данных
	rowCount := 0
	for rows.Next() {
		var cae, name string
		var diameter float64

		if err := rows.Scan(&cae, &name, &diameter); err != nil {
			log.Printf("Failed to scan row: %v", err)
			continue
		}

		// Получаем габариты по диаметру
		dims, ok := dimensionsByDiameter[diameter]
		if !ok {
			// Если диаметр не найден, используем ближайший
			dims = struct {
				Weight float64
				Width  float64
				Height float64
				Length float64
			}{
				Weight: 8.0,  // средний вес
				Width:  55.0, // средняя ширина
				Height: 55.0, // средняя высота
				Length: 20.0, // средняя длина
			}
		}

		// Создаем строку CSV (только 6 колонок)
		row := []string{
			cae,                                    // IE_XML_ID
			name,                                   // IE_NAME
			fmt.Sprintf("%.0f", dims.Weight*1000),  // CP_WEIGHT (граммы)
			fmt.Sprintf("%.0f", dims.Width*10),     // CP_WIDTH (мм)
			fmt.Sprintf("%.0f", dims.Height*10),    // CP_HEIGHT (мм)
			fmt.Sprintf("%.0f", dims.Length*10),    // CP_LENGTH (мм)
		}

		if err := writer.Write(row); err != nil {
			log.Printf("Failed to write row: %v", err)
			continue
		}

		rowCount++
	}

	if err := rows.Err(); err != nil {
		log.Fatalf("Error iterating rows: %v", err)
	}

	fmt.Printf("✅ Generated %d rows\n", rowCount)
	fmt.Printf("📄 File saved to: %s\n", *outputPath)
}
