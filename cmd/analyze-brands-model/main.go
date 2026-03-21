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

	fmt.Println("\n📊 Анализ совпадения по БРЕНДУ + МОДЕЛИ\n")

	// 1. JOIN по бренду + модель (точное совпадение)
	var exactBrandModel int
	err = pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT nt.id)
		FROM nomenclature_tyres nt
		INNER JOIN nomenclature_tyres_severavto nts
			ON UPPER(TRIM(nt.brand)) = UPPER(TRIM(nts.brand))
			AND UPPER(TRIM(nt.model)) = UPPER(TRIM(nts.model))
		INNER JOIN foto_tyres ft ON nts.commodity_id = ft.commodity_id
		WHERE ft.picture_url IS NOT NULL AND ft.picture_url != ''
	`).Scan(&exactBrandModel)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✅ Точное совпадение (бренд + модель): %d товаров\n", exactBrandModel)

	// 2. JOIN по бренду + модель (нормализованное - без пробелов/дефисов)
	var normalizedBrandModel int
	err = pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT nt.id)
		FROM nomenclature_tyres nt
		INNER JOIN nomenclature_tyres_severavto nts
			ON UPPER(TRIM(nt.brand)) = UPPER(TRIM(nts.brand))
			AND REGEXP_REPLACE(UPPER(nt.model), '[^A-Z0-9]', '', 'g') =
			    REGEXP_REPLACE(UPPER(nts.model), '[^A-Z0-9]', '', 'g')
		INNER JOIN foto_tyres ft ON nts.commodity_id = ft.commodity_id
		WHERE ft.picture_url IS NOT NULL AND ft.picture_url != ''
	`).Scan(&normalizedBrandModel)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✅ Нормализованное (бренд + модель без спецсимволов): %d товаров\n", normalizedBrandModel)

	// 3. JOIN по бренду + модель + ширина
	var brandModelWidth int
	err = pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT nt.id)
		FROM nomenclature_tyres nt
		INNER JOIN nomenclature_tyres_severavto nts
			ON UPPER(TRIM(nt.brand)) = UPPER(TRIM(nts.brand))
			AND REGEXP_REPLACE(UPPER(nt.model), '[^A-Z0-9]', '', 'g') =
			    REGEXP_REPLACE(UPPER(nts.model), '[^A-Z0-9]', '', 'g')
			AND nt.width::float = nts.width::float
		INNER JOIN foto_tyres ft ON nts.commodity_id = ft.commodity_id
		WHERE ft.picture_url IS NOT NULL AND ft.picture_url != ''
	`).Scan(&brandModelWidth)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✅ Бренд + модель + ширина: %d товаров\n", brandModelWidth)

	// 4. JOIN по бренду + модель + все размеры
	var brandModelAllSizes int
	err = pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT nt.id)
		FROM nomenclature_tyres nt
		INNER JOIN nomenclature_tyres_severavto nts
			ON UPPER(TRIM(nt.brand)) = UPPER(TRIM(nts.brand))
			AND REGEXP_REPLACE(UPPER(nt.model), '[^A-Z0-9]', '', 'g') =
			    REGEXP_REPLACE(UPPER(nts.model), '[^A-Z0-9]', '', 'g')
			AND nt.width::float = nts.width::float
			AND nt.height::float = nts.height::float
			AND REGEXP_REPLACE(nt.diameter, '[^0-9.]', '', 'g') = nts.diameter
		INNER JOIN foto_tyres ft ON nts.commodity_id = ft.commodity_id
		WHERE ft.picture_url IS NOT NULL AND ft.picture_url != ''
	`).Scan(&brandModelAllSizes)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✅ Бренд + модель + размеры (W/H/D): %d товаров\n", brandModelAllSizes)

	// 5. JOIN по бренду + модель + размеры + индексы
	var brandModelSizesIndices int
	err = pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT nt.id)
		FROM nomenclature_tyres nt
		INNER JOIN nomenclature_tyres_severavto nts
			ON UPPER(TRIM(nt.brand)) = UPPER(TRIM(nts.brand))
			AND REGEXP_REPLACE(UPPER(nt.model), '[^A-Z0-9]', '', 'g') =
			    REGEXP_REPLACE(UPPER(nts.model), '[^A-Z0-9]', '', 'g')
			AND nt.width::float = nts.width::float
			AND nt.height::float = nts.height::float
			AND REGEXP_REPLACE(nt.diameter, '[^0-9.]', '', 'g') = nts.diameter
			AND nt.load_index = nts.load_index
			AND nt.speed_index = nts.speed_index
		INNER JOIN foto_tyres ft ON nts.commodity_id = ft.commodity_id
		WHERE ft.picture_url IS NOT NULL AND ft.picture_url != ''
	`).Scan(&brandModelSizesIndices)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✅ Полное совпадение (бренд+модель+размеры+индексы): %d товаров\n", brandModelSizesIndices)

	// 6. Только по бренду (для сравнения)
	var onlyBrand int
	err = pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT nt.id)
		FROM nomenclature_tyres nt
		INNER JOIN nomenclature_tyres_severavto nts
			ON UPPER(TRIM(nt.brand)) = UPPER(TRIM(nts.brand))
		INNER JOIN foto_tyres ft ON nts.commodity_id = ft.commodity_id
		WHERE ft.picture_url IS NOT NULL AND ft.picture_url != ''
	`).Scan(&onlyBrand)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n📈 Сравнение методов:\n")
	fmt.Printf("   Только бренд:                          %7d (100.0%%)\n", onlyBrand)
	fmt.Printf("   Бренд + модель (точное):               %7d (%.1f%%)\n", exactBrandModel, float64(exactBrandModel)/float64(onlyBrand)*100)
	fmt.Printf("   Бренд + модель (норм.):                %7d (%.1f%%)\n", normalizedBrandModel, float64(normalizedBrandModel)/float64(onlyBrand)*100)
	fmt.Printf("   Бренд + модель + ширина:               %7d (%.1f%%)\n", brandModelWidth, float64(brandModelWidth)/float64(onlyBrand)*100)
	fmt.Printf("   Бренд + модель + размеры:              %7d (%.1f%%)\n", brandModelAllSizes, float64(brandModelAllSizes)/float64(onlyBrand)*100)
	fmt.Printf("   Бренд + модель + размеры + индексы:    %7d (%.1f%%)\n", brandModelSizesIndices, float64(brandModelSizesIndices)/float64(onlyBrand)*100)

	fmt.Println("\n🎯 Текущий результат экспорта (с приоритетами): 21,979 товаров")
	fmt.Printf("   Покрытие от потенциала (только бренд): %.1f%%\n", float64(21979)/float64(onlyBrand)*100)

	// 7. Примеры несовпадений моделей
	fmt.Println("\n📝 Примеры различий в написании моделей (топ-10):\n")
	rows, err := pool.Query(ctx, `
		SELECT DISTINCT
			nt.brand,
			nt.model as model_ft,
			nts.model as model_sv,
			COUNT(*) as cnt
		FROM nomenclature_tyres nt
		INNER JOIN nomenclature_tyres_severavto nts
			ON UPPER(TRIM(nt.brand)) = UPPER(TRIM(nts.brand))
			AND UPPER(TRIM(nt.model)) != UPPER(TRIM(nts.model))
			AND REGEXP_REPLACE(UPPER(nt.model), '[^A-Z0-9]', '', 'g') =
			    REGEXP_REPLACE(UPPER(nts.model), '[^A-Z0-9]', '', 'g')
		INNER JOIN foto_tyres ft ON nts.commodity_id = ft.commodity_id
		WHERE ft.picture_url IS NOT NULL AND ft.picture_url != ''
		GROUP BY nt.brand, nt.model, nts.model
		ORDER BY cnt DESC
		LIMIT 10
	`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var brand, modelFT, modelSV string
		var cnt int
		rows.Scan(&brand, &modelFT, &modelSV, &cnt)
		fmt.Printf("   %s:\n", brand)
		fmt.Printf("      Форточки:  %s\n", modelFT)
		fmt.Printf("      Северавто: %s\n", modelSV)
		fmt.Printf("      Товаров:   %d\n\n", cnt)
	}

	fmt.Println()
}
