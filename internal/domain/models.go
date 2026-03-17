package domain

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// JSONB represents a PostgreSQL JSONB field
type JSONB map[string]interface{}

// Value implements the driver.Valuer interface
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan JSONB value")
	}

	return json.Unmarshal(bytes, j)
}

// Tyre represents a tire product in the database
type Tyre struct {
	ID          int64     `json:"id"`
	Code        string    `json:"code"`         // Article/SKU from supplier
	Supplier    string    `json:"supplier"`     // Supplier name (e.g., "4tochki")
	Brand       string    `json:"brand"`        // Tire brand
	Model       string    `json:"model"`        // Tire model
	Width       int       `json:"width"`        // Tire width in mm
	Height      int       `json:"height"`       // Aspect ratio
	Diameter    int       `json:"diameter"`     // Rim diameter in inches
	LoadIndex   string    `json:"load_index"`   // Load index
	SpeedIndex  string    `json:"speed_index"`  // Speed rating
	Season      string    `json:"season"`       // summer, winter, all-season
	RunFlat     bool      `json:"run_flat"`     // Run-flat tire
	Studded     bool      `json:"studded"`      // Studded tire
	Description string    `json:"description"`  // Full description
	RawPayload  JSONB     `json:"raw_payload"`  // Original data from supplier
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Rim represents a wheel rim product in the database
type Rim struct {
	ID          int64     `json:"id"`
	Code        string    `json:"code"`       // Article/SKU from supplier
	Supplier    string    `json:"supplier"`   // Supplier name
	Brand       string    `json:"brand"`      // Rim brand
	Model       string    `json:"model"`      // Rim model
	Width       float64   `json:"width"`      // Rim width in inches
	Diameter    float64   `json:"diameter"`   // Rim diameter in inches
	BoltPattern string    `json:"bolt_pattern"` // e.g., "5x114.3"
	Offset      int       `json:"offset"`     // ET offset in mm
	CenterBore  float64   `json:"center_bore"` // Center bore diameter in mm
	Color       string    `json:"color"`      // Rim color/finish
	Description string    `json:"description"` // Full description
	RawPayload  JSONB     `json:"raw_payload"` // Original data from supplier
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ProductOffer represents pricing and stock information for a product
type ProductOffer struct {
	ID              int64     `json:"id"`
	Code            string    `json:"code"`             // Product code
	Supplier        string    `json:"supplier"`         // Supplier name
	ProductType     string    `json:"product_type"`     // "tyre" or "rim"
	Price           float64   `json:"price"`            // Price in RUB
	Currency        string    `json:"currency"`         // Currency code
	Stock           int       `json:"stock"`            // Available quantity
	WarehouseCode   string    `json:"warehouse_code"`   // Warehouse identifier
	WarehouseName   string    `json:"warehouse_name"`   // Warehouse name
	DeliveryDays    int       `json:"delivery_days"`    // Estimated delivery time
	MinOrderQty     int       `json:"min_order_qty"`    // Minimum order quantity
	AvailableForOrder bool    `json:"available_for_order"` // Can be ordered
	RawPayload      JSONB     `json:"raw_payload"`      // Original data from supplier
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// SyncRun represents a synchronization job execution
type SyncRun struct {
	ID              int64     `json:"id"`
	Provider        string    `json:"provider"`         // Provider name
	Status          string    `json:"status"`           // pending, running, completed, failed
	StartedAt       time.Time `json:"started_at"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	TotalProducts   int       `json:"total_products"`   // Total products processed
	SuccessProducts int       `json:"success_products"` // Successfully synced
	FailedProducts  int       `json:"failed_products"`  // Failed to sync
	ErrorMessage    string    `json:"error_message,omitempty"` // Error details if failed
	Metadata        JSONB     `json:"metadata"`         // Additional sync metadata
	CreatedAt       time.Time `json:"created_at"`
}

// Sync statuses
const (
	SyncStatusPending   = "pending"
	SyncStatusRunning   = "running"
	SyncStatusCompleted = "completed"
	SyncStatusFailed    = "failed"
)

// Product types
const (
	ProductTypeTyre = "tyre"
	ProductTypeRim  = "rim"
)

// Seasons
const (
	SeasonSummer     = "summer"
	SeasonWinter     = "winter"
	SeasonAllSeason  = "all-season"
)
