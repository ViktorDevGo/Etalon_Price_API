package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/prokoleso/etalon-price-api/internal/domain"
)

// RimsRepository handles rim database operations
type RimsRepository struct {
	db DBTX
}

// NewRimsRepository creates a new rims repository
func NewRimsRepository(db DBTX) *RimsRepository {
	return &RimsRepository{db: db}
}

// UpsertRims inserts or updates multiple rims
func (r *RimsRepository) UpsertRims(ctx context.Context, rims []domain.Rim) error {
	if len(rims) == 0 {
		return nil
	}

	query := `
		INSERT INTO rims_api (
			code, supplier, brand, model, width, diameter,
			bolt_pattern, offset, center_bore, color,
			description, raw_payload, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW(), NOW()
		)
		ON CONFLICT (code, supplier)
		DO UPDATE SET
			brand = EXCLUDED.brand,
			model = EXCLUDED.model,
			width = EXCLUDED.width,
			diameter = EXCLUDED.diameter,
			bolt_pattern = EXCLUDED.bolt_pattern,
			offset = EXCLUDED.offset,
			center_bore = EXCLUDED.center_bore,
			color = EXCLUDED.color,
			description = EXCLUDED.description,
			raw_payload = EXCLUDED.raw_payload,
			updated_at = NOW()
	`

	batch := &pgx.Batch{}
	for _, rim := range rims {
		batch.Queue(
			query,
			rim.Code,
			rim.Supplier,
			rim.Brand,
			rim.Model,
			rim.Width,
			rim.Diameter,
			rim.BoltPattern,
			rim.Offset,
			rim.CenterBore,
			rim.Color,
			rim.Description,
			rim.RawPayload,
		)
	}

	br := r.db.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < len(rims); i++ {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("failed to upsert rim %d: %w", i, err)
		}
	}

	return nil
}

// GetByCode retrieves a rim by code and supplier
func (r *RimsRepository) GetByCode(ctx context.Context, code, supplier string) (*domain.Rim, error) {
	query := `
		SELECT id, code, supplier, brand, model, width, diameter,
			   bolt_pattern, offset, center_bore, color,
			   description, raw_payload, created_at, updated_at
		FROM rims_api
		WHERE code = $1 AND supplier = $2
	`

	var rim domain.Rim
	err := r.db.QueryRow(ctx, query, code, supplier).Scan(
		&rim.ID,
		&rim.Code,
		&rim.Supplier,
		&rim.Brand,
		&rim.Model,
		&rim.Width,
		&rim.Diameter,
		&rim.BoltPattern,
		&rim.Offset,
		&rim.CenterBore,
		&rim.Color,
		&rim.Description,
		&rim.RawPayload,
		&rim.CreatedAt,
		&rim.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get rim: %w", err)
	}

	return &rim, nil
}

// List retrieves rims with pagination
func (r *RimsRepository) List(ctx context.Context, supplier string, limit, offset int) ([]domain.Rim, error) {
	query := `
		SELECT id, code, supplier, brand, model, width, diameter,
			   bolt_pattern, offset, center_bore, color,
			   description, raw_payload, created_at, updated_at
		FROM rims_api
		WHERE supplier = $1
		ORDER BY updated_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, supplier, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list rims: %w", err)
	}
	defer rows.Close()

	var rims []domain.Rim
	for rows.Next() {
		var rim domain.Rim
		err := rows.Scan(
			&rim.ID,
			&rim.Code,
			&rim.Supplier,
			&rim.Brand,
			&rim.Model,
			&rim.Width,
			&rim.Diameter,
			&rim.BoltPattern,
			&rim.Offset,
			&rim.CenterBore,
			&rim.Color,
			&rim.Description,
			&rim.RawPayload,
			&rim.CreatedAt,
			&rim.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan rim: %w", err)
		}
		rims = append(rims, rim)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return rims, nil
}

// Count returns total number of rims for a supplier
func (r *RimsRepository) Count(ctx context.Context, supplier string) (int64, error) {
	query := `SELECT COUNT(*) FROM rims_api WHERE supplier = $1`

	var count int64
	err := r.db.QueryRow(ctx, query, supplier).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count rims: %w", err)
	}

	return count, nil
}
