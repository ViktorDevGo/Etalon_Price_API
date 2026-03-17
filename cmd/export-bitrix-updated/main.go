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
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	// Флаги
	outputDir := flag.String("output-dir", "/Users/viktor/Desktop", "Директория для выходных CSV файлов")
	limit := flag.Int("limit", 0, "Количество позиций (0 = все)")
	flag.Parse()

	_ = godotenv.Load()

	pool, err := pgxpool.New(context.Background(), os.Getenv("DB_DSN"))
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	ctx := context.Background()

	fmt.Println("================================================================")
	fmt.Println("Выгрузка каталога шин")
	fmt.Println("================================================================")
	fmt.Println()

	// Генерируем имя файла с текущей датой
	dateStr := time.Now().Format("20060102") // YYYYMMDD
	filename := fmt.Sprintf("Каталог_шины_%s.csv", dateStr)

	fmt.Printf("📦 Экспорт легковых шин (tiretype='Легковая')\n")
	fmt.Println()

	exported := exportFile(ctx, pool, *outputDir, filename, "Легковая", *limit)

	fmt.Println()
	fmt.Println("================================================================")
	fmt.Printf("✅ Экспортировано: %d позиций (все товары в наличии)\n", exported)
	fmt.Printf("📁 Файл: %s/%s\n", *outputDir, filename)
	fmt.Println()
	fmt.Println("✅ Файл готов для импорта в 1C-Битрикс!")
	fmt.Println("================================================================")
}

func exportFile(ctx context.Context, pool *pgxpool.Pool, outputDir, filename, tiretype string, limit int) int {
	// Запрос с фильтром по tiretype
	query := buildQuery(tiretype, limit)

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
	defer writer.Flush()

	// Заголовок (74 колонки)
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
		"CP_QUANTITY", "CATALOG_QUANTITY", "CP_QUANTITY_TRACE",
		"CP_AVAILABLE", "CP_CATALOG_TYPE",
		"CP_WEIGHT", "CP_WIDTH", "CP_HEIGHT", "CP_LENGTH",
		"CV_QUANTITY_FROM", "CV_QUANTITY_TO", "CV_PRICE_1", "CV_CURRENCY_1",
	}

	if err := writer.Write(header); err != nil {
		log.Fatal(err)
	}

	fmt.Println("✅ Логика экспорта:")
	fmt.Println("  ✓ ВСЕ товары из nomenclature_tyres (Легковая)")
	fmt.Println("  ✓ CP_QUANTITY: если остаток ≤ 4 → 0, если ≥ 5 → реальный остаток")
	fmt.Println("  ✓ CV_PRICE_1: MIN(price) * 1.12 (если нет цены → 0)")
	fmt.Println("  ✓ Цена в рублях (БД уже хранит цены в рублях)")
	fmt.Println()
	fmt.Println("Экспорт данных...")

	count := 0

	for rows.Next() {
		var (
			cae, brand, model, season, studded, loadIndex, speedIndex, runflat string
			width, height                                                       *float64
			diameter                                                            *string
			minPrice, totalStock, hasSpecialProvider                            int
		)

		if err := rows.Scan(&cae, &brand, &model, &width, &height, &diameter,
			&season, &studded, &loadIndex, &speedIndex, &runflat,
			&minPrice, &totalStock, &hasSpecialProvider); err != nil {
			log.Fatal(err)
		}

		record := make([]string, 74)

		// 1. IE_XML_ID (index 0)
		record[0] = cae

		// 2. IE_NAME (index 1)
		record[1] = formatProductName(brand, model, width, height, diameter, loadIndex, speedIndex)

		// 6. IP_PROP139 - RunFlat (index 5)
		record[5] = runflat

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

		// 11. IP_PROP144 - сезон (index 10)
		record[10] = season

		// 12. IP_PROP145 - шипы (index 11)
		record[11] = studded

		// 14. IP_PROP147 - индекс нагрузки (index 14)
		record[14] = loadIndex

		// 15. IP_PROP148 - индекс скорости (index 15)
		record[15] = speedIndex

		// 19. IP_PROP151 - артикул дубль (index 18)
		record[18] = cae

		// 22. IP_PROP154 - модель (index 21)
		record[21] = model

		// 39. IP_PROP171 - специальный поставщик (index 38)
		if hasSpecialProvider > 0 {
			record[38] = "1 день"
		}

		// 59. IC_GROUP0 - категория (index 58)
		record[58] = "1"

		// 60. IC_GROUP1 - бренд (index 59)
		record[59] = brand

		// 61. IC_GROUP2 - модель (index 60)
		record[60] = model

		// 62. CP_QUANTITY (index 61) - суммарный остаток
		// Если остаток <= 4, ставим 0, иначе реальное значение
		quantity := 0
		if totalStock >= 5 {
			quantity = totalStock
		}
		record[61] = strconv.Itoa(quantity)

		// 63. CATALOG_QUANTITY (index 62)
		record[62] = strconv.Itoa(quantity)

		// 64. CP_QUANTITY_TRACE (index 63)
		record[63] = "Y"

		// 65. CP_AVAILABLE (index 64)
		record[64] = "Y"

		// 66. CP_CATALOG_TYPE (index 65)
		record[65] = "product"

		// 67. CP_WEIGHT (index 66)
		record[66] = "0"

		// 73. CV_PRICE_1 (index 72) - минимальная цена + 12%
		// Цена в БД уже в рублях, просто добавляем наценку 12%
		priceRub := float64(minPrice)
		priceWithMarkup := priceRub * 1.12
		record[72] = fmt.Sprintf("%.2f", priceWithMarkup)

		// 74. CV_CURRENCY_1 (index 73)
		record[73] = "RUB"

		if err := writer.Write(record); err != nil {
			log.Fatal(err)
		}

		count++
	}

	return count
}

func buildQuery(tiretype string, limit int) string {
	// Всегда используем таблицы шин
	nomenclatureTable := "nomenclature_tyres"
	pricesTable := "tyres_prices_stock"

	query := fmt.Sprintf(`
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
			COALESCE(MIN(p.price), 0) as min_price,
			COALESCE(SUM(p.stock), 0) as total_stock,
			COALESCE(MAX(CASE WHEN p.provider IN ('Группа Бринекс', 'Северавто') THEN 1 ELSE 0 END), 0) as has_special_provider
		FROM %s t
		LEFT JOIN %s p ON t.cae = p.cae
		WHERE t.tiretype = '%s'
		GROUP BY t.id, t.cae, t.brand, t.model, t.width, t.height, t.diameter,
				 t.season, t.is_studded, t.load_index, t.speed_index, t.runflat
		ORDER BY t.id
	`, nomenclatureTable, pricesTable, tiretype)

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	return query
}

func formatProductName(brand, model string, width, height *float64, diameter *string, loadIndex, speedIndex string) string {
	parts := []string{}

	// Бренд
	if brand != "" {
		parts = append(parts, brand)
	}

	// Модель
	if model != "" {
		parts = append(parts, model)
	}

	// Размер: Ширина/Профиль
	sizeStr := ""
	if width != nil {
		sizeStr = strconv.FormatFloat(*width, 'f', 0, 64)
	}
	if height != nil {
		sizeStr += "/" + strconv.FormatFloat(*height, 'f', 0, 64)
	}

	// Диаметр: RДиаметр
	if diameter != nil && *diameter != "" {
		cleanDiam := normalizeDiameter(*diameter)
		if cleanDiam != "" {
			sizeStr += " R" + cleanDiam
		}
	}

	if sizeStr != "" {
		parts = append(parts, sizeStr)
	}

	// Индексы нагрузки и скорости
	if loadIndex != "" || speedIndex != "" {
		parts = append(parts, loadIndex+speedIndex)
	}

	return strings.Join(parts, " ")
}

func normalizeDiameter(s string) string {
	// Убираем префикс "R" и пробелы
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "R")
	s = strings.TrimPrefix(s, "r")

	// Заменяем запятую на точку
	s = strings.Replace(s, ",", ".", -1)

	// Проверяем, есть ли суффикс типа C
	hasSuffix := ""
	if strings.HasSuffix(s, "C") {
		hasSuffix = "C"
		s = strings.TrimSuffix(s, "C")
	}

	// Парсим число
	if num, err := strconv.ParseFloat(s, 64); err == nil {
		// Если число целое (например 17.00 или 17.0), убираем дробную часть
		if num == float64(int(num)) {
			return fmt.Sprintf("%d%s", int(num), hasSuffix)
		}
		// Иначе оставляем с одним знаком после запятой (например 22.5)
		return fmt.Sprintf("%.1f%s", num, hasSuffix)
	}

	// Если не удалось распарсить, возвращаем как есть
	return s + hasSuffix
}
