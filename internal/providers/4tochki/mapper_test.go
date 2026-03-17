package fourtochki

import (
	"testing"

	"github.com/prokoleso/etalon-price-api/internal/domain"
)

func TestNormalizeSeason(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"SUMMER", domain.SeasonSummer},
		{"summer", domain.SeasonSummer},
		{"ЛЕТНИЕ", domain.SeasonSummer},
		{"WINTER", domain.SeasonWinter},
		{"winter", domain.SeasonWinter},
		{"ЗИМНИЕ", domain.SeasonWinter},
		{"ALL_SEASON", domain.SeasonAllSeason},
		{"ALLSEASON", domain.SeasonAllSeason},
		{"ВСЕСЕЗОННЫЕ", domain.SeasonAllSeason},
		{"unknown", domain.SeasonAllSeason},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizeSeason(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeSeason(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestMapToTyre(t *testing.T) {
	item := GoodsInfoItem{
		Code:        "TEST123",
		GoodType:    "TIRE",
		Brand:       "Michelin",
		Model:       "Pilot Sport 4",
		Width:       225,
		Height:      45,
		Diameter:    17,
		Season:      "SUMMER",
		Thorns:      "N",
		RunFlat:     "Y",
		LoadIndex:   "94",
		SpeedIndex:  "W",
		Description: "Test tire",
	}

	tyre := mapToTyre(item)

	if tyre.Code != "TEST123" {
		t.Errorf("expected code TEST123, got %s", tyre.Code)
	}
	if tyre.Brand != "Michelin" {
		t.Errorf("expected brand Michelin, got %s", tyre.Brand)
	}
	if tyre.Width != 225 {
		t.Errorf("expected width 225, got %d", tyre.Width)
	}
	if tyre.Season != domain.SeasonSummer {
		t.Errorf("expected season %s, got %s", domain.SeasonSummer, tyre.Season)
	}
	if !tyre.RunFlat {
		t.Error("expected RunFlat to be true")
	}
	if tyre.Studded {
		t.Error("expected Studded to be false")
	}
}

func TestMapToRim(t *testing.T) {
	item := GoodsInfoItem{
		Code:         "RIM456",
		GoodType:     "DISK",
		Brand:        "BBS",
		Model:        "CH-R",
		DiskWidth:    8.0,
		DiskDiameter: 18.0,
		BoltPattern:  "5x112",
		Offset:       35,
		CenterBore:   66.6,
		Color:        "Black",
		Description:  "Test rim",
	}

	rim := mapToRim(item)

	if rim.Code != "RIM456" {
		t.Errorf("expected code RIM456, got %s", rim.Code)
	}
	if rim.Brand != "BBS" {
		t.Errorf("expected brand BBS, got %s", rim.Brand)
	}
	if rim.Width != 8.0 {
		t.Errorf("expected width 8.0, got %f", rim.Width)
	}
	if rim.BoltPattern != "5x112" {
		t.Errorf("expected bolt pattern 5x112, got %s", rim.BoltPattern)
	}
}

func TestSplitByType(t *testing.T) {
	items := []GoodsInfoItem{
		{
			Code:     "TIRE1",
			GoodType: "TIRE",
			Width:    225,
			Height:   45,
			Diameter: 17,
		},
		{
			Code:         "RIM1",
			GoodType:     "DISK",
			DiskWidth:    8.0,
			DiskDiameter: 18.0,
		},
		{
			Code:     "TIRE2",
			GoodType: "TYRE",
			Width:    205,
			Height:   55,
			Diameter: 16,
		},
	}

	tyres, rims := splitByType(items)

	if len(tyres) != 2 {
		t.Errorf("expected 2 tyres, got %d", len(tyres))
	}
	if len(rims) != 1 {
		t.Errorf("expected 1 rim, got %d", len(rims))
	}
}

func TestDetermineProductType(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		goodType string
		expected string
	}{
		{"tire", "T1", "TIRE", domain.ProductTypeTyre},
		{"disk", "D1", "DISK", domain.ProductTypeRim},
		{"unknown", "U1", "UNKNOWN", domain.ProductTypeTyre},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goodsMap := map[string]GoodsInfoItem{
				tt.code: {
					Code:     tt.code,
					GoodType: tt.goodType,
				},
			}

			result := determineProductType(tt.code, goodsMap)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}
