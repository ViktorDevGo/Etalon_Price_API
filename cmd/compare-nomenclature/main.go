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
	fmt.Println("Сравнение таблиц: etalon_nomenclature vs nomenclature_tyres")
	fmt.Println("================================================================")
	fmt.Println()

	// 1. Уникальные значения в etalon_nomenclature.article
	var uniqueArticles int
	err = pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT article) 
		FROM etalon_nomenclature
		WHERE article IS NOT NULL AND article != ''
	`).Scan(&uniqueArticles)
	
	if err != nil {
		log.Fatalf("Failed to count unique articles: %v", err)
	}

	fmt.Printf("📊 Уникальных article в etalon_nomenclature: %d\n", uniqueArticles)

	// 2. Всего записей в etalon_nomenclature
	var totalEtalon int
	err = pool.QueryRow(ctx, `
		SELECT COUNT(*) 
		FROM etalon_nomenclature
	`).Scan(&totalEtalon)
	
	if err != nil {
		log.Fatalf("Failed to count total: %v", err)
	}

	fmt.Printf("📋 Всего записей в etalon_nomenclature: %d\n", totalEtalon)
	fmt.Println()

	// 3. Уникальные значения в nomenclature_tyres.cae
	var uniqueCAEs int
	err = pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT cae) 
		FROM nomenclature_tyres
		WHERE cae IS NOT NULL AND cae != ''
	`).Scan(&uniqueCAEs)
	
	if err != nil {
		log.Fatalf("Failed to count unique CAEs: %v", err)
	}

	fmt.Printf("📊 Уникальных cae в nomenclature_tyres: %d\n", uniqueCAEs)

	// 4. Всего записей в nomenclature_tyres
	var totalTyres int
	err = pool.QueryRow(ctx, `
		SELECT COUNT(*) 
		FROM nomenclature_tyres
	`).Scan(&totalTyres)
	
	if err != nil {
		log.Fatalf("Failed to count total tyres: %v", err)
	}

	fmt.Printf("📋 Всего записей в nomenclature_tyres: %d\n", totalTyres)
	fmt.Println()

	// 5. Сколько article из etalon_nomenclature есть в nomenclature_tyres.cae
	var matchingCount int
	err = pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT e.article)
		FROM etalon_nomenclature e
		WHERE e.article IS NOT NULL 
		  AND e.article != ''
		  AND EXISTS (
			SELECT 1 
			FROM nomenclature_tyres t 
			WHERE t.cae = e.article
		  )
	`).Scan(&matchingCount)
	
	if err != nil {
		log.Fatalf("Failed to count matching: %v", err)
	}

	fmt.Println("🔗 Совпадения (article == cae):")
	fmt.Printf("   Найдено совпадений: %d\n", matchingCount)
	fmt.Printf("   Процент от уникальных article: %.2f%%\n", 
		float64(matchingCount)/float64(uniqueArticles)*100)
	fmt.Println()

	// 6. Сколько article НЕТ в nomenclature_tyres.cae
	notInTyres := uniqueArticles - matchingCount
	fmt.Println("❌ Не найдены в nomenclature_tyres:")
	fmt.Printf("   Количество: %d\n", notInTyres)
	fmt.Printf("   Процент от уникальных article: %.2f%%\n", 
		float64(notInTyres)/float64(uniqueArticles)*100)
	fmt.Println()

	// 7. Примеры article которых нет в nomenclature_tyres (первые 10)
	if notInTyres > 0 {
		fmt.Println("📋 Примеры article, которых нет в nomenclature_tyres (первые 10):")
		
		rows, err := pool.Query(ctx, `
			SELECT DISTINCT e.article
			FROM etalon_nomenclature e
			WHERE e.article IS NOT NULL 
			  AND e.article != ''
			  AND NOT EXISTS (
				SELECT 1 
				FROM nomenclature_tyres t 
				WHERE t.cae = e.article
			  )
			ORDER BY e.article
			LIMIT 10
		`)
		
		if err != nil {
			log.Fatalf("Failed to get examples: %v", err)
		}
		defer rows.Close()
		
		i := 1
		for rows.Next() {
			var article string
			if err := rows.Scan(&article); err != nil {
				log.Printf("Failed to scan: %v", err)
				continue
			}
			fmt.Printf("   %d. %s\n", i, article)
			i++
		}
		fmt.Println()
	}

	// 8. Сколько cae из nomenclature_tyres есть в etalon_nomenclature.article
	var reverseMatchingCount int
	err = pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT t.cae)
		FROM nomenclature_tyres t
		WHERE t.cae IS NOT NULL 
		  AND t.cae != ''
		  AND EXISTS (
			SELECT 1 
			FROM etalon_nomenclature e 
			WHERE e.article = t.cae
		  )
	`).Scan(&reverseMatchingCount)
	
	if err != nil {
		log.Fatalf("Failed to count reverse matching: %v", err)
	}

	fmt.Println("🔄 Обратное сравнение (cae → article):")
	fmt.Printf("   CAE найдено в etalon_nomenclature: %d\n", reverseMatchingCount)
	fmt.Printf("   Процент от уникальных cae: %.2f%%\n", 
		float64(reverseMatchingCount)/float64(uniqueCAEs)*100)
	fmt.Println()

	// 9. Сколько cae НЕТ в etalon_nomenclature.article
	notInEtalon := uniqueCAEs - reverseMatchingCount
	fmt.Println("❌ CAE не найдены в etalon_nomenclature:")
	fmt.Printf("   Количество: %d\n", notInEtalon)
	fmt.Printf("   Процент от уникальных cae: %.2f%%\n", 
		float64(notInEtalon)/float64(uniqueCAEs)*100)
	fmt.Println()

	// 10. Итоговая статистика
	fmt.Println("================================================================")
	fmt.Println("📊 ИТОГОВАЯ СТАТИСТИКА")
	fmt.Println("================================================================")
	fmt.Println()
	
	fmt.Println("Таблица etalon_nomenclature:")
	fmt.Printf("  • Всего записей: %d\n", totalEtalon)
	fmt.Printf("  • Уникальных article: %d\n", uniqueArticles)
	fmt.Printf("  • Совпадений с nomenclature_tyres: %d (%.2f%%)\n", 
		matchingCount, float64(matchingCount)/float64(uniqueArticles)*100)
	fmt.Printf("  • Не найдено в nomenclature_tyres: %d (%.2f%%)\n", 
		notInTyres, float64(notInTyres)/float64(uniqueArticles)*100)
	fmt.Println()
	
	fmt.Println("Таблица nomenclature_tyres:")
	fmt.Printf("  • Всего записей: %d\n", totalTyres)
	fmt.Printf("  • Уникальных cae: %d\n", uniqueCAEs)
	fmt.Printf("  • Совпадений с etalon_nomenclature: %d (%.2f%%)\n", 
		reverseMatchingCount, float64(reverseMatchingCount)/float64(uniqueCAEs)*100)
	fmt.Printf("  • Не найдено в etalon_nomenclature: %d (%.2f%%)\n", 
		notInEtalon, float64(notInEtalon)/float64(uniqueCAEs)*100)
	fmt.Println()
	
	fmt.Println("================================================================")
}
