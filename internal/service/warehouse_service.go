package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/prokoleso/etalon-price-api/internal/domain"
	fourtochki "github.com/prokoleso/etalon-price-api/internal/providers/4tochki"
	"github.com/prokoleso/etalon-price-api/internal/repository"
)

// WarehouseService handles warehouse synchronization
type WarehouseService struct {
	repo          *repository.WarehouseRepository
	fourtochkiAPI *fourtochki.Client
	logger        *slog.Logger
}

// NewWarehouseService creates a new warehouse service
func NewWarehouseService(
	repo *repository.WarehouseRepository,
	fourtochkiAPI *fourtochki.Client,
	logger *slog.Logger,
) *WarehouseService {
	return &WarehouseService{
		repo:          repo,
		fourtochkiAPI: fourtochkiAPI,
		logger:        logger,
	}
}

// SyncWarehouses downloads and syncs warehouses from API
func (s *WarehouseService) SyncWarehouses(ctx context.Context) error {
	s.logger.Info("Starting warehouses sync")

	// Create sync run
	runID, err := s.repo.CreateWarehousesSyncRun(ctx)
	if err != nil {
		return fmt.Errorf("failed to create sync run: %w", err)
	}

	// Fetch warehouses from API
	s.logger.Info("Fetching warehouses from API")
	resp, err := s.fourtochkiAPI.GetWarehouses(ctx)
	if err != nil {
		s.repo.UpdateWarehousesSyncRun(ctx, runID, "failed", 0, 0, 0, err.Error())
		return fmt.Errorf("failed to fetch warehouses from API: %w", err)
	}

	if resp.Result.Error != nil && resp.Result.Error.Code != 0 {
		errMsg := fmt.Sprintf("API error: code=%d, msg=%s", resp.Result.Error.Code, resp.Result.Error.Comment)
		s.repo.UpdateWarehousesSyncRun(ctx, runID, "failed", 0, 0, 0, errMsg)
		return fmt.Errorf(errMsg)
	}

	if !resp.Result.Success {
		errMsg := "API returned success=false"
		s.repo.UpdateWarehousesSyncRun(ctx, runID, "failed", 0, 0, 0, errMsg)
		return fmt.Errorf(errMsg)
	}

	s.logger.Info("Fetched warehouses from API", "count", len(resp.Result.Warehouses))

	// Convert to domain models
	warehouses := make([]domain.Warehouse, 0, len(resp.Result.Warehouses))
	for _, w := range resp.Result.Warehouses {
		warehouses = append(warehouses, domain.Warehouse{
			ID:              w.ID,
			Key:             w.Key,
			Name:            w.Name,
			ShortName:       w.ShortName,
			StoID:           w.StoID,
			BackgroundColor: w.BackgroundColor,
			HaveDelivery:    w.HaveDelivery,
			HavePickup:      w.HavePickup,
			IsPaidDelivery:  w.IsPaidDelivery,
			LogisticDays:    w.LogisticDays,
		})
	}

	// Save to database
	newCount, updatedCount, err := s.repo.UpsertWarehouses(ctx, warehouses)
	if err != nil {
		s.repo.UpdateWarehousesSyncRun(ctx, runID, "failed", len(warehouses), 0, 0, err.Error())
		return fmt.Errorf("failed to save warehouses: %w", err)
	}

	s.logger.Info("Saved warehouses",
		"total", len(warehouses),
		"new", newCount,
		"updated", updatedCount,
	)

	// Update sync run
	if err := s.repo.UpdateWarehousesSyncRun(ctx, runID, "completed", len(warehouses), newCount, updatedCount, ""); err != nil {
		return fmt.Errorf("failed to update sync run: %w", err)
	}

	s.logger.Info("Warehouses sync completed",
		"total", len(warehouses),
		"new", newCount,
		"updated", updatedCount,
	)

	return nil
}
