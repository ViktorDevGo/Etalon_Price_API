package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

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

	fmt.Println("\n📊 Анализ совпадения брендов Форточки ↔ Северавто\n")

	// 1. Уникальные бренды Форточки
	var fourtochkiBrands int
	err = pool.QueryRow(ctx, "SELECT COUNT(DISTINCT brand) FROM nomenclature_tyres").Scan(&fourtochkiBrands)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✅ Форточки (уникальные бренды): %d\n", fourtochkiBrands)

	// 2. Уникальные бренды Северавто
	var severavtoBrands int
	err = pool.QueryRow(ctx, "SELECT COUNT(DISTINCT brand) FROM nomenclature_tyres_severavto").Scan(&severavtoBrands)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✅ Северавто (уникальные бренды): %d\n", severavtoBrands)

	// 3. Уникальные бренды Северавто с фото
	var severavtoWithPhotoBrands int
	err = pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT nts.brand)
		FROM nomenclature_tyres_severavto nts
		INNER JOIN foto_tyres ft ON nts.commodity_id = ft.commodity_id
		WHERE ft.picture_url IS NOT NULL AND ft.picture_url != ''
	`).Scan(&severavtoWithPhotoBrands)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✅ Северавто с фото (уникальные бренды): %d\n", severavtoWithPhotoBrands)

	// 4. Прямое совпадение брендов
	var directMatch int
	err = pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT nt.brand)
		FROM nomenclature_tyres nt
		INNER JOIN nomenclature_tyres_severavto nts
			ON UPPER(TRIM(nt.brand)) = UPPER(TRIM(nts.brand))
	`).Scan(&directMatch)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✅ Прямое совпадение брендов: %d\n", directMatch)

	// 5. Совпадение с фото
	var matchWithPhoto int
	err = pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT nt.brand)
		FROM nomenclature_tyres nt
		INNER JOIN nomenclature_tyres_severavto nts
			ON UPPER(TRIM(nt.brand)) = UPPER(TRIM(nts.brand))
		INNER JOIN foto_tyres ft ON nts.commodity_id = ft.commodity_id
		WHERE ft.picture_url IS NOT NULL AND ft.picture_url != ''
	`).Scan(&matchWithPhoto)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✅ Совпадение брендов с фото: %d\n", matchWithPhoto)

	// 6. Товары с совпадением по бренду
	var productsDirectMatch int
	err = pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM nomenclature_tyres nt
		INNER JOIN nomenclature_tyres_severavto nts
			ON UPPER(TRIM(nt.brand)) = UPPER(TRIM(nts.brand))
	`).Scan(&productsDirectMatch)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\n📦 Товары Форточки с совпадением бренда: %d\n", productsDirectMatch)

	// 7. Товары с совпадением по бренду и фото
	var productsWithPhoto int
	err = pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM nomenclature_tyres nt
		INNER JOIN nomenclature_tyres_severavto nts
			ON UPPER(TRIM(nt.brand)) = UPPER(TRIM(nts.brand))
		INNER JOIN foto_tyres ft ON nts.commodity_id = ft.commodity_id
		WHERE ft.picture_url IS NOT NULL AND ft.picture_url != ''
	`).Scan(&productsWithPhoto)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("📦 Товары Форточки с совпадением бренда И фото: %d\n", productsWithPhoto)

	// 8. Список брендов которые совпадают напрямую
	fmt.Println("\n🏷️  Совпадающие бренды:")
	rows, err := pool.Query(ctx, `
		SELECT DISTINCT nt.brand, COUNT(nt.id) as cnt
		FROM nomenclature_tyres nt
		INNER JOIN nomenclature_tyres_severavto nts
			ON UPPER(TRIM(nt.brand)) = UPPER(TRIM(nts.brand))
		INNER JOIN foto_tyres ft ON nts.commodity_id = ft.commodity_id
		WHERE ft.picture_url IS NOT NULL AND ft.picture_url != ''
		GROUP BY nt.brand
		ORDER BY cnt DESC
		LIMIT 20
	`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var brand string
		var count int
		rows.Scan(&brand, &count)
		fmt.Printf("   - %s: %d товаров\n", brand, count)
	}

	// 9. Brand mapping check
	brandMapping := map[string][]string{
		"PIRELLI":     {"ПИРЕЛЛИ", "Пирелли"},
		"MICHELIN":    {"МИШЕЛИН", "Мишелин"},
		"YOKOHAMA":    {"ЙОКОХАМА", "Йокохама"},
		"BRIDGESTONE": {"БРИДЖСТОУН", "Бриджстоун"},
		"CONTINENTAL": {"КОНТИНЕНТАЛЬ", "Континенталь"},
		"GOODYEAR":    {"ГУДИЕР", "Гудиер"},
		"HANKOOK":     {"КУМХО", "Кумхо", "ХАНКУК"},
		"NOKIAN":      {"НОКИАН", "Нокиан"},
		"NOKIANTYRES": {"НОКИАН", "Нокиан"},
		"TOYO":        {"ТОЙО", "Тойо"},
		"DUNLOP":      {"ДАНЛОП", "Данлоп"},
		"KUMHO":       {"КУМХО", "Кумхо"},
		"CORDIANT":    {"КОРДИАНТ", "Кордиант"},
		"KORDIANT":    {"КОРДИАНТ", "Кордиант"},
	}

	fmt.Println("\n🔄 Проверка маппинга брендов (Форточки → Северавто):")
	for engBrand, ruBrands := range brandMapping {
		// Check if Fourtochki has this brand
		var ftCount int
		pool.QueryRow(ctx, "SELECT COUNT(*) FROM nomenclature_tyres WHERE UPPER(TRIM(brand)) = $1", engBrand).Scan(&ftCount)

		// Check if Severavto has Russian variants
		var svCount int
		placeholders := make([]string, len(ruBrands))
		args := make([]interface{}, len(ruBrands))
		for i, rb := range ruBrands {
			placeholders[i] = fmt.Sprintf("$%d", i+1)
			args[i] = strings.ToUpper(strings.TrimSpace(rb))
		}
		query := fmt.Sprintf(`
			SELECT COUNT(*)
			FROM nomenclature_tyres_severavto nts
			INNER JOIN foto_tyres ft ON nts.commodity_id = ft.commodity_id
			WHERE ft.picture_url IS NOT NULL
			  AND ft.picture_url != ''
			  AND UPPER(TRIM(nts.brand)) IN (%s)
		`, strings.Join(placeholders, ","))
		pool.QueryRow(ctx, query, args...).Scan(&svCount)

		if ftCount > 0 && svCount > 0 {
			fmt.Printf("   ✅ %s (%d) ↔ %v (%d с фото)\n", engBrand, ftCount, ruBrands, svCount)
		} else if ftCount > 0 {
			fmt.Printf("   ⚠️  %s (%d) ↔ %v (0 с фото)\n", engBrand, ftCount, ruBrands)
		}
	}

	fmt.Println()
}
