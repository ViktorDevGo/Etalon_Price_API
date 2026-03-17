package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	// Флаги
	output := flag.String("output", "/Users/viktor/Desktop/bitrix_catalog.csv", "Путь к выходному CSV")
	limit := flag.Int("limit", 10, "Количество позиций (0 = все)")
	withoutPrices := flag.Bool("include-no-price", false, "Включить товары без цен (с ценой 0.00)")
	flag.Parse()

	_ = godotenv.Load()

	pool, err := pgxpool.New(context.Background(), os.Getenv("DB_DSN"))
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	ctx := context.Background()

	fmt.Println("================================================================")
	fmt.Println("Экспорт номенклатуры для 1C-Битрикс")
	fmt.Println("================================================================")
	fmt.Println()

	// Проверки качества данных
	runQualityChecks(ctx, pool)

	// Основной запрос
	query := buildQuery(*withoutPrices, *limit)

	rows, err := pool.Query(ctx, query)
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	defer rows.Close()

	// Создаем CSV
	file, err := os.Create(*output)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Заголовок Bitrix (70 колонок)
	header := []string{
		"IE_XML_ID", "IE_NAME", "IE_PREVIEW_TEXT", "IE_DETAIL_TEXT",
		"IP_PROP138", "IP_PROP139", "IP_PROP140", "IP_PROP141", "IP_PROP142",
		"IP_PROP143", "IP_PROP144", "IP_PROP145", "IP_PROP189", "IP_PROP146",
		"IP_PROP147", "IP_PROP148", "IP_PROP149", "IP_PROP150", "IP_PROP151",
		"IP_PROP152", "IP_PROP153", "IP_PROP154", "IP_PROP155", "IP_PROP156",
		"IP_PROP157", "IP_PROP158", "IP_PROP159", "IP_PROP160", "IP_PROP161",
		"IP_PROP162", "IP_PROP163", "IP_PROP164", "IP_PROP165", "IP_PROP166",
		"IP_PROP167", "IP_PROP168", "IP_PROP169", "IP_PROP170", "IP_PROP171",
		"IP_PROP172", "IP_PROP173", "IP_PROP174", "IP_PROP175", "IP_PROP176",
		"IP_PROP177", "IP_PROP178", "IP_PROP179", "IP_PROP180", "IP_PROP181",
		"IP_PROP182", "IP_PROP183", "IP_PROP184", "IP_PROP185", "IP_PROP186",
		"IP_PROP187", "IP_PROP481", "IP_PROP482", "IP_PROP188", "IC_GROUP0",
		"IC_GROUP1", "IC_GROUP2", "CP_QUANTITY", "CP_WEIGHT", "CP_WIDTH",
		"CP_HEIGHT", "CP_LENGTH", "CV_QUANTITY_FROM", "CV_QUANTITY_TO",
		"CV_PRICE_1", "CV_CURRENCY_1",
	}

	if err := writer.Write(header); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Экспорт данных...")
	count := 0

	for rows.Next() {
		var (
			cae, brand, model, season, studded, loadIndex, speedIndex, runflat string
			width, height                                                       *float64
			diameter                                                            *string
			priceKopeks, totalStock                                             int
		)

		if err := rows.Scan(&cae, &brand, &model, &width, &height, &diameter,
			&season, &studded, &loadIndex, &speedIndex, &runflat,
			&priceKopeks, &totalStock); err != nil {
			log.Fatal(err)
		}

		record := make([]string, 70)

		// IE_XML_ID - артикул
		record[0] = cae

		// IE_NAME - форматированное название
		record[1] = formatProductName(brand, model, width, height, diameter, loadIndex, speedIndex)

		// IP_PROP140 - ширина
		if width != nil {
			record[6] = strconv.FormatFloat(*width, 'f', 0, 64)
		}

		// IP_PROP141 - высота
		if height != nil {
			record[7] = strconv.FormatFloat(*height, 'f', 0, 64)
		}

		// IP_PROP142 - диаметр (очистить от "R" и ",")
		if diameter != nil {
			record[8] = normalizeDiameter(*diameter)
		}

		// IP_PROP145 - сезон
		record[11] = season

		// IP_PROP146 - шипы
		record[13] = studded

		// IP_PROP149 - индекс нагрузки
		record[16] = loadIndex

		// IP_PROP150 - индекс скорости
		record[17] = speedIndex

		// IP_PROP151 - артикул (дубль)
		record[18] = cae

		// IP_PROP163 - RunFlat
		record[30] = runflat

		// IC_GROUP0 - категория (1 = шины)
		record[58] = "1"

		// IC_GROUP1 - бренд
		record[59] = brand

		// IC_GROUP2 - модель
		record[60] = model

		// CP_QUANTITY - остаток
		record[61] = strconv.Itoa(totalStock)

		// CP_WEIGHT - вес (нет данных)
		record[62] = "0"

		// CV_PRICE_1 - цена в рублях (копейки → рубли)
		priceRub := float64(priceKopeks) / 100.0
		record[68] = fmt.Sprintf("%.2f", priceRub)

		// CV_CURRENCY_1 - валюта
		record[69] = "RUB"

		if err := writer.Write(record); err != nil {
			log.Fatal(err)
		}

		count++
		if count%1000 == 0 {
			fmt.Printf("  Обработано: %d позиций...\n", count)
		}
	}

	fmt.Println()
	fmt.Println("================================================================")
	fmt.Printf("✅ Экспортировано %d позиций\n", count)
	fmt.Printf("📁 Файл: %s\n", *output)
	fmt.Println("================================================================")
}

func buildQuery(includeNoPrices bool, limit int) string {
	query := `
		SELECT
			t.cae,
			COALESCE(t.brand, '') as brand,
			COALESCE(t.model, '') as model,
			t.width,
			t.height,
			t.diameter,
			COALESCE(t.season, '') as season,
			CASE
				WHEN t.is_studded = 'шип' THEN 'Шипованные'
				WHEN t.is_studded = 'не шип' THEN 'Нешипованные'
				ELSE ''
			END as studded,
			COALESCE(t.load_index, '') as load_index,
			COALESCE(t.speed_index, '') as speed_index,
			COALESCE(t.runflat, '') as runflat,
			COALESCE(MIN(p.price), 0) as price_kopeks,
			COALESCE(SUM(p.stock), 0) as total_stock
		FROM nomenclature_tyres t
	`

	if includeNoPrices {
		query += `LEFT JOIN tyres_prices_stock p ON t.cae = p.cae`
	} else {
		query += `INNER JOIN tyres_prices_stock p ON t.cae = p.cae`
	}

	query += `
		GROUP BY t.id, t.cae, t.brand, t.model, t.width, t.height, t.diameter,
				 t.season, t.is_studded, t.load_index, t.speed_index, t.runflat
		ORDER BY t.id
	`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	return query
}

func formatProductName(brand, model string, width, height *float64, diameter *string, loadIndex, speedIndex string) string {
	parts := []string{}

	if brand != "" {
		parts = append(parts, brand)
	}
	if model != "" {
		parts = append(parts, model)
	}

	// Размер
	sizeStr := ""
	if width != nil {
		sizeStr += strconv.FormatFloat(*width, 'f', 0, 64)
	}
	if height != nil {
		sizeStr += "/" + strconv.FormatFloat(*height, 'f', 0, 64)
	}
	if diameter != nil && *diameter != "" {
		cleanDiam := normalizeDiameter(*diameter)
		if cleanDiam != "" {
			sizeStr += " R" + cleanDiam
		}
	}
	if sizeStr != "" {
		parts = append(parts, sizeStr)
	}

	// Индексы
	if loadIndex != "" || speedIndex != "" {
		parts = append(parts, loadIndex+speedIndex)
	}

	return strings.Join(parts, " ")
}

func normalizeDiameter(s string) string {
	// "R16,00" → "16"
	// "16" → "16"
	// "R16" → "16"
	re := regexp.MustCompile(`[R,\s]`)
	cleaned := re.ReplaceAllString(s, "")

	// Убрать дробную часть если она .00
	if strings.Contains(cleaned, ".") {
		parts := strings.Split(cleaned, ".")
		if len(parts) == 2 && parts[1] == "00" {
			return parts[0]
		}
	}

	return cleaned
}

func runQualityChecks(ctx context.Context, pool *pgxpool.Pool) {
	fmt.Println("📊 Проверка качества данных...")
	fmt.Println("----------------------------------------------------------------")

	// 1. Товары без артикула
	var noCAE int
	pool.QueryRow(ctx, "SELECT COUNT(*) FROM nomenclature_tyres WHERE cae IS NULL OR cae = ''").Scan(&noCAE)
	if noCAE > 0 {
		fmt.Printf("⚠️  Товары без артикула (CAE): %d\n", noCAE)
	} else {
		fmt.Printf("✅ Товары без артикула: %d\n", noCAE)
	}

	// 2. Товары без цены
	var noPrices int
	pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM nomenclature_tyres t
		LEFT JOIN tyres_prices_stock p ON t.cae = p.cae
		WHERE p.cae IS NULL
	`).Scan(&noPrices)
	fmt.Printf("⚠️  Товары без цены: %d (%.1f%%)\n",
		noPrices, float64(noPrices)/60637.0*100)

	// 3. Товары с нулевым остатком
	var zeroStock int
	pool.QueryRow(ctx, "SELECT COUNT(*) FROM tyres_prices_stock WHERE stock = 0").Scan(&zeroStock)
	if zeroStock > 0 {
		fmt.Printf("⚠️  Позиций с нулевым остатком: %d\n", zeroStock)
	}

	// 4. Товары без бренда/модели
	var noBrandModel int
	pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM nomenclature_tyres
		WHERE brand IS NULL OR brand = '' OR model IS NULL OR model = ''
	`).Scan(&noBrandModel)
	if noBrandModel > 0 {
		fmt.Printf("⚠️  Товары без бренда/модели: %d\n", noBrandModel)
	}

	// 5. Товары без размеров
	var noSizes int
	pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM nomenclature_tyres
		WHERE width IS NULL OR height IS NULL OR diameter IS NULL
	`).Scan(&noSizes)
	if noSizes > 0 {
		fmt.Printf("⚠️  Товары без размеров: %d\n", noSizes)
	}

	// 6. Дубли по CAE (не должно быть из-за UNIQUE)
	var dupes int
	pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM (
			SELECT cae FROM nomenclature_tyres GROUP BY cae HAVING COUNT(*) > 1
		) sub
	`).Scan(&dupes)
	if dupes > 0 {
		fmt.Printf("❌ КРИТИЧНО: Дубли по CAE: %d\n", dupes)
	} else {
		fmt.Printf("✅ Дубли по CAE: %d\n", dupes)
	}

	// 7. Товары на нескольких складах
	var multiWarehouse int
	pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM (
			SELECT cae FROM tyres_prices_stock GROUP BY cae HAVING COUNT(*) > 1
		) sub
	`).Scan(&multiWarehouse)
	fmt.Printf("ℹ️  Товары на нескольких складах: %d (используем MIN цену, SUM остатков)\n", multiWarehouse)

	fmt.Println("----------------------------------------------------------------")
	fmt.Println()
}
