package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/prokoleso/etalon-price-api/internal/domain"
)

// TyresRepository handles tyre database operations
type TyresRepository struct {
	db DBTX
}

// NewTyresRepository creates a new tyres repository
func NewTyresRepository(db DBTX) *TyresRepository {
	return &TyresRepository{db: db}
}

// UpsertTyres inserts or updates multiple tyres
func (r *TyresRepository) UpsertTyres(ctx context.Context, tyres []domain.Tyre) error {
	if len(tyres) == 0 {
		return nil
	}

	query := `
		INSERT INTO tyres_api (
			code, supplier, brand, model, width, height, diameter,
			load_index, speed_index, season, run_flat, studded,
			description, raw_payload, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, NOW(), NOW()
		)
		ON CONFLICT (code, supplier)
		DO UPDATE SET
			brand = EXCLUDED.brand,
			model = EXCLUDED.model,
			width = EXCLUDED.width,
			height = EXCLUDED.height,
			diameter = EXCLUDED.diameter,
			load_index = EXCLUDED.load_index,
			speed_index = EXCLUDED.speed_index,
			season = EXCLUDED.season,
			run_flat = EXCLUDED.run_flat,
			studded = EXCLUDED.studded,
			description = EXCLUDED.description,
			raw_payload = EXCLUDED.raw_payload,
			updated_at = NOW()
	`

	batch := &pgx.Batch{}
	for _, tyre := range tyres {
		batch.Queue(
			query,
			tyre.Code,
			tyre.Supplier,
			tyre.Brand,
			tyre.Model,
			tyre.Width,
			tyre.Height,
			tyre.Diameter,
			tyre.LoadIndex,
			tyre.SpeedIndex,
			tyre.Season,
			tyre.RunFlat,
			tyre.Studded,
			tyre.Description,
			tyre.RawPayload,
		)
	}

	br := r.db.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < len(tyres); i++ {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("failed to upsert tyre %d: %w", i, err)
		}
	}

	return nil
}

// GetByCode retrieves a tyre by code and supplier
func (r *TyresRepository) GetByCode(ctx context.Context, code, supplier string) (*domain.Tyre, error) {
	query := `
		SELECT id, code, supplier, brand, model, width, height, diameter,
			   load_index, speed_index, season, run_flat, studded,
			   description, raw_payload, created_at, updated_at
		FROM tyres_api
		WHERE code = $1 AND supplier = $2
	`

	var tyre domain.Tyre
	err := r.db.QueryRow(ctx, query, code, supplier).Scan(
		&tyre.ID,
		&tyre.Code,
		&tyre.Supplier,
		&tyre.Brand,
		&tyre.Model,
		&tyre.Width,
		&tyre.Height,
		&tyre.Diameter,
		&tyre.LoadIndex,
		&tyre.SpeedIndex,
		&tyre.Season,
		&tyre.RunFlat,
		&tyre.Studded,
		&tyre.Description,
		&tyre.RawPayload,
		&tyre.CreatedAt,
		&tyre.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get tyre: %w", err)
	}

	return &tyre, nil
}

// List retrieves tyres with pagination
func (r *TyresRepository) List(ctx context.Context, supplier string, limit, offset int) ([]domain.Tyre, error) {
	query := `
		SELECT id, code, supplier, brand, model, width, height, diameter,
			   load_index, speed_index, season, run_flat, studded,
			   description, raw_payload, created_at, updated_at
		FROM tyres_api
		WHERE supplier = $1
		ORDER BY updated_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, supplier, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list tyres: %w", err)
	}
	defer rows.Close()

	var tyres []domain.Tyre
	for rows.Next() {
		var tyre domain.Tyre
		err := rows.Scan(
			&tyre.ID,
			&tyre.Code,
			&tyre.Supplier,
			&tyre.Brand,
			&tyre.Model,
			&tyre.Width,
			&tyre.Height,
			&tyre.Diameter,
			&tyre.LoadIndex,
			&tyre.SpeedIndex,
			&tyre.Season,
			&tyre.RunFlat,
			&tyre.Studded,
			&tyre.Description,
			&tyre.RawPayload,
			&tyre.CreatedAt,
			&tyre.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tyre: %w", err)
		}
		tyres = append(tyres, tyre)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return tyres, nil
}

// Count returns total number of tyres for a supplier
func (r *TyresRepository) Count(ctx context.Context, supplier string) (int64, error) {
	query := `SELECT COUNT(*) FROM tyres_api WHERE supplier = $1`

	var count int64
	err := r.db.QueryRow(ctx, query, supplier).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count tyres: %w", err)
	}

	return count, nil
}
