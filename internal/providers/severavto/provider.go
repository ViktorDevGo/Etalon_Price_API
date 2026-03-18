package severavto

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/prokoleso/etalon-price-api/internal/domain/severavto"
)

// Provider представляет провайдера данных от Северавто
type Provider struct {
	client *Client
	logger *slog.Logger
}

// ProviderConfig содержит конфигурацию для создания провайдера
type ProviderConfig struct {
	BaseURL string
	APIKey  string
	Timeout string
	Logger  *slog.Logger
}

// NewProvider создает новый провайдер Северавто
func NewProvider(cfg ProviderConfig) *Provider {
	clientCfg := ClientConfig{
		BaseURL: cfg.BaseURL,
		APIKey:  cfg.APIKey,
		Logger:  cfg.Logger,
	}

	return &Provider{
		client: NewClient(clientCfg),
		logger: cfg.Logger.With("component", "severavto_provider"),
	}
}

// Name возвращает название провайдера
func (p *Provider) Name() string {
	return "Северавто"
}

// FetchTyresWithStock получает номенклатуру и цены/остатки шин одним запросом
// Северавто возвращает все данные в одном XML (nomenclature + prices + stock)
func (p *Provider) FetchTyresWithStock(ctx context.Context) (
	nomenclature []severavto.NomenclatureTyre,
	pricesStock []severavto.TyrePriceStock,
	warehouses []severavto.Warehouse,
	err error,
) {
	logger := p.logger.With("operation", "fetch_tyres_with_stock")

	// 1. Скачать XML с данными
	logger.Info("Fetching tyres XML from Severavto API")
	xml, err := p.client.FetchTyres(ctx)
	if err != nil {
		// Обработка специальных ошибок
		if errors.Is(err, ErrRateLimitExceeded) {
			logger.Warn("Rate limit exceeded, skipping fetch")
			return nil, nil, nil, err
		}
		if errors.Is(err, ErrNotModified) {
			logger.Info("Data not modified since last fetch")
			return nil, nil, nil, err
		}

		return nil, nil, nil, fmt.Errorf("fetch tyres xml: %w", err)
	}

	logger.Info("XML fetched successfully", "total_commodities", len(xml.Commodities))

	// 2. Фильтровать только склады (не пункты выдачи)
	filtered := filterWarehouses(xml.Commodities)
	logger.Info("Filtered warehouses", "before", len(xml.Commodities), "after", len(filtered))

	// 3. Маппинг в domain модели
	nomenclature = mapToNomenclatureTyres(filtered)
	pricesStock = mapToTyresPricesStock(filtered)
	warehouses = extractUniqueWarehouses(filtered)

	logger.Info("Mapping completed",
		"nomenclature_count", len(nomenclature),
		"prices_stock_count", len(pricesStock),
		"warehouses_count", len(warehouses),
	)

	return nomenclature, pricesStock, warehouses, nil
}

// FetchRimsWithStock получает номенклатуру и цены/остатки дисков одним запросом
func (p *Provider) FetchRimsWithStock(ctx context.Context) (
	nomenclature []severavto.NomenclatureRim,
	pricesStock []severavto.RimPriceStock,
	warehouses []severavto.Warehouse,
	err error,
) {
	logger := p.logger.With("operation", "fetch_rims_with_stock")

	// 1. Скачать XML с данными
	logger.Info("Fetching rims XML from Severavto API")
	xml, err := p.client.FetchRims(ctx)
	if err != nil {
		// Обработка специальных ошибок
		if errors.Is(err, ErrRateLimitExceeded) {
			logger.Warn("Rate limit exceeded, skipping fetch")
			return nil, nil, nil, err
		}
		if errors.Is(err, ErrNotModified) {
			logger.Info("Data not modified since last fetch")
			return nil, nil, nil, err
		}

		return nil, nil, nil, fmt.Errorf("fetch rims xml: %w", err)
	}

	logger.Info("XML fetched successfully", "total_commodities", len(xml.Commodities))

	// 2. Фильтровать только склады
	filtered := filterWarehouses(xml.Commodities)
	logger.Info("Filtered warehouses", "before", len(xml.Commodities), "after", len(filtered))

	// 3. Маппинг в domain модели
	nomenclature = mapToNomenclatureRims(filtered)
	pricesStock = mapToRimsPricesStock(filtered)
	warehouses = extractUniqueWarehouses(filtered)

	logger.Info("Mapping completed",
		"nomenclature_count", len(nomenclature),
		"prices_stock_count", len(pricesStock),
		"warehouses_count", len(warehouses),
	)

	return nomenclature, pricesStock, warehouses, nil
}
