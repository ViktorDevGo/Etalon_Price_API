package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	// Флаги
	outputDir := flag.String("output-dir", "/Users/viktor/Desktop", "Директория для выходного CSV файла")
	limit := flag.Int("limit", 0, "Ограничение количества строк для тестирования (0 = без ограничений)")
	flag.Parse()

	_ = godotenv.Load()

	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_DSN"))
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	ctx := context.Background()

	fmt.Println("================================================================")
	fmt.Println("Выгрузка цен и остатков для каталога шин")
	fmt.Println("================================================================")
	fmt.Println()

	// Генерируем имя файла с текущей датой
	dateStr := time.Now().Format("20060102") // YYYYMMDD
	filename := fmt.Sprintf("Переоценка_шины_%s.csv", dateStr)
	if *limit > 0 {
		filename = fmt.Sprintf("Переоценка_шины_%s_TEST_%d.csv", dateStr, *limit)
	}

	fmt.Printf("📦 Экспорт цен и остатков\n")
	if *limit > 0 {
		fmt.Printf("⚠️  ТЕСТОВЫЙ РЕЖИМ: ограничение %d строк\n", *limit)
	}
	fmt.Println()

	exported := exportPrices(ctx, pool, *outputDir, filename, *limit)

	fmt.Println()
	fmt.Println("================================================================")
	fmt.Printf("✅ Экспортировано: %d позиций\n", exported)
	fmt.Printf("📁 Файл: %s/%s\n", *outputDir, filename)
	fmt.Println()
	fmt.Println("✅ Файл готов для импорта в 1C-Битрикс!")
	fmt.Println("================================================================")
}

func exportPrices(ctx context.Context, pool *pgxpool.Pool, outputDir, filename string, limit int) int {
	// SQL запрос для получения цен и остатков
	limitClause := ""
	if limit > 0 {
		limitClause = fmt.Sprintf("LIMIT %d", limit)
	}

	query := fmt.Sprintf(`
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
				-- Проверяем наличие быстрых поставщиков (case-insensitive)
				SUM(CASE
					WHEN LOWER(provider) IN ('запаска', 'форточки', 'бигмашин')
					THEN 1
					ELSE 0
				END) as has_fast_provider
			FROM filtered_stock
			GROUP BY cae
			HAVING
				SUM(stock) > 4
				AND MIN(isimport) = 0
		)
		SELECT
			cae,
			total_stock,
			min_price,
			has_fast_provider
		FROM grouped_stock
		ORDER BY cae
		%s
	`, limitClause)

	rows, err := pool.Query(ctx, query)
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	defer rows.Close()

	// Создаем CSV
	filepath := fmt.Sprintf("%s/%s", outputDir, filename)
	file, err := os.Create(filepath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	writer.Comma = ',' // Используем запятую как разделитель
	defer writer.Flush()

	// Заголовок (5 колонок)
	header := []string{
		"IE_XML_ID", "IP_PROP171", "CP_QUANTITY", "CV_PRICE_1", "CV_CURRENCY_1",
	}

	if err := writer.Write(header); err != nil {
		log.Fatal(err)
	}

	fmt.Println("✅ Логика экспорта:")
	fmt.Println("  ✓ Источник: tyres_prices_stock + nomenclature_tyres")
	fmt.Println("  ✓ Фильтр: tiretype = 'Легковая' AND SUM(stock) > 4 AND MIN(isimport) = 0")
	fmt.Println("  ✓ QUANTITY: SUM(stock) по CAE")
	fmt.Println("  ✓ CV_PRICE_1: CEIL(MIN(price) * 1.12) - округление вверх")
	fmt.Println("  ✓ IP_PROP171:")
	fmt.Println("      - 'Доставим курьером Сегодня и позже' - если есть Запаска/Форточки/Бигмашин")
	fmt.Println("      - 'Доставим курьером Завтра и позже' - если только Группа Бринекс/Северавто")
	fmt.Println()
	fmt.Println("Экспорт данных...")

	count := 0
	exportedCAEs := []string{}

	for rows.Next() {
		var (
			cae                      string
			totalStock, hasFastProvider int
			minPrice                 int
		)

		if err := rows.Scan(&cae, &totalStock, &minPrice, &hasFastProvider); err != nil {
			log.Fatal(err)
		}

		exportedCAEs = append(exportedCAEs, cae)

		record := make([]string, 5)

		// 1. IE_XML_ID (CAE)
		record[0] = cae

		// 2. IP_PROP171 - срок доставки
		if hasFastProvider > 0 {
			// Есть быстрые поставщики (Запаска/Форточки/Бигмашин)
			record[1] = "Доставим курьером Сегодня и позже"
		} else {
			// Только медленные поставщики (Группа Бринекс/Северавто)
			record[1] = "Доставим курьером Завтра и позже"
		}

		// 3. QUANTITY - суммарный остаток
		record[2] = fmt.Sprintf("%d", totalStock)

		// 4. CV_PRICE_1 - минимальная цена + 12%, округление вверх
		priceRub := float64(minPrice)
		priceWithMarkup := priceRub * 1.12
		// Округление вверх до целого числа
		priceRounded := int(math.Ceil(priceWithMarkup))
		record[3] = fmt.Sprintf("%d", priceRounded)

		// 5. CV_CURRENCY_1 - валюта
		record[4] = "RUB"

		if err := writer.Write(record); err != nil {
			log.Fatal(err)
		}

		count++
	}

	// После экспорта - обновить isimport = 1 для всех экспортированных товаров
	// В тестовом режиме (limit > 0) НЕ обновляем isimport
	if len(exportedCAEs) > 0 && limit == 0 {
		fmt.Println()
		fmt.Printf("📝 Обновление isimport для %d товаров...\n", len(exportedCAEs))

		updateQuery := `
			UPDATE tyres_prices_stock
			SET isimport = 1
			WHERE cae = ANY($1) AND isimport = 0
		`

		result, err := pool.Exec(ctx, updateQuery, exportedCAEs)
		if err != nil {
			log.Printf("⚠️  Ошибка обновления isimport: %v", err)
		} else {
			rowsAffected := result.RowsAffected()
			fmt.Printf("✅ Обновлено %d строк (isimport = 1)\n", rowsAffected)
		}
	} else if limit > 0 {
		fmt.Println()
		fmt.Println("⚠️  ТЕСТОВЫЙ РЕЖИМ: isimport НЕ обновлен")
	}

	return count
}
