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
	output := flag.String("output", "/Users/viktor/Desktop/export_nomenclature_tyres.csv", "Путь к выходному CSV файлу")
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
			cae, name, width, height, diameter,
			load_index, speed_index, model, brand, season,
			is_studded, tiretype, runflat, reinforced
		FROM nomenclature_tyres
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

	// Заголовок (как в примере export_file_112522.csv)
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
	fmt.Println("Экспорт номенклатуры шин в CSV")
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
		var width, height *float64
		var diameter, loadIndex, speedIndex, model, brand, season, isStudded, tiretype, runflat, reinforced *string

		if err := rows.Scan(&cae, &name, &width, &height, &diameter,
			&loadIndex, &speedIndex, &model, &brand, &season,
			&isStudded, &tiretype, &runflat, &reinforced); err != nil {
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
		heightStr := formatFloat(height)
		diameterStr := valueOrEmpty(diameter)
		loadStr := valueOrEmpty(loadIndex)
		speedStr := valueOrEmpty(speedIndex)

		fullName := fmt.Sprintf("%s %s %s/%s %s %s%s",
			brandStr, modelStr, widthStr, heightStr, diameterStr, loadStr, speedStr)
		record[1] = fullName

		// IP_PROP140 - ширина
		record[6] = widthStr

		// IP_PROP141 - высота
		record[7] = heightStr

		// IP_PROP142 - диаметр
		record[8] = diameterStr

		// IP_PROP145 - сезон
		record[11] = valueOrEmpty(season)

		// IP_PROP146 - шипы
		if isStudded != nil {
			if *isStudded == "шип" {
				record[12] = "Шипованные"
			} else {
				record[12] = "Нешипованные"
			}
		}

		// IP_PROP149 - индекс нагрузки
		record[16] = loadStr

		// IP_PROP150 - индекс скорости
		record[17] = speedStr

		// IP_PROP151 - артикул (дублируем)
		record[18] = cae

		// IP_PROP163 - RunFlat
		record[28] = valueOrEmpty(runflat)

		// IC_GROUP0 - категория (1 = шины)
		record[58] = "1"

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
		return strconv.FormatFloat(*f, 'f', 0, 64)
	}
	return ""
}
