package fourtochki

import "encoding/xml"

// EmptyFilter represents an empty filter element
type EmptyFilter struct {
	XMLName xml.Name `xml:"wcf:filter"`
}

// GetFindTyreRequest represents GetFindTyre SOAP request
type GetFindTyreRequest struct {
	XMLName  xml.Name    `xml:"wcf:GetFindTyre"`
	Login    string      `xml:"wcf:login"`
	Password string      `xml:"wcf:password"`
	Filter   EmptyFilter `xml:"wcf:filter"`
	Page     int         `xml:"wcf:page"`
	PageSize int         `xml:"wcf:pageSize"`
}

// GetFindTyreResponse represents GetFindTyre SOAP response
type GetFindTyreResponse struct {
	XMLName xml.Name     `xml:"GetFindTyreResponse"`
	Result  FindTyreList `xml:"GetFindTyreResult"`
}

// FindTyreList contains list of tyres with prices and stock
type FindTyreList struct {
	Items []TyrePriceRest `xml:"price_rest_list>TyrePriceRest,omitempty"`
	Error *APIError       `xml:"error,omitempty"`
}

// TyrePriceRest represents a tyre with prices and stock information
type TyrePriceRest struct {
	Code   string               `xml:"code"`
	Marka  string               `xml:"marka"`
	Model  string               `xml:"model"`
	Name   string               `xml:"name"`
	Season string               `xml:"season"`
	Thorn  string               `xml:"thorn"`
	Type   string               `xml:"type"`
	Whpr   []WarehousePriceRest `xml:"whpr>wh_price_rest,omitempty"`
}

// GetFindDiskRequest represents GetFindDisk SOAP request
type GetFindDiskRequest struct {
	XMLName  xml.Name    `xml:"wcf:GetFindDisk"`
	Login    string      `xml:"wcf:login"`
	Password string      `xml:"wcf:password"`
	Filter   EmptyFilter `xml:"wcf:filter"`
	Page     int         `xml:"wcf:page"`
	PageSize int         `xml:"wcf:pageSize"`
}

// GetFindDiskResponse represents GetFindDisk SOAP response
type GetFindDiskResponse struct {
	XMLName xml.Name     `xml:"GetFindDiskResponse"`
	Result  FindDiskList `xml:"GetFindDiskResult"`
}

// FindDiskList contains list of disks with prices and stock
type FindDiskList struct {
	Items []DiskPriceRest `xml:"price_rest_list>DiskPriceRest,omitempty"`
	Error *APIError       `xml:"error,omitempty"`
}

// DiskPriceRest represents a disk with prices and stock information
type DiskPriceRest struct {
	Code   string               `xml:"code"`
	Marka  string               `xml:"marka"`
	Model  string               `xml:"model"`
	Name   string               `xml:"name"`
	Type   string               `xml:"type"`
	Whpr   []WarehousePriceRest `xml:"whpr>wh_price_rest,omitempty"`
}

// WarehousePriceRest represents price and stock at a specific warehouse
type WarehousePriceRest struct {
	Price int `xml:"price"` // Price in kopecks or minimal currency units
	Rest  int `xml:"rest"`  // Stock quantity
	Wrh   int `xml:"wrh"`   // Warehouse ID
}
