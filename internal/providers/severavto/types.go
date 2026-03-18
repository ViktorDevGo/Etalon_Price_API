package severavto

import "encoding/xml"

// CommoditiesXML представляет корневой элемент XML ответа от API Северавто
type CommoditiesXML struct {
	XMLName     xml.Name     `xml:"COMMODITIES"`
	Description Description  `xml:"DESCRIPTION"` // Описание полей (meta)
	Commodities []Commodity  `xml:"COMMODITY"`   // Список товаров
}

// Description содержит описание полей в XML (не используется в обработке)
type Description struct {
	XMLName xml.Name `xml:"DESCRIPTION"`
}

// Commodity представляет один товар с остатками (одна запись = товар на одной территории)
type Commodity struct {
	// Идентификаторы
	ID      string `xml:"NNCOMMODIF"` // Уникальный ID товара (RN, commodityId)
	Article string `xml:"SARTICLE"`   // Артикул товара

	// Основная информация
	Brand string `xml:"SBREND"` // Бренд (производитель)
	Model string `xml:"SMODEL"` // Модель

	// Параметры шины
	Width  string `xml:"SWIDTH"`  // Ширина (195, 205 и т.д.)
	Height string `xml:"SHEIGHT"` // Профиль/высота (65, 55 и т.д.)
	Radius string `xml:"SRADIUS"` // Радиус (R15, R16 и т.д.)

	// Дополнительные параметры шины
	Thorning      string `xml:"STHORNING"`    // Шипы (Да/Нет)
	Speed         string `xml:"SSPEED"`       // Индекс скорости (T, H, V и т.д.)
	Load          string `xml:"SLOAD"`        // Индекс нагрузки (91, 95 и т.д.)
	ThornType     string `xml:"STHORNTYPE"`   // Вид ошиповки
	ManufactureCode string `xml:"SMNFCODE"`   // Код производителя

	// Параметры диска (если это диск)
	// ET, DIA, PCD и т.д. (аналогично шинам, но другие названия полей)

	// Территория и остатки
	PlaceType     string `xml:"SPLACE"`         // Тип места (Склад / Пункт выдачи)
	TerritoryName string `xml:"TERRITORY_NAME"` // Название территории (Новосибирск, Москва и т.д.)
	Rest          string `xml:"NREST"`          // Остаток (количество)

	// Цены (в рублях!)
	PriceDelayed    string `xml:"NPRICE_DELAYED"`    // Цена по отсрочке
	PricePrepayment string `xml:"NPRICE_PREPAYMENT"` // Цена по предоплате
	PriceRRP        string `xml:"NPRICE_RRP"`        // Рекомендованная розничная цена
	PreorderPrice   string `xml:"NPREORDER_PRICE"`   // Цена под заказ
	PriceCliD       string `xml:"NPRICE_CLI_D"`      // Клиентская цена (отсрочка)
	PriceCliP       string `xml:"NPRICE_CLI_P"`      // Клиентская цена (предоплата)

	// Дополнительная информация
	Sellout        string `xml:"SSELLOUT"`    // Распродажа (Да/Нет)
	Picture        string `xml:"SPICTURE"`    // Картинка (URL)
	ManufactureYear string `xml:"SMNFYEAR"`   // Год производства
}
