package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/prokoleso/etalon-price-api/internal/domain"
	fourtochki "github.com/prokoleso/etalon-price-api/internal/providers/4tochki"
	"github.com/prokoleso/etalon-price-api/internal/repository"
)

// PricesStockService handles prices and stock synchronization
type PricesStockService struct {
	repo          *repository.PricesStockRepository
	warehouseRepo *repository.WarehouseRepository
	fourtochkiAPI *fourtochki.Client
	logger        *slog.Logger
}

// NewPricesStockService creates a new prices stock service
func NewPricesStockService(
	repo *repository.PricesStockRepository,
	warehouseRepo *repository.WarehouseRepository,
	fourtochkiAPI *fourtochki.Client,
	logger *slog.Logger,
) *PricesStockService {
	return &PricesStockService{
		repo:          repo,
		warehouseRepo: warehouseRepo,
		fourtochkiAPI: fourtochkiAPI,
		logger:        logger,
	}
}

// SyncTyresPricesStock downloads and syncs tyres prices and stock
func (s *PricesStockService) SyncTyresPricesStock(ctx context.Context) error {
	s.logger.Info("Starting tyres prices/stock sync")

	// Create sync run
	runID, err := s.repo.CreatePricesStockSyncRun(ctx, "tyres")
	if err != nil {
		return fmt.Errorf("failed to create sync run: %w", err)
	}

	// Fetch all tyres with prices via API (with pagination)
	allTyres := []fourtochki.TyrePriceRest{}
	pageSize := 2000
	page := 0

	for {
		s.logger.Info("Fetching tyres from API", "page", page, "page_size", pageSize)

		resp, err := s.fourtochkiAPI.GetFindTyre(ctx, page, pageSize)
		if err != nil {
			s.repo.UpdatePricesStockSyncRun(ctx, runID, "failed", 0, 0, 0, 0, err.Error())
			return fmt.Errorf("failed to fetch tyres from API: %w", err)
		}

		if resp.Result.Error != nil && resp.Result.Error.Code != 0 {
			errMsg := fmt.Sprintf("API error: code=%d, msg=%s", resp.Result.Error.Code, resp.Result.Error.Comment)
			s.repo.UpdatePricesStockSyncRun(ctx, runID, "failed", 0, 0, 0, 0, errMsg)
			return fmt.Errorf(errMsg)
		}

		allTyres = append(allTyres, resp.Result.Items...)

		s.logger.Info("Fetched tyres page", "page", page, "items", len(resp.Result.Items), "total_so_far", len(allTyres))

		// If we got less than pageSize, we've reached the end
		if len(resp.Result.Items) < pageSize {
			break
		}

		page++
	}

	s.logger.Info("Fetched all tyres from API", "total", len(allTyres))

	// Get warehouse ID to name mapping
	warehouses, err := s.warehouseRepo.GetAllWarehouses(ctx)
	if err != nil {
		s.repo.UpdatePricesStockSyncRun(ctx, runID, "failed", 0, 0, 0, 0, err.Error())
		return fmt.Errorf("failed to get warehouses: %w", err)
	}
	warehouseMap := make(map[int]string)
	for _, w := range warehouses {
		warehouseMap[w.ID] = w.Name
	}

	// Convert to domain models (flatten warehouse data)
	pricesStock := []domain.TyrePriceStock{}
	for _, tyre := range allTyres {
		for _, wh := range tyre.Whpr {
			warehouseName, ok := warehouseMap[wh.Wrh]
			if !ok {
				s.logger.Warn("Unknown warehouse ID", "warehouse_id", wh.Wrh)
				continue // Skip unknown warehouses
			}
			pricesStock = append(pricesStock, domain.TyrePriceStock{
				CAE:           tyre.Code,
				WarehouseName: warehouseName,
				Price:         wh.Price,
				Stock:         wh.Rest,
				IsImport:      0, // 0 - loaded
				Provider:      "Форточки",
			})
		}
	}

	s.logger.Info("Converted to warehouse records", "total_warehouses", len(pricesStock))

	// Save to database in batches
	batchSize := 5000
	totalNew := 0
	totalUpdated := 0

	for i := 0; i < len(pricesStock); i += batchSize {
		end := i + batchSize
		if end > len(pricesStock) {
			end = len(pricesStock)
		}

		batch := pricesStock[i:end]
		newCount, updatedCount, err := s.repo.UpsertTyresPricesStock(ctx, batch)
		if err != nil {
			s.repo.UpdatePricesStockSyncRun(ctx, runID, "failed", len(allTyres), len(pricesStock), totalNew, totalUpdated, err.Error())
			return fmt.Errorf("failed to save prices/stock batch: %w", err)
		}

		totalNew += newCount
		totalUpdated += updatedCount

		s.logger.Info("Saved prices/stock batch",
			"batch", fmt.Sprintf("%d-%d", i, end),
			"new", newCount,
			"updated", updatedCount,
		)
	}

	// Update sync run
	if err := s.repo.UpdatePricesStockSyncRun(ctx, runID, "completed", len(allTyres), len(pricesStock), totalNew, totalUpdated, ""); err != nil {
		return fmt.Errorf("failed to update sync run: %w", err)
	}

	s.logger.Info("Tyres prices/stock sync completed",
		"total_products", len(allTyres),
		"total_warehouses", len(pricesStock),
		"new", totalNew,
		"updated", totalUpdated,
	)

	return nil
}

// SyncRimsPricesStock downloads and syncs rims prices and stock
func (s *PricesStockService) SyncRimsPricesStock(ctx context.Context) error {
	s.logger.Info("Starting rims prices/stock sync")

	// Create sync run
	runID, err := s.repo.CreatePricesStockSyncRun(ctx, "rims")
	if err != nil {
		return fmt.Errorf("failed to create sync run: %w", err)
	}

	// Fetch all rims with prices via API (with pagination)
	allRims := []fourtochki.DiskPriceRest{}
	pageSize := 2000
	page := 0

	for {
		s.logger.Info("Fetching rims from API", "page", page, "page_size", pageSize)

		resp, err := s.fourtochkiAPI.GetFindDisk(ctx, page, pageSize)
		if err != nil {
			s.repo.UpdatePricesStockSyncRun(ctx, runID, "failed", 0, 0, 0, 0, err.Error())
			return fmt.Errorf("failed to fetch rims from API: %w", err)
		}

		if resp.Result.Error != nil && resp.Result.Error.Code != 0 {
			errMsg := fmt.Sprintf("API error: code=%d, msg=%s", resp.Result.Error.Code, resp.Result.Error.Comment)
			s.repo.UpdatePricesStockSyncRun(ctx, runID, "failed", 0, 0, 0, 0, errMsg)
			return fmt.Errorf(errMsg)
		}

		allRims = append(allRims, resp.Result.Items...)

		s.logger.Info("Fetched rims page", "page", page, "items", len(resp.Result.Items), "total_so_far", len(allRims))

		// If we got less than pageSize, we've reached the end
		if len(resp.Result.Items) < pageSize {
			break
		}

		page++
	}

	s.logger.Info("Fetched all rims from API", "total", len(allRims))

	// Get warehouse ID to name mapping
	warehouses, err := s.warehouseRepo.GetAllWarehouses(ctx)
	if err != nil {
		s.repo.UpdatePricesStockSyncRun(ctx, runID, "failed", 0, 0, 0, 0, err.Error())
		return fmt.Errorf("failed to get warehouses: %w", err)
	}
	warehouseMap := make(map[int]string)
	for _, w := range warehouses {
		warehouseMap[w.ID] = w.Name
	}

	// Convert to domain models (flatten warehouse data)
	pricesStock := []domain.RimPriceStock{}
	for _, rim := range allRims {
		for _, wh := range rim.Whpr {
			warehouseName, ok := warehouseMap[wh.Wrh]
			if !ok {
				s.logger.Warn("Unknown warehouse ID", "warehouse_id", wh.Wrh)
				continue // Skip unknown warehouses
			}
			pricesStock = append(pricesStock, domain.RimPriceStock{
				CAE:           rim.Code,
				WarehouseName: warehouseName,
				Price:         wh.Price,
				Stock:         wh.Rest,
				IsImport:      0, // 0 - loaded
				Provider:      "Форточки",
			})
		}
	}

	s.logger.Info("Converted to warehouse records", "total_warehouses", len(pricesStock))

	// Save to database in batches
	batchSize := 5000
	totalNew := 0
	totalUpdated := 0

	for i := 0; i < len(pricesStock); i += batchSize {
		end := i + batchSize
		if end > len(pricesStock) {
			end = len(pricesStock)
		}

		batch := pricesStock[i:end]
		newCount, updatedCount, err := s.repo.UpsertRimsPricesStock(ctx, batch)
		if err != nil {
			s.repo.UpdatePricesStockSyncRun(ctx, runID, "failed", len(allRims), len(pricesStock), totalNew, totalUpdated, err.Error())
			return fmt.Errorf("failed to save prices/stock batch: %w", err)
		}

		totalNew += newCount
		totalUpdated += updatedCount

		s.logger.Info("Saved prices/stock batch",
			"batch", fmt.Sprintf("%d-%d", i, end),
			"new", newCount,
			"updated", updatedCount,
		)
	}

	// Update sync run
	if err := s.repo.UpdatePricesStockSyncRun(ctx, runID, "completed", len(allRims), len(pricesStock), totalNew, totalUpdated, ""); err != nil {
		return fmt.Errorf("failed to update sync run: %w", err)
	}

	s.logger.Info("Rims prices/stock sync completed",
		"total_products", len(allRims),
		"total_warehouses", len(pricesStock),
		"new", totalNew,
		"updated", totalUpdated,
	)

	return nil
}

// SyncAll syncs both tyres and rims prices/stock
func (s *PricesStockService) SyncAll(ctx context.Context) error {
	if err := s.SyncTyresPricesStock(ctx); err != nil {
		return fmt.Errorf("failed to sync tyres prices/stock: %w", err)
	}

	if err := s.SyncRimsPricesStock(ctx); err != nil {
		return fmt.Errorf("failed to sync rims prices/stock: %w", err)
	}

	return nil
}
