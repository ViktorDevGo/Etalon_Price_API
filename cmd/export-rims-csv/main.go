package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	// Флаги
	output := flag.String("output", "/Users/viktor/Desktop/export_nomenclature_rims.csv", "Путь к выходному CSV файлу")
	limit := flag.Int("limit", 10, "Количество позиций (0 = все)")
	flag.Parse()

	// Load .env
	_ = godotenv.Load()

	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("DB_DSN not set")
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	// Формируем запрос
	query := `
		SELECT
			cae, name, width, diameter, bolts_count, bolts_spacing,
			et, dia, model, brand, color, rim_type
		FROM nomenclature_rims
		WHERE cae IS NOT NULL AND cae != ''
		ORDER BY id
	`
	if *limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", *limit)
	}

	rows, err := pool.Query(context.Background(), query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	// Создаем CSV файл
	file, err := os.Create(*output)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Заголовок
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

	fmt.Println("================================================================")
	fmt.Println("Экспорт номенклатуры дисков в CSV")
	fmt.Println("================================================================")
	fmt.Printf("Выходной файл: %s\n", *output)
	if *limit > 0 {
		fmt.Printf("Лимит: %d позиций\n", *limit)
	} else {
		fmt.Println("Лимит: все позиции")
	}
	fmt.Println()

	count := 0
	for rows.Next() {
		var cae, name string
		var width, diameter, boltsSpacing, et, dia *float64
		var boltsCount *int
		var model, brand, color, rimType *string

		if err := rows.Scan(&cae, &name, &width, &diameter, &boltsCount, &boltsSpacing,
			&et, &dia, &model, &brand, &color, &rimType); err != nil {
			log.Fatal(err)
		}

		// Формируем строку CSV
		record := make([]string, 69)

		// IE_XML_ID - артикул
		record[0] = cae

		// IE_NAME - полное название
		brandStr := valueOrEmpty(brand)
		modelStr := valueOrEmpty(model)
		widthStr := formatFloat(width)
		diameterStr := formatFloat(diameter)
		boltsStr := formatInt(boltsCount)
		spacingStr := formatFloat(boltsSpacing)
		etStr := formatFloat(et)
		diaStr := formatFloat(dia)
		colorStr := valueOrEmpty(color)

		fullName := fmt.Sprintf("%s %s %sx%s %s/%s ET%s DIA%s %s",
			brandStr, modelStr, widthStr, diameterStr, boltsStr, spacingStr, etStr, diaStr, colorStr)
		record[1] = fullName

		// IP_PROP140 - ширина диска
		record[6] = widthStr

		// IP_PROP142 - диаметр диска
		record[8] = diameterStr

		// IP_PROP151 - артикул (дублируем)
		record[18] = cae

		// IP_PROP152 - количество болтов
		record[19] = boltsStr

		// IP_PROP153 - разболтовка
		record[20] = spacingStr

		// IP_PROP154 - ET (вылет)
		record[21] = etStr

		// IP_PROP155 - DIA
		record[22] = diaStr

		// IP_PROP156 - цвет
		record[23] = colorStr

		// IC_GROUP0 - категория (2 = диски)
		record[58] = "2"

		// IC_GROUP1 - бренд
		record[59] = brandStr

		// IC_GROUP2 - модель
		record[60] = modelStr

		// CP_QUANTITY - количество
		record[61] = "1000"

		// CV_CURRENCY_1 - валюта
		record[68] = "RUB"

		if err := writer.Write(record); err != nil {
			log.Fatal(err)
		}
		count++

		if count%1000 == 0 {
			fmt.Printf("Обработано: %d позиций...\n", count)
		}
	}

	fmt.Println()
	fmt.Println("================================================================")
	fmt.Printf("✅ Экспортировано %d позиций\n", count)
	fmt.Println("================================================================")
}

func valueOrEmpty(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

func formatFloat(f *float64) string {
	if f != nil {
		return strconv.FormatFloat(*f, 'f', 1, 64)
	}
	return ""
}

func formatInt(i *int) string {
	if i != nil {
		return strconv.Itoa(*i)
	}
	return ""
}
