package domain

import "time"

// TyrePriceStock represents price and stock for a tyre at a specific warehouse
type TyrePriceStock struct {
	ID            int64
	CAE           string // Product code
	WarehouseName string // Warehouse name
	Price         int    // Price in kopecks
	Stock         int    // Stock quantity
	IsImport      int    // 0 - loaded, 1 - imported/used
	Provider      string // Provider name (e.g., "Форточки")
	UpdatedAt     time.Time
}

// RimPriceStock represents price and stock for a rim at a specific warehouse
type RimPriceStock struct {
	ID            int64
	CAE           string // Product code
	WarehouseName string // Warehouse name
	Price         int    // Price in kopecks
	Stock         int    // Stock quantity
	IsImport      int    // 0 - loaded, 1 - imported/used
	Provider      string // Provider name (e.g., "Форточки")
	UpdatedAt     time.Time
}

// PricesStockSyncRun represents a prices/stock synchronization run
type PricesStockSyncRun struct {
	ID              int64
	SyncType        string // 'tyres' or 'rims'
	StartedAt       time.Time
	CompletedAt     *time.Time
	Status          string // 'running', 'completed', 'failed'
	TotalItems      int    // Total products processed
	TotalWarehouses int    // Total warehouse positions
	NewItems        int    // New records
	UpdatedItems    int    // Updated records
	ErrorMessage    string
}
