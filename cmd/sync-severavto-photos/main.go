package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/prokoleso/etalon-price-api/internal/providers/severavto"
)

type PhotoRecord struct {
	CommodityID string
	Brand       string
	Model       string
	Size        string
	Season      string // только для шин
	PictureURL  string
}

func main() {
	// Load environment variables
	_ = godotenv.Load()

	dbDSN := os.Getenv("DATABASE_DSN")
	if dbDSN == "" {
		log.Fatal("DATABASE_DSN must be set")
	}

	baseURL := os.Getenv("SEVERAVTO_BASE_URL")
	apiKey := os.Getenv("SEVERAVTO_API_KEY")
	if baseURL == "" || apiKey == "" {
		log.Fatal("SEVERAVTO_BASE_URL and SEVERAVTO_API_KEY must be set")
	}

	// Connect to database
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbDSN)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	fmt.Println("✅ Connected to database")

	// Create Severavto client
	client := severavto.NewClient(severavto.ClientConfig{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Timeout: 120 * time.Second,
	})

	// Sync tyres photos
	fmt.Println("\n📸 Syncing tyres photos...")
	if err := syncTyresPhotos(ctx, pool, client); err != nil {
		log.Fatalf("Failed to sync tyres photos: %v", err)
	}

	// Sync rims photos
	fmt.Println("\n📸 Syncing rims photos...")
	if err := syncRimsPhotos(ctx, pool, client); err != nil {
		log.Fatalf("Failed to sync rims photos: %v", err)
	}

	fmt.Println("\n🎉 All photos synced successfully!")
}

func syncTyresPhotos(ctx context.Context, pool *pgxpool.Pool, client *severavto.Client) error {
	// Fetch data from API
	fmt.Println("📥 Fetching tyres from Severavto API...")
	data, err := client.FetchTyres(ctx)
	if err != nil {
		return fmt.Errorf("fetch tyres: %w", err)
	}

	// Extract unique photos
	uniquePhotos := make(map[string]*PhotoRecord)
	for _, item := range data.Commodities {
		if _, exists := uniquePhotos[item.ID]; !exists {
			size := fmt.Sprintf("%s/%s R%s", item.Width, item.Height, item.Diameter)
			uniquePhotos[item.ID] = &PhotoRecord{
				CommodityID: item.ID,
				Brand:       item.Brand,
				Model:       item.Model,
				Size:        size,
				Season:      item.Season,
				PictureURL:  item.Picture,
			}
		}
	}

	fmt.Printf("📊 Found %d unique tyres\n", len(uniquePhotos))

	// Bulk upsert using single connection
	withPhotos := 0
	withoutPhotos := 0

	// Get connection from pool
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire connection: %w", err)
	}
	defer conn.Release()

	// Begin transaction
	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Create temp table
	_, err = tx.Exec(ctx, `
		CREATE TEMP TABLE temp_foto_tyres (
			commodity_id VARCHAR(50),
			brand VARCHAR(255),
			model VARCHAR(255),
			size VARCHAR(50),
			season VARCHAR(50),
			picture_url TEXT
		) ON COMMIT DROP
	`)
	if err != nil {
		return fmt.Errorf("create temp table: %w", err)
	}

	// Prepare data for COPY
	rows := make([][]interface{}, 0, len(uniquePhotos))
	for _, photo := range uniquePhotos {
		rows = append(rows, []interface{}{
			photo.CommodityID,
			photo.Brand,
			photo.Model,
			photo.Size,
			photo.Season,
			photo.PictureURL,
		})

		if photo.PictureURL != "" {
			withPhotos++
		} else {
			withoutPhotos++
		}
	}

	// Copy to temp table
	_, err = tx.CopyFrom(
		ctx,
		pgx.Identifier{"temp_foto_tyres"},
		[]string{"commodity_id", "brand", "model", "size", "season", "picture_url"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return fmt.Errorf("copy to temp table: %w", err)
	}

	// Upsert from temp to main table
	result, err := tx.Exec(ctx, `
		INSERT INTO foto_tyres (commodity_id, brand, model, size, season, picture_url, created_at, updated_at)
		SELECT commodity_id, brand, model, size, season, picture_url, NOW(), NOW()
		FROM temp_foto_tyres
		ON CONFLICT (commodity_id) DO UPDATE SET
			brand = EXCLUDED.brand,
			model = EXCLUDED.model,
			size = EXCLUDED.size,
			season = EXCLUDED.season,
			picture_url = EXCLUDED.picture_url,
			updated_at = NOW()
	`)
	if err != nil {
		return fmt.Errorf("upsert: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	fmt.Printf("✅ Tyres photos synced: %d records\n", result.RowsAffected())
	fmt.Printf("   - With photos: %d (%.1f%%)\n", withPhotos, float64(withPhotos)/float64(len(uniquePhotos))*100)
	fmt.Printf("   - Without photos: %d (%.1f%%)\n", withoutPhotos, float64(withoutPhotos)/float64(len(uniquePhotos))*100)

	return nil
}

func syncRimsPhotos(ctx context.Context, pool *pgxpool.Pool, client *severavto.Client) error {
	// Fetch data from API
	fmt.Println("📥 Fetching rims from Severavto API...")
	data, err := client.FetchRims(ctx)
	if err != nil {
		return fmt.Errorf("fetch rims: %w", err)
	}

	// Extract unique photos
	uniquePhotos := make(map[string]*PhotoRecord)
	for _, item := range data.Commodities {
		if _, exists := uniquePhotos[item.ID]; !exists {
			size := fmt.Sprintf("%sx%s", item.Width, item.Diameter)
			uniquePhotos[item.ID] = &PhotoRecord{
				CommodityID: item.ID,
				Brand:       item.Brand,
				Model:       item.Model,
				Size:        size,
				PictureURL:  item.Picture,
			}
		}
	}

	fmt.Printf("📊 Found %d unique rims\n", len(uniquePhotos))

	// Bulk upsert using single connection
	withPhotos := 0
	withoutPhotos := 0

	// Get connection from pool
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire connection: %w", err)
	}
	defer conn.Release()

	// Begin transaction
	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Create temp table
	_, err = tx.Exec(ctx, `
		CREATE TEMP TABLE temp_foto_rims (
			commodity_id VARCHAR(50),
			brand VARCHAR(255),
			model VARCHAR(255),
			size VARCHAR(50),
			picture_url TEXT
		) ON COMMIT DROP
	`)
	if err != nil {
		return fmt.Errorf("create temp table: %w", err)
	}

	// Prepare data for COPY
	rows := make([][]interface{}, 0, len(uniquePhotos))
	for _, photo := range uniquePhotos {
		rows = append(rows, []interface{}{
			photo.CommodityID,
			photo.Brand,
			photo.Model,
			photo.Size,
			photo.PictureURL,
		})

		if photo.PictureURL != "" {
			withPhotos++
		} else {
			withoutPhotos++
		}
	}

	// Copy to temp table
	_, err = tx.CopyFrom(
		ctx,
		pgx.Identifier{"temp_foto_rims"},
		[]string{"commodity_id", "brand", "model", "size", "picture_url"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return fmt.Errorf("copy to temp table: %w", err)
	}

	// Upsert from temp to main table
	result, err := tx.Exec(ctx, `
		INSERT INTO foto_rims (commodity_id, brand, model, size, picture_url, created_at, updated_at)
		SELECT commodity_id, brand, model, size, picture_url, NOW(), NOW()
		FROM temp_foto_rims
		ON CONFLICT (commodity_id) DO UPDATE SET
			brand = EXCLUDED.brand,
			model = EXCLUDED.model,
			size = EXCLUDED.size,
			picture_url = EXCLUDED.picture_url,
			updated_at = NOW()
	`)
	if err != nil {
		return fmt.Errorf("upsert: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	fmt.Printf("✅ Rims photos synced: %d records\n", result.RowsAffected())
	fmt.Printf("   - With photos: %d (%.1f%%)\n", withPhotos, float64(withPhotos)/float64(len(uniquePhotos))*100)
	fmt.Printf("   - Without photos: %d (%.1f%%)\n", withoutPhotos, float64(withoutPhotos)/float64(len(uniquePhotos))*100)

	return nil
}
