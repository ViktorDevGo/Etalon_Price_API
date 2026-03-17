# Запрос в техподдержку 4tochki

**Дата:** 2026-03-13
**Аккаунт:** sa69263

---

## Проблема

Методы **GetFindTyre** и **GetFindDisk** возвращают ошибку 50 для нашего аккаунта.

При этом метод **GetGoodsInfo** работает отлично с теми же учетными данными.

---

## Вопросы к поддержке

### 1. Есть ли у аккаунта sa69263 доступ к методам GetFindTyre и GetFindDisk?

Если нет - как можно получить доступ?

### 2. Правильный ли формат наших запросов?

Ниже приведены точные SOAP запросы, которые мы отправляем.

---

## Наши SOAP запросы

### GetFindTyre (ошибка 50)

```xml
<?xml version="1.0" encoding="UTF-8"?>
<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/"
                  xmlns:wcf="Wcf.ClientService.Client.WebAPI.TS3"
                  xmlns:arr="http://schemas.microsoft.com/2003/10/Serialization/Arrays"
                  xmlns:ts3="http://schemas.datacontract.org/2004/07/TS3.Domain.Models.Client.ClientSoapService.SearchProductsAndPricesInfo">
  <soapenv:Body>
    <wcf:GetFindTyre>
      <wcf:login>sa69263</wcf:login>
      <wcf:password>jkHP4Nj3)z</wcf:password>
    </wcf:GetFindTyre>
  </soapenv:Body>
</soapenv:Envelope>
```

**Ответ:**
```xml
<a:error>
  <b:code>50</b:code>
  <b:comment>Неизвестная ошибка. Обратитесь к администратору веб-сервиса.</b:comment>
</a:error>
```

---

### GetFindDisk (ошибка 50)

```xml
<?xml version="1.0" encoding="UTF-8"?>
<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/"
                  xmlns:wcf="Wcf.ClientService.Client.WebAPI.TS3"
                  xmlns:arr="http://schemas.microsoft.com/2003/10/Serialization/Arrays"
                  xmlns:ts3="http://schemas.datacontract.org/2004/07/TS3.Domain.Models.Client.ClientSoapService.SearchProductsAndPricesInfo">
  <soapenv:Body>
    <wcf:GetFindDisk>
      <wcf:login>sa69263</wcf:login>
      <wcf:password>jkHP4Nj3)z</wcf:password>
    </wcf:GetFindDisk>
  </soapenv:Body>
</soapenv:Envelope>
```

**Ответ:**
```xml
<a:error>
  <b:code>50</b:code>
  <b:comment>Неизвестная ошибка. Обратитесь к администратору веб-сервиса.</b:comment>
</a:error>
```

---

### GetGoodsInfo (РАБОТАЕТ ✅)

Для сравнения - этот метод работает без проблем:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/"
                  xmlns:wcf="Wcf.ClientService.Client.WebAPI.TS3"
                  xmlns:arr="http://schemas.microsoft.com/2003/10/Serialization/Arrays">
  <soapenv:Body>
    <wcf:GetGoodsInfo>
      <wcf:login>sa69263</wcf:login>
      <wcf:password>jkHP4Nj3)z</wcf:password>
      <wcf:code_list>
        <arr:string>1027210</arr:string>
        <arr:string>1026634</arr:string>
      </wcf:code_list>
    </wcf:GetGoodsInfo>
  </soapenv:Body>
</soapenv:Envelope>
```

**Ответ:** HTTP 200 OK, без ошибок, получаем полную информацию о товарах.

---

## Что нам нужно

Согласно вашим рекомендациям, мы понимаем что:

1. **GetFindTyre** → возвращает ResultFindTyre с артикулами, ценами и остатками по шинам
2. **GetFindDisk** → возвращает ResultFindDisk с артикулами, ценами и остатками по дискам
3. **GetGoodsInfo** → получение детальной информации по конкретным артикулам

Нам нужно получить доступ к методам GetFindTyre и GetFindDisk для загрузки полного каталога товаров.

---

## Просим помочь

1. **Проверить права доступа** аккаунта sa69263 к методам GetFindTyre и GetFindDisk
2. **Активировать доступ** к этим методам, если он отключен
3. **Подтвердить корректность** наших SOAP запросов
4. Если возможно - **предоставить тестовый успешный ответ** от GetFindTyre (хотя бы с 1-2 товарами)

---

## Контакты

**Аккаунт:** sa69263
**Email:** [ваш email]
**Телефон:** [ваш телефон]

---

Спасибо за помощь!
