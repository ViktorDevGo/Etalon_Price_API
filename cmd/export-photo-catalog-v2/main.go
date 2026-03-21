package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

// Словарь соответствий брендов (английский → русский)
var brandMapping = map[string][]string{
	"PIRELLI":       {"ПИРЕЛЛИ", "Пирелли"},
	"MICHELIN":      {"МИШЕЛИН", "Мишелин"},
	"YOKOHAMA":      {"ЙОКОХАМА", "Йокохама"},
	"BRIDGESTONE":   {"БРИДЖСТОУН", "Бриджстоун"},
	"CONTINENTAL":   {"КОНТИНЕНТАЛЬ", "Континенталь"},
	"GOODYEAR":      {"ГУДИЕР", "Гудиер"},
	"HANKOOK":       {"КУМХО", "Кумхо", "ХАНКУК", "Хан Кук"},
	"NOKIAN":        {"НОКИАН", "Нокиан"},
	"NOKIANTYRES":   {"НОКИАН", "Нокиан"},
	"TOYO":          {"ТОЙО", "Тойо"},
	"DUNLOP":        {"ДАНЛОП", "Данлоп"},
	"KUMHO":         {"КУМХО", "Кумхо"},
	"NEXEN":         {"NEXEN", "Нексен"},
	"SAILUN":        {"SAILUN", "Сайлун"},
	"TRIANGLE":      {"ТРЕУГОЛЬНИК", "Triangle"},
	"MARSHAL":       {"МАРШАЛ", "Маршал"},
	"ROADSTONE":     {"РОУДСТОУН", "Роудстоун"},
	"GISLAVED":      {"ГИСЛАВЕД", "Гиславед"},
	"CORDIANT":      {"КОРДИАНТ", "Кордиант"},
	"KORDIANT":      {"КОРДИАНТ", "Кордиант"},
	"DOUBLESTAR":    {"ДАБЛСТАР", "ДаблСтар"},
	"VIATTI":        {"VIATTI", "Виатти"},
	"LANDSAIL":      {"LANDSAIL"},
	"XTRIKE":        {"XTRIKE", "Xtrike"},
	"IKON":          {"IKONTYRES", "Ikon Tyres"},
	"GENERALTIRETIRE": {"ДЖЕНЕРАЛ"},
	"MATADOR":       {"МАТАДОР"},
	"NITTO":         {"НИТТО"},
	"BFGOODRICH":    {"БФ ГУДРИЧ"},
	"COOPER":        {"КУПЕР"},
	"FALKEN":        {"ФАЛКЕН"},
	"FIRESTONE":     {"ФАЙРСТОУН"},
	"GOODRIDE":      {"ГУДРАЙД"},
	"HIFLY":         {"ХАЙФЛАЙ"},
	"FORTUNE":       {"ФОРТУНА"},
	"TRACMAX":       {"ТРАКМАКС"},
	"ANTARES":       {"АНТАРЕС"},
}

func main() {
	inputFile := flag.String("input", "/Users/viktor/Desktop/Каталог_шины_20260316.csv", "Input catalog CSV")
	outputFile := flag.String("output", "/Users/viktor/Desktop/Фото_NEW.csv", "Output photo CSV")
	flag.Parse()

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

	fmt.Println("✅ Connected to database")

	// Read input CSV
	fmt.Printf("📥 Reading input file: %s\n", *inputFile)
	inFile, err := os.Open(*inputFile)
	if err != nil {
		log.Fatalf("Failed to open input file: %v", err)
	}
	defer inFile.Close()

	reader := csv.NewReader(inFile)
	reader.Comma = ','
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Failed to read CSV: %v", err)
	}

	if len(records) < 2 {
		log.Fatal("CSV file is empty")
	}

	header := records[0]
	rows := records[1:]

	// Find column indices
	xmlIDIdx := findColumn(header, "IE_XML_ID")
	nameIdx := findColumn(header, "IE_NAME")

	if xmlIDIdx == -1 || nameIdx == -1 {
		log.Fatal("Required columns not found in CSV")
	}

	fmt.Printf("📊 Found %d records in catalog\n", len(rows))

	// Create output CSV
	outFile, err := os.Create(*outputFile)
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer outFile.Close()

	// Write UTF-8 BOM
	outFile.Write([]byte{0xEF, 0xBB, 0xBF})

	writer := csv.NewWriter(outFile)
	writer.Comma = ','
	defer writer.Flush()

	// Write header
	writer.Write([]string{"IE_XML_ID", "IE_NAME", "IP_PROP180", "IP_PROP181", "IP_PROP187"})

	// Process each record
	matched := 0
	notMatched := 0

	for i, row := range rows {
		if i%1000 == 0 {
			fmt.Printf("   Processing: %d/%d (matched: %d)\r", i, len(rows), matched)
		}

		xmlID := row[xmlIDIdx]
		name := row[nameIdx]

		// Get product details from nomenclature_tyres
		var brand, model string
		var width, height float64
		var diameter, loadIndex, speedIndex string

		err := pool.QueryRow(ctx, `
			SELECT brand, model, width::float, height::float,
				   REGEXP_REPLACE(diameter, '[^0-9.]', '', 'g') as diameter_clean,
				   load_index, speed_index
			FROM nomenclature_tyres
			WHERE cae = $1
		`, xmlID).Scan(&brand, &model, &width, &height, &diameter, &loadIndex, &speedIndex)

		if err != nil {
			// No match in nomenclature_tyres, skip
			notMatched++
			continue
		}

		// Try to find match in Severavto data
		pictureURL := findPhotoURL(ctx, pool, brand, model, width, height, diameter, loadIndex, speedIndex)

		if pictureURL != "" {
			// Write only if photo found
			matched++
			writer.Write([]string{xmlID, name, pictureURL, pictureURL, pictureURL})
		} else {
			// Skip records without photo
			notMatched++
		}
	}

	fmt.Printf("\n✅ Export completed!\n")
	fmt.Printf("   Total records: %d\n", len(rows))
	fmt.Printf("   With photos: %d (%.1f%%)\n", matched, float64(matched)/float64(len(rows))*100)
	fmt.Printf("   Without photos: %d (%.1f%%)\n", notMatched, float64(notMatched)/float64(len(rows))*100)
	fmt.Printf("   Output file: %s\n", *outputFile)
}

func findPhotoURL(ctx context.Context, pool *pgxpool.Pool, brand, model string, width, height float64, diameter, loadIndex, speedIndex string) string {
	// Get Russian brand variants from mapping
	brandKey := strings.ToUpper(strings.TrimSpace(brand))
	brandKey = strings.ReplaceAll(brandKey, " ", "")

	russianBrands := brandMapping[brandKey]
	if len(russianBrands) == 0 {
		russianBrands = []string{brand}
	}

	// Normalize model for fuzzy matching
	modelNorm := strings.ToUpper(strings.TrimSpace(model))
	modelNorm = strings.ReplaceAll(modelNorm, " ", "")
	modelNorm = strings.ReplaceAll(modelNorm, "-", "")

	// Priority 1: Exact match (brand + model + all sizes + indices)
	for _, ruBrand := range russianBrands {
		var url string
		err := pool.QueryRow(ctx, `
			SELECT ft.picture_url
			FROM nomenclature_tyres_severavto nts
			INNER JOIN foto_tyres ft ON nts.commodity_id = ft.commodity_id
			WHERE UPPER(TRIM(nts.brand)) = UPPER(TRIM($1))
				AND UPPER(REPLACE(REPLACE(nts.model, ' ', ''), '-', '')) LIKE '%' || $2 || '%'
				AND nts.width::float = $3
				AND nts.height::float = $4
				AND nts.diameter = $5
				AND nts.load_index = $6
				AND nts.speed_index = $7
				AND ft.picture_url IS NOT NULL
				AND ft.picture_url != ''
			LIMIT 1
		`, ruBrand, modelNorm, width, height, diameter, loadIndex, speedIndex).Scan(&url)

		if err == nil && url != "" {
			return url
		}
	}

	// Priority 2: Brand + model + width + indices (ignore height/diameter)
	for _, ruBrand := range russianBrands {
		var url string
		err := pool.QueryRow(ctx, `
			SELECT ft.picture_url
			FROM nomenclature_tyres_severavto nts
			INNER JOIN foto_tyres ft ON nts.commodity_id = ft.commodity_id
			WHERE UPPER(TRIM(nts.brand)) = UPPER(TRIM($1))
				AND UPPER(REPLACE(REPLACE(nts.model, ' ', ''), '-', '')) LIKE '%' || $2 || '%'
				AND nts.width::float = $3
				AND nts.load_index = $4
				AND nts.speed_index = $5
				AND ft.picture_url IS NOT NULL
				AND ft.picture_url != ''
			LIMIT 1
		`, ruBrand, modelNorm, width, loadIndex, speedIndex).Scan(&url)

		if err == nil && url != "" {
			return url
		}
	}

	// Priority 3: Brand + width + indices (ignore model)
	for _, ruBrand := range russianBrands {
		var url string
		err := pool.QueryRow(ctx, `
			SELECT ft.picture_url
			FROM nomenclature_tyres_severavto nts
			INNER JOIN foto_tyres ft ON nts.commodity_id = ft.commodity_id
			WHERE UPPER(TRIM(nts.brand)) = UPPER(TRIM($1))
				AND nts.width::float = $2
				AND nts.load_index = $3
				AND nts.speed_index = $4
				AND ft.picture_url IS NOT NULL
				AND ft.picture_url != ''
			LIMIT 1
		`, ruBrand, width, loadIndex, speedIndex).Scan(&url)

		if err == nil && url != "" {
			return url
		}
	}

	// Priority 4: Brand + all sizes (old logic, fallback)
	for _, ruBrand := range russianBrands {
		var url string
		err := pool.QueryRow(ctx, `
			SELECT ft.picture_url
			FROM nomenclature_tyres_severavto nts
			INNER JOIN foto_tyres ft ON nts.commodity_id = ft.commodity_id
			WHERE UPPER(TRIM(nts.brand)) = UPPER(TRIM($1))
				AND nts.width::float = $2
				AND nts.height::float = $3
				AND nts.diameter = $4
				AND ft.picture_url IS NOT NULL
				AND ft.picture_url != ''
			LIMIT 1
		`, ruBrand, width, height, diameter).Scan(&url)

		if err == nil && url != "" {
			return url
		}
	}

	// Priority 5: Try original brand (for NEXEN, SAILUN, etc)
	var url string
	pool.QueryRow(ctx, `
		SELECT ft.picture_url
		FROM nomenclature_tyres_severavto nts
		INNER JOIN foto_tyres ft ON nts.commodity_id = ft.commodity_id
		WHERE UPPER(TRIM(nts.brand)) = UPPER(TRIM($1))
			AND nts.width::float = $2
			AND nts.load_index = $3
			AND nts.speed_index = $4
			AND ft.picture_url IS NOT NULL
			AND ft.picture_url != ''
		LIMIT 1
	`, brand, width, loadIndex, speedIndex).Scan(&url)

	return url
}

func findColumn(header []string, name string) int {
	for i, col := range header {
		if col == name {
			return i
		}
	}
	return -1
}
