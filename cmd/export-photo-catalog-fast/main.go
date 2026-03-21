package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

// Product from nomenclature_tyres
type Product struct {
	Brand      string
	Model      string
	Width      float64
	Height     float64
	Diameter   string
	LoadIndex  string
	SpeedIndex string
}

// SeveravtoProduct from Severavto with photo
type SeveravtoProduct struct {
	Brand      string
	Model      string
	Width      float64
	Height     float64
	Diameter   string
	LoadIndex  string
	SpeedIndex string
	PictureURL string
}

var brandMapping = map[string][]string{
	"PIRELLI":       {"ПИРЕЛЛИ", "Пирелли"},
	"MICHELIN":      {"МИШЕЛИН", "Мишелин"},
	"YOKOHAMA":      {"ЙОКОХАМА", "Йокохама"},
	"BRIDGESTONE":   {"БРИДЖСТОУН", "Бриджстоун"},
	"CONTINENTAL":   {"КОНТИНЕНТАЛЬ", "Континенталь"},
	"GOODYEAR":      {"ГУДИЕР", "Гудиер"},
	"HANKOOK":       {"КУМХО", "Кумхо", "ХАНКУК"},
	"NOKIAN":        {"НОКИАН", "Нокиан"},
	"NOKIANTYRES":   {"НОКИАН", "Нокиан"},
	"TOYO":          {"ТОЙО", "Тойо"},
	"DUNLOP":        {"ДАНЛОП", "Данлоп"},
	"KUMHO":         {"КУМХО", "Кумхо"},
	"NEXEN":         {"NEXEN", "Нексен"},
	"SAILUN":        {"SAILUN", "Сайлун"},
	"TRIANGLE":      {"ТРЕУГОЛЬНИК", "Triangle"},
	"ROADSTONE":     {"РОУДСТОУН", "Роудстоун"},
	"GISLAVED":      {"ГИСЛАВЕД", "Гиславед"},
	"CORDIANT":      {"КОРДИАНТ", "Кордиант"},
	"KORDIANT":      {"КОРДИАНТ", "Кордиант"},
	"DOUBLESTAR":    {"ДАБЛСТАР", "ДаблСтар"},
	"VIATTI":        {"VIATTI", "Виатти"},
	"LANDSAIL":      {"LANDSAIL"},
	"XTRIKE":        {"XTRIKE", "Xtrike"},
	"IKON":          {"IKONTYRES", "Ikon Tyres"},
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

	// Load all data into memory
	fmt.Println("📥 Loading nomenclature_tyres into memory...")
	products := loadProducts(ctx, pool)
	fmt.Printf("   Loaded %d products\n", len(products))

	fmt.Println("📥 Loading Severavto products with photos into memory...")
	severavtoProducts := loadSeveravtoProducts(ctx, pool)
	fmt.Printf("   Loaded %d products with photos\n", len(severavtoProducts))

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

	xmlIDIdx := findColumn(header, "IE_XML_ID")
	nameIdx := findColumn(header, "IE_NAME")

	if xmlIDIdx == -1 || nameIdx == -1 {
		log.Fatal("Required columns not found in CSV")
	}

	fmt.Printf("📊 Processing %d records...\n", len(rows))

	// Create output CSV
	outFile, err := os.Create(*outputFile)
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer outFile.Close()

	outFile.Write([]byte{0xEF, 0xBB, 0xBF})

	writer := csv.NewWriter(outFile)
	writer.Comma = ','
	defer writer.Flush()

	writer.Write([]string{"IE_XML_ID", "IE_NAME", "IP_PROP180", "IP_PROP181", "IP_PROP187", "IP_PROP146"})

	matched := 0
	notMatched := 0
	start := time.Now()

	for i, row := range rows {
		if i%5000 == 0 && i > 0 {
			elapsed := time.Since(start).Seconds()
			speed := float64(i) / elapsed
			remaining := float64(len(rows)-i) / speed
			fmt.Printf("   Progress: %d/%d (%.1f%%) | Matched: %d | Speed: %.0f rec/sec | ETA: %.0f sec\n",
				i, len(rows), float64(i)/float64(len(rows))*100, matched, speed, remaining)
		}

		xmlID := row[xmlIDIdx]
		name := row[nameIdx]

		product, exists := products[xmlID]
		if !exists {
			notMatched++
			continue
		}

		pictureURL := findPhotoInMemory(product, severavtoProducts)
		if pictureURL != "" {
			matched++
			writer.Write([]string{xmlID, name, pictureURL, pictureURL, pictureURL, pictureURL})
		} else {
			notMatched++
		}
	}

	elapsed := time.Since(start)
	fmt.Printf("\n✅ Export completed in %.1f seconds!\n", elapsed.Seconds())
	fmt.Printf("   Total records: %d\n", len(rows))
	fmt.Printf("   With photos: %d (%.1f%%)\n", matched, float64(matched)/float64(len(rows))*100)
	fmt.Printf("   Without photos: %d (%.1f%%)\n", notMatched, float64(notMatched)/float64(len(rows))*100)
	fmt.Printf("   Speed: %.0f records/second\n", float64(len(rows))/elapsed.Seconds())
	fmt.Printf("   Output file: %s\n", *outputFile)
}

func loadProducts(ctx context.Context, pool *pgxpool.Pool) map[string]Product {
	products := make(map[string]Product)

	rows, err := pool.Query(ctx, `
		SELECT cae, brand, model, width::float, height::float,
		       REGEXP_REPLACE(diameter, '[^0-9.]', '', 'g') as diameter_clean,
		       load_index, speed_index
		FROM nomenclature_tyres
	`)
	if err != nil {
		log.Fatalf("Failed to load products: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var cae string
		var p Product
		rows.Scan(&cae, &p.Brand, &p.Model, &p.Width, &p.Height, &p.Diameter, &p.LoadIndex, &p.SpeedIndex)
		products[cae] = p
	}

	return products
}

func loadSeveravtoProducts(ctx context.Context, pool *pgxpool.Pool) []SeveravtoProduct {
	var products []SeveravtoProduct

	rows, err := pool.Query(ctx, `
		SELECT nts.brand, nts.model, nts.width::float, nts.height::float,
		       nts.diameter, nts.load_index, nts.speed_index,
		       ft.picture_url
		FROM nomenclature_tyres_severavto nts
		INNER JOIN foto_tyres ft ON nts.commodity_id = ft.commodity_id
		WHERE ft.picture_url IS NOT NULL AND ft.picture_url != ''
	`)
	if err != nil {
		log.Fatalf("Failed to load Severavto products: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var p SeveravtoProduct
		rows.Scan(&p.Brand, &p.Model, &p.Width, &p.Height, &p.Diameter, &p.LoadIndex, &p.SpeedIndex, &p.PictureURL)
		products = append(products, p)
	}

	return products
}

func findPhotoInMemory(product Product, severavtoProducts []SeveravtoProduct) string {
	brandKey := normalizeBrand(product.Brand)
	russianBrands := brandMapping[brandKey]
	if len(russianBrands) == 0 {
		russianBrands = []string{product.Brand}
	}

	modelNorm := normalizeModel(product.Model)

	// Priority 1: Exact match (brand + model + all sizes + indices)
	for _, p := range severavtoProducts {
		if matchesBrand(p.Brand, russianBrands, product.Brand) &&
			matchesModel(p.Model, modelNorm) &&
			p.Width == product.Width &&
			p.Height == product.Height &&
			p.Diameter == product.Diameter &&
			p.LoadIndex == product.LoadIndex &&
			p.SpeedIndex == product.SpeedIndex {
			return p.PictureURL
		}
	}

	// Priority 2: Brand + model + width + indices
	for _, p := range severavtoProducts {
		if matchesBrand(p.Brand, russianBrands, product.Brand) &&
			matchesModel(p.Model, modelNorm) &&
			p.Width == product.Width &&
			p.LoadIndex == product.LoadIndex &&
			p.SpeedIndex == product.SpeedIndex {
			return p.PictureURL
		}
	}

	// Priority 3: Brand + width + indices
	for _, p := range severavtoProducts {
		if matchesBrand(p.Brand, russianBrands, product.Brand) &&
			p.Width == product.Width &&
			p.LoadIndex == product.LoadIndex &&
			p.SpeedIndex == product.SpeedIndex {
			return p.PictureURL
		}
	}

	// Priority 4: Brand + all sizes
	for _, p := range severavtoProducts {
		if matchesBrand(p.Brand, russianBrands, product.Brand) &&
			p.Width == product.Width &&
			p.Height == product.Height &&
			p.Diameter == product.Diameter {
			return p.PictureURL
		}
	}

	return ""
}

func matchesBrand(severavtoBrand string, russianBrands []string, originalBrand string) bool {
	severavtoNorm := strings.ToUpper(strings.TrimSpace(severavtoBrand))

	for _, ruBrand := range russianBrands {
		if severavtoNorm == strings.ToUpper(strings.TrimSpace(ruBrand)) {
			return true
		}
	}

	// Try original brand
	if severavtoNorm == strings.ToUpper(strings.TrimSpace(originalBrand)) {
		return true
	}

	return false
}

func matchesModel(severavtoModel, normalizedModel string) bool {
	severavtoNorm := normalizeModel(severavtoModel)
	return strings.Contains(severavtoNorm, normalizedModel) || strings.Contains(normalizedModel, severavtoNorm)
}

func normalizeBrand(brand string) string {
	s := strings.ToUpper(strings.TrimSpace(brand))
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "-", "")
	return s
}

func normalizeModel(model string) string {
	s := strings.ToUpper(strings.TrimSpace(model))
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "-", "")
	// Remove special characters
	reg := regexp.MustCompile(`[^A-Z0-9]`)
	s = reg.ReplaceAllString(s, "")
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
