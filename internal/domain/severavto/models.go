package severavto

import "time"

// NomenclatureTyre представляет номенклатуру шины от поставщика Северавто
type NomenclatureTyre struct {
	CommodityID     string  // NNCOMMODIF - уникальный ID товара
	Article         string  // SARTICLE - артикул
	Name            string  // полное название товара
	Brand           string  // SBREND - бренд (производитель)
	Model           string  // SMODEL - модель шины
	Width           float64 // SWIDTH - ширина шины (мм)
	Height          float64 // SHEIGHT - профиль/высота (%)
	Diameter        string  // SRADIUS - радиус (R13, R14, R15 и т.д.)
	LoadIndex       string  // SLOAD - индекс нагрузки
	SpeedIndex      string  // SSPEED - индекс скорости
	Season          string  // сезон (Зимняя/Летняя/Всесезонная) - определяется из STHORNING + модели
	IsStudded       string  // STHORNING - шипы (Да/Нет)
	TireType        string  // тип шины (по умолчанию "Легковая")
	RunFlat         string  // RunFlat - определяется из названия/модели
	ManufactureYear int     // SMNFYEAR - год производства
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// NomenclatureRim представляет номенклатуру диска от поставщика Северавто
type NomenclatureRim struct {
	CommodityID     string  // NNCOMMODIF - уникальный ID товара
	Article         string  // SARTICLE - артикул
	Name            string  // полное название товара
	Brand           string  // SBREND - бренд (производитель)
	Model           string  // SMODEL - модель диска
	Width           float64 // ширина диска (дюймы)
	Diameter        float64 // диаметр диска (дюймы)
	BoltsCount      int     // количество отверстий под болты (PCD)
	BoltsSpacing    float64 // расстояние между отверстиями (мм)
	BoltsSpacing2   float64 // дополнительное расстояние (для двойного PCD)
	ET              float64 // ET - вылет диска (мм)
	DIA             float64 // DIA - диаметр центрального отверстия (мм)
	Color           string  // цвет диска
	RimType         string  // тип диска (литой/кованый/штампованный)
	ManufactureYear int     // год производства
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// TyrePriceStock представляет цену и остаток шины на конкретной территории
type TyrePriceStock struct {
	CommodityID   string // NNCOMMODIF - ID товара
	TerritoryName string // TERRITORY_NAME - название территории/склада
	Price         int    // NPRICE_PREPAYMENT - цена по предоплате (в рублях!)
	PriceDelayed  int    // NPRICE_DELAYED - цена по отсрочке (в рублях)
	PriceRRP      int    // NPRICE_RRP - рекомендованная розничная цена
	Stock         int    // NREST - остаток (количество)
	PlaceType     string // SPLACE - тип места (Склад / Пункт выдачи)
	IsSellout     bool   // SSELLOUT - распродажа (Да/Нет)
	Provider      string // провайдер ("Северавто")
	UpdatedAt     time.Time
}

// RimPriceStock представляет цену и остаток диска на конкретной территории
type RimPriceStock struct {
	CommodityID   string // NNCOMMODIF - ID товара
	TerritoryName string // TERRITORY_NAME - название территории/склада
	Price         int    // NPRICE_PREPAYMENT - цена по предоплате (в рублях!)
	PriceDelayed  int    // NPRICE_DELAYED - цена по отсрочке (в рублях)
	PriceRRP      int    // NPRICE_RRP - рекомендованная розничная цена
	Stock         int    // NREST - остаток (количество)
	PlaceType     string // SPLACE - тип места (Склад / Пункт выдачи)
	IsSellout     bool   // SSELLOUT - распродажа (Да/Нет)
	Provider      string // провайдер ("Северавто")
	UpdatedAt     time.Time
}

// Warehouse представляет территорию/склад Северавто
type Warehouse struct {
	TerritoryName string // TERRITORY_NAME - уникальное название территории
	Region        string // регион (Новосибирский, Московский и т.д.)
	PlaceType     string // SPLACE - тип (Склад / Пункт выдачи)
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
