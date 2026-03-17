package providers

import (
	"context"
	"time"

	"github.com/prokoleso/etalon-price-api/internal/domain"
)

// SupplierProvider defines the interface that all supplier providers must implement
type SupplierProvider interface {
	// Name returns the provider name
	Name() string

	// FetchGoodsInfo retrieves detailed product information (specs, descriptions)
	// for the given product codes
	FetchGoodsInfo(ctx context.Context, codes []string) (*GoodsInfoResult, error)

	// FetchPricesAndStocks retrieves pricing and stock availability information
	// for the given product codes
	FetchPricesAndStocks(ctx context.Context, codes []string) (*PriceStockResult, error)
}

// GoodsInfoResult contains detailed product information
type GoodsInfoResult struct {
	Tyres      []domain.Tyre `json:"tyres"`
	Rims       []domain.Rim  `json:"rims"`
	FetchedAt  time.Time     `json:"fetched_at"`
	TotalCount int           `json:"total_count"`
}

// PriceStockResult contains pricing and stock information
type PriceStockResult struct {
	Offers     []domain.ProductOffer `json:"offers"`
	FetchedAt  time.Time             `json:"fetched_at"`
	TotalCount int                   `json:"total_count"`
}

// ProviderError represents an error from a provider
type ProviderError struct {
	Provider string
	Code     string
	Message  string
	Err      error
}

func (e *ProviderError) Error() string {
	if e.Err != nil {
		return e.Provider + ": " + e.Message + ": " + e.Err.Error()
	}
	return e.Provider + ": " + e.Message
}

func (e *ProviderError) Unwrap() error {
	return e.Err
}

// NewProviderError creates a new provider error
func NewProviderError(provider, code, message string, err error) *ProviderError {
	return &ProviderError{
		Provider: provider,
		Code:     code,
		Message:  message,
		Err:      err,
	}
}
