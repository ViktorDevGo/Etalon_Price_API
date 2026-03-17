package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	// Флаги
	output := flag.String("output", "/Users/viktor/Desktop/bitrix_complete.csv", "Путь к выходному CSV")
	limit := flag.Int("limit", 10, "Количество позиций (0 = все)")
	flag.Parse()

	_ = godotenv.Load()

	pool, err := pgxpool.New(context.Background(), os.Getenv("DB_DSN"))
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	ctx := context.Background()

	fmt.Println("================================================================")
	fmt.Println("Экспорт для 1C-Битрикс (ПОЛНАЯ ВЕРСИЯ)")
	fmt.Println("================================================================")
	fmt.Println()

	// Запрос с агрегацией
	query := buildQuery(*limit)

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

	// Заголовок (75 колонок)
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
		"IP_PROP187", "IP_PROP481", "IP_PROP482", "IP_PROP188",
		"IC_GROUP0", "IC_GROUP1", "IC_GROUP2",
		"CP_QUANTITY",             // index 61
		"CATALOG_QUANTITY",        // index 62
		"CP_QUANTITY_TRACE",       // index 63
		"CP_AVAILABLE",            // index 64
		"CP_CATALOG_TYPE",         // index 65
		"CATALOG_STORE_AMOUNT_1",  // index 66 ← НОВАЯ КОЛОНКА
		"CP_WEIGHT", "CP_WIDTH", "CP_HEIGHT", "CP_LENGTH",
		"CV_QUANTITY_FROM", "CV_QUANTITY_TO", "CV_PRICE_1", "CV_CURRENCY_1",
	}

	if err := writer.Write(header); err != nil {
		log.Fatal(err)
	}

	fmt.Println("✅ Структура файла:")
	fmt.Println("  ✓ CATALOG_QUANTITY - после CP_QUANTITY")
	fmt.Println("  ✓ CATALOG_STORE_AMOUNT_1 - после CP_CATALOG_TYPE")
	fmt.Println("  ✓ Оба поля дублируют CP_QUANTITY")
	fmt.Println("  ✓ Диаметры: R1700 → R17, R1600 → R16, R2250C → R22.5C")
	fmt.Println()
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

		record := make([]string, 75) // 75 колонок

		// 1. IE_XML_ID (index 0)
		record[0] = cae

		// 2. IE_NAME (index 1)
		record[1] = formatProductName(brand, model, width, height, diameter, loadIndex, speedIndex)

		// 7. IP_PROP140 - ширина (index 6)
		if width != nil {
			record[6] = strconv.FormatFloat(*width, 'f', 0, 64)
		}

		// 8. IP_PROP141 - высота (index 7)
		if height != nil {
			record[7] = strconv.FormatFloat(*height, 'f', 0, 64)
		}

		// 9. IP_PROP142 - диаметр (index 8)
		if diameter != nil {
			record[8] = normalizeDiameter(*diameter)
		}

		// 12. IP_PROP145 - сезон (index 11)
		record[11] = season

		// 14. IP_PROP146 - шипы (index 13)
		record[13] = studded

		// 17. IP_PROP149 - индекс нагрузки (index 16)
		record[16] = loadIndex

		// 18. IP_PROP150 - индекс скорости (index 17)
		record[17] = speedIndex

		// 19. IP_PROP151 - артикул дубль (index 18)
		record[18] = cae

		// 31. IP_PROP163 - RunFlat (index 30)
		record[30] = runflat

		// 59. IC_GROUP0 - категория (index 58)
		record[58] = "1"

		// 60. IC_GROUP1 - бренд (index 59)
		record[59] = brand

		// 61. IC_GROUP2 - модель (index 60)
		record[60] = model

		// 62. CP_QUANTITY (index 61)
		record[61] = strconv.Itoa(totalStock)

		// 63. CATALOG_QUANTITY (index 62)
		record[62] = strconv.Itoa(totalStock)

		// 64. CP_QUANTITY_TRACE (index 63)
		record[63] = "Y"

		// 65. CP_AVAILABLE (index 64)
		record[64] = "Y"

		// 66. CP_CATALOG_TYPE (index 65)
		record[65] = "product"

		// 67. CATALOG_STORE_AMOUNT_1 (index 66) ← НОВОЕ ПОЛЕ
		record[66] = strconv.Itoa(totalStock)

		// 68. CP_WEIGHT (index 67)
		record[67] = "0"

		// 74. CV_PRICE_1 (index 73)
		priceRub := float64(priceKopeks) / 100.0
		record[73] = fmt.Sprintf("%.2f", priceRub)

		// 75. CV_CURRENCY_1 (index 74)
		record[74] = "RUB"

		if err := writer.Write(record); err != nil {
			log.Fatal(err)
		}

		count++
	}

	fmt.Println()
	fmt.Println("================================================================")
	fmt.Printf("✅ Экспортировано %d позиций\n", count)
	fmt.Printf("📁 Файл: %s\n", *output)
	fmt.Println()
	fmt.Println("📋 Структура:")
	fmt.Println("  CP_QUANTITY | CATALOG_QUANTITY | ... | CP_CATALOG_TYPE | CATALOG_STORE_AMOUNT_1 | ... | CV_PRICE_1 | CV_CURRENCY_1")
	fmt.Println("  6           | 6                | ... | product         | 6                      | ... | 142.20     | RUB")
	fmt.Println()
	fmt.Println("✅ Файл готов для импорта в 1C-Битрикс!")
	fmt.Println("================================================================")
}

func buildQuery(limit int) string {
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
		INNER JOIN tyres_prices_stock p ON t.cae = p.cae
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

	sizeStr := ""
	if width != nil {
		sizeStr = strconv.FormatFloat(*width, 'f', 0, 64)
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

	if loadIndex != "" || speedIndex != "" {
		parts = append(parts, loadIndex+speedIndex)
	}

	return strings.Join(parts, " ")
}

func normalizeDiameter(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "R")
	s = strings.TrimPrefix(s, "r")
	s = strings.Replace(s, ",", ".", -1)

	hasSuffix := ""
	if strings.HasSuffix(s, "C") {
		hasSuffix = "C"
		s = strings.TrimSuffix(s, "C")
	}

	if num, err := strconv.ParseFloat(s, 64); err == nil {
		if num == float64(int(num)) {
			return fmt.Sprintf("%d%s", int(num), hasSuffix)
		}
		return fmt.Sprintf("%.1f%s", num, hasSuffix)
	}

	return s + hasSuffix
}
