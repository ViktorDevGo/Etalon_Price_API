package main

import (
	"context"
	"fmt"
	"log"

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
	fmt.Println("ПОЛНОЕ СРАВНЕНИЕ: etalon_nomenclature vs nomenclature_tyres/rims")
	fmt.Println("================================================================")
	fmt.Println()

	// 1. Статистика по etalon_nomenclature
	var uniqueArticles, totalEtalon int
	pool.QueryRow(ctx, "SELECT COUNT(DISTINCT article) FROM etalon_nomenclature WHERE article IS NOT NULL AND article != ''").Scan(&uniqueArticles)
	pool.QueryRow(ctx, "SELECT COUNT(*) FROM etalon_nomenclature").Scan(&totalEtalon)

	fmt.Println("📦 ETALON_NOMENCLATURE:")
	fmt.Printf("   Всего записей: %d\n", totalEtalon)
	fmt.Printf("   Уникальных article: %d\n", uniqueArticles)
	fmt.Println()

	// 2. Статистика по nomenclature_tyres
	var uniqueTyres, totalTyres int
	pool.QueryRow(ctx, "SELECT COUNT(DISTINCT cae) FROM nomenclature_tyres").Scan(&uniqueTyres)
	pool.QueryRow(ctx, "SELECT COUNT(*) FROM nomenclature_tyres").Scan(&totalTyres)

	fmt.Println("🚗 NOMENCLATURE_TYRES (шины):")
	fmt.Printf("   Всего записей: %d\n", totalTyres)
	fmt.Printf("   Уникальных cae: %d\n", uniqueTyres)
	fmt.Println()

	// 3. Статистика по nomenclature_rims
	var uniqueRims, totalRims int
	pool.QueryRow(ctx, "SELECT COUNT(DISTINCT cae) FROM nomenclature_rims").Scan(&uniqueRims)
	pool.QueryRow(ctx, "SELECT COUNT(*) FROM nomenclature_rims").Scan(&totalRims)

	fmt.Println("⭕ NOMENCLATURE_RIMS (диски):")
	fmt.Printf("   Всего записей: %d\n", totalRims)
	fmt.Printf("   Уникальных cae: %d\n", uniqueRims)
	fmt.Println()

	// 4. Совпадения article → tyres.cae
	var matchTyres int
	pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT e.article)
		FROM etalon_nomenclature e
		WHERE e.article IS NOT NULL AND e.article != ''
		  AND EXISTS (SELECT 1 FROM nomenclature_tyres t WHERE t.cae = e.article)
	`).Scan(&matchTyres)

	// 5. Совпадения article → rims.cae
	var matchRims int
	pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT e.article)
		FROM etalon_nomenclature e
		WHERE e.article IS NOT NULL AND e.article != ''
		  AND EXISTS (SELECT 1 FROM nomenclature_rims r WHERE r.cae = e.article)
	`).Scan(&matchRims)

	// 6. Совпадения в обеих таблицах
	var matchBoth int
	pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT e.article)
		FROM etalon_nomenclature e
		WHERE e.article IS NOT NULL AND e.article != ''
		  AND EXISTS (SELECT 1 FROM nomenclature_tyres t WHERE t.cae = e.article)
		  AND EXISTS (SELECT 1 FROM nomenclature_rims r WHERE r.cae = e.article)
	`).Scan(&matchBoth)

	// 7. Совпадения хотя бы в одной таблице
	var matchAny int
	pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT e.article)
		FROM etalon_nomenclature e
		WHERE e.article IS NOT NULL AND e.article != ''
		  AND (
			EXISTS (SELECT 1 FROM nomenclature_tyres t WHERE t.cae = e.article)
			OR EXISTS (SELECT 1 FROM nomenclature_rims r WHERE r.cae = e.article)
		  )
	`).Scan(&matchAny)

	// 8. Не найдено ни в одной таблице
	notFound := uniqueArticles - matchAny

	fmt.Println("================================================================")
	fmt.Println("🔗 СОВПАДЕНИЯ (article → cae)")
	fmt.Println("================================================================")
	fmt.Println()

	fmt.Printf("✅ Найдено в nomenclature_tyres:     %6d (%.2f%%)\n",
		matchTyres, float64(matchTyres)/float64(uniqueArticles)*100)
	fmt.Printf("✅ Найдено в nomenclature_rims:      %6d (%.2f%%)\n",
		matchRims, float64(matchRims)/float64(uniqueArticles)*100)
	fmt.Printf("⚠️  Найдено в ОБЕИХ таблицах:        %6d (%.2f%%)\n",
		matchBoth, float64(matchBoth)/float64(uniqueArticles)*100)
	fmt.Printf("✅ Найдено ХОТЯ БЫ в одной:          %6d (%.2f%%)\n",
		matchAny, float64(matchAny)/float64(uniqueArticles)*100)
	fmt.Printf("❌ НЕ найдено ни в одной:            %6d (%.2f%%)\n",
		notFound, float64(notFound)/float64(uniqueArticles)*100)
	fmt.Println()

	// 9. Только в шинах (не в дисках)
	onlyTyres := matchTyres - matchBoth
	fmt.Printf("🚗 Только в шинах (не в дисках):     %6d (%.2f%%)\n",
		onlyTyres, float64(onlyTyres)/float64(uniqueArticles)*100)

	// 10. Только в дисках (не в шинах)
	onlyRims := matchRims - matchBoth
	fmt.Printf("⭕ Только в дисках (не в шинах):     %6d (%.2f%%)\n",
		onlyRims, float64(onlyRims)/float64(uniqueArticles)*100)
	fmt.Println()

	// 11. Примеры не найденных
	if notFound > 0 {
		fmt.Println("📋 Примеры article НЕ найденных ни в шинах, ни в дисках (первые 10):")

		rows, err := pool.Query(ctx, `
			SELECT DISTINCT e.article
			FROM etalon_nomenclature e
			WHERE e.article IS NOT NULL AND e.article != ''
			  AND NOT EXISTS (SELECT 1 FROM nomenclature_tyres t WHERE t.cae = e.article)
			  AND NOT EXISTS (SELECT 1 FROM nomenclature_rims r WHERE r.cae = e.article)
			ORDER BY e.article
			LIMIT 10
		`)

		if err == nil {
			defer rows.Close()
			i := 1
			for rows.Next() {
				var article string
				rows.Scan(&article)
				fmt.Printf("   %2d. %s\n", i, article)
				i++
			}
		}
		fmt.Println()
	}

	// 12. Обратное сравнение - сколько из tyres/rims есть в etalon
	var tyresInEtalon, rimsInEtalon int
	pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT t.cae)
		FROM nomenclature_tyres t
		WHERE EXISTS (SELECT 1 FROM etalon_nomenclature e WHERE e.article = t.cae)
	`).Scan(&tyresInEtalon)

	pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT r.cae)
		FROM nomenclature_rims r
		WHERE EXISTS (SELECT 1 FROM etalon_nomenclature e WHERE e.article = r.cae)
	`).Scan(&rimsInEtalon)

	tyresNotInEtalon := uniqueTyres - tyresInEtalon
	rimsNotInEtalon := uniqueRims - rimsInEtalon

	fmt.Println("================================================================")
	fmt.Println("🔄 ОБРАТНОЕ СРАВНЕНИЕ (cae → article)")
	fmt.Println("================================================================")
	fmt.Println()

	fmt.Println("🚗 Шины (nomenclature_tyres):")
	fmt.Printf("   ✅ Найдено в etalon_nomenclature: %6d (%.2f%%)\n",
		tyresInEtalon, float64(tyresInEtalon)/float64(uniqueTyres)*100)
	fmt.Printf("   ❌ НЕ найдено:                    %6d (%.2f%%)\n",
		tyresNotInEtalon, float64(tyresNotInEtalon)/float64(uniqueTyres)*100)
	fmt.Println()

	fmt.Println("⭕ Диски (nomenclature_rims):")
	fmt.Printf("   ✅ Найдено в etalon_nomenclature: %6d (%.2f%%)\n",
		rimsInEtalon, float64(rimsInEtalon)/float64(uniqueRims)*100)
	fmt.Printf("   ❌ НЕ найдено:                    %6d (%.2f%%)\n",
		rimsNotInEtalon, float64(rimsNotInEtalon)/float64(uniqueRims)*100)
	fmt.Println()

	// 13. Итоговая диаграмма
	fmt.Println("================================================================")
	fmt.Println("📊 ИТОГОВАЯ ДИАГРАММА")
	fmt.Println("================================================================")
	fmt.Println()

	fmt.Println("Распределение article из etalon_nomenclature:")
	fmt.Println()
	fmt.Printf("  Всего уникальных article: %d\n", uniqueArticles)
	fmt.Println("  ├─ 🚗 Только шины:        " + fmt.Sprintf("%6d (%.1f%%)", onlyTyres, float64(onlyTyres)/float64(uniqueArticles)*100))
	fmt.Println("  ├─ ⭕ Только диски:        " + fmt.Sprintf("%6d (%.1f%%)", onlyRims, float64(onlyRims)/float64(uniqueArticles)*100))
	fmt.Println("  ├─ 🔄 И шины, и диски:    " + fmt.Sprintf("%6d (%.1f%%)", matchBoth, float64(matchBoth)/float64(uniqueArticles)*100))
	fmt.Println("  └─ ❌ Не найдено:         " + fmt.Sprintf("%6d (%.1f%%)", notFound, float64(notFound)/float64(uniqueArticles)*100))
	fmt.Println()

	// 14. Общая статистика по всем таблицам
	totalUnique := uniqueTyres + uniqueRims
	totalRecords := totalTyres + totalRims

	fmt.Println("================================================================")
	fmt.Println("📈 ОБЩАЯ СТАТИСТИКА")
	fmt.Println("================================================================")
	fmt.Println()

	fmt.Printf("Шины + Диски (nomenclature_tyres + nomenclature_rims):\n")
	fmt.Printf("  • Всего записей: %d (шины: %d + диски: %d)\n", totalRecords, totalTyres, totalRims)
	fmt.Printf("  • Всего уникальных cae: %d (шины: %d + диски: %d)\n", totalUnique, uniqueTyres, uniqueRims)
	fmt.Println()

	fmt.Printf("Эталонная таблица (etalon_nomenclature):\n")
	fmt.Printf("  • Всего записей: %d\n", totalEtalon)
	fmt.Printf("  • Уникальных article: %d\n", uniqueArticles)
	fmt.Printf("  • Покрытие: %.2f%% (найдено %d из %d)\n",
		float64(matchAny)/float64(uniqueArticles)*100, matchAny, uniqueArticles)
	fmt.Println()

	coverage := float64(matchAny) / float64(uniqueArticles) * 100
	if coverage >= 90 {
		fmt.Println("✅ Отличное покрытие! Большинство товаров из эталона есть в nomenclature.")
	} else if coverage >= 70 {
		fmt.Println("⚠️  Хорошее покрытие, но есть товары не синхронизированные.")
	} else if coverage >= 50 {
		fmt.Println("⚠️  Среднее покрытие. Много товаров из эталона отсутствует в nomenclature.")
	} else {
		fmt.Println("❌ Низкое покрытие. Большинство товаров из эталона не синхронизировано.")
	}

	fmt.Println()
	fmt.Println("================================================================")
}
