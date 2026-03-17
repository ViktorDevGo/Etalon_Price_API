package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/prokoleso/etalon-price-api/internal/domain"
)

// SyncRunsRepository handles sync runs database operations
type SyncRunsRepository struct {
	db DBTX
}

// NewSyncRunsRepository creates a new sync runs repository
func NewSyncRunsRepository(db DBTX) *SyncRunsRepository {
	return &SyncRunsRepository{db: db}
}

// Create creates a new sync run
func (r *SyncRunsRepository) Create(ctx context.Context, provider string, metadata domain.JSONB) (int64, error) {
	query := `
		INSERT INTO sync_runs_api (
			provider, status, started_at, metadata, created_at
		) VALUES (
			$1, $2, NOW(), $3, NOW()
		)
		RETURNING id
	`

	var id int64
	err := r.db.QueryRow(ctx, query, provider, domain.SyncStatusRunning, metadata).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create sync run: %w", err)
	}

	return id, nil
}

// UpdateProgress updates sync run progress
func (r *SyncRunsRepository) UpdateProgress(ctx context.Context, id int64, totalProducts, successProducts, failedProducts int) error {
	query := `
		UPDATE sync_runs_api
		SET total_products = $2,
			success_products = $3,
			failed_products = $4
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, id, totalProducts, successProducts, failedProducts)
	if err != nil {
		return fmt.Errorf("failed to update sync run progress: %w", err)
	}

	return nil
}

// Complete marks sync run as completed
func (r *SyncRunsRepository) Complete(ctx context.Context, id int64, totalProducts, successProducts, failedProducts int) error {
	query := `
		UPDATE sync_runs_api
		SET status = $2,
			completed_at = NOW(),
			total_products = $3,
			success_products = $4,
			failed_products = $5
		WHERE id = $1
	`

	_, err := r.db.Exec(
		ctx,
		query,
		id,
		domain.SyncStatusCompleted,
		totalProducts,
		successProducts,
		failedProducts,
	)
	if err != nil {
		return fmt.Errorf("failed to complete sync run: %w", err)
	}

	return nil
}

// Fail marks sync run as failed
func (r *SyncRunsRepository) Fail(ctx context.Context, id int64, errorMessage string) error {
	query := `
		UPDATE sync_runs_api
		SET status = $2,
			completed_at = NOW(),
			error_message = $3
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, id, domain.SyncStatusFailed, errorMessage)
	if err != nil {
		return fmt.Errorf("failed to mark sync run as failed: %w", err)
	}

	return nil
}

// GetByID retrieves a sync run by ID
func (r *SyncRunsRepository) GetByID(ctx context.Context, id int64) (*domain.SyncRun, error) {
	query := `
		SELECT id, provider, status, started_at, completed_at,
			   total_products, success_products, failed_products,
			   error_message, metadata, created_at
		FROM sync_runs_api
		WHERE id = $1
	`

	var syncRun domain.SyncRun
	err := r.db.QueryRow(ctx, query, id).Scan(
		&syncRun.ID,
		&syncRun.Provider,
		&syncRun.Status,
		&syncRun.StartedAt,
		&syncRun.CompletedAt,
		&syncRun.TotalProducts,
		&syncRun.SuccessProducts,
		&syncRun.FailedProducts,
		&syncRun.ErrorMessage,
		&syncRun.Metadata,
		&syncRun.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get sync run: %w", err)
	}

	return &syncRun, nil
}

// List retrieves sync runs with pagination
func (r *SyncRunsRepository) List(ctx context.Context, provider string, limit, offset int) ([]domain.SyncRun, error) {
	query := `
		SELECT id, provider, status, started_at, completed_at,
			   total_products, success_products, failed_products,
			   error_message, metadata, created_at
		FROM sync_runs_api
		WHERE provider = $1 OR $1 = ''
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, provider, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list sync runs: %w", err)
	}
	defer rows.Close()

	var syncRuns []domain.SyncRun
	for rows.Next() {
		var syncRun domain.SyncRun
		err := rows.Scan(
			&syncRun.ID,
			&syncRun.Provider,
			&syncRun.Status,
			&syncRun.StartedAt,
			&syncRun.CompletedAt,
			&syncRun.TotalProducts,
			&syncRun.SuccessProducts,
			&syncRun.FailedProducts,
			&syncRun.ErrorMessage,
			&syncRun.Metadata,
			&syncRun.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan sync run: %w", err)
		}
		syncRuns = append(syncRuns, syncRun)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return syncRuns, nil
}

// GetLatest retrieves the latest sync run for a provider
func (r *SyncRunsRepository) GetLatest(ctx context.Context, provider string) (*domain.SyncRun, error) {
	query := `
		SELECT id, provider, status, started_at, completed_at,
			   total_products, success_products, failed_products,
			   error_message, metadata, created_at
		FROM sync_runs_api
		WHERE provider = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	var syncRun domain.SyncRun
	err := r.db.QueryRow(ctx, query, provider).Scan(
		&syncRun.ID,
		&syncRun.Provider,
		&syncRun.Status,
		&syncRun.StartedAt,
		&syncRun.CompletedAt,
		&syncRun.TotalProducts,
		&syncRun.SuccessProducts,
		&syncRun.FailedProducts,
		&syncRun.ErrorMessage,
		&syncRun.Metadata,
		&syncRun.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get latest sync run: %w", err)
	}

	return &syncRun, nil
}

// Count returns total number of sync runs for a provider
func (r *SyncRunsRepository) Count(ctx context.Context, provider string) (int64, error) {
	query := `SELECT COUNT(*) FROM sync_runs_api WHERE provider = $1 OR $1 = ''`

	var count int64
	err := r.db.QueryRow(ctx, query, provider).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count sync runs: %w", err)
	}

	return count, nil
}

// CleanupOld deletes sync runs older than the specified duration
func (r *SyncRunsRepository) CleanupOld(ctx context.Context, olderThan time.Duration) (int64, error) {
	query := `
		DELETE FROM sync_runs_api
		WHERE created_at < NOW() - $1::interval
	`

	tag, err := r.db.Exec(ctx, query, olderThan.String())
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old sync runs: %w", err)
	}

	return tag.RowsAffected(), nil
}
