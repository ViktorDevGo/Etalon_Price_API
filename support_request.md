# Обращение в техподдержку 4tochki

## Текст письма для поддержки

---

**Тема:** Ошибка 50 при запросе информации о товарах через SOAP API

Здравствуйте!

Мы интегрируем наш сервис с вашим SOAP API и столкнулись с проблемой. При выполнении запросов `GetGoodsInfo` и `GetGoodsPriceRestByCode` API возвращает ошибку с кодом 50:

```
<b:code>50</b:code>
<b:comment>Неизвестная ошибка. Обратитесь к администратору веб-сервиса.</b:comment>
```

**Данные учетной записи:**
- Логин: `sa69263`
- WSDL URL: `http://api-b2b.4tochki.ru/WCF/ClientService.svc?wsdl`

**Коды товаров, которые мы пытаемся запросить:**
- 1027210
- 6924590270688
- 6924590215702
- ETL28403000
- ETL00015200
- RA149901
- RA185101
- 1026634
- ETL00097900

**Вопросы:**

1. **Существуют ли эти коды товаров в вашей системе?** Или они недействительны?

2. **Имеет ли наша учетная запись (`sa69263`) доступ к этим товарам?** Возможно, требуется настройка прав доступа?

3. **Какой правильный формат кодов товаров?** Может быть, нужен префикс или другой формат?

4. **Можете ли вы предоставить несколько тестовых кодов товаров**, которые точно существуют и доступны для нашей учетной записи? Это поможет нам завершить интеграцию.

5. **Требуются ли дополнительные настройки** для работы с SOAP API?

---

## Технические детали для разработчиков

### SOAP Запрос GetGoodsInfo

**HTTP Headers:**
```
POST http://api-b2b.4tochki.ru/WCF/ClientService.svc?wsdl
Content-Type: text/xml; charset=utf-8
SOAPAction: Wcf.ClientService.Client.WebAPI.TS3/ClientService/GetGoodsInfo
```

**Request Body (XML):**
```xml
<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
  <s:Body>
    <GetGoodsInfo xmlns="Wcf.ClientService.Client.WebAPI.TS3">
      <login>sa69263</login>
      <password>jkHP4Nj3)z</password>
      <codes>
        <string>1027210</string>
        <string>6924590270688</string>
      </codes>
    </GetGoodsInfo>
  </s:Body>
</s:Envelope>
```

**Response (XML):**
```xml
<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
  <s:Body>
    <GetGoodsInfoResponse xmlns="Wcf.ClientService.Client.WebAPI.TS3">
      <GetGoodsInfoResult
          xmlns:a="http://schemas.datacontract.org/2004/07/TS3.Domain.Models.Client.ClientSoapService.SearchProductsDetailedInfo"
          xmlns:i="http://www.w3.org/2001/XMLSchema-instance">
        <a:cameraList i:nil="true"/>
        <a:fastenerList i:nil="true"/>
        <a:oilList i:nil="true"/>
        <a:pressureSensorList i:nil="true"/>
        <a:rimList i:nil="true"/>
        <a:sparePartList i:nil="true"/>
        <a:tyreList i:nil="true"/>
        <a:error xmlns:b="http://schemas.datacontract.org/2004/07/TS3.Domain.Models.Client.ClientSoapService">
          <b:code>50</b:code>
          <b:comment>Неизвестная ошибка. Обратитесь к администратору веб-сервиса.</b:comment>
        </a:error>
        <a:wheelList i:nil="true"/>
      </GetGoodsInfoResult>
    </GetGoodsInfoResponse>
  </s:Body>
</s:Envelope>
```

**HTTP Status:** 200 OK

**Проблема:** Все списки товаров (`tyreList`, `rimList`, и т.д.) возвращаются как `nil`, а в ответе присутствует элемент `error` с кодом 50.

---

### SOAP Запрос GetGoodsPriceRestByCode

**HTTP Headers:**
```
POST http://api-b2b.4tochki.ru/WCF/ClientService.svc?wsdl
Content-Type: text/xml; charset=utf-8
SOAPAction: Wcf.ClientService.Client.WebAPI.TS3/ClientService/GetGoodsPriceRestByCode
```

**Request Body (XML):**
```xml
<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
  <s:Body>
    <GetGoodsPriceRestByCode xmlns="Wcf.ClientService.Client.WebAPI.TS3">
      <login>sa69263</login>
      <password>jkHP4Nj3)z</password>
      <codes>
        <string>1027210</string>
        <string>6924590270688</string>
      </codes>
    </GetGoodsPriceRestByCode>
  </s:Body>
</s:Envelope>
```

**Response (XML):**
```xml
<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
  <s:Body>
    <GetGoodsPriceRestByCodeResponse xmlns="Wcf.ClientService.Client.WebAPI.TS3">
      <GetGoodsPriceRestByCodeResult
          xmlns:a="http://schemas.datacontract.org/2004/07/TS3.Domain.Models.Client.ClientSoapService.SearchProductsAndPricesInfo"
          xmlns:i="http://www.w3.org/2001/XMLSchema-instance">
        <a:error xmlns:b="http://schemas.datacontract.org/2004/07/TS3.Domain.Models.Client.ClientSoapService">
          <b:code>50</b:code>
          <b:comment>Неизвестная ошибка. Обратитесь к администратору веб-сервиса.</b:comment>
        </a:error>
        <a:price_rest_list i:nil="true"/>
      </GetGoodsPriceRestByCodeResult>
    </GetGoodsPriceRestByCodeResponse>
  </s:Body>
</s:Envelope>
```

**HTTP Status:** 200 OK

**Проблема:** Список цен и остатков (`price_rest_list`) возвращается как `nil`, а в ответе присутствует элемент `error` с кодом 50.

---

## Что мы уже проверили

✅ SOAP запросы формируются правильно (корректные namespace и SOAPAction)
✅ Аутентификация проходит успешно (API не возвращает ошибки аутентификации)
✅ HTTP соединение работает (получаем HTTP 200 OK)
✅ XML структура соответствует спецификации WSDL

❌ API возвращает ошибку 50 для всех запрошенных кодов товаров

---

## Ожидаемый результат

Мы ожидаем получить в ответе информацию о товарах (шины, диски) с заполненными полями:
- `tyreList` или `rimList` с деталями товаров
- `price_rest_list` с ценами и остатками
- **Без** элемента `error` в ответе

---

Просим помочь разобраться в причине ошибки 50 и предоставить корректные тестовые данные для завершения интеграции.

С уважением,
[Ваше имя]
[Контактная информация]
