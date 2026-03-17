package fourtochki

import "encoding/xml"

// GetWarehousesRequest represents GetWarehouses SOAP request
type GetWarehousesRequest struct {
	XMLName  xml.Name `xml:"wcf:GetWarehouses"`
	Login    string   `xml:"wcf:login"`
	Password string   `xml:"wcf:password"`
}

// GetWarehousesResponse represents GetWarehouses SOAP response
type GetWarehousesResponse struct {
	XMLName xml.Name       `xml:"GetWarehousesResponse"`
	Result  WarehousesList `xml:"GetWarehousesResult"`
}

// WarehousesList contains list of warehouses
type WarehousesList struct {
	Warehouses []WarehouseInfo `xml:"warehouses>WarehouseInfo"`
	Success    bool            `xml:"success"`
	Error      *APIError       `xml:"error"`
}

// WarehouseInfo represents warehouse information
type WarehouseInfo struct {
	ID              int    `xml:"id"`
	Key             string `xml:"key"`
	Name            string `xml:"name"`
	ShortName       string `xml:"shortName"`
	StoID           int    `xml:"stoID"`
	BackgroundColor string `xml:"background-color"`
	HaveDelivery    bool   `xml:"haveDelivery"`
	HavePickup      bool   `xml:"havePickup"`
	IsPaidDelivery  bool   `xml:"isPaidDelivery"`
	LogisticDays    int    `xml:"logisticDays"`
}
