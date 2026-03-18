package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/prokoleso/etalon-price-api/internal/domain"
	"github.com/prokoleso/etalon-price-api/internal/providers"
	"github.com/prokoleso/etalon-price-api/internal/repository/postgres"
)

// SyncService handles synchronization logic
type SyncService struct {
	providerRegistry *providers.Registry
	db               *postgres.DB
	logger           *slog.Logger
}

// SyncServiceConfig holds sync service configuration
type SyncServiceConfig struct {
	ProviderRegistry *providers.Registry
	DB               *postgres.DB
	Logger           *slog.Logger
}

// NewSyncService creates a new sync service
func NewSyncService(cfg SyncServiceConfig) *SyncService {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}

	return &SyncService{
		providerRegistry: cfg.ProviderRegistry,
		db:               cfg.DB,
		logger:           cfg.Logger,
	}
}

// SyncOptions contains options for synchronization
type SyncOptions struct {
	Codes      []string // Specific codes to sync (empty = sync all)
	BatchSize  int      // Batch size for processing
	BatchDelay time.Duration // Delay between batches
}

// SupplierSyncResult contains synchronization results
type SupplierSyncResult struct {
	SyncRunID       int64
	Provider        string
	TotalProducts   int
	SuccessProducts int
	FailedProducts  int
	Duration        time.Duration
	Errors          []error
}

// SyncSupplier synchronizes data from a specific supplier
func (s *SyncService) SyncSupplier(ctx context.Context, providerName string, opts SyncOptions) (*SupplierSyncResult, error) {
	s.logger.Info("Starting supplier sync",
		"provider", providerName,
		"codes_count", len(opts.Codes),
		"batch_size", opts.BatchSize,
	)

	start := time.Now()

	// Get provider
	provider, err := s.providerRegistry.Get(providerName)
	if err != nil {
		return nil, fmt.Errorf("provider not found: %w", err)
	}

	// Create sync run
	syncRunID, err := s.createSyncRun(ctx, providerName)
	if err != nil {
		return nil, fmt.Errorf("failed to create sync run: %w", err)
	}

	result := &SupplierSyncResult{
		SyncRunID: syncRunID,
		Provider:  providerName,
		Errors:    []error{},
	}

	// Determine codes to sync
	codes := opts.Codes
	if len(codes) == 0 {
		s.logger.Info("No specific codes provided, will sync available products")
		codes = []string{}
	}

	// Process in batches
	if err := s.processBatches(ctx, provider, codes, opts, result); err != nil {
		s.logger.Error("Sync failed",
			"provider", providerName,
			"sync_run_id", syncRunID,
			"error", err,
		)

		// Mark sync run as failed
		if failErr := s.failSyncRun(ctx, syncRunID, err.Error()); failErr != nil {
			s.logger.Error("Failed to mark sync run as failed", "error", failErr)
		}

		return result, fmt.Errorf("sync failed: %w", err)
	}

	result.Duration = time.Since(start)

	// Complete sync run
	if err := s.completeSyncRun(ctx, syncRunID, result); err != nil {
		s.logger.Error("Failed to complete sync run", "error", err)
		return result, fmt.Errorf("failed to complete sync run: %w", err)
	}

	s.logger.Info("Supplier sync completed",
		"provider", providerName,
		"sync_run_id", syncRunID,
		"total_products", result.TotalProducts,
		"success_products", result.SuccessProducts,
		"failed_products", result.FailedProducts,
		"duration", result.Duration,
	)

	return result, nil
}

// processBatches processes codes in batches
func (s *SyncService) processBatches(
	ctx context.Context,
	provider providers.SupplierProvider,
	codes []string,
	opts SyncOptions,
	result *SupplierSyncResult,
) error {
	batchSize := opts.BatchSize
	if batchSize <= 0 {
		batchSize = 100
	}
	if batchSize > 200 {
		batchSize = 200 // Maximum batch size
	}

	// If no codes provided, we need at least one batch to fetch available products
	if len(codes) == 0 {
		codes = []string{""} // Empty code to trigger a general fetch
	}

	// Deduplicate codes
	codes = deduplicateCodes(codes)

	// Split into batches
	batches := splitIntoBatches(codes, batchSize)

	s.logger.Info("Processing batches",
		"total_codes", len(codes),
		"batch_count", len(batches),
		"batch_size", batchSize,
	)

	for i, batch := range batches {
		if i > 0 && opts.BatchDelay > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(opts.BatchDelay):
			}
		}

		s.logger.Debug("Processing batch",
			"batch_number", i+1,
			"batch_size", len(batch),
		)

		// Filter empty codes for actual API calls
		filteredBatch := filterEmptyCodes(batch)
		if len(filteredBatch) == 0 {
			continue
		}

		if err := s.processBatch(ctx, provider, filteredBatch, result); err != nil {
			s.logger.Error("Batch processing failed",
				"batch_number", i+1,
				"error", err,
			)
			result.Errors = append(result.Errors, err)
			// Continue with next batch even if this one fails
		}
	}

	return nil
}

// processBatch processes a single batch
func (s *SyncService) processBatch(
	ctx context.Context,
	provider providers.SupplierProvider,
	codes []string,
	result *SupplierSyncResult,
) error {
	// Fetch goods info
	goodsResult, err := provider.FetchGoodsInfo(ctx, codes)
	if err != nil {
		return fmt.Errorf("failed to fetch goods info: %w", err)
	}

	// Fetch prices and stocks (continue even if this fails)
	pricesResult, err := provider.FetchPricesAndStocks(ctx, codes)
	if err != nil {
		s.logger.Warn("Failed to fetch prices and stocks, continuing with goods info only",
			"error", err,
			"codes_count", len(codes),
		)
		// Create empty prices result to avoid nil pointer
		pricesResult = &providers.PriceStockResult{
			Offers:     []domain.ProductOffer{},
			FetchedAt:  time.Now(),
			TotalCount: 0,
		}
	}

	// Save to database in transaction
	err = s.db.WithTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// Save tyres
		if len(goodsResult.Tyres) > 0 {
			repo := postgres.NewTyresRepository(tx)
			if err := repo.UpsertTyres(ctx, goodsResult.Tyres); err != nil {
				return fmt.Errorf("failed to upsert tyres: %w", err)
			}
			result.SuccessProducts += len(goodsResult.Tyres)
		}

		// Save rims
		if len(goodsResult.Rims) > 0 {
			repo := postgres.NewRimsRepository(tx)
			if err := repo.UpsertRims(ctx, goodsResult.Rims); err != nil {
				return fmt.Errorf("failed to upsert rims: %w", err)
			}
			result.SuccessProducts += len(goodsResult.Rims)
		}

		// Save offers
		if len(pricesResult.Offers) > 0 {
			repo := postgres.NewOffersRepository(tx)
			if err := repo.UpsertOffers(ctx, pricesResult.Offers); err != nil {
				return fmt.Errorf("failed to upsert offers: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		result.FailedProducts += len(goodsResult.Tyres) + len(goodsResult.Rims)
		return err
	}

	result.TotalProducts += len(goodsResult.Tyres) + len(goodsResult.Rims)

	return nil
}

// createSyncRun creates a new sync run record
func (s *SyncService) createSyncRun(ctx context.Context, provider string) (int64, error) {
	repo := postgres.NewSyncRunsRepository(s.db.Pool())

	metadata := domain.JSONB{
		"started_by": "system",
		"version":    "1.0",
	}

	return repo.Create(ctx, provider, metadata)
}

// completeSyncRun marks sync run as completed
func (s *SyncService) completeSyncRun(ctx context.Context, syncRunID int64, result *SupplierSyncResult) error {
	repo := postgres.NewSyncRunsRepository(s.db.Pool())
	return repo.Complete(ctx, syncRunID, result.TotalProducts, result.SuccessProducts, result.FailedProducts)
}

// failSyncRun marks sync run as failed
func (s *SyncService) failSyncRun(ctx context.Context, syncRunID int64, errorMessage string) error {
	repo := postgres.NewSyncRunsRepository(s.db.Pool())
	return repo.Fail(ctx, syncRunID, errorMessage)
}

// deduplicateCodes removes duplicate codes
func deduplicateCodes(codes []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, code := range codes {
		if !seen[code] {
			seen[code] = true
			result = append(result, code)
		}
	}

	return result
}

// splitIntoBatches splits codes into batches
func splitIntoBatches(codes []string, batchSize int) [][]string {
	var batches [][]string

	for i := 0; i < len(codes); i += batchSize {
		end := i + batchSize
		if end > len(codes) {
			end = len(codes)
		}
		batches = append(batches, codes[i:end])
	}

	return batches
}

// filterEmptyCodes removes empty strings from codes
func filterEmptyCodes(codes []string) []string {
	result := []string{}
	for _, code := range codes {
		if code != "" {
			result = append(result, code)
		}
	}
	return result
}
