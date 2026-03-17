package fourtochki

import (
	"encoding/xml"
)

// TyresXMLRoot represents root element of Tires.xml
type TyresXMLRoot struct {
	XMLName xml.Name `xml:"root"`
	Tyres   []TyreXML `xml:"tires"`
}

// TyreXML represents single tyre from XML export
type TyreXML struct {
	CAE         string  `xml:"cae"`
	Name        string  `xml:"name"`
	Width       float64 `xml:"width"`
	Height      float64 `xml:"height"`
	Diameter    string  `xml:"diameter"`
	DiameterOut string  `xml:"diameter_out"`
	LoadIndex   string  `xml:"load_index"`
	SpeedIndex  string  `xml:"speed_index"`
	Model       string  `xml:"model"`
	Brand       string  `xml:"brand"`
	Season      string  `xml:"season"`
	IsStudded   string  `xml:"is_studded"`
	TireType    string  `xml:"tiretype"`
	RunFlat     string  `xml:"runflat"`
	Reinforced  string  `xml:"reinforced"`
}

// RimsXMLRoot represents root element of Rims.xml
type RimsXMLRoot struct {
	XMLName xml.Name `xml:"root"`
	Rims    []RimXML `xml:"rims"`
}

// RimXML represents single rim from XML export
type RimXML struct {
	CAE           string  `xml:"cae"`
	Name          string  `xml:"name"`
	Width         float64 `xml:"width"`
	Diameter      float64 `xml:"diameter"`
	BoltsCount    int     `xml:"bolts_count"`
	BoltsSpacing  float64 `xml:"bolts_spacing"`
	BoltsSpacing2 float64 `xml:"bolts_spacing2"`
	ET            float64 `xml:"et"`
	DIA           float64 `xml:"dia"`
	Model         string  `xml:"model"`
	Brand         string  `xml:"brand"`
	Color         string  `xml:"color"`
	RimType       string  `xml:"rim_type"`
}
