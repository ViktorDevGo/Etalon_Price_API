package main

import (
	"context"
	"fmt"
	"log"
	"os"

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

	fmt.Println("\n📊 Аналитика пересечения артикулов провайдеров с nomenclature_tyres\n")

	providers := []string{"ГРУППА БРИНЕКС", "ЗАПАСКА", "БИГМАШИН", "Северавто", "Форточки"}

	// 1. Общая статистика по каждому провайдеру
	fmt.Println("📋 Общая статистика по провайдерам:\n")

	for _, provider := range providers {
		fmt.Printf("🔹 %s:\n", provider)

		// Всего уникальных CAE в tyres_prices_stock
		var totalCAE int
		pool.QueryRow(ctx, `
			SELECT COUNT(DISTINCT cae)
			FROM tyres_prices_stock
			WHERE provider = $1
		`, provider).Scan(&totalCAE)

		// Сколько из них есть в nomenclature_tyres
		var inNomenclature int
		pool.QueryRow(ctx, `
			SELECT COUNT(DISTINCT tps.cae)
			FROM tyres_prices_stock tps
			INNER JOIN nomenclature_tyres nt ON tps.cae = nt.cae
			WHERE tps.provider = $1
		`, provider).Scan(&inNomenclature)

		// Сколько НЕТ в nomenclature_tyres (orphan)
		orphan := totalCAE - inNomenclature

		// Процент покрытия
		coverage := 0.0
		if totalCAE > 0 {
			coverage = float64(inNomenclature) / float64(totalCAE) * 100
		}

		fmt.Printf("   Всего уникальных CAE:        %d\n", totalCAE)
		fmt.Printf("   Есть в nomenclature_tyres:   %d (%.1f%%)\n", inNomenclature, coverage)
		fmt.Printf("   НЕТ в nomenclature_tyres:    %d (%.1f%%)\n\n", orphan, 100-coverage)
	}

	// 2. Сравнение с nomenclature_tyres
	fmt.Println("📊 Покрытие nomenclature_tyres провайдерами:\n")

	var totalNomenclature int
	pool.QueryRow(ctx, "SELECT COUNT(*) FROM nomenclature_tyres WHERE tiretype = 'Легковая'").Scan(&totalNomenclature)
	fmt.Printf("   Всего в nomenclature_tyres (Легковая): %d\n\n", totalNomenclature)

	for _, provider := range providers {
		var coveredByProvider int
		pool.QueryRow(ctx, `
			SELECT COUNT(DISTINCT nt.cae)
			FROM nomenclature_tyres nt
			INNER JOIN tyres_prices_stock tps ON nt.cae = tps.cae
			WHERE tps.provider = $1
			  AND nt.tiretype = 'Легковая'
		`, provider).Scan(&coveredByProvider)

		coverage := float64(coveredByProvider) / float64(totalNomenclature) * 100
		fmt.Printf("   %s: %d CAE (%.1f%% от номенклатуры)\n",
			provider, coveredByProvider, coverage)
	}

	// 3. Перекрытие между провайдерами
	fmt.Println("\n🔄 Пересечение артикулов между провайдерами:\n")

	// Группа Бринекс ∩ Запаска
	var brinexZapacka int
	pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT b.cae)
		FROM tyres_prices_stock b
		INNER JOIN tyres_prices_stock z ON b.cae = z.cae
		WHERE b.provider = 'ГРУППА БРИНЕКС'
		  AND z.provider = 'ЗАПАСКА'
	`).Scan(&brinexZapacka)

	// Группа Бринекс ∩ Бигмашин
	var brinexBigmashin int
	pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT b.cae)
		FROM tyres_prices_stock b
		INNER JOIN tyres_prices_stock big ON b.cae = big.cae
		WHERE b.provider = 'ГРУППА БРИНЕКС'
		  AND big.provider = 'БИГМАШИН'
	`).Scan(&brinexBigmashin)

	// Запаска ∩ Бигмашин
	var zapackaBigmashin int
	pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT z.cae)
		FROM tyres_prices_stock z
		INNER JOIN tyres_prices_stock big ON z.cae = big.cae
		WHERE z.provider = 'ЗАПАСКА'
		  AND big.provider = 'БИГМАШИН'
	`).Scan(&zapackaBigmashin)

	// Все три вместе
	var allThree int
	pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT b.cae)
		FROM tyres_prices_stock b
		INNER JOIN tyres_prices_stock z ON b.cae = z.cae
		INNER JOIN tyres_prices_stock big ON b.cae = big.cae
		WHERE b.provider = 'ГРУППА БРИНЕКС'
		  AND z.provider = 'ЗАПАСКА'
		  AND big.provider = 'БИГМАШИН'
	`).Scan(&allThree)

	fmt.Printf("   ГРУППА БРИНЕКС ∩ ЗАПАСКА:     %d CAE\n", brinexZapacka)
	fmt.Printf("   ГРУППА БРИНЕКС ∩ БИГМАШИН:    %d CAE\n", brinexBigmashin)
	fmt.Printf("   ЗАПАСКА ∩ БИГМАШИН:           %d CAE\n", zapackaBigmashin)
	fmt.Printf("   Все три провайдера:           %d CAE\n", allThree)

	// 4. Пересечение с Северавто
	fmt.Println("\n🆕 Пересечение с Северавто:\n")

	for _, provider := range []string{"ГРУППА БРИНЕКС", "ЗАПАСКА", "БИГМАШИН", "Форточки"} {
		var overlap int
		pool.QueryRow(ctx, `
			SELECT COUNT(DISTINCT p.cae)
			FROM tyres_prices_stock p
			INNER JOIN tyres_prices_stock s ON p.cae = s.cae
			WHERE p.provider = $1
			  AND s.provider = 'Северавто'
		`, provider).Scan(&overlap)

		fmt.Printf("   %s ∩ Северавто: %d CAE\n", provider, overlap)
	}

	// 5. Уникальные артикулы (только у одного провайдера)
	fmt.Println("\n🎯 Уникальные артикулы (есть только у одного провайдера):\n")

	for _, provider := range providers {
		var unique int
		pool.QueryRow(ctx, `
			SELECT COUNT(DISTINCT cae)
			FROM (
				SELECT cae
				FROM tyres_prices_stock
				WHERE provider = $1
				GROUP BY cae
				HAVING COUNT(DISTINCT provider) = 1
			) t
		`, provider).Scan(&unique)

		// Процент от всех артикулов провайдера
		var totalProvider int
		pool.QueryRow(ctx, "SELECT COUNT(DISTINCT cae) FROM tyres_prices_stock WHERE provider = $1", provider).Scan(&totalProvider)

		pct := 0.0
		if totalProvider > 0 {
			pct = float64(unique) / float64(totalProvider) * 100
		}

		fmt.Printf("   %s: %d CAE (%.1f%% от своих)\n", provider, unique, pct)
	}

	// 6. Orphan артикулы (есть в prices, но нет в nomenclature)
	fmt.Println("\n⚠️  Orphan артикулы (есть в tyres_prices_stock, но НЕТ в nomenclature_tyres):\n")

	for _, provider := range providers {
		var orphans int
		pool.QueryRow(ctx, `
			SELECT COUNT(DISTINCT tps.cae)
			FROM tyres_prices_stock tps
			LEFT JOIN nomenclature_tyres nt ON tps.cae = nt.cae
			WHERE tps.provider = $1
			  AND nt.cae IS NULL
		`, provider).Scan(&orphans)

		if orphans > 0 {
			fmt.Printf("   %s: %d orphan CAE\n", provider, orphans)

			// Показать примеры
			fmt.Println("      Примеры:")
			rows, _ := pool.Query(ctx, `
				SELECT DISTINCT tps.cae
				FROM tyres_prices_stock tps
				LEFT JOIN nomenclature_tyres nt ON tps.cae = nt.cae
				WHERE tps.provider = $1
				  AND nt.cae IS NULL
				LIMIT 5
			`, provider)
			for rows.Next() {
				var cae string
				rows.Scan(&cae)
				fmt.Printf("         - %s\n", cae)
			}
			rows.Close()
			fmt.Println()
		}
	}

	fmt.Println("✅ Анализ завершен!")
}
