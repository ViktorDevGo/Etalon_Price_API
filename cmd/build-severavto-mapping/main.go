package main

import (
	"context"
	"encoding/json"
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
	CAE        string
	Brand      string
	Model      string
	Width      float64
	Height     float64
	Diameter   string
	LoadIndex  string
	SpeedIndex string
}

// SeveravtoProduct from Severavto
type SeveravtoProduct struct {
	CommodityID string
	Brand       string
	Model       string
	Width       float64
	Height      float64
	Diameter    string
	LoadIndex   string
	SpeedIndex  string
}

type MappingResult struct {
	CAE         string
	CommodityID string
	Priority    int
	Details     map[string]interface{}
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

	fmt.Println("📥 Loading Severavto products into memory...")
	severavtoProducts := loadSeveravtoProducts(ctx, pool)
	fmt.Printf("   Loaded %d products\n", len(severavtoProducts))

	// Build mapping
	fmt.Println("🔄 Building CAE ↔ commodity_id mapping...")
	start := time.Now()

	mappings := make([]MappingResult, 0)
	matched := 0
	notMatched := 0

	for i, product := range products {
		if i%5000 == 0 && i > 0 {
			elapsed := time.Since(start).Seconds()
			speed := float64(i) / elapsed
			remaining := float64(len(products)-i) / speed
			fmt.Printf("   Progress: %d/%d (%.1f%%) | Matched: %d | Speed: %.0f rec/sec | ETA: %.0f sec\n",
				i, len(products), float64(i)/float64(len(products))*100, matched, speed, remaining)
		}

		commodityID, priority := findBestMatch(product, severavtoProducts)
		if commodityID != "" {
			matched++
			mappings = append(mappings, MappingResult{
				CAE:         product.CAE,
				CommodityID: commodityID,
				Priority:    priority,
				Details: map[string]interface{}{
					"brand":       product.Brand,
					"model":       product.Model,
					"width":       product.Width,
					"height":      product.Height,
					"diameter":    product.Diameter,
					"load_index":  product.LoadIndex,
					"speed_index": product.SpeedIndex,
				},
			})
		} else {
			notMatched++
		}
	}

	elapsed := time.Since(start)
	fmt.Printf("\n✅ Mapping completed in %.1f seconds!\n", elapsed.Seconds())
	fmt.Printf("   Total products: %d\n", len(products))
	fmt.Printf("   Matched: %d (%.1f%%)\n", matched, float64(matched)/float64(len(products))*100)
	fmt.Printf("   Not matched: %d (%.1f%%)\n", notMatched, float64(notMatched)/float64(len(products))*100)

	// Save to database
	fmt.Println("\n💾 Saving mapping to database...")
	// Reconnect to avoid connection timeout after long processing
	pool.Close()
	pool, err = pgxpool.New(ctx, dbDSN)
	if err != nil {
		log.Fatalf("Failed to reconnect to database: %v", err)
	}
	defer pool.Close()
	saveMappings(ctx, pool, mappings)

	// Statistics by priority
	fmt.Println("\n📊 Mapping statistics by priority:")
	stats := make(map[int]int)
	for _, m := range mappings {
		stats[m.Priority]++
	}
	for priority := 1; priority <= 4; priority++ {
		count := stats[priority]
		if count > 0 {
			fmt.Printf("   Priority %d: %d mappings (%.1f%%)\n",
				priority, count, float64(count)/float64(matched)*100)
		}
	}

	fmt.Println("\n✅ Done!")
}

func loadProducts(ctx context.Context, pool *pgxpool.Pool) []Product {
	var products []Product

	rows, err := pool.Query(ctx, `
		SELECT cae, brand, model, width::float, height::float,
		       REGEXP_REPLACE(diameter, '[^0-9.]', '', 'g') as diameter_clean,
		       load_index, speed_index
		FROM nomenclature_tyres
		WHERE tiretype = 'Легковая'
	`)
	if err != nil {
		log.Fatalf("Failed to load products: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var p Product
		rows.Scan(&p.CAE, &p.Brand, &p.Model, &p.Width, &p.Height, &p.Diameter, &p.LoadIndex, &p.SpeedIndex)
		products = append(products, p)
	}

	return products
}

func loadSeveravtoProducts(ctx context.Context, pool *pgxpool.Pool) []SeveravtoProduct {
	var products []SeveravtoProduct

	rows, err := pool.Query(ctx, `
		SELECT commodity_id, brand, model, width::float, height::float,
		       diameter, load_index, speed_index
		FROM nomenclature_tyres_severavto
	`)
	if err != nil {
		log.Fatalf("Failed to load Severavto products: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var p SeveravtoProduct
		rows.Scan(&p.CommodityID, &p.Brand, &p.Model, &p.Width, &p.Height, &p.Diameter, &p.LoadIndex, &p.SpeedIndex)
		products = append(products, p)
	}

	return products
}

func findBestMatch(product Product, severavtoProducts []SeveravtoProduct) (string, int) {
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
			return p.CommodityID, 1
		}
	}

	// Priority 2: Brand + model + width + indices
	for _, p := range severavtoProducts {
		if matchesBrand(p.Brand, russianBrands, product.Brand) &&
			matchesModel(p.Model, modelNorm) &&
			p.Width == product.Width &&
			p.LoadIndex == product.LoadIndex &&
			p.SpeedIndex == product.SpeedIndex {
			return p.CommodityID, 2
		}
	}

	// Priority 3: Brand + width + indices
	for _, p := range severavtoProducts {
		if matchesBrand(p.Brand, russianBrands, product.Brand) &&
			p.Width == product.Width &&
			p.LoadIndex == product.LoadIndex &&
			p.SpeedIndex == product.SpeedIndex {
			return p.CommodityID, 3
		}
	}

	// Priority 4: Brand + all sizes
	for _, p := range severavtoProducts {
		if matchesBrand(p.Brand, russianBrands, product.Brand) &&
			p.Width == product.Width &&
			p.Height == product.Height &&
			p.Diameter == product.Diameter {
			return p.CommodityID, 4
		}
	}

	return "", 0
}

func matchesBrand(severavtoBrand string, russianBrands []string, originalBrand string) bool {
	severavtoNorm := strings.ToUpper(strings.TrimSpace(severavtoBrand))

	for _, ruBrand := range russianBrands {
		if severavtoNorm == strings.ToUpper(strings.TrimSpace(ruBrand)) {
			return true
		}
	}

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
	reg := regexp.MustCompile(`[^A-Z0-9]`)
	s = reg.ReplaceAllString(s, "")
	return s
}

func saveMappings(ctx context.Context, pool *pgxpool.Pool, mappings []MappingResult) {
	start := time.Now()

	// Insert new mappings in batches (ON CONFLICT will update existing)
	batchSize := 1000
	for i := 0; i < len(mappings); i += batchSize {
		end := i + batchSize
		if end > len(mappings) {
			end = len(mappings)
		}

		batch := mappings[i:end]
		err := insertBatch(ctx, pool, batch)
		if err != nil {
			log.Fatalf("Failed to insert batch: %v", err)
		}

		fmt.Printf("   Saved %d/%d mappings\r", end, len(mappings))
	}

	elapsed := time.Since(start)
	fmt.Printf("\n✅ Saved %d mappings in %.1f seconds\n", len(mappings), elapsed.Seconds())
}

func insertBatch(ctx context.Context, pool *pgxpool.Pool, batch []MappingResult) error {
	// Build multi-value INSERT
	if len(batch) == 0 {
		return nil
	}

	values := make([]interface{}, 0, len(batch)*4)
	placeholders := make([]string, 0, len(batch))

	for i, m := range batch {
		detailsJSON, _ := json.Marshal(m.Details)
		offset := i * 4
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d)", offset+1, offset+2, offset+3, offset+4))
		values = append(values, m.CAE, m.CommodityID, m.Priority, string(detailsJSON))
	}

	query := fmt.Sprintf(`
		INSERT INTO severavto_tyres_mapping (cae, commodity_id, match_priority, match_details)
		VALUES %s
		ON CONFLICT (cae) DO UPDATE SET
			commodity_id = EXCLUDED.commodity_id,
			match_priority = EXCLUDED.match_priority,
			match_details = EXCLUDED.match_details,
			updated_at = NOW()
	`, strings.Join(placeholders, ","))

	_, err := pool.Exec(ctx, query, values...)
	return err
}
