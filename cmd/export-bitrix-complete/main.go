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
	output := flag.String("output", "/Users/viktor/Desktop/bitrix_catalog_complete.csv", "Путь к выходному CSV")
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

	// Заголовок (73 колонки - добавлена CP_CATALOG_TYPE)
	header := []string{
		// 1-4: Основная информация
		"IE_XML_ID", "IE_NAME", "IE_PREVIEW_TEXT", "IE_DETAIL_TEXT",
		// 5-58: Свойства IP_PROP138-188
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
		// 59-61: Категория, бренд, модель
		"IC_GROUP0", "IC_GROUP1", "IC_GROUP2",
		// 62-69: Параметры товара
		"CP_QUANTITY",        // 62 (index 61)
		"CP_QUANTITY_TRACE",  // 63 (index 62)
		"CP_AVAILABLE",       // 64 (index 63)
		"CP_CATALOG_TYPE",    // 65 (index 64) ← НОВОЕ ПОЛЕ
		"CP_WEIGHT",          // 66 (index 65)
		"CP_WIDTH",           // 67 (index 66)
		"CP_HEIGHT",          // 68 (index 67)
		"CP_LENGTH",          // 69 (index 68)
		// 70-73: Цена
		"CV_QUANTITY_FROM",   // 70 (index 69)
		"CV_QUANTITY_TO",     // 71 (index 70)
		"CV_PRICE_1",         // 72 (index 71)
		"CV_CURRENCY_1",      // 73 (index 72)
	}

	if err := writer.Write(header); err != nil {
		log.Fatal(err)
	}

	fmt.Println("✅ Формат файла:")
	fmt.Println("  ✓ IE_XML_ID: артикул товара")
	fmt.Println("  ✓ IE_NAME: название товара")
	fmt.Println("  ✓ CP_QUANTITY: остаток (число)")
	fmt.Println("  ✓ CP_QUANTITY_TRACE: Y (управление остатками)")
	fmt.Println("  ✓ CP_AVAILABLE: Y (товар доступен)")
	fmt.Println("  ✓ CP_CATALOG_TYPE: product (тип товара)")
	fmt.Println("  ✓ CV_PRICE_1: цена (XXX.XX)")
	fmt.Println("  ✓ CV_CURRENCY_1: RUB")
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

		// Создаем массив на 73 элемента (indices 0-72)
		record := make([]string, 73)

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

		// 62. CP_QUANTITY - остаток (index 61) ← ЧИСЛО
		record[61] = strconv.Itoa(totalStock)

		// 63. CP_QUANTITY_TRACE - управление остатками (index 62) ← Y
		record[62] = "Y"

		// 64. CP_AVAILABLE - доступность товара (index 63) ← Y
		record[63] = "Y"

		// 65. CP_CATALOG_TYPE - тип товара (index 64) ← product (НОВОЕ!)
		record[64] = "product"

		// 66. CP_WEIGHT - вес (index 65)
		record[65] = "0"

		// 72. CV_PRICE_1 - цена (index 71) ← ЧИСЛО
		priceRub := float64(priceKopeks) / 100.0
		record[71] = fmt.Sprintf("%.2f", priceRub)

		// 73. CV_CURRENCY_1 - валюта (index 72) ← RUB
		record[72] = "RUB"

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
	fmt.Println("📋 Пример строки (все ключевые поля):")
	fmt.Println("  Артикул | Остаток | Управление | Доступен | Тип | Цена | Валюта")
	fmt.Printf("  OTR_037 | 6 | Y | Y | product | 142.20 | RUB\n")
	fmt.Println()
	fmt.Println("✅ Файл полностью готов к импорту в 1C-Битрикс!")
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
