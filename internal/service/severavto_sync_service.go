package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/prokoleso/etalon-price-api/internal/email"
	"github.com/prokoleso/etalon-price-api/internal/providers/severavto"
	"github.com/prokoleso/etalon-price-api/internal/repository"
)

// SeveravtoSyncService orchestrates synchronization with Severavto API
type SeveravtoSyncService struct {
	provider         *severavto.Provider
	nomenclRepo      *repository.NomenclatureRepository
	pricesRepo       *repository.PricesStockRepository
	warehouseRepo    *repository.WarehouseRepository
	emailService     *email.Client
	notificationEmail string
	logger           *slog.Logger
}

// NewSeveravtoSyncService creates a new Severavto sync service
func NewSeveravtoSyncService(
	provider *severavto.Provider,
	nomenclRepo *repository.NomenclatureRepository,
	pricesRepo *repository.PricesStockRepository,
	warehouseRepo *repository.WarehouseRepository,
	emailService *email.Client,
	notificationEmail string,
	logger *slog.Logger,
) *SeveravtoSyncService {
	return &SeveravtoSyncService{
		provider:         provider,
		nomenclRepo:      nomenclRepo,
		pricesRepo:       pricesRepo,
		warehouseRepo:    warehouseRepo,
		emailService:     emailService,
		notificationEmail: notificationEmail,
		logger:           logger.With("service", "severavto_sync"),
	}
}

// SyncTyres синхронизирует шины: номенклатуру, цены и остатки
func (s *SeveravtoSyncService) SyncTyres(ctx context.Context) error {
	logger := s.logger.With("operation", "sync_tyres")
	logger.Info("Starting Severavto tyres synchronization")

	// 1. Fetch data from Severavto XML API
	nomenclature, pricesStock, warehouses, err := s.provider.FetchTyresWithStock(ctx)
	if err != nil {
		// Обработка специальных ошибок
		if errors.Is(err, severavto.ErrRateLimitExceeded) {
			logger.Warn("Rate limit exceeded, skipping sync")
			return nil // не ошибка, просто пропускаем
		}
		if errors.Is(err, severavto.ErrNotModified) {
			logger.Info("Data not modified since last fetch, skipping sync")
			return nil
		}

		logger.Error("Failed to fetch tyres data", "error", err)
		s.sendErrorEmail("Северавто - Шины", err.Error())
		return fmt.Errorf("fetch tyres: %w", err)
	}

	logger.Info("Fetched data from Severavto",
		"nomenclature_count", len(nomenclature),
		"prices_stock_count", len(pricesStock),
		"warehouses_count", len(warehouses),
	)

	// 2. Upsert nomenclature
	newNom, skippedNom, err := s.nomenclRepo.UpsertSeveravtoTyres(ctx, nomenclature)
	if err != nil {
		logger.Error("Failed to upsert nomenclature", "error", err)
		s.sendErrorEmail("Северавто - Шины (Номенклатура)", err.Error())
		return fmt.Errorf("upsert nomenclature: %w", err)
	}

	logger.Info("Nomenclature upserted",
		"new", newNom,
		"skipped", skippedNom,
	)

	// 3. Upsert prices/stock
	newPrices, updatedPrices, err := s.pricesRepo.UpsertSeveravtoTyresPricesStock(ctx, pricesStock)
	if err != nil {
		logger.Error("Failed to upsert prices/stock", "error", err)
		s.sendErrorEmail("Северавто - Шины (Цены/Остатки)", err.Error())
		return fmt.Errorf("upsert prices/stock: %w", err)
	}

	logger.Info("Prices/Stock upserted",
		"new", newPrices,
		"updated", updatedPrices,
	)

	// 4. Upsert warehouses (territories)
	newWh, updatedWh, err := s.warehouseRepo.UpsertSeveravtoWarehouses(ctx, warehouses)
	if err != nil {
		logger.Error("Failed to upsert warehouses", "error", err)
		// Не критично, продолжаем
	} else {
		logger.Info("Warehouses upserted",
			"new", newWh,
			"updated", updatedWh,
		)
	}

	// 5. Send success email notification
	s.sendSuccessEmail("Северавто - Шины", map[string]interface{}{
		"nomenclature_new":     newNom,
		"nomenclature_skipped": skippedNom,
		"prices_new":           newPrices,
		"prices_updated":       updatedPrices,
		"territories":          len(warehouses),
	})

	logger.Info("Severavto tyres synchronization completed successfully")
	return nil
}

// SyncRims синхронизирует диски: номенклатуру, цены и остатки
func (s *SeveravtoSyncService) SyncRims(ctx context.Context) error {
	logger := s.logger.With("operation", "sync_rims")
	logger.Info("Starting Severavto rims synchronization")

	// 1. Fetch data from Severavto XML API
	nomenclature, pricesStock, warehouses, err := s.provider.FetchRimsWithStock(ctx)
	if err != nil {
		// Обработка специальных ошибок
		if errors.Is(err, severavto.ErrRateLimitExceeded) {
			logger.Warn("Rate limit exceeded, skipping sync")
			return nil
		}
		if errors.Is(err, severavto.ErrNotModified) {
			logger.Info("Data not modified since last fetch, skipping sync")
			return nil
		}

		logger.Error("Failed to fetch rims data", "error", err)
		s.sendErrorEmail("Северавто - Диски", err.Error())
		return fmt.Errorf("fetch rims: %w", err)
	}

	logger.Info("Fetched data from Severavto",
		"nomenclature_count", len(nomenclature),
		"prices_stock_count", len(pricesStock),
		"warehouses_count", len(warehouses),
	)

	// 2. Upsert nomenclature
	newNom, skippedNom, err := s.nomenclRepo.UpsertSeveravtoRims(ctx, nomenclature)
	if err != nil {
		logger.Error("Failed to upsert nomenclature", "error", err)
		s.sendErrorEmail("Северавто - Диски (Номенклатура)", err.Error())
		return fmt.Errorf("upsert nomenclature: %w", err)
	}

	logger.Info("Nomenclature upserted",
		"new", newNom,
		"skipped", skippedNom,
	)

	// 3. Upsert prices/stock
	newPrices, updatedPrices, err := s.pricesRepo.UpsertSeveravtoRimsPricesStock(ctx, pricesStock)
	if err != nil {
		logger.Error("Failed to upsert prices/stock", "error", err)
		s.sendErrorEmail("Северавто - Диски (Цены/Остатки)", err.Error())
		return fmt.Errorf("upsert prices/stock: %w", err)
	}

	logger.Info("Prices/Stock upserted",
		"new", newPrices,
		"updated", updatedPrices,
	)

	// 4. Upsert warehouses (territories)
	newWh, updatedWh, err := s.warehouseRepo.UpsertSeveravtoWarehouses(ctx, warehouses)
	if err != nil {
		logger.Error("Failed to upsert warehouses", "error", err)
		// Не критично, продолжаем
	} else {
		logger.Info("Warehouses upserted",
			"new", newWh,
			"updated", updatedWh,
		)
	}

	// 5. Send success email notification
	s.sendSuccessEmail("Северавто - Диски", map[string]interface{}{
		"nomenclature_new":     newNom,
		"nomenclature_skipped": skippedNom,
		"prices_new":           newPrices,
		"prices_updated":       updatedPrices,
		"territories":          len(warehouses),
	})

	logger.Info("Severavto rims synchronization completed successfully")
	return nil
}

// sendSuccessEmail отправляет email уведомление об успешной синхронизации
func (s *SeveravtoSyncService) sendSuccessEmail(subject string, stats map[string]interface{}) {
	if s.emailService == nil {
		return
	}

	body := fmt.Sprintf(`
<h2>✅ Синхронизация завершена: %s</h2>

<h3>Статистика:</h3>
<ul>
	<li><b>Номенклатура:</b> %d новых, %d пропущено (дублей)</li>
	<li><b>Цены/Остатки:</b> %d новых, %d обновлено</li>
	<li><b>Территорий:</b> %d</li>
</ul>

<p><small>Сообщение отправлено автоматически системой Etalon Price API</small></p>
`,
		subject,
		stats["nomenclature_new"],
		stats["nomenclature_skipped"],
		stats["prices_new"],
		stats["prices_updated"],
		stats["territories"],
	)

	if err := s.emailService.Send(email.Message{
		To:      []string{s.notificationEmail},
		Subject: subject + " - Успешно",
		Body:    body,
		HTML:    true,
	}); err != nil {
		s.logger.Error("Failed to send success email", "error", err)
	}
}

// sendErrorEmail отправляет email уведомление об ошибке синхронизации
func (s *SeveravtoSyncService) sendErrorEmail(subject string, errorMsg string) {
	if s.emailService == nil {
		return
	}

	body := fmt.Sprintf(`
<h2>❌ Ошибка синхронизации: %s</h2>

<h3>Детали ошибки:</h3>
<pre style="background: #f5f5f5; padding: 10px; border-radius: 5px;">%s</pre>

<p><small>Сообщение отправлено автоматически системой Etalon Price API</small></p>
`,
		subject,
		errorMsg,
	)

	if err := s.emailService.Send(email.Message{
		To:      []string{s.notificationEmail},
		Subject: subject + " - Ошибка",
		Body:    body,
		HTML:    true,
	}); err != nil {
		s.logger.Error("Failed to send error email", "error", err)
	}
}
