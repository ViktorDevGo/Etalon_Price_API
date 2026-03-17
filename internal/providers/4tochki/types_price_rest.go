package fourtochki

import "encoding/xml"

// PriceRestFilter represents filter for GetGoodsPriceRestByCode
type PriceRestFilter struct {
	XMLName              xml.Name        `xml:"wcf:filter"`
	CodeList             StringArrayTS3  `xml:"ts3:code_list"`
	IncludePaidDelivery  bool            `xml:"ts3:include_paid_delivery"`
}

// StringArrayTS3 represents array of strings with ts3 namespace
type StringArrayTS3 struct {
	XMLName xml.Name `xml:"ts3:code_list"`
	Items   []string `xml:"arr:string,omitempty"`
}

// GetGoodsPriceRestByCodeRequest represents GetGoodsPriceRestByCode SOAP request
type GetGoodsPriceRestByCodeRequest struct {
	XMLName  xml.Name        `xml:"wcf:GetGoodsPriceRestByCode"`
	Login    string          `xml:"wcf:login"`
	Password string          `xml:"wcf:password"`
	Filter   PriceRestFilter `xml:"wcf:filter"`
}

// GetGoodsPriceRestByCodeResponse represents GetGoodsPriceRestByCode SOAP response
type GetGoodsPriceRestByCodeResponse struct {
	XMLName xml.Name          `xml:"GetGoodsPriceRestByCodeResponse"`
	Result  PriceRestInfoList `xml:"GetGoodsPriceRestByCodeResult"`
}

// PriceRestInfoList contains array of price/stock information
type PriceRestInfoList struct {
	Items []PriceRestInfoItem `xml:"price_rest_list>PriceRestInfo,omitempty"`
	Error *APIError           `xml:"error,omitempty"`
}

// PriceRestInfoItem represents price and stock information for a product
type PriceRestInfoItem struct {
	Code          string  `xml:"Code"`          // Article/SKU
	Price         float64 `xml:"Price"`         // Price in RUB
	Currency      string  `xml:"Currency"`      // Currency code (RUB, USD, EUR)
	Quantity      int     `xml:"Quantity"`      // Available quantity
	WarehouseCode string  `xml:"WarehouseCode"` // Warehouse identifier
	WarehouseName string  `xml:"WarehouseName"` // Warehouse name
	DeliveryDays  int     `xml:"DeliveryDays"`  // Estimated delivery time in days
	MinOrder      int     `xml:"MinOrder"`      // Minimum order quantity
	Available     string  `xml:"Available"`     // "Y" if available for order
	Stock         string  `xml:"Stock"`         // Stock status: "IN_STOCK", "ON_ORDER", "OUT_OF_STOCK"
	Updated       string  `xml:"Updated"`       // Last update timestamp
	Region        string  `xml:"Region"`        // Region/city
}
