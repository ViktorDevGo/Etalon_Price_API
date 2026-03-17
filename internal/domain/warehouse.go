package domain

import "time"

// Warehouse represents a warehouse/distribution center
type Warehouse struct {
	ID              int    // ID склада
	Key             string // Ключ склада
	Name            string // Полное название
	ShortName       string // Короткое название
	StoID           int    // ID в системе STO
	BackgroundColor string // Цвет для UI
	HaveDelivery    bool   // Есть доставка
	HavePickup      bool   // Есть самовывоз
	IsPaidDelivery  bool   // Платная доставка
	LogisticDays    int    // Дни логистики
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// WarehousesSyncRun represents a warehouses synchronization run
type WarehousesSyncRun struct {
	ID                int64
	StartedAt         time.Time
	CompletedAt       *time.Time
	Status            string // 'running', 'completed', 'failed'
	TotalWarehouses   int
	NewWarehouses     int
	UpdatedWarehouses int
	ErrorMessage      string
}
