package fourtochki

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/prokoleso/etalon-price-api/internal/domain"
)

const providerName = "4tochki"

// mapToTyre converts GoodsInfoItem to domain.Tyre
func mapToTyre(item GoodsInfoItem) domain.Tyre {
	season := normalizeSeason(item.Season)
	studded := item.Thorn == "true" || item.Thorn == "1" || strings.ToUpper(item.Thorn) == "TRUE"
	// RunFlat detection could be based on special markers or model name
	runFlat := strings.Contains(strings.ToUpper(item.Model), "RUNFLAT") || strings.Contains(strings.ToUpper(item.Model), "RFT")

	rawPayload := makeRawPayload(item)

	return domain.Tyre{
		Code:        item.Code,
		Supplier:    providerName,
		Brand:       item.Brand,
		Model:       item.Model,
		Width:       int(item.Width),
		Height:      int(item.Height),
		Diameter:    int(item.Diameter),
		LoadIndex:   item.LoadIndex,
		SpeedIndex:  item.SpeedIndex,
		Season:      season,
		RunFlat:     runFlat,
		Studded:     studded,
		Description: item.Name,
		RawPayload:  rawPayload,
		UpdatedAt:   time.Now(),
	}
}

// mapToRim converts GoodsInfoItem to domain.Rim
func mapToRim(item GoodsInfoItem) domain.Rim {
	rawPayload := makeRawPayload(item)

	// Parse offset from radius field if needed
	var offset int
	if item.Radius != "" {
		// Try to extract numeric value from radius string
		// This might need more sophisticated parsing depending on format
	}

	// Parse center bore from central_hole field if needed
	var centerBore float64
	if item.CentralHole != "" {
		// Try to extract numeric value
	}

	return domain.Rim{
		Code:        item.Code,
		Supplier:    providerName,
		Brand:       item.Brand,
		Model:       item.Model,
		Width:       item.DiskWidth,
		Diameter:    item.DiskDiameter,
		BoltPattern: item.Drilling,
		Offset:      offset,
		CenterBore:  centerBore,
		Color:       item.Color,
		Description: item.Name,
		RawPayload:  rawPayload,
		UpdatedAt:   time.Now(),
	}
}

// mapToProductOffer converts PriceRestInfoItem to domain.ProductOffer
func mapToProductOffer(item PriceRestInfoItem, productType string) domain.ProductOffer {
	currency := item.Currency
	if currency == "" {
		currency = "RUB"
	}

	available := item.Available == "Y" || item.Available == "1" || strings.ToUpper(item.Available) == "TRUE"

	minOrder := item.MinOrder
	if minOrder <= 0 {
		minOrder = 1
	}

	rawPayload := makeRawPayload(item)

	return domain.ProductOffer{
		Code:              item.Code,
		Supplier:          providerName,
		ProductType:       productType,
		Price:             item.Price,
		Currency:          currency,
		Stock:             item.Quantity,
		WarehouseCode:     item.WarehouseCode,
		WarehouseName:     item.WarehouseName,
		DeliveryDays:      item.DeliveryDays,
		MinOrderQty:       minOrder,
		AvailableForOrder: available,
		RawPayload:        rawPayload,
		UpdatedAt:         time.Now(),
	}
}

// normalizeSeason converts provider season to domain season
func normalizeSeason(season string) string {
	season = strings.ToLower(strings.TrimSpace(season))

	switch season {
	case "s", "summer", "летние", "leto":
		return domain.SeasonSummer
	case "w", "winter", "зимние", "zima":
		return domain.SeasonWinter
	case "m", "all_season", "allseason", "всесезонные", "all":
		return domain.SeasonAllSeason
	default:
		if strings.Contains(season, "summer") || strings.Contains(season, "летн") {
			return domain.SeasonSummer
		}
		if strings.Contains(season, "winter") || strings.Contains(season, "зимн") {
			return domain.SeasonWinter
		}
		return domain.SeasonAllSeason
	}
}

// makeRawPayload converts any struct to JSONB
func makeRawPayload(data interface{}) domain.JSONB {
	bytes, err := json.Marshal(data)
	if err != nil {
		return domain.JSONB{}
	}

	var result domain.JSONB
	if err := json.Unmarshal(bytes, &result); err != nil {
		return domain.JSONB{}
	}

	return result
}

// splitByType splits goods into tyres and rims
func splitByType(items []GoodsInfoItem) (tyres []domain.Tyre, rims []domain.Rim) {
	for _, item := range items {
		prodType := strings.ToLower(strings.TrimSpace(item.Type))

		// Determine type based on product type field or available dimensions
		if prodType == "car" || prodType == "vned" || prodType == "truck" || prodType == "moto" {
			// These are tire types
			tyres = append(tyres, mapToTyre(item))
		} else if prodType == "disk" || prodType == "disc" || prodType == "rim" || prodType == "wheel" {
			// These are rim types
			rims = append(rims, mapToRim(item))
		} else {
			// Try to determine by fields
			if item.Width > 0 && item.Height > 0 && item.Diameter > 0 {
				// Has tire dimensions
				tyres = append(tyres, mapToTyre(item))
			} else if item.DiskWidth > 0 && item.DiskDiameter > 0 {
				// Has rim dimensions
				rims = append(rims, mapToRim(item))
			}
		}
	}

	return tyres, rims
}

// determineProductType determines if code is for tire or rim based on goods info
func determineProductType(code string, goodsMap map[string]GoodsInfoItem) string {
	if item, exists := goodsMap[code]; exists {
		prodType := strings.ToLower(strings.TrimSpace(item.Type))

		if prodType == "car" || prodType == "vned" || prodType == "truck" || prodType == "moto" {
			return domain.ProductTypeTyre
		} else if prodType == "disk" || prodType == "disc" || prodType == "rim" || prodType == "wheel" {
			return domain.ProductTypeRim
		}

		// Try to determine by fields
		if item.Width > 0 && item.Height > 0 && item.Diameter > 0 {
			return domain.ProductTypeTyre
		} else if item.DiskWidth > 0 && item.DiskDiameter > 0 {
			return domain.ProductTypeRim
		}
	}

	// Default to tyre if unknown
	return domain.ProductTypeTyre
}
