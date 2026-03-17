package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
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

	pool, err := pgxpool.New(context.Background(), os.Getenv("DB_DSN"))
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	ctx := context.Background()

	fmt.Println("================================================================")
	fmt.Println("Выгрузка каталога дисков для 1C-Битрикс")
	fmt.Println("================================================================")
	fmt.Println()

	// Генерируем имя файла с текущей датой
	dateStr := time.Now().Format("20060102") // YYYYMMDD
	filename := fmt.Sprintf("Каталог_диски_%s.csv", dateStr)
	if *limit > 0 {
		filename = fmt.Sprintf("Каталог_диски_%s_TEST_%d.csv", dateStr, *limit)
	}

	fmt.Printf("📦 Экспорт каталога дисков\n")
	if *limit > 0 {
		fmt.Printf("⚠️  ТЕСТОВЫЙ РЕЖИМ: ограничение %d строк\n", *limit)
	}
	fmt.Println()

	exported := exportRims(ctx, pool, *outputDir, filename, *limit)

	fmt.Println()
	fmt.Println("================================================================")
	fmt.Printf("✅ Экспортировано: %d позиций\n", exported)
	fmt.Printf("📁 Файл: %s/%s\n", *outputDir, filename)
	fmt.Println()
	fmt.Println("✅ Файл готов для импорта в 1C-Битрикс!")
	fmt.Println("================================================================")
}

func exportRims(ctx context.Context, pool *pgxpool.Pool, outputDir, filename string, limit int) int {
	// Список разрешенных брендов
	allowedBrands := []string{
		"Advan", "Advanti", "Aero", "Alcasta", "Alutec", "Asterro", "BBS",
		"Better", "Black Rhino", "Buffalo", "COX", "Carwel", "CrossStreet",
		"Enkei", "FF", "FM", "FR design", "FR replica", "Fuel", "Inforget",
		"K&K", "Khomen Wheels", "KoKo", "Konig", "Kosei", "Kronprinz/Accuride",
		"LS", "LS FlowForming", "LS Forged", "LegeArtis", "LegeArtis Concept",
		"Lenso", "Lizardo", "MAK", "Magnetto", "Mefro", "Megami", "Momo",
		"N2O", "NZ", "Neo", "Next", "OZ", "Off-Road Wheels", "PDW",
		"Premium Series", "RST", "Race Ready", "Remain", "Replay",
		"SKAD Original", "SRW", "Sakura", "Tech Line", "Trebl", "Venti",
		"Vossen", "X-Race", "X`trike", "YST", "Yamato", "Yamato Samurai",
		"iFree", "iFree Original", "Евродиск", "СКАД", "ТЗСК",
	}

	// SQL запрос для получения дисков с ценами и остатками
	limitClause := ""
	if limit > 0 {
		limitClause = fmt.Sprintf("LIMIT %d", limit)
	}

	// Формируем список брендов для SQL IN clause
	brandList := "("
	for i, brand := range allowedBrands {
		if i > 0 {
			brandList += ","
		}
		brandList += fmt.Sprintf("'%s'", brand)
	}
	brandList += ")"

	query := fmt.Sprintf(`
		WITH rim_aggregates AS (
			SELECT
				cae,
				SUM(stock) as total_stock,
				MIN(price) as min_price
			FROM rims_prices_stock
			GROUP BY cae
		)
		SELECT
			n.cae,
			n.name,
			n.width,
			n.diameter,
			n.bolts_count,
			n.bolts_spacing,
			n.bolts_spacing2,
			n.et,
			n.dia,
			n.model,
			n.brand,
			n.color,
			CASE
				WHEN COALESCE(p.total_stock, 0) <= 4 THEN 0
				ELSE COALESCE(p.total_stock, 0)
			END as display_stock,
			COALESCE(p.min_price, 0) as min_price
		FROM nomenclature_rims n
		LEFT JOIN rim_aggregates p ON n.cae = p.cae
		WHERE n.brand IN %s
		ORDER BY n.cae
		%s
	`, brandList, limitClause)

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

	// Заголовок (64 колонки как в примере)
	header := []string{
		"IE_XML_ID", "IE_NAME", "IE_PREVIEW_TEXT", "IE_DETAIL_TEXT",
		"IP_PROP281", "IP_PROP282", "IP_PROP283", "IP_PROP284", "IP_PROP285",
		"IP_PROP286", "IP_PROP287", "IP_PROP288", "IP_PROP289", "IP_PROP290",
		"IP_PROP291", "IP_PROP292", "IP_PROP293", "IP_PROP294", "IP_PROP295",
		"IP_PROP296", "IP_PROP297", "IP_PROP298", "IP_PROP299", "IP_PROP300",
		"IP_PROP301", "IP_PROP302", "IP_PROP303", "IP_PROP304", "IP_PROP305",
		"IP_PROP306", "IP_PROP307", "IP_PROP308", "IP_PROP309", "IP_PROP310",
		"IP_PROP311", "IP_PROP312", "IP_PROP313", "IP_PROP314", "IP_PROP315",
		"IP_PROP316", "IP_PROP317", "IP_PROP318", "IP_PROP319", "IP_PROP320",
		"IP_PROP321", "IP_PROP322", "IP_PROP323", "IP_PROP324",
		"IP_PROP479", "IP_PROP480",
		"IP_PROP325", "IP_PROP326",
		"IC_GROUP0", "IC_GROUP1", "IC_GROUP2",
		"CP_QUANTITY", "CP_WEIGHT", "CP_WIDTH", "CP_HEIGHT", "CP_LENGTH",
		"CV_QUANTITY_FROM", "CV_QUANTITY_TO", "CV_PRICE_1", "CV_CURRENCY_1",
	}

	if err := writer.Write(header); err != nil {
		log.Fatal(err)
	}

	fmt.Println("✅ Логика экспорта:")
	fmt.Println("  ✓ Источник: nomenclature_rims + rims_prices_stock")
	fmt.Println("  ✓ Фильтр: только разрешенные бренды (68 брендов)")
	fmt.Println("  ✓ CP_QUANTITY: SUM(stock) по CAE, если ≤ 4 → 0, иначе факт")
	fmt.Println("  ✓ CV_PRICE_1: MIN(price) по CAE")
	fmt.Println()
	fmt.Println("Экспорт данных...")

	count := 0

	for rows.Next() {
		var (
			cae                                                       string
			name                                                      string
			width, diameter, boltSpacing, boltSpacing2, et, dia       *float64
			boltsCount                                                *int
			model, brand, color                                       *string
			totalStock, minPrice                                      int
		)

		if err := rows.Scan(
			&cae, &name, &width, &diameter, &boltsCount, &boltSpacing,
			&boltSpacing2, &et, &dia, &model, &brand, &color,
			&totalStock, &minPrice,
		); err != nil {
			log.Fatal(err)
		}

		record := make([]string, 64)

		// 1. IE_XML_ID (CAE)
		record[0] = cae

		// 2. IE_NAME (Название)
		record[1] = name

		// 3-4. IE_PREVIEW_TEXT, IE_DETAIL_TEXT (пусто)
		record[2] = ""
		record[3] = ""

		// 5. IP_PROP281 (Метка - пусто)
		record[4] = ""

		// 6. IP_PROP282 (Ширина)
		if width != nil {
			record[5] = fmt.Sprintf("%.1f", *width)
		}

		// 7. IP_PROP283 (Диаметр)
		if diameter != nil {
			record[6] = fmt.Sprintf("%.0f", *diameter)
		}

		// 8. IP_PROP284 (Количество болтов)
		if boltsCount != nil {
			record[7] = fmt.Sprintf("%d", *boltsCount)
		}

		// 9. IP_PROP285 (Разболтовка)
		if boltSpacing != nil {
			record[8] = fmt.Sprintf("%.1f", *boltSpacing)
		}

		// 10. IP_PROP286 (ET - Вылет)
		if et != nil {
			record[9] = fmt.Sprintf("%.0f", *et)
		}

		// 11. IP_PROP287 (DIA)
		if dia != nil {
			record[10] = fmt.Sprintf("%.1f", *dia)
		}

		// 12-25. IP_PROP288-301 (Пустые колонки)
		for i := 11; i < 25; i++ {
			record[i] = ""
		}

		// 26. IP_PROP302 (CAE - дублируем)
		record[25] = cae

		// 27-50. IP_PROP303-324 (Пустые колонки)
		for i := 26; i < 50; i++ {
			record[i] = ""
		}

		// 51. IP_PROP325 (Бренд)
		if brand != nil {
			record[50] = *brand
		}

		// 52. IP_PROP326 (Цвет)
		if color != nil {
			record[51] = *color
		}

		// 53. IC_GROUP0 (Бренд)
		if brand != nil {
			record[52] = *brand
		}

		// 54. IC_GROUP1 (Модель)
		if model != nil {
			record[53] = *model
		}

		// 55. IC_GROUP2 (пусто)
		record[54] = ""

		// 56. CP_QUANTITY (Суммарный остаток)
		record[55] = fmt.Sprintf("%d", totalStock)

		// 57. CP_WEIGHT (не заполняем)
		record[56] = ""

		// 58-60. CP_WIDTH, CP_HEIGHT, CP_LENGTH (не заполняем)
		record[57] = ""
		record[58] = ""
		record[59] = ""

		// 61-62. CV_QUANTITY_FROM, CV_QUANTITY_TO (пусто)
		record[60] = ""
		record[61] = ""

		// 63. CV_PRICE_1 (Минимальная цена)
		record[62] = fmt.Sprintf("%d", minPrice)

		// 64. CV_CURRENCY_1 (Валюта)
		record[63] = "RUB"

		if err := writer.Write(record); err != nil {
			log.Fatal(err)
		}

		count++
	}

	return count
}
