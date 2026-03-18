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

// NomenclatureRepository handles nomenclature database operations
type NomenclatureRepository struct {
	pool *pgxpool.Pool
}

// NewNomenclatureRepository creates a new nomenclature repository
func NewNomenclatureRepository(pool *pgxpool.Pool) *NomenclatureRepository {
	return &NomenclatureRepository{pool: pool}
}

// UpsertTyres inserts new tyres (skips if CAE already exists)
func (r *NomenclatureRepository) UpsertTyres(ctx context.Context, tyres []domain.NomenclatureTyre) (newCount, skippedCount int, err error) {
	if len(tyres) == 0 {
		return 0, 0, nil
	}

	// Bulk upsert using CopyFrom
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Create temporary table
	_, err = tx.Exec(ctx, `
		CREATE TEMP TABLE temp_tyres (LIKE nomenclature_tyres INCLUDING ALL) ON COMMIT DROP
	`)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to create temp table: %w", err)
	}

	// Prepare data for copy
	rows, err := tx.Query(ctx, "SELECT column_name FROM information_schema.columns WHERE table_name = 'temp_tyres' AND column_name NOT IN ('id', 'created_at', 'updated_at') ORDER BY ordinal_position")
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get columns: %w", err)
	}
	defer rows.Close()

	columns := []string{}
	for rows.Next() {
		var col string
		if err := rows.Scan(&col); err != nil {
			return 0, 0, fmt.Errorf("failed to scan column: %w", err)
		}
		columns = append(columns, col)
	}

	// Copy data to temp table
	_, err = tx.CopyFrom(
		ctx,
		pgx.Identifier{"temp_tyres"},
		columns,
		pgx.CopyFromSlice(len(tyres), func(i int) ([]interface{}, error) {
			t := tyres[i]
			return []interface{}{
				t.CAE, t.Name, t.Width, t.Height, t.Diameter,
				t.DiameterOut, t.LoadIndex, t.SpeedIndex, t.Model,
				t.Brand, t.Season, t.IsStudded, t.TireType,
				t.RunFlat, t.Reinforced, 0, // isimport = 0 при загрузке
			}, nil
		}),
	)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to copy data: %w", err)
	}

	// Insert only records with new CAE (artіcle code)
	// Skip if CAE already exists in database
	result, err := tx.Exec(ctx, `
		INSERT INTO nomenclature_tyres (
			cae, name, width, height, diameter, diameter_out,
			load_index, speed_index, model, brand, season,
			is_studded, tiretype, runflat, reinforced, isimport
		)
		SELECT
			t.cae, t.name, t.width, t.height, t.diameter, t.diameter_out,
			t.load_index, t.speed_index, t.model, t.brand, t.season,
			t.is_studded, t.tiretype, t.runflat, t.reinforced, t.isimport
		FROM temp_tyres t
		WHERE NOT EXISTS (
			SELECT 1
			FROM nomenclature_tyres nt
			WHERE nt.cae = t.cae
		)
	`)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to insert data: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return 0, 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Get actual inserted count
	newCount = int(result.RowsAffected())
	skippedCount = len(tyres) - newCount

	return newCount, skippedCount, nil
}

// UpsertRims inserts new rims (skips if CAE already exists)
func (r *NomenclatureRepository) UpsertRims(ctx context.Context, rims []domain.NomenclatureRim) (newCount, skippedCount int, err error) {
	if len(rims) == 0 {
		return 0, 0, nil
	}

	// Bulk upsert using CopyFrom
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Create temporary table
	_, err = tx.Exec(ctx, `
		CREATE TEMP TABLE temp_rims (LIKE nomenclature_rims INCLUDING ALL) ON COMMIT DROP
	`)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to create temp table: %w", err)
	}

	// Prepare columns list
	columns := []string{
		"cae", "name", "width", "diameter", "bolts_count",
		"bolts_spacing", "bolts_spacing2", "et", "dia",
		"model", "brand", "color", "rim_type",
	}

	// Copy data to temp table
	_, err = tx.CopyFrom(
		ctx,
		pgx.Identifier{"temp_rims"},
		columns,
		pgx.CopyFromSlice(len(rims), func(i int) ([]interface{}, error) {
			r := rims[i]
			return []interface{}{
				r.CAE, r.Name, r.Width, r.Diameter, r.BoltsCount,
				r.BoltsSpacing, r.BoltsSpacing2, r.ET, r.DIA,
				r.Model, r.Brand, r.Color, r.RimType,
			}, nil
		}),
	)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to copy data: %w", err)
	}

	// Insert only records with new CAE (article code)
	// Skip if CAE already exists in database
	result, err := tx.Exec(ctx, `
		INSERT INTO nomenclature_rims (
			cae, name, width, diameter, bolts_count,
			bolts_spacing, bolts_spacing2, et, dia,
			model, brand, color, rim_type
		)
		SELECT
			r.cae, r.name, r.width, r.diameter, r.bolts_count,
			r.bolts_spacing, r.bolts_spacing2, r.et, r.dia,
			r.model, r.brand, r.color, r.rim_type
		FROM temp_rims r
		WHERE NOT EXISTS (
			SELECT 1
			FROM nomenclature_rims nr
			WHERE nr.cae = r.cae
		)
	`)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to insert data: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return 0, 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Get actual inserted count
	newCount = int(result.RowsAffected())
	skippedCount = len(rims) - newCount

	return newCount, skippedCount, nil
}

// CreateSyncRun creates a new sync run record
func (r *NomenclatureRepository) CreateSyncRun(ctx context.Context, syncType string) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx, `
		INSERT INTO nomenclature_sync_runs (sync_type, started_at, status)
		VALUES ($1, $2, 'running')
		RETURNING id
	`, syncType, time.Now()).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to create sync run: %w", err)
	}

	return id, nil
}

// UpdateSyncRun updates sync run status
func (r *NomenclatureRepository) UpdateSyncRun(ctx context.Context, id int64, status string, totalItems, newItems, updatedItems int, errorMsg string) error {
	now := time.Now()
	_, err := r.pool.Exec(ctx, `
		UPDATE nomenclature_sync_runs SET
			completed_at = $2,
			status = $3,
			total_items = $4,
			new_items = $5,
			updated_items = $6,
			error_message = $7
		WHERE id = $1
	`, id, now, status, totalItems, newItems, updatedItems, errorMsg)

	if err != nil {
		return fmt.Errorf("failed to update sync run: %w", err)
	}

	return nil
}

// GetTyreByCAE retrieves a tyre by CAE code
func (r *NomenclatureRepository) GetTyreByCAE(ctx context.Context, cae string) (*domain.NomenclatureTyre, error) {
	var tyre domain.NomenclatureTyre
	err := r.pool.QueryRow(ctx, `
		SELECT id, cae, name, width, height, diameter, diameter_out,
		       load_index, speed_index, model, brand, season,
		       is_studded, tiretype, runflat, reinforced, isimport,
		       created_at, updated_at
		FROM nomenclature_tyres
		WHERE cae = $1
	`, cae).Scan(
		&tyre.ID, &tyre.CAE, &tyre.Name, &tyre.Width, &tyre.Height,
		&tyre.Diameter, &tyre.DiameterOut, &tyre.LoadIndex, &tyre.SpeedIndex,
		&tyre.Model, &tyre.Brand, &tyre.Season, &tyre.IsStudded,
		&tyre.TireType, &tyre.RunFlat, &tyre.Reinforced, &tyre.IsImport,
		&tyre.CreatedAt, &tyre.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tyre: %w", err)
	}

	return &tyre, nil
}

// GetRimByCAE retrieves a rim by CAE code
func (r *NomenclatureRepository) GetRimByCAE(ctx context.Context, cae string) (*domain.NomenclatureRim, error) {
	var rim domain.NomenclatureRim
	err := r.pool.QueryRow(ctx, `
		SELECT id, cae, name, width, diameter, bolts_count,
		       bolts_spacing, bolts_spacing2, et, dia,
		       model, brand, color, rim_type,
		       created_at, updated_at
		FROM nomenclature_rims
		WHERE cae = $1
	`, cae).Scan(
		&rim.ID, &rim.CAE, &rim.Name, &rim.Width, &rim.Diameter,
		&rim.BoltsCount, &rim.BoltsSpacing, &rim.BoltsSpacing2,
		&rim.ET, &rim.DIA, &rim.Model, &rim.Brand, &rim.Color,
		&rim.RimType, &rim.CreatedAt, &rim.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get rim: %w", err)
	}

	return &rim, nil
}

// CountTyres returns total tyres count
func (r *NomenclatureRepository) CountTyres(ctx context.Context) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM nomenclature_tyres").Scan(&count)
	return count, err
}

// CountRims returns total rims count
func (r *NomenclatureRepository) CountRims(ctx context.Context) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM nomenclature_rims").Scan(&count)
	return count, err
}

// ====================
// Severavto Repository Methods
// ====================

// UpsertSeveravtoTyres inserts new Severavto tyres (skips if commodity_id already exists)
func (r *NomenclatureRepository) UpsertSeveravtoTyres(ctx context.Context, tyres []severavto.NomenclatureTyre) (newCount, skippedCount int, err error) {
	if len(tyres) == 0 {
		return 0, 0, nil
	}

	// Bulk upsert using CopyFrom
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Create temporary table
	_, err = tx.Exec(ctx, `
		CREATE TEMP TABLE temp_severavto_tyres (LIKE nomenclature_tyres_severavto INCLUDING ALL) ON COMMIT DROP
	`)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to create temp table: %w", err)
	}

	// Prepare columns list
	columns := []string{
		"commodity_id", "article", "name", "brand", "model",
		"width", "height", "diameter", "load_index", "speed_index",
		"season", "is_studded", "tiretype", "runflat", "manufacture_year",
	}

	// Copy data to temp table
	_, err = tx.CopyFrom(
		ctx,
		pgx.Identifier{"temp_severavto_tyres"},
		columns,
		pgx.CopyFromSlice(len(tyres), func(i int) ([]interface{}, error) {
			t := tyres[i]
			return []interface{}{
				t.CommodityID, t.Article, t.Name, t.Brand, t.Model,
				t.Width, t.Height, t.Diameter, t.LoadIndex, t.SpeedIndex,
				t.Season, t.IsStudded, t.TireType, t.RunFlat, t.ManufactureYear,
			}, nil
		}),
	)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to copy data: %w", err)
	}

	// Insert only records with new commodity_id
	// Skip if commodity_id already exists in database
	result, err := tx.Exec(ctx, `
		INSERT INTO nomenclature_tyres_severavto (
			commodity_id, article, name, brand, model,
			width, height, diameter, load_index, speed_index,
			season, is_studded, tiretype, runflat, manufacture_year
		)
		SELECT
			t.commodity_id, t.article, t.name, t.brand, t.model,
			t.width, t.height, t.diameter, t.load_index, t.speed_index,
			t.season, t.is_studded, t.tiretype, t.runflat, t.manufacture_year
		FROM temp_severavto_tyres t
		WHERE NOT EXISTS (
			SELECT 1
			FROM nomenclature_tyres_severavto nt
			WHERE nt.commodity_id = t.commodity_id
		)
	`)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to insert data: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return 0, 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Get actual inserted count
	newCount = int(result.RowsAffected())
	skippedCount = len(tyres) - newCount

	return newCount, skippedCount, nil
}

// UpsertSeveravtoRims inserts new Severavto rims (skips if commodity_id already exists)
func (r *NomenclatureRepository) UpsertSeveravtoRims(ctx context.Context, rims []severavto.NomenclatureRim) (newCount, skippedCount int, err error) {
	if len(rims) == 0 {
		return 0, 0, nil
	}

	// Bulk upsert using CopyFrom
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Create temporary table
	_, err = tx.Exec(ctx, `
		CREATE TEMP TABLE temp_severavto_rims (LIKE nomenclature_rims_severavto INCLUDING ALL) ON COMMIT DROP
	`)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to create temp table: %w", err)
	}

	// Prepare columns list
	columns := []string{
		"commodity_id", "article", "name", "brand", "model",
		"width", "diameter", "bolts_count", "bolts_spacing", "bolts_spacing2",
		"et", "dia", "color", "rim_type", "manufacture_year",
	}

	// Copy data to temp table
	_, err = tx.CopyFrom(
		ctx,
		pgx.Identifier{"temp_severavto_rims"},
		columns,
		pgx.CopyFromSlice(len(rims), func(i int) ([]interface{}, error) {
			rim := rims[i]
			return []interface{}{
				rim.CommodityID, rim.Article, rim.Name, rim.Brand, rim.Model,
				rim.Width, rim.Diameter, rim.BoltsCount, rim.BoltsSpacing, rim.BoltsSpacing2,
				rim.ET, rim.DIA, rim.Color, rim.RimType, rim.ManufactureYear,
			}, nil
		}),
	)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to copy data: %w", err)
	}

	// Insert only records with new commodity_id
	// Skip if commodity_id already exists in database
	result, err := tx.Exec(ctx, `
		INSERT INTO nomenclature_rims_severavto (
			commodity_id, article, name, brand, model,
			width, diameter, bolts_count, bolts_spacing, bolts_spacing2,
			et, dia, color, rim_type, manufacture_year
		)
		SELECT
			r.commodity_id, r.article, r.name, r.brand, r.model,
			r.width, r.diameter, r.bolts_count, r.bolts_spacing, r.bolts_spacing2,
			r.et, r.dia, r.color, r.rim_type, r.manufacture_year
		FROM temp_severavto_rims r
		WHERE NOT EXISTS (
			SELECT 1
			FROM nomenclature_rims_severavto nr
			WHERE nr.commodity_id = r.commodity_id
		)
	`)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to insert data: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return 0, 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Get actual inserted count
	newCount = int(result.RowsAffected())
	skippedCount = len(rims) - newCount

	return newCount, skippedCount, nil
}
