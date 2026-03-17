package fourtochki

import "encoding/xml"

// StringArray represents array of strings in SOAP
type StringArray struct {
	XMLName xml.Name `xml:"wcf:code_list"`
	Items   []string `xml:"arr:string,omitempty"`
}

// GetGoodsInfoRequest represents GetGoodsInfo SOAP request
type GetGoodsInfoRequest struct {
	XMLName  xml.Name    `xml:"wcf:GetGoodsInfo"`
	Login    string      `xml:"wcf:login"`
	Password string      `xml:"wcf:password"`
	CodeList StringArray `xml:"wcf:code_list"`
}

// APIError represents API error in response
type APIError struct {
	Code    int    `xml:"code"`
	Comment string `xml:"comment"`
}

// GetGoodsInfoResponse represents GetGoodsInfo SOAP response
type GetGoodsInfoResponse struct {
	XMLName xml.Name      `xml:"GetGoodsInfoResponse"`
	Result  GoodsInfoList `xml:"GetGoodsInfoResult"`
}

// GoodsInfoList contains array of goods
type GoodsInfoList struct {
	TyreList []GoodsInfoItem `xml:"tyreList>TyreContainer,omitempty"`
	RimList  []GoodsInfoItem `xml:"rimList>RimContainer,omitempty"`
	Error    *APIError       `xml:"error,omitempty"`
}

// GoodsInfoItem represents a single product (tire or rim)
type GoodsInfoItem struct {
	Code        string  `xml:"code"`        // Article/SKU
	Type        string  `xml:"type"`        // Product type: "car", "vned", etc
	Brand       string  `xml:"brand"`       // Brand name
	Model       string  `xml:"model"`       // Model name
	Name        string  `xml:"name"`        // Full product name

	// Tire specific fields
	Width       float64 `xml:"width,omitempty"`       // Width in mm
	Height      float64 `xml:"height,omitempty"`      // Aspect ratio
	Diameter    float64 `xml:"diameter,omitempty"`    // Rim diameter in inches
	Season      string  `xml:"season,omitempty"`      // "s" (summer), "w" (winter), "m" (all-season)
	Thorn       string  `xml:"thorn,omitempty"`       // "true" if studded
	Constr      string  `xml:"constr,omitempty"`      // Construction: "R", "ZR"
	LoadIndex   string  `xml:"load_index,omitempty"`  // Load index
	SpeedIndex  string  `xml:"speed_index,omitempty"` // Speed rating
	Tonnage     string  `xml:"tonnage,omitempty"`     // Extra load: "XL"
	Camera      string  `xml:"camera,omitempty"`      // Tube type: "TL" (tubeless)
	Weight      float64 `xml:"weight,omitempty"`      // Weight in kg
	Volume      float64 `xml:"volume,omitempty"`      // Volume in m³

	// Rim specific fields
	DiskWidth    float64 `xml:"disk_width,omitempty"`    // Rim width in inches
	DiskDiameter float64 `xml:"disk_diameter,omitempty"` // Rim diameter in inches
	Drilling     string  `xml:"drilling,omitempty"`      // Bolt pattern
	Radius       string  `xml:"radius,omitempty"`        // Offset
	CentralHole  string  `xml:"central_hole,omitempty"`  // Center bore
	Color        string  `xml:"color,omitempty"`         // Rim color/finish
}
