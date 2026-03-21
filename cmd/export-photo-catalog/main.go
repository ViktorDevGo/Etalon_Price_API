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

func main() {
	inputFile := flag.String("input", "/Users/viktor/Desktop/Каталог_шины_20260316.csv", "Input catalog CSV")
	outputFile := flag.String("output", "/Users/viktor/Desktop/Фото.csv", "Output photo CSV")
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
	writer.Comma = ';'
	defer writer.Flush()

	// Write header
	writer.Write([]string{"IE_XML_ID", "IE_NAME", "IP_PROP190"})

	// Process each record
	matched := 0
	notMatched := 0

	for i, row := range rows {
		if i%1000 == 0 {
			fmt.Printf("   Processing: %d/%d\r", i, len(rows))
		}

		xmlID := row[xmlIDIdx]
		name := row[nameIdx]

		// Get product details from nomenclature_tyres
		var brand, model string
		var width, height float64
		var diameter string

		err := pool.QueryRow(ctx, `
			SELECT brand, model, width::float, height::float, diameter
			FROM nomenclature_tyres
			WHERE cae = $1
		`, xmlID).Scan(&brand, &model, &width, &height, &diameter)

		if err != nil {
			// No match in nomenclature_tyres, skip
			writer.Write([]string{xmlID, name, ""})
			notMatched++
			continue
		}

		// Try to find match in Severavto data
		pictureURL := findPhotoURL(ctx, pool, brand, model, width, height, diameter)

		if pictureURL != "" {
			matched++
		} else {
			notMatched++
		}

		writer.Write([]string{xmlID, name, pictureURL})
	}

	fmt.Printf("\n✅ Export completed!\n")
	fmt.Printf("   Total records: %d\n", len(rows))
	fmt.Printf("   With photos: %d (%.1f%%)\n", matched, float64(matched)/float64(len(rows))*100)
	fmt.Printf("   Without photos: %d (%.1f%%)\n", notMatched, float64(notMatched)/float64(len(rows))*100)
	fmt.Printf("   Output file: %s\n", *outputFile)
}

func findPhotoURL(ctx context.Context, pool *pgxpool.Pool, brand, model string, width, height float64, diameter string) string {
	// Strategy 1: Exact match (brand, model, sizes)
	var url string
	err := pool.QueryRow(ctx, `
		SELECT ft.picture_url
		FROM nomenclature_tyres_severavto nts
		INNER JOIN foto_tyres ft ON nts.commodity_id = ft.commodity_id
		WHERE UPPER(TRIM(nts.brand)) = UPPER(TRIM($1))
			AND UPPER(TRIM(nts.model)) = UPPER(TRIM($2))
			AND nts.width::float = $3
			AND nts.height::float = $4
			AND nts.diameter = $5
			AND ft.picture_url IS NOT NULL
			AND ft.picture_url != ''
		LIMIT 1
	`, brand, model, width, height, diameter).Scan(&url)

	if err == nil && url != "" {
		return url
	}

	// Strategy 2: Match by brand and sizes only (ignore model differences)
	err = pool.QueryRow(ctx, `
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
	`, brand, width, height, diameter).Scan(&url)

	if err == nil && url != "" {
		return url
	}

	// Strategy 3: Fuzzy brand match + exact sizes
	normalizedBrand := normalizeBrand(brand)
	err = pool.QueryRow(ctx, `
		SELECT ft.picture_url
		FROM nomenclature_tyres_severavto nts
		INNER JOIN foto_tyres ft ON nts.commodity_id = ft.commodity_id
		WHERE nts.width::float = $1
			AND nts.height::float = $2
			AND nts.diameter = $3
			AND (
				UPPER(REPLACE(REPLACE(TRIM(nts.brand), ' ', ''), '-', '')) = $4
				OR UPPER(TRIM(nts.brand)) LIKE $5
			)
			AND ft.picture_url IS NOT NULL
			AND ft.picture_url != ''
		LIMIT 1
	`, width, height, diameter, normalizedBrand, "%"+strings.ToUpper(brand[:min(5, len(brand))])+"%").Scan(&url)

	if err == nil && url != "" {
		return url
	}

	return ""
}

func normalizeBrand(brand string) string {
	s := strings.ToUpper(strings.TrimSpace(brand))
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, "_", "")
	return s
}

func findColumn(header []string, name string) int {
	for i, col := range header {
		if col == name {
			return i
		}
	}
	return -1
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
