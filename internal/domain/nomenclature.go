package domain

import "time"

// NomenclatureTyre represents a tyre in nomenclature catalog
type NomenclatureTyre struct {
	ID          int64
	CAE         string  // Код товара
	Name        string  // Полное название
	Width       float64 // Ширина
	Height      float64 // Высота профиля
	Diameter    string  // Диаметр (R16,00)
	DiameterOut string  // Внешний диаметр
	LoadIndex   string  // Индекс нагрузки
	SpeedIndex  string  // Индекс скорости
	Model       string  // Модель
	Brand       string  // Бренд
	Season      string  // Сезон (Зимняя, Летняя, Всесезонная)
	IsStudded   string  // Шипы (шип / не шип)
	TireType    string  // Тип шины
	RunFlat     string  // Ранфлет
	Reinforced  string  // Усиленная
	IsImport    int     // 0 - загружено, 1 - использовано
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NomenclatureRim represents a rim in nomenclature catalog
type NomenclatureRim struct {
	ID            int64
	CAE           string  // Код товара
	Name          string  // Полное название
	Width         float64 // Ширина
	Diameter      float64 // Диаметр
	BoltsCount    int     // Количество болтов
	BoltsSpacing  float64 // Разболтовка
	BoltsSpacing2 float64 // Разболтовка 2
	ET            float64 // Вылет
	DIA           float64 // DIA
	Model         string  // Модель
	Brand         string  // Бренд
	Color         string  // Цвет
	RimType       string  // Тип диска
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// NomenclatureSyncRun represents a nomenclature synchronization run
type NomenclatureSyncRun struct {
	ID           int64
	SyncType     string // 'tyres' или 'rims'
	StartedAt    time.Time
	CompletedAt  *time.Time
	Status       string // 'running', 'completed', 'failed'
	TotalItems   int
	NewItems     int
	UpdatedItems int
	ErrorMessage string
}
