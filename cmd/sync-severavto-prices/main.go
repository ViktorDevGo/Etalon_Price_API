package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

type PriceRecord struct {
	CAE          string
	CommodityID  string
	TerritoryName string
	Price        int
	Stock        int
	Priority     int
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

	fmt.Println("🔄 Синхронизация цен/остатков Северавто")
	fmt.Printf("   Время: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))

	// 1. Загрузить данные из tyres_prices_stock_severavto через маппинг
	fmt.Println("📥 Загрузка данных из tyres_prices_stock_severavto...")
	start := time.Now()

	rows, err := pool.Query(ctx, `
		SELECT
			sm.cae,
			tps.commodity_id,
			tps.territory_name,
			tps.price,
			tps.stock,
			sm.match_priority
		FROM tyres_prices_stock_severavto tps
		INNER JOIN severavto_tyres_mapping sm ON tps.commodity_id = sm.commodity_id
		WHERE tps.stock > 0
		  AND tps.price > 0
		ORDER BY sm.cae, sm.match_priority
	`)
	if err != nil {
		log.Fatalf("Failed to load prices: %v", err)
	}
	defer rows.Close()

	records := make([]PriceRecord, 0)
	for rows.Next() {
		var r PriceRecord
		rows.Scan(&r.CAE, &r.CommodityID, &r.TerritoryName, &r.Price, &r.Stock, &r.Priority)
		records = append(records, r)
	}

	elapsed := time.Since(start)
	fmt.Printf("   Загружено: %d записей за %.1f сек\n\n", len(records), elapsed.Seconds())

	if len(records) == 0 {
		fmt.Println("⚠️  Нет данных для синхронизации")
		return
	}

	// Группировка по CAE для статистики
	uniqueCAE := make(map[string]bool)
	uniqueWarehouses := make(map[string]bool)
	for _, r := range records {
		uniqueCAE[r.CAE] = true
		uniqueWarehouses[r.TerritoryName] = true
	}

	fmt.Printf("   Уникальных товаров (CAE): %d\n", len(uniqueCAE))
	fmt.Printf("   Уникальных складов: %d\n\n", len(uniqueWarehouses))

	// 2. Bulk upsert в tyres_prices_stock
	fmt.Println("💾 Сохранение в tyres_prices_stock...")
	start = time.Now()

	newCount, updatedCount, err := bulkUpsert(ctx, pool, records)
	if err != nil {
		log.Fatalf("Failed to upsert: %v", err)
	}

	elapsed = time.Since(start)
	fmt.Printf("   Сохранено за %.1f сек\n\n", elapsed.Seconds())

	// 3. Статистика
	fmt.Println("📊 Результат:")
	fmt.Printf("   ✅ Новые записи: %d\n", newCount)
	fmt.Printf("   🔄 Обновлено: %d\n", updatedCount)
	fmt.Printf("   📦 Всего обработано: %d\n", newCount+updatedCount)

	// 4. Проверка общего количества записей Северавто в БД
	var totalSeveravto int
	pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM tyres_prices_stock
		WHERE provider = 'Северавто'
	`).Scan(&totalSeveravto)

	fmt.Printf("\n📈 Всего записей Северавто в БД: %d\n", totalSeveravto)
	fmt.Println("\n✅ Синхронизация завершена!")
}

func bulkUpsert(ctx context.Context, pool *pgxpool.Pool, records []PriceRecord) (int, int, error) {
	batchSize := 5000
	totalNew := 0
	totalUpdated := 0

	for i := 0; i < len(records); i += batchSize {
		end := i + batchSize
		if end > len(records) {
			end = len(records)
		}

		batch := records[i:end]
		newCount, updatedCount, err := upsertBatch(ctx, pool, batch)
		if err != nil {
			return totalNew, totalUpdated, err
		}

		totalNew += newCount
		totalUpdated += updatedCount

		if (i+batchSize)%10000 == 0 || end == len(records) {
			fmt.Printf("   Обработано: %d/%d (новые: %d, обновлено: %d)\r",
				end, len(records), totalNew, totalUpdated)
		}
	}

	fmt.Println()
	return totalNew, totalUpdated, nil
}

func upsertBatch(ctx context.Context, pool *pgxpool.Pool, batch []PriceRecord) (int, int, error) {
	// Use transaction
	tx, err := pool.Begin(ctx)
	if err != nil {
		return 0, 0, err
	}
	defer tx.Rollback(ctx)

	// Create temp table
	_, err = tx.Exec(ctx, `
		CREATE TEMP TABLE IF NOT EXISTS temp_severavto_prices (
			cae VARCHAR(50),
			warehouse_name VARCHAR(255),
			price INTEGER,
			stock INTEGER
		) ON COMMIT DROP
	`)
	if err != nil {
		return 0, 0, err
	}

	// Truncate temp table (if exists from previous batch)
	_, err = tx.Exec(ctx, "DELETE FROM temp_severavto_prices")
	if err != nil {
		return 0, 0, err
	}

	// Copy data to temp table
	_, err = tx.CopyFrom(
		ctx,
		[]string{"temp_severavto_prices"},
		[]string{"cae", "warehouse_name", "price", "stock"},
		&priceSource{records: batch, index: 0},
	)
	if err != nil {
		return 0, 0, err
	}

	// Count existing records
	var existingCount int
	err = tx.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM tyres_prices_stock tps
		INNER JOIN temp_severavto_prices tmp
			ON tps.cae = tmp.cae
			AND tps.warehouse_name = tmp.warehouse_name
		WHERE tps.provider = 'Северавто'
	`).Scan(&existingCount)
	if err != nil {
		return 0, 0, err
	}

	// Upsert
	_, err = tx.Exec(ctx, `
		INSERT INTO tyres_prices_stock (cae, warehouse_name, price, stock, provider, isimport, updated_at)
		SELECT cae, warehouse_name, price, stock, 'Северавто', 0, NOW()
		FROM temp_severavto_prices
		ON CONFLICT (cae, warehouse_name, provider)
		DO UPDATE SET
			price = EXCLUDED.price,
			stock = EXCLUDED.stock,
			isimport = 0,
			updated_at = NOW()
	`)
	if err != nil {
		return 0, 0, err
	}

	// Commit transaction
	err = tx.Commit(ctx)
	if err != nil {
		return 0, 0, err
	}

	newCount := len(batch) - existingCount
	updatedCount := existingCount

	return newCount, updatedCount, nil
}

type priceSource struct {
	records []PriceRecord
	index   int
}

func (ps *priceSource) Next() bool {
	ps.index++
	return ps.index <= len(ps.records)
}

func (ps *priceSource) Values() ([]interface{}, error) {
	r := ps.records[ps.index-1]
	return []interface{}{r.CAE, r.TerritoryName, r.Price, r.Stock}, nil
}

func (ps *priceSource) Err() error {
	return nil
}
