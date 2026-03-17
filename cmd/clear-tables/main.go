package main

import (
	"context"
	"fmt"
	"log"

	"github.com/prokoleso/etalon-price-api/internal/config"
	"github.com/prokoleso/etalon-price-api/internal/repository/postgres"
)

func main() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	ctx := context.Background()
	db, err := postgres.New(ctx, postgres.Config{
		DSN: cfg.Database.DSN,
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	pool := db.Pool()

	fmt.Println("=================================================================")
	fmt.Println("ОЧИСТКА ТАБЛИЦ")
	fmt.Println("=================================================================")
	fmt.Println()

	tables := []string{
		"etalon_nomenclature",
		"price_disks",
		"price_tires",
		"processed_emails",
	}

	// Show counts before deletion
	fmt.Println("Количество записей ДО очистки:")
	fmt.Println("-----------------------------------------------------------------")
	for _, table := range tables {
		var count int
		err := pool.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
		if err != nil {
			log.Printf("⚠️  Не удалось подсчитать записи в %s: %v", table, err)
			continue
		}
		fmt.Printf("  %-25s: %d записей\n", table, count)
	}
	fmt.Println()

	// Clear tables
	fmt.Println("Очистка таблиц...")
	fmt.Println("-----------------------------------------------------------------")

	for _, table := range tables {
		fmt.Printf("Очищаю %s... ", table)

		_, err := pool.Exec(ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			log.Printf("❌ ОШИБКА: %v\n", err)
			continue
		}

		fmt.Println("✅ Успешно")
	}

	fmt.Println()

	// Show counts after deletion
	fmt.Println("Количество записей ПОСЛЕ очистки:")
	fmt.Println("-----------------------------------------------------------------")
	for _, table := range tables {
		var count int
		err := pool.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
		if err != nil {
			log.Printf("⚠️  Не удалось подсчитать записи в %s: %v", table, err)
			continue
		}
		fmt.Printf("  %-25s: %d записей\n", table, count)
	}

	fmt.Println()
	fmt.Println("=================================================================")
	fmt.Println("ОЧИСТКА ЗАВЕРШЕНА")
	fmt.Println("=================================================================")
}
