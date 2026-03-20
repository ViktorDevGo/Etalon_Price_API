package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prokoleso/etalon-price-api/internal/domain"
	"github.com/prokoleso/etalon-price-api/internal/domain/severavto"
)

// PricesStockRepository handles prices and stock database operations
type PricesStockRepository struct {
	pool *pgxpool.Pool
}

// NewPricesStockRepository creates a new prices stock repository
func NewPricesStockRepository(pool *pgxpool.Pool) *PricesStockRepository {
	return &PricesStockRepository{pool: pool}
}

// UpsertTyresPricesStock inserts or updates tyres prices and stock in batch
func (r *PricesStockRepository) UpsertTyresPricesStock(ctx context.Context, items []domain.TyrePriceStock) (newCount, updatedCount int, err error) {
	if len(items) == 0 {
		return 0, 0, nil
	}

	// Build list of (cae, warehouse_name, provider) tuples to check
	type keyTuple struct {
		CAE           string
		WarehouseName string
		Provider      string
	}
	tuples := make([]keyTuple, len(items))
	for i, item := range items {
		tuples[i] = keyTuple{CAE: item.CAE, WarehouseName: item.WarehouseName, Provider: item.Provider}
	}

	// Get existing items in one query
	existingKeys := make(map[keyTuple]bool)
	rows, err := r.pool.Query(ctx, `
		SELECT DISTINCT cae, warehouse_name, provider
		FROM tyres_prices_stock
		WHERE (cae, warehouse_name, provider) IN (
			SELECT unnest($1::text[]), unnest($2::text[]), unnest($3::text[])
		)
	`,
		func() []string {
			caeList := make([]string, len(tuples))
			for i, t := range tuples {
				caeList[i] = t.CAE
			}
			return caeList
		}(),
		func() []string {
			whList := make([]string, len(tuples))
			for i, t := range tuples {
				whList[i] = t.WarehouseName
			}
			return whList
		}(),
		func() []string {
			provList := make([]string, len(tuples))
			for i, t := range tuples {
				provList[i] = t.Provider
			}
			return provList
		}(),
	)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var cae, whName, prov string
			if err := rows.Scan(&cae, &whName, &prov); err == nil {
				existingKeys[keyTuple{CAE: cae, WarehouseName: whName, Provider: prov}] = true
			}
		}
	}

	// Count new vs updated
	for _, item := range items {
		if existingKeys[keyTuple{CAE: item.CAE, WarehouseName: item.WarehouseName, Provider: item.Provider}] {
			updatedCount++
		} else {
			newCount++
		}
	}

	// Bulk upsert
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Create temporary table
	_, err = tx.Exec(ctx, `
		CREATE TEMP TABLE temp_tyres_prices_stock (
			cae VARCHAR(50),
			warehouse_name VARCHAR(255),
			price INTEGER,
			stock INTEGER,
			isimport INTEGER,
			provider VARCHAR(100)
		) ON COMMIT DROP
	`)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to create temp table: %w", err)
	}

	// Copy data to temp table
	_, err = tx.CopyFrom(
		ctx,
		pgx.Identifier{"temp_tyres_prices_stock"},
		[]string{"cae", "warehouse_name", "price", "stock", "isimport", "provider"},
		pgx.CopyFromSlice(len(items), func(i int) ([]interface{}, error) {
			item := items[i]
			return []interface{}{item.CAE, item.WarehouseName, item.Price, item.Stock, item.IsImport, item.Provider}, nil
		}),
	)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to copy data: %w", err)
	}

	// Merge from temp table to main table
	_, err = tx.Exec(ctx, `
		INSERT INTO tyres_prices_stock (cae, warehouse_name, price, stock, isimport, provider)
		SELECT cae, warehouse_name, price, stock, isimport, provider
		FROM temp_tyres_prices_stock
		ON CONFLICT (cae, warehouse_name, provider) DO UPDATE SET
			price = EXCLUDED.price,
			stock = EXCLUDED.stock,
			isimport = 0,
			updated_at = CURRENT_TIMESTAMP
	`)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to merge data: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return 0, 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return newCount, updatedCount, nil
}

// UpsertRimsPricesStock inserts or updates rims prices and stock in batch
func (r *PricesStockRepository) UpsertRimsPricesStock(ctx context.Context, items []domain.RimPriceStock) (newCount, updatedCount int, err error) {
	if len(items) == 0 {
		return 0, 0, nil
	}

	// Build list of (cae, warehouse_name, provider) tuples to check
	type keyTuple struct {
		CAE           string
		WarehouseName string
		Provider      string
	}
	tuples := make([]keyTuple, len(items))
	for i, item := range items {
		tuples[i] = keyTuple{CAE: item.CAE, WarehouseName: item.WarehouseName, Provider: item.Provider}
	}

	// Get existing items in one query
	existingKeys := make(map[keyTuple]bool)
	rows, err := r.pool.Query(ctx, `
		SELECT DISTINCT cae, warehouse_name, provider
		FROM rims_prices_stock
		WHERE (cae, warehouse_name, provider) IN (
			SELECT unnest($1::text[]), unnest($2::text[]), unnest($3::text[])
		)
	`,
		func() []string {
			caeList := make([]string, len(tuples))
			for i, t := range tuples {
				caeList[i] = t.CAE
			}
			return caeList
		}(),
		func() []string {
			whList := make([]string, len(tuples))
			for i, t := range tuples {
				whList[i] = t.WarehouseName
			}
			return whList
		}(),
		func() []string {
			provList := make([]string, len(tuples))
			for i, t := range tuples {
				provList[i] = t.Provider
			}
			return provList
		}(),
	)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var cae, whName, prov string
			if err := rows.Scan(&cae, &whName, &prov); err == nil {
				existingKeys[keyTuple{CAE: cae, WarehouseName: whName, Provider: prov}] = true
			}
		}
	}

	// Count new vs updated
	for _, item := range items {
		if existingKeys[keyTuple{CAE: item.CAE, WarehouseName: item.WarehouseName, Provider: item.Provider}] {
			updatedCount++
		} else {
			newCount++
		}
	}

	// Bulk upsert
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Create temporary table
	_, err = tx.Exec(ctx, `
		CREATE TEMP TABLE temp_rims_prices_stock (
			cae VARCHAR(50),
			warehouse_name VARCHAR(255),
			price INTEGER,
			stock INTEGER,
			isimport INTEGER,
			provider VARCHAR(100)
		) ON COMMIT DROP
	`)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to create temp table: %w", err)
	}

	// Copy data to temp table
	_, err = tx.CopyFrom(
		ctx,
		pgx.Identifier{"temp_rims_prices_stock"},
		[]string{"cae", "warehouse_name", "price", "stock", "isimport", "provider"},
		pgx.CopyFromSlice(len(items), func(i int) ([]interface{}, error) {
			item := items[i]
			return []interface{}{item.CAE, item.WarehouseName, item.Price, item.Stock, item.IsImport, item.Provider}, nil
		}),
	)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to copy data: %w", err)
	}

	// Merge from temp table to main table
	_, err = tx.Exec(ctx, `
		INSERT INTO rims_prices_stock (cae, warehouse_name, price, stock, isimport, provider)
		SELECT cae, warehouse_name, price, stock, isimport, provider
		FROM temp_rims_prices_stock
		ON CONFLICT (cae, warehouse_name, provider) DO UPDATE SET
			price = EXCLUDED.price,
			stock = EXCLUDED.stock,
			isimport = 0,
			updated_at = CURRENT_TIMESTAMP
	`)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to merge data: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return 0, 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return newCount, updatedCount, nil
}

// CreatePricesStockSyncRun creates a new sync run record
func (r *PricesStockRepository) CreatePricesStockSyncRun(ctx context.Context, syncType string) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx, `
		INSERT INTO prices_stock_sync_runs (sync_type, started_at, status)
		VALUES ($1, $2, 'running')
		RETURNING id
	`, syncType, time.Now()).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to create sync run: %w", err)
	}

	return id, nil
}

// UpdatePricesStockSyncRun updates sync run status
func (r *PricesStockRepository) UpdatePricesStockSyncRun(ctx context.Context, id int64, status string, totalItems, totalWarehouses, newItems, updatedItems int, errorMsg string) error {
	now := time.Now()
	_, err := r.pool.Exec(ctx, `
		UPDATE prices_stock_sync_runs SET
			completed_at = $2,
			status = $3,
			total_items = $4,
			total_warehouses = $5,
			new_items = $6,
			updated_items = $7,
			error_message = $8
		WHERE id = $1
	`, id, now, status, totalItems, totalWarehouses, newItems, updatedItems, errorMsg)

	if err != nil {
		return fmt.Errorf("failed to update sync run: %w", err)
	}

	return nil
}

// GetTyrePricesStockByCAE retrieves prices and stock for a tyre
func (r *PricesStockRepository) GetTyrePricesStockByCAE(ctx context.Context, cae string) ([]domain.TyrePriceStock, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, cae, warehouse_name, price, stock, isimport, provider, updated_at
		FROM tyres_prices_stock
		WHERE cae = $1
		ORDER BY warehouse_name
	`, cae)
	if err != nil {
		return nil, fmt.Errorf("failed to query tyre prices: %w", err)
	}
	defer rows.Close()

	var items []domain.TyrePriceStock
	for rows.Next() {
		var item domain.TyrePriceStock
		if err := rows.Scan(&item.ID, &item.CAE, &item.WarehouseName, &item.Price, &item.Stock, &item.IsImport, &item.Provider, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		items = append(items, item)
	}

	return items, nil
}

// GetRimPricesStockByCAE retrieves prices and stock for a rim
func (r *PricesStockRepository) GetRimPricesStockByCAE(ctx context.Context, cae string) ([]domain.RimPriceStock, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, cae, warehouse_name, price, stock, isimport, provider, updated_at
		FROM rims_prices_stock
		WHERE cae = $1
		ORDER BY warehouse_name
	`, cae)
	if err != nil {
		return nil, fmt.Errorf("failed to query rim prices: %w", err)
	}
	defer rows.Close()

	var items []domain.RimPriceStock
	for rows.Next() {
		var item domain.RimPriceStock
		if err := rows.Scan(&item.ID, &item.CAE, &item.WarehouseName, &item.Price, &item.Stock, &item.IsImport, &item.Provider, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		items = append(items, item)
	}

	return items, nil
}

// CountTyresPricesStock returns total tyres prices/stock records
func (r *PricesStockRepository) CountTyresPricesStock(ctx context.Context) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM tyres_prices_stock").Scan(&count)
	return count, err
}

// CountRimsPricesStock returns total rims prices/stock records
func (r *PricesStockRepository) CountRimsPricesStock(ctx context.Context) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM rims_prices_stock").Scan(&count)
	return count, err
}

// CountUniqueCAEsInTyresPricesStock returns count of unique products
func (r *PricesStockRepository) CountUniqueCAEsInTyresPricesStock(ctx context.Context) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(DISTINCT cae) FROM tyres_prices_stock").Scan(&count)
	return count, err
}

// CountUniqueCAEsInRimsPricesStock returns count of unique products
func (r *PricesStockRepository) CountUniqueCAEsInRimsPricesStock(ctx context.Context) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(DISTINCT cae) FROM rims_prices_stock").Scan(&count)
	return count, err
}

// ====================
// Severavto Repository Methods
// ====================

// UpsertSeveravtoTyresPricesStock inserts or updates Severavto tyres prices and stock in batch
func (r *PricesStockRepository) UpsertSeveravtoTyresPricesStock(ctx context.Context, items []severavto.TyrePriceStock) (newCount, updatedCount int, err error) {
	if len(items) == 0 {
		return 0, 0, nil
	}

	// Build list of (commodity_id, territory_name, provider) tuples to check
	type keyTuple struct {
		CommodityID   string
		TerritoryName string
		Provider      string
	}
	tuples := make([]keyTuple, len(items))
	for i, item := range items {
		tuples[i] = keyTuple{CommodityID: item.CommodityID, TerritoryName: item.TerritoryName, Provider: item.Provider}
	}

	// Get existing items in one query
	existingKeys := make(map[keyTuple]bool)
	rows, err := r.pool.Query(ctx, `
		SELECT DISTINCT commodity_id, territory_name, provider
		FROM tyres_prices_stock_severavto
		WHERE (commodity_id, territory_name, provider) IN (
			SELECT unnest($1::text[]), unnest($2::text[]), unnest($3::text[])
		)
	`,
		func() []string {
			idList := make([]string, len(tuples))
			for i, t := range tuples {
				idList[i] = t.CommodityID
			}
			return idList
		}(),
		func() []string {
			terrList := make([]string, len(tuples))
			for i, t := range tuples {
				terrList[i] = t.TerritoryName
			}
			return terrList
		}(),
		func() []string {
			provList := make([]string, len(tuples))
			for i, t := range tuples {
				provList[i] = t.Provider
			}
			return provList
		}(),
	)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var commodityID, terrName, prov string
			if err := rows.Scan(&commodityID, &terrName, &prov); err == nil {
				existingKeys[keyTuple{CommodityID: commodityID, TerritoryName: terrName, Provider: prov}] = true
			}
		}
	}

	// Count new vs updated
	for _, item := range items {
		if existingKeys[keyTuple{CommodityID: item.CommodityID, TerritoryName: item.TerritoryName, Provider: item.Provider}] {
			updatedCount++
		} else {
			newCount++
		}
	}

	// Bulk upsert
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Create temporary table
	_, err = tx.Exec(ctx, `
		CREATE TEMP TABLE temp_severavto_tyres_prices (
			commodity_id VARCHAR(50),
			territory_name VARCHAR(255),
			price INTEGER,
			price_delayed INTEGER,
			price_rrp INTEGER,
			stock INTEGER,
			place_type VARCHAR(50),
			is_sellout BOOLEAN,
			provider VARCHAR(100)
		) ON COMMIT DROP
	`)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to create temp table: %w", err)
	}

	// Copy data to temp table
	_, err = tx.CopyFrom(
		ctx,
		pgx.Identifier{"temp_severavto_tyres_prices"},
		[]string{"commodity_id", "territory_name", "price", "price_delayed", "price_rrp", "stock", "place_type", "is_sellout", "provider"},
		pgx.CopyFromSlice(len(items), func(i int) ([]interface{}, error) {
			item := items[i]
			return []interface{}{
				item.CommodityID, item.TerritoryName, item.Price, item.PriceDelayed,
				item.PriceRRP, item.Stock, item.PlaceType, item.IsSellout, item.Provider,
			}, nil
		}),
	)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to copy data: %w", err)
	}

	// Merge from temp table to main table (with deduplication)
	// Use DISTINCT ON to handle duplicates in source data
	_, err = tx.Exec(ctx, `
		INSERT INTO tyres_prices_stock_severavto (
			commodity_id, territory_name, price, price_delayed, price_rrp,
			stock, place_type, is_sellout, provider
		)
		SELECT DISTINCT ON (commodity_id, territory_name, provider)
			commodity_id, territory_name, price, price_delayed, price_rrp,
			stock, place_type, is_sellout, provider
		FROM temp_severavto_tyres_prices
		ORDER BY commodity_id, territory_name, provider, price ASC
		ON CONFLICT (commodity_id, territory_name, provider) DO UPDATE SET
			price = EXCLUDED.price,
			price_delayed = EXCLUDED.price_delayed,
			price_rrp = EXCLUDED.price_rrp,
			stock = EXCLUDED.stock,
			place_type = EXCLUDED.place_type,
			is_sellout = EXCLUDED.is_sellout,
			updated_at = CURRENT_TIMESTAMP
	`)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to merge data: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return 0, 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return newCount, updatedCount, nil
}

// UpsertSeveravtoRimsPricesStock inserts or updates Severavto rims prices and stock in batch
func (r *PricesStockRepository) UpsertSeveravtoRimsPricesStock(ctx context.Context, items []severavto.RimPriceStock) (newCount, updatedCount int, err error) {
	if len(items) == 0 {
		return 0, 0, nil
	}

	// Build list of (commodity_id, territory_name, provider) tuples to check
	type keyTuple struct {
		CommodityID   string
		TerritoryName string
		Provider      string
	}
	tuples := make([]keyTuple, len(items))
	for i, item := range items {
		tuples[i] = keyTuple{CommodityID: item.CommodityID, TerritoryName: item.TerritoryName, Provider: item.Provider}
	}

	// Get existing items in one query
	existingKeys := make(map[keyTuple]bool)
	rows, err := r.pool.Query(ctx, `
		SELECT DISTINCT commodity_id, territory_name, provider
		FROM rims_prices_stock_severavto
		WHERE (commodity_id, territory_name, provider) IN (
			SELECT unnest($1::text[]), unnest($2::text[]), unnest($3::text[])
		)
	`,
		func() []string {
			idList := make([]string, len(tuples))
			for i, t := range tuples {
				idList[i] = t.CommodityID
			}
			return idList
		}(),
		func() []string {
			terrList := make([]string, len(tuples))
			for i, t := range tuples {
				terrList[i] = t.TerritoryName
			}
			return terrList
		}(),
		func() []string {
			provList := make([]string, len(tuples))
			for i, t := range tuples {
				provList[i] = t.Provider
			}
			return provList
		}(),
	)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var commodityID, terrName, prov string
			if err := rows.Scan(&commodityID, &terrName, &prov); err == nil {
				existingKeys[keyTuple{CommodityID: commodityID, TerritoryName: terrName, Provider: prov}] = true
			}
		}
	}

	// Count new vs updated
	for _, item := range items {
		if existingKeys[keyTuple{CommodityID: item.CommodityID, TerritoryName: item.TerritoryName, Provider: item.Provider}] {
			updatedCount++
		} else {
			newCount++
		}
	}

	// Bulk upsert
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Create temporary table
	_, err = tx.Exec(ctx, `
		CREATE TEMP TABLE temp_severavto_rims_prices (
			commodity_id VARCHAR(50),
			territory_name VARCHAR(255),
			price INTEGER,
			price_delayed INTEGER,
			price_rrp INTEGER,
			stock INTEGER,
			place_type VARCHAR(50),
			is_sellout BOOLEAN,
			provider VARCHAR(100)
		) ON COMMIT DROP
	`)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to create temp table: %w", err)
	}

	// Copy data to temp table
	_, err = tx.CopyFrom(
		ctx,
		pgx.Identifier{"temp_severavto_rims_prices"},
		[]string{"commodity_id", "territory_name", "price", "price_delayed", "price_rrp", "stock", "place_type", "is_sellout", "provider"},
		pgx.CopyFromSlice(len(items), func(i int) ([]interface{}, error) {
			item := items[i]
			return []interface{}{
				item.CommodityID, item.TerritoryName, item.Price, item.PriceDelayed,
				item.PriceRRP, item.Stock, item.PlaceType, item.IsSellout, item.Provider,
			}, nil
		}),
	)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to copy data: %w", err)
	}

	// Merge from temp table to main table (with deduplication)
	// Use DISTINCT ON to handle duplicates in source data
	_, err = tx.Exec(ctx, `
		INSERT INTO rims_prices_stock_severavto (
			commodity_id, territory_name, price, price_delayed, price_rrp,
			stock, place_type, is_sellout, provider
		)
		SELECT DISTINCT ON (commodity_id, territory_name, provider)
			commodity_id, territory_name, price, price_delayed, price_rrp,
			stock, place_type, is_sellout, provider
		FROM temp_severavto_rims_prices
		ORDER BY commodity_id, territory_name, provider, price ASC
		ON CONFLICT (commodity_id, territory_name, provider) DO UPDATE SET
			price = EXCLUDED.price,
			price_delayed = EXCLUDED.price_delayed,
			price_rrp = EXCLUDED.price_rrp,
			stock = EXCLUDED.stock,
			place_type = EXCLUDED.place_type,
			is_sellout = EXCLUDED.is_sellout,
			updated_at = CURRENT_TIMESTAMP
	`)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to merge data: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return 0, 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return newCount, updatedCount, nil
}
