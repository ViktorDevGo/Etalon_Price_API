package severavto

import (
	"strconv"
	"strings"
	"time"

	"github.com/prokoleso/etalon-price-api/internal/domain/severavto"
)

// mapToNomenclatureTyres преобразует XML commodities в domain модели шин
// Выполняет дедупликацию по commodity_id (один товар может быть на нескольких территориях)
func mapToNomenclatureTyres(commodities []Commodity) []severavto.NomenclatureTyre {
	// Дедупликация: один commodity_id = одна запись
	unique := make(map[string]Commodity)
	for _, c := range commodities {
		if _, exists := unique[c.ID]; !exists {
			unique[c.ID] = c
		}
	}

	// Маппинг в domain модели
	result := make([]severavto.NomenclatureTyre, 0, len(unique))
	for _, c := range unique {
		result = append(result, severavto.NomenclatureTyre{
			CommodityID:     c.ID,
			Article:         c.Article,
			Name:            buildTyreName(c),
			Brand:           c.Brand,
			Model:           c.Model,
			Width:           parseFloat(c.Width),
			Height:          parseFloat(c.Height),
			Diameter:        normalizeDiameter(c.Radius),
			LoadIndex:       c.Load,
			SpeedIndex:      c.Speed,
			Season:          detectSeason(c.Thorning, c.Model),
			IsStudded:       normalizeStudded(c.Thorning),
			TireType:        "Легковая", // по умолчанию
			RunFlat:         detectRunFlat(c.Model),
			ManufactureYear: parseInt(c.ManufactureYear),
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		})
	}

	return result
}

// mapToNomenclatureRims преобразует XML commodities в domain модели дисков
func mapToNomenclatureRims(commodities []Commodity) []severavto.NomenclatureRim {
	// Дедупликация по commodity_id
	unique := make(map[string]Commodity)
	for _, c := range commodities {
		if _, exists := unique[c.ID]; !exists {
			unique[c.ID] = c
		}
	}

	// Маппинг в domain модели
	result := make([]severavto.NomenclatureRim, 0, len(unique))
	for _, c := range unique {
		result = append(result, severavto.NomenclatureRim{
			CommodityID:     c.ID,
			Article:         c.Article,
			Name:            buildRimName(c),
			Brand:           c.Brand,
			Model:           c.Model,
			Width:           parseFloat(c.Width),
			Diameter:        parseFloat(c.Radius),
			ManufactureYear: parseInt(c.ManufactureYear),
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		})
	}

	return result
}

// mapToTyresPricesStock преобразует XML commodities в записи цен/остатков шин
// НЕ дедуплицирует - каждая территория = отдельная запись
func mapToTyresPricesStock(commodities []Commodity) []severavto.TyrePriceStock {
	result := make([]severavto.TyrePriceStock, 0, len(commodities))

	for _, c := range commodities {
		result = append(result, severavto.TyrePriceStock{
			CommodityID:   c.ID,
			TerritoryName: c.TerritoryName,
			Price:         parseInt(c.PricePrepayment), // основная цена (предоплата)
			PriceDelayed:  parseInt(c.PriceDelayed),
			PriceRRP:      parseInt(c.PriceRRP),
			Stock:         parseInt(c.Rest),
			PlaceType:     c.PlaceType,
			IsSellout:     c.Sellout == "Да",
			Provider:      "Северавто",
			UpdatedAt:     time.Now(),
		})
	}

	return result
}

// mapToRimsPricesStock преобразует XML commodities в записи цен/остатков дисков
func mapToRimsPricesStock(commodities []Commodity) []severavto.RimPriceStock {
	result := make([]severavto.RimPriceStock, 0, len(commodities))

	for _, c := range commodities {
		result = append(result, severavto.RimPriceStock{
			CommodityID:   c.ID,
			TerritoryName: c.TerritoryName,
			Price:         parseInt(c.PricePrepayment),
			PriceDelayed:  parseInt(c.PriceDelayed),
			PriceRRP:      parseInt(c.PriceRRP),
			Stock:         parseInt(c.Rest),
			PlaceType:     c.PlaceType,
			IsSellout:     c.Sellout == "Да",
			Provider:      "Северавто",
			UpdatedAt:     time.Now(),
		})
	}

	return result
}

// extractUniqueWarehouses извлекает уникальные территории из commodities
func extractUniqueWarehouses(commodities []Commodity) []severavto.Warehouse {
	unique := make(map[string]severavto.Warehouse)

	for _, c := range commodities {
		if _, exists := unique[c.TerritoryName]; !exists {
			unique[c.TerritoryName] = severavto.Warehouse{
				TerritoryName: c.TerritoryName,
				PlaceType:     c.PlaceType,
				Region:        extractRegion(c.TerritoryName),
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			}
		}
	}

	result := make([]severavto.Warehouse, 0, len(unique))
	for _, w := range unique {
		result = append(result, w)
	}

	return result
}

// filterWarehouses фильтрует только склады (исключает пункты выдачи)
func filterWarehouses(commodities []Commodity) []Commodity {
	result := make([]Commodity, 0, len(commodities))

	for _, c := range commodities {
		if c.PlaceType == "Склад" {
			result = append(result, c)
		}
	}

	return result
}

// detectSeason определяет сезон шины из поля "шипы" и модели
func detectSeason(thorning, model string) string {
	model = strings.ToLower(model)

	// Если есть шипы → Зимняя
	if thorning == "Да" {
		return "Зимняя"
	}

	// Ключевые слова в модели для зимы
	winterKeywords := []string{"winter", "ice", "snow", "blizzak", "x-ice", "alpin"}
	for _, keyword := range winterKeywords {
		if strings.Contains(model, keyword) {
			return "Зимняя"
		}
	}

	// Ключевые слова для лета
	summerKeywords := []string{"summer", "sport", "pilot"}
	for _, keyword := range summerKeywords {
		if strings.Contains(model, keyword) {
			return "Летняя"
		}
	}

	// Ключевые слова для всесезонки
	allSeasonKeywords := []string{"all season", "4 season", "4s", "as", "crossclimate"}
	for _, keyword := range allSeasonKeywords {
		if strings.Contains(model, keyword) {
			return "Всесезонная"
		}
	}

	// По умолчанию - пусто (не определено)
	return ""
}

// normalizeStudded нормализует значение поля "шипы"
func normalizeStudded(thorning string) string {
	if thorning == "Да" {
		return "шип"
	}
	if thorning == "Нет" {
		return "не шип"
	}
	return ""
}

// detectRunFlat определяет RunFlat из модели
func detectRunFlat(model string) string {
	model = strings.ToLower(model)
	if strings.Contains(model, "runflat") || strings.Contains(model, "run flat") || strings.Contains(model, "rft") {
		return "RunFlat"
	}
	return ""
}

// normalizeDiameter нормализует диаметр (R15, R16 и т.д.)
func normalizeDiameter(radius string) string {
	// Убираем пробелы
	radius = strings.TrimSpace(radius)

	// Если уже в формате R15, R16 - оставляем как есть
	if strings.HasPrefix(radius, "R") {
		return radius
	}

	// Если просто число (15, 16) - добавляем R
	if radius != "" && radius != "0" {
		return "R" + radius
	}

	return radius
}

// buildTyreName формирует полное название шины
func buildTyreName(c Commodity) string {
	// Формат: Brand Model Width/Height RRadius LoadIndex SpeedIndex
	// Пример: Michelin X-Ice North 4 195/65 R15 91T
	parts := []string{c.Brand, c.Model}

	if c.Width != "" && c.Height != "" {
		parts = append(parts, c.Width+"/"+c.Height)
	}

	if c.Radius != "" {
		parts = append(parts, normalizeDiameter(c.Radius))
	}

	if c.Load != "" && c.Speed != "" {
		parts = append(parts, c.Load+c.Speed)
	}

	return strings.Join(parts, " ")
}

// buildRimName формирует полное название диска
func buildRimName(c Commodity) string {
	// Формат: Brand Model Width x Diameter
	// Пример: K&K КС681 6.5x16
	parts := []string{c.Brand, c.Model}

	if c.Width != "" && c.Radius != "" {
		parts = append(parts, c.Width+"x"+c.Radius)
	}

	return strings.Join(parts, " ")
}

// extractRegion извлекает регион из названия территории
// Пример: "Новосибирск" → "Новосибирский регион"
func extractRegion(territoryName string) string {
	// Упрощенная логика - можно улучшить
	return territoryName
}

// parseFloat парсит строку в float64
func parseFloat(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" || s == "0" {
		return 0
	}

	// Заменяем запятую на точку
	s = strings.ReplaceAll(s, ",", ".")

	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}

	return f
}

// parseInt парсит строку в int
func parseInt(s string) int {
	s = strings.TrimSpace(s)
	if s == "" || s == "0" {
		return 0
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}

	return i
}
