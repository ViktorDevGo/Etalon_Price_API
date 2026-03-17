package main

import (
	"context"
	"fmt"
	"log"

	"github.com/prokoleso/etalon-price-api/internal/config"
	"github.com/prokoleso/etalon-price-api/internal/repository/postgres"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx := context.Background()
	db, err := postgres.New(ctx, postgres.Config{DSN: cfg.Database.DSN})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	pool := db.Pool()

	fmt.Println("=================================================================")
	fmt.Println("Данные в таблице tyres_api")
	fmt.Println("=================================================================")
	fmt.Println()

	query := `
		SELECT
			code,
			supplier,
			brand,
			model,
			width,
			height,
			diameter,
			load_index,
			speed_index,
			season,
			description
		FROM tyres_api
		WHERE supplier = '4tochki'
		ORDER BY code
	`

	rows, err := pool.Query(ctx, query)
	if err != nil {
		log.Fatalf("Failed to query: %v", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var code, supplier, brand, model, loadIndex, speedIndex, season, description string
		var width, height, diameter int

		err := rows.Scan(&code, &supplier, &brand, &model, &width, &height, &diameter, &loadIndex, &speedIndex, &season, &description)
		if err != nil {
			log.Printf("Failed to scan row: %v", err)
			continue
		}

		count++
		fmt.Printf("Товар #%d:\n", count)
		fmt.Printf("  Код:         %s\n", code)
		fmt.Printf("  Поставщик:   %s\n", supplier)
		fmt.Printf("  Бренд:       %s\n", brand)
		fmt.Printf("  Модель:      %s\n", model)
		fmt.Printf("  Размер:      %d/%d R%d\n", width, height, diameter)
		fmt.Printf("  Индексы:     %s/%s\n", loadIndex, speedIndex)
		fmt.Printf("  Сезон:       %s\n", season)
		fmt.Printf("  Описание:    %s\n", description)
		fmt.Println()
	}

	if err = rows.Err(); err != nil {
		log.Fatalf("Error iterating rows: %v", err)
	}

	fmt.Println("=================================================================")
	fmt.Printf("Всего найдено записей: %d\n", count)
	fmt.Println("=================================================================")
}
