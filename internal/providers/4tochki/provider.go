package fourtochki

import (
	"context"
	"log/slog"
	"time"

	"github.com/prokoleso/etalon-price-api/internal/domain"
	"github.com/prokoleso/etalon-price-api/internal/providers"
)

// Provider implements SupplierProvider interface for 4tochki
type Provider struct {
	client *Client
	logger *slog.Logger
}

// ProviderConfig holds provider configuration
type ProviderConfig struct {
	WSDLURL    string
	Login      string
	Password   string
	Timeout    time.Duration
	RetryCount int
	RetryDelay time.Duration
	Logger     *slog.Logger
}

// NewProvider creates a new 4tochki provider
func NewProvider(cfg ProviderConfig) *Provider {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}

	client := NewClient(ClientConfig{
		WSDLURL:    cfg.WSDLURL,
		Login:      cfg.Login,
		Password:   cfg.Password,
		Timeout:    cfg.Timeout,
		RetryCount: cfg.RetryCount,
		RetryDelay: cfg.RetryDelay,
		Logger:     cfg.Logger,
	})

	return &Provider{
		client: client,
		logger: cfg.Logger,
	}
}

// Name returns the provider name
func (p *Provider) Name() string {
	return providerName
}

// FetchGoodsInfo retrieves detailed product information
func (p *Provider) FetchGoodsInfo(ctx context.Context, codes []string) (*providers.GoodsInfoResult, error) {
	if len(codes) == 0 {
		return &providers.GoodsInfoResult{
			Tyres:      []domain.Tyre{},
			Rims:       []domain.Rim{},
			FetchedAt:  time.Now(),
			TotalCount: 0,
		}, nil
	}

	p.logger.Info("Fetching goods info from 4tochki",
		"codes_count", len(codes),
	)

	start := time.Now()

	// Call SOAP API
	resp, err := p.client.GetGoodsInfo(ctx, codes)
	if err != nil {
		p.logger.Error("Failed to fetch goods info",
			"error", err,
			"codes_count", len(codes),
		)
		return nil, providers.NewProviderError(
			p.Name(),
			"FETCH_GOODS_INFO_FAILED",
			"Failed to fetch goods information",
			err,
		)
	}

	// Check for API errors (only if error code is non-zero or comment is not empty)
	if resp.Result.Error != nil && (resp.Result.Error.Code != 0 || resp.Result.Error.Comment != "") {
		p.logger.Error("API returned error",
			"error_code", resp.Result.Error.Code,
			"error_comment", resp.Result.Error.Comment,
			"codes_count", len(codes),
		)
		return nil, providers.NewProviderError(
			p.Name(),
			"API_ERROR",
			resp.Result.Error.Comment,
			nil,
		)
	}

	// Convert SOAP response to domain models
	items := append(resp.Result.TyreList, resp.Result.RimList...)
	tyres, rims := splitByType(items)

	duration := time.Since(start)

	p.logger.Info("Successfully fetched goods info",
		"codes_count", len(codes),
		"tyres_count", len(tyres),
		"rims_count", len(rims),
		"duration", duration,
	)

	result := &providers.GoodsInfoResult{
		Tyres:      tyres,
		Rims:       rims,
		FetchedAt:  time.Now(),
		TotalCount: len(tyres) + len(rims),
	}

	return result, nil
}

// FetchPricesAndStocks retrieves pricing and stock availability information
func (p *Provider) FetchPricesAndStocks(ctx context.Context, codes []string) (*providers.PriceStockResult, error) {
	if len(codes) == 0 {
		return &providers.PriceStockResult{
			Offers:     []domain.ProductOffer{},
			FetchedAt:  time.Now(),
			TotalCount: 0,
		}, nil
	}

	p.logger.Info("Fetching prices and stocks from 4tochki",
		"codes_count", len(codes),
	)

	start := time.Now()

	// First, fetch goods info to determine product types
	goodsResp, err := p.client.GetGoodsInfo(ctx, codes)
	if err != nil {
		p.logger.Warn("Failed to fetch goods info for type determination, continuing with defaults",
			"error", err,
		)
	}

	// Build a map of code -> goods info for type determination
	goodsMap := make(map[string]GoodsInfoItem)
	if goodsResp != nil && goodsResp.Result.Error == nil {
		allItems := append(goodsResp.Result.TyreList, goodsResp.Result.RimList...)
		for _, item := range allItems {
			goodsMap[item.Code] = item
		}
	}

	// Call SOAP API for prices and stocks
	priceResp, err := p.client.GetGoodsPriceRestByCode(ctx, codes)
	if err != nil {
		p.logger.Error("Failed to fetch prices and stocks",
			"error", err,
			"codes_count", len(codes),
		)
		return nil, providers.NewProviderError(
			p.Name(),
			"FETCH_PRICES_STOCKS_FAILED",
			"Failed to fetch prices and stocks",
			err,
		)
	}

	// Check for API errors (only if error code is non-zero or comment is not empty)
	if priceResp.Result.Error != nil && (priceResp.Result.Error.Code != 0 || priceResp.Result.Error.Comment != "") {
		p.logger.Error("API returned error",
			"error_code", priceResp.Result.Error.Code,
			"error_comment", priceResp.Result.Error.Comment,
			"codes_count", len(codes),
		)
		return nil, providers.NewProviderError(
			p.Name(),
			"API_ERROR",
			priceResp.Result.Error.Comment,
			nil,
		)
	}

	// Convert SOAP response to domain models
	offers := make([]domain.ProductOffer, 0, len(priceResp.Result.Items))
	for _, item := range priceResp.Result.Items {
		productType := determineProductType(item.Code, goodsMap)
		offer := mapToProductOffer(item, productType)
		offers = append(offers, offer)
	}

	duration := time.Since(start)

	p.logger.Info("Successfully fetched prices and stocks",
		"codes_count", len(codes),
		"offers_count", len(offers),
		"duration", duration,
	)

	result := &providers.PriceStockResult{
		Offers:     offers,
		FetchedAt:  time.Now(),
		TotalCount: len(offers),
	}

	return result, nil
}
