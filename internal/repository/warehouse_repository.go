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

// WarehouseRepository handles warehouse database operations
type WarehouseRepository struct {
	pool *pgxpool.Pool
}

// NewWarehouseRepository creates a new warehouse repository
func NewWarehouseRepository(pool *pgxpool.Pool) *WarehouseRepository {
	return &WarehouseRepository{pool: pool}
}

// UpsertWarehouses inserts or updates warehouses in batch
func (r *WarehouseRepository) UpsertWarehouses(ctx context.Context, warehouses []domain.Warehouse) (newCount, updatedCount int, err error) {
	if len(warehouses) == 0 {
		return 0, 0, nil
	}

	// Get existing warehouse IDs
	existingIDs := make(map[int]bool)
	ids := make([]int, len(warehouses))
	for i, w := range warehouses {
		ids[i] = w.ID
	}

	rows, err := r.pool.Query(ctx, "SELECT id FROM warehouses WHERE id = ANY($1)", ids)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var id int
			if err := rows.Scan(&id); err == nil {
				existingIDs[id] = true
			}
		}
	}

	// Count new vs updated
	for _, w := range warehouses {
		if existingIDs[w.ID] {
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
		CREATE TEMP TABLE temp_warehouses (
			id INTEGER,
			key VARCHAR(100),
			name TEXT,
			short_name VARCHAR(100),
			sto_id INTEGER,
			background_color VARCHAR(20),
			have_delivery BOOLEAN,
			have_pickup BOOLEAN,
			is_paid_delivery BOOLEAN,
			logistic_days INTEGER
		) ON COMMIT DROP
	`)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to create temp table: %w", err)
	}

	// Copy data to temp table
	_, err = tx.CopyFrom(
		ctx,
		pgx.Identifier{"temp_warehouses"},
		[]string{"id", "key", "name", "short_name", "sto_id", "background_color",
			"have_delivery", "have_pickup", "is_paid_delivery", "logistic_days"},
		pgx.CopyFromSlice(len(warehouses), func(i int) ([]interface{}, error) {
			w := warehouses[i]
			return []interface{}{
				w.ID, w.Key, w.Name, w.ShortName, w.StoID, w.BackgroundColor,
				w.HaveDelivery, w.HavePickup, w.IsPaidDelivery, w.LogisticDays,
			}, nil
		}),
	)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to copy data: %w", err)
	}

	// Merge from temp table to main table
	_, err = tx.Exec(ctx, `
		INSERT INTO warehouses (
			id, key, name, short_name, sto_id, background_color,
			have_delivery, have_pickup, is_paid_delivery, logistic_days
		)
		SELECT
			id, key, name, short_name, sto_id, background_color,
			have_delivery, have_pickup, is_paid_delivery, logistic_days
		FROM temp_warehouses
		ON CONFLICT (id) DO UPDATE SET
			key = EXCLUDED.key,
			name = EXCLUDED.name,
			short_name = EXCLUDED.short_name,
			sto_id = EXCLUDED.sto_id,
			background_color = EXCLUDED.background_color,
			have_delivery = EXCLUDED.have_delivery,
			have_pickup = EXCLUDED.have_pickup,
			is_paid_delivery = EXCLUDED.is_paid_delivery,
			logistic_days = EXCLUDED.logistic_days,
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

// GetWarehouseByID retrieves a warehouse by ID
func (r *WarehouseRepository) GetWarehouseByID(ctx context.Context, id int) (*domain.Warehouse, error) {
	var w domain.Warehouse
	err := r.pool.QueryRow(ctx, `
		SELECT id, key, name, short_name, sto_id, background_color,
		       have_delivery, have_pickup, is_paid_delivery, logistic_days,
		       created_at, updated_at
		FROM warehouses
		WHERE id = $1
	`, id).Scan(
		&w.ID, &w.Key, &w.Name, &w.ShortName, &w.StoID, &w.BackgroundColor,
		&w.HaveDelivery, &w.HavePickup, &w.IsPaidDelivery, &w.LogisticDays,
		&w.CreatedAt, &w.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get warehouse: %w", err)
	}

	return &w, nil
}

// GetAllWarehouses retrieves all warehouses
func (r *WarehouseRepository) GetAllWarehouses(ctx context.Context) ([]domain.Warehouse, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, key, name, short_name, sto_id, background_color,
		       have_delivery, have_pickup, is_paid_delivery, logistic_days,
		       created_at, updated_at
		FROM warehouses
		ORDER BY name
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query warehouses: %w", err)
	}
	defer rows.Close()

	var warehouses []domain.Warehouse
	for rows.Next() {
		var w domain.Warehouse
		if err := rows.Scan(
			&w.ID, &w.Key, &w.Name, &w.ShortName, &w.StoID, &w.BackgroundColor,
			&w.HaveDelivery, &w.HavePickup, &w.IsPaidDelivery, &w.LogisticDays,
			&w.CreatedAt, &w.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan warehouse: %w", err)
		}
		warehouses = append(warehouses, w)
	}

	return warehouses, nil
}

// CountWarehouses returns total warehouses count
func (r *WarehouseRepository) CountWarehouses(ctx context.Context) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM warehouses").Scan(&count)
	return count, err
}

// CreateWarehousesSyncRun creates a new sync run record
func (r *WarehouseRepository) CreateWarehousesSyncRun(ctx context.Context) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx, `
		INSERT INTO warehouses_sync_runs (started_at, status)
		VALUES ($1, 'running')
		RETURNING id
	`, time.Now()).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to create sync run: %w", err)
	}

	return id, nil
}

// UpdateWarehousesSyncRun updates sync run status
func (r *WarehouseRepository) UpdateWarehousesSyncRun(ctx context.Context, id int64, status string, total, newCount, updatedCount int, errorMsg string) error {
	now := time.Now()
	_, err := r.pool.Exec(ctx, `
		UPDATE warehouses_sync_runs SET
			completed_at = $2,
			status = $3,
			total_warehouses = $4,
			new_warehouses = $5,
			updated_warehouses = $6,
			error_message = $7
		WHERE id = $1
	`, id, now, status, total, newCount, updatedCount, errorMsg)

	if err != nil {
		return fmt.Errorf("failed to update sync run: %w", err)
	}

	return nil
}

// ====================
// Severavto Repository Methods
// ====================

// UpsertSeveravtoWarehouses inserts or updates Severavto warehouses (territories)
func (r *WarehouseRepository) UpsertSeveravtoWarehouses(ctx context.Context, warehouses []severavto.Warehouse) (newCount, updatedCount int, err error) {
	if len(warehouses) == 0 {
		return 0, 0, nil
	}

	// Check which warehouses already exist
	territoryNames := make([]string, len(warehouses))
	for i, w := range warehouses {
		territoryNames[i] = w.TerritoryName
	}

	existingNames := make(map[string]bool)
	rows, err := r.pool.Query(ctx, `
		SELECT territory_name
		FROM warehouses_severavto
		WHERE territory_name = ANY($1)
	`, territoryNames)

	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err == nil {
				existingNames[name] = true
			}
		}
	}

	// Count new vs updated
	for _, w := range warehouses {
		if existingNames[w.TerritoryName] {
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
		CREATE TEMP TABLE temp_severavto_warehouses (
			territory_name VARCHAR(255),
			region VARCHAR(255),
			place_type VARCHAR(50)
		) ON COMMIT DROP
	`)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to create temp table: %w", err)
	}

	// Copy data to temp table
	_, err = tx.CopyFrom(
		ctx,
		pgx.Identifier{"temp_severavto_warehouses"},
		[]string{"territory_name", "region", "place_type"},
		pgx.CopyFromSlice(len(warehouses), func(i int) ([]interface{}, error) {
			w := warehouses[i]
			return []interface{}{w.TerritoryName, w.Region, w.PlaceType}, nil
		}),
	)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to copy data: %w", err)
	}

	// Merge from temp table to main table
	_, err = tx.Exec(ctx, `
		INSERT INTO warehouses_severavto (territory_name, region, place_type)
		SELECT territory_name, region, place_type
		FROM temp_severavto_warehouses
		ON CONFLICT (territory_name) DO UPDATE SET
			region = EXCLUDED.region,
			place_type = EXCLUDED.place_type,
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
