package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/prokoleso/etalon-price-api/internal/domain"
)

// OffersRepository handles product offers database operations
type OffersRepository struct {
	db DBTX
}

// NewOffersRepository creates a new offers repository
func NewOffersRepository(db DBTX) *OffersRepository {
	return &OffersRepository{db: db}
}

// UpsertOffers inserts or updates multiple product offers
func (r *OffersRepository) UpsertOffers(ctx context.Context, offers []domain.ProductOffer) error {
	if len(offers) == 0 {
		return nil
	}

	query := `
		INSERT INTO product_offers_api (
			code, supplier, product_type, price, currency, stock,
			warehouse_code, warehouse_name, delivery_days, min_order_qty,
			available_for_order, raw_payload, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW(), NOW()
		)
		ON CONFLICT (code, supplier, warehouse_code)
		DO UPDATE SET
			product_type = EXCLUDED.product_type,
			price = EXCLUDED.price,
			currency = EXCLUDED.currency,
			stock = EXCLUDED.stock,
			warehouse_name = EXCLUDED.warehouse_name,
			delivery_days = EXCLUDED.delivery_days,
			min_order_qty = EXCLUDED.min_order_qty,
			available_for_order = EXCLUDED.available_for_order,
			raw_payload = EXCLUDED.raw_payload,
			updated_at = NOW()
	`

	batch := &pgx.Batch{}
	for _, offer := range offers {
		batch.Queue(
			query,
			offer.Code,
			offer.Supplier,
			offer.ProductType,
			offer.Price,
			offer.Currency,
			offer.Stock,
			offer.WarehouseCode,
			offer.WarehouseName,
			offer.DeliveryDays,
			offer.MinOrderQty,
			offer.AvailableForOrder,
			offer.RawPayload,
		)
	}

	br := r.db.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < len(offers); i++ {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("failed to upsert offer %d: %w", i, err)
		}
	}

	return nil
}

// GetByCode retrieves all offers for a product code from a supplier
func (r *OffersRepository) GetByCode(ctx context.Context, code, supplier string) ([]domain.ProductOffer, error) {
	query := `
		SELECT id, code, supplier, product_type, price, currency, stock,
			   warehouse_code, warehouse_name, delivery_days, min_order_qty,
			   available_for_order, raw_payload, created_at, updated_at
		FROM product_offers_api
		WHERE code = $1 AND supplier = $2
		ORDER BY price ASC
	`

	rows, err := r.db.Query(ctx, query, code, supplier)
	if err != nil {
		return nil, fmt.Errorf("failed to get offers: %w", err)
	}
	defer rows.Close()

	var offers []domain.ProductOffer
	for rows.Next() {
		var offer domain.ProductOffer
		err := rows.Scan(
			&offer.ID,
			&offer.Code,
			&offer.Supplier,
			&offer.ProductType,
			&offer.Price,
			&offer.Currency,
			&offer.Stock,
			&offer.WarehouseCode,
			&offer.WarehouseName,
			&offer.DeliveryDays,
			&offer.MinOrderQty,
			&offer.AvailableForOrder,
			&offer.RawPayload,
			&offer.CreatedAt,
			&offer.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan offer: %w", err)
		}
		offers = append(offers, offer)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return offers, nil
}

// List retrieves offers with pagination
func (r *OffersRepository) List(ctx context.Context, supplier string, limit, offset int) ([]domain.ProductOffer, error) {
	query := `
		SELECT id, code, supplier, product_type, price, currency, stock,
			   warehouse_code, warehouse_name, delivery_days, min_order_qty,
			   available_for_order, raw_payload, created_at, updated_at
		FROM product_offers_api
		WHERE supplier = $1
		ORDER BY updated_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, supplier, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list offers: %w", err)
	}
	defer rows.Close()

	var offers []domain.ProductOffer
	for rows.Next() {
		var offer domain.ProductOffer
		err := rows.Scan(
			&offer.ID,
			&offer.Code,
			&offer.Supplier,
			&offer.ProductType,
			&offer.Price,
			&offer.Currency,
			&offer.Stock,
			&offer.WarehouseCode,
			&offer.WarehouseName,
			&offer.DeliveryDays,
			&offer.MinOrderQty,
			&offer.AvailableForOrder,
			&offer.RawPayload,
			&offer.CreatedAt,
			&offer.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan offer: %w", err)
		}
		offers = append(offers, offer)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return offers, nil
}

// Count returns total number of offers for a supplier
func (r *OffersRepository) Count(ctx context.Context, supplier string) (int64, error) {
	query := `SELECT COUNT(*) FROM product_offers_api WHERE supplier = $1`

	var count int64
	err := r.db.QueryRow(ctx, query, supplier).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count offers: %w", err)
	}

	return count, nil
}

// DeleteOldOffers deletes offers older than the specified time
func (r *OffersRepository) DeleteOldOffers(ctx context.Context, supplier string, olderThan string) (int64, error) {
	query := `
		DELETE FROM product_offers_api
		WHERE supplier = $1 AND updated_at < NOW() - $2::interval
	`

	tag, err := r.db.Exec(ctx, query, supplier, olderThan)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old offers: %w", err)
	}

	return tag.RowsAffected(), nil
}
