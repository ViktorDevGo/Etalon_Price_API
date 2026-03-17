package main

import (
	"context"
	"fmt"
	"log"
	"os"

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
	fmt.Println("Применение миграции 013: Ведение истории номенклатуры")
	fmt.Println("================================================================")
	fmt.Println()

	// Read migration file
	sql, err := os.ReadFile("migrations/013_modify_nomenclature_for_history.up.sql")
	if err != nil {
		log.Fatalf("Failed to read migration file: %v", err)
	}

	// Execute migration
	fmt.Println("⏳ Применение миграции...")
	_, err = pool.Exec(ctx, string(sql))
	if err != nil {
		log.Fatalf("❌ Ошибка применения миграции: %v", err)
	}

	fmt.Println("✅ Миграция 013 применена успешно!")
	fmt.Println()
	fmt.Println("Изменения:")
	fmt.Println("  ✅ UNIQUE constraints на cae удалены")
	fmt.Println("  ✅ Добавлены индексы для проверки дублей:")
	fmt.Println("     - idx_nomenclature_tyres_duplicate_check")
	fmt.Println("     - idx_nomenclature_rims_duplicate_check")
	fmt.Println("  ✅ Добавлены индексы для поиска по cae:")
	fmt.Println("     - idx_nomenclature_tyres_cae_lookup")
	fmt.Println("     - idx_nomenclature_rims_cae_lookup")
	fmt.Println()
	fmt.Println("Таблицы готовы к ведению истории изменений!")
	fmt.Println("================================================================")
}
