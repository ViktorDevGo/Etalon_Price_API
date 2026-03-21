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

	fmt.Println("\n🔍 Проверка: попадает ли Северавто в переоценку\n")

	// 1. Сколько товаров Северавто с остатком > 0
	var severavtoWithStock int
	pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT cae)
		FROM tyres_prices_stock p
		INNER JOIN nomenclature_tyres n ON p.cae = n.cae
		WHERE p.provider = 'Северавто'
		  AND p.stock > 0
		  AND n.tiretype = 'Легковая'
	`).Scan(&severavtoWithStock)

	fmt.Printf("📊 Северавто с остатком > 0: %d CAE\n", severavtoWithStock)

	// 2. Проверим логику экспорта (точно такую же как в export-bitrix-prices)
	var exportedFromSeveravto int
	pool.QueryRow(ctx, `
		WITH filtered_stock AS (
			SELECT
				p.cae,
				p.provider,
				p.stock,
				p.price,
				p.isimport
			FROM tyres_prices_stock p
			INNER JOIN nomenclature_tyres n ON p.cae = n.cae
			WHERE p.stock > 0 AND n.tiretype = 'Легковая'
		),
		grouped_stock AS (
			SELECT
				cae,
				SUM(stock) as total_stock,
				MIN(price) as min_price,
				MIN(isimport) as has_isimport_0,
				SUM(CASE
					WHEN LOWER(provider) IN ('запаска', 'форточки', 'бигмашин')
					THEN 1
					ELSE 0
				END) as has_fast_provider,
				-- Добавим подсчет записей от Северавто
				SUM(CASE
					WHEN provider = 'Северавто'
					THEN 1
					ELSE 0
				END) as has_severavto
			FROM filtered_stock
			GROUP BY cae
			HAVING
				SUM(stock) > 4
				AND MIN(isimport) = 0
		)
		SELECT COUNT(*)
		FROM grouped_stock
		WHERE has_severavto > 0
	`).Scan(&exportedFromSeveravto)

	fmt.Printf("📊 В переоценке (SUM(stock) > 4): %d CAE от Северавто\n\n", exportedFromSeveravto)

	// 3. Проверим примеры товаров
	fmt.Println("📝 Примеры товаров Северавто в переоценке:\n")
	rows, _ := pool.Query(ctx, `
		WITH filtered_stock AS (
			SELECT
				p.cae,
				p.provider,
				p.stock,
				p.price,
				p.isimport
			FROM tyres_prices_stock p
			INNER JOIN nomenclature_tyres n ON p.cae = n.cae
			WHERE p.stock > 0 AND n.tiretype = 'Легковая'
		),
		grouped_stock AS (
			SELECT
				cae,
				SUM(stock) as total_stock,
				MIN(price) as min_price,
				MIN(isimport) as has_isimport_0,
				SUM(CASE
					WHEN LOWER(provider) IN ('запаска', 'форточки', 'бигмашин')
					THEN 1
					ELSE 0
				END) as has_fast_provider,
				SUM(CASE
					WHEN provider = 'Северавто'
					THEN 1
					ELSE 0
				END) as has_severavto,
				-- Получим список провайдеров
				STRING_AGG(DISTINCT provider, ', ' ORDER BY provider) as providers
			FROM filtered_stock
			GROUP BY cae
			HAVING
				SUM(stock) > 4
				AND MIN(isimport) = 0
				AND SUM(CASE WHEN provider = 'Северавто' THEN 1 ELSE 0 END) > 0
		)
		SELECT
			cae,
			total_stock,
			min_price,
			has_fast_provider,
			has_severavto,
			providers,
			CASE
				WHEN has_fast_provider > 0 THEN 'Доставим курьером Сегодня и позже'
				ELSE 'Доставим курьером Завтра и позже'
			END as delivery_time
		FROM grouped_stock
		ORDER BY total_stock DESC
		LIMIT 10
	`)
	defer rows.Close()

	for rows.Next() {
		var cae, providers, delivery string
		var totalStock, hasFast, hasSeveravto, minPrice int
		rows.Scan(&cae, &totalStock, &minPrice, &hasFast, &hasSeveravto, &providers, &delivery)
		fmt.Printf("   CAE: %s\n", cae)
		fmt.Printf("   Провайдеры: %s\n", providers)
		fmt.Printf("   Остаток: %d | Цена: %d₽ | Срок: %s\n\n", totalStock, minPrice, delivery)
	}

	// 4. Статистика по срокам доставки
	fmt.Println("📊 Распределение по срокам доставки:\n")
	rows, _ = pool.Query(ctx, `
		WITH filtered_stock AS (
			SELECT
				p.cae,
				p.provider,
				p.stock,
				p.price,
				p.isimport
			FROM tyres_prices_stock p
			INNER JOIN nomenclature_tyres n ON p.cae = n.cae
			WHERE p.stock > 0 AND n.tiretype = 'Легковая'
		),
		grouped_stock AS (
			SELECT
				cae,
				SUM(stock) as total_stock,
				SUM(CASE
					WHEN LOWER(provider) IN ('запаска', 'форточки', 'бигмашин')
					THEN 1
					ELSE 0
				END) as has_fast_provider,
				SUM(CASE
					WHEN provider = 'Северавто'
					THEN 1
					ELSE 0
				END) as has_severavto
			FROM filtered_stock
			GROUP BY cae
			HAVING
				SUM(stock) > 4
				AND MIN(isimport) = 0
		)
		SELECT
			CASE
				WHEN has_fast_provider > 0 THEN 'Сегодня и позже (быстрые)'
				ELSE 'Завтра и позже (медленные)'
			END as delivery,
			COUNT(*) as cnt,
			SUM(CASE WHEN has_severavto > 0 THEN 1 ELSE 0 END) as with_severavto
		FROM grouped_stock
		GROUP BY
			CASE
				WHEN has_fast_provider > 0 THEN 'Сегодня и позже (быстрые)'
				ELSE 'Завтра и позже (медленные)'
			END
	`)
	defer rows.Close()

	for rows.Next() {
		var delivery string
		var cnt, withSeveravto int
		rows.Scan(&delivery, &cnt, &withSeveravto)
		fmt.Printf("   %s:\n", delivery)
		fmt.Printf("      Всего: %d товаров\n", cnt)
		fmt.Printf("      С Северавто: %d товаров (%.1f%%)\n\n",
			withSeveravto, float64(withSeveravto)/float64(cnt)*100)
	}

	fmt.Println("💡 Выводы:")
	fmt.Println("   ✅ Северавто ПОПАДАЕТ в переоценку")
	fmt.Println("   📦 Учитывается в SUM(stock) и MIN(price)")
	fmt.Println("   🚚 Срок доставки:")
	fmt.Println("      - 'Сегодня и позже' - если есть Запаска/Форточки/Бигмашин")
	fmt.Println("      - 'Завтра и позже' - если только Северавто/Группа Бринекс")
	fmt.Println()
}
