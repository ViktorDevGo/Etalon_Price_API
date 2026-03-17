# Руководство по экспорту в 1C-Битрикс

## Быстрый старт

```bash
# Экспорт 10 позиций для теста
go run cmd/export-bitrix-csv/main.go -limit=10

# Экспорт всех товаров С ценами (~8,272 позиций)
go run cmd/export-bitrix-csv/main.go -limit=0

# Экспорт всех товаров, ВКЛЮЧАЯ без цен (~60,637 позиций)
go run cmd/export-bitrix-csv/main.go -limit=0 -include-no-price

# Свой путь к файлу
go run cmd/export-bitrix-csv/main.go -output=/path/to/catalog.csv -limit=100
```

## Параметры

| Флаг | Значение по умолчанию | Описание |
|------|----------------------|----------|
| `-output` | `/Users/viktor/Desktop/bitrix_catalog.csv` | Путь к выходному CSV |
| `-limit` | `10` | Количество позиций (`0` = все) |
| `-include-no-price` | `false` | Включить товары без цен (с ценой `0.00`) |

## Проверки качества данных

При запуске автоматически выполняются проверки:

```
📊 Проверка качества данных...
----------------------------------------------------------------
✅ Товары без артикула: 0
⚠️  Товары без цены: 52365 (86.4%)
✅ Дубли по CAE: 0
ℹ️  Товары на нескольких складах: 4502 (используем MIN цену, SUM остатков)
----------------------------------------------------------------
```

### SQL запросы для ручной проверки

```sql
-- 1. Товары без артикула (КРИТИЧНО)
SELECT COUNT(*) FROM nomenclature_tyres WHERE cae IS NULL OR cae = '';

-- 2. Товары без цены
SELECT COUNT(*) FROM nomenclature_tyres t
LEFT JOIN tyres_prices_stock p ON t.cae = p.cae
WHERE p.cae IS NULL;

-- 3. Товары с нулевым остатком
SELECT COUNT(*) FROM tyres_prices_stock WHERE stock = 0;

-- 4. Товары без бренда/модели
SELECT COUNT(*) FROM nomenclature_tyres
WHERE brand IS NULL OR brand = '' OR model IS NULL OR model = '';

-- 5. Товары без размеров
SELECT COUNT(*) FROM nomenclature_tyres
WHERE width IS NULL OR height IS NULL OR diameter IS NULL;

-- 6. Дубли по CAE (не должно быть)
SELECT cae, COUNT(*) FROM nomenclature_tyres GROUP BY cae HAVING COUNT(*) > 1;

-- 7. Товары на нескольких складах
SELECT cae, COUNT(*) as warehouses,
       MIN(price)/100.0 as min_price,
       MAX(price)/100.0 as max_price,
       SUM(stock) as total_stock
FROM tyres_prices_stock
GROUP BY cae
HAVING COUNT(*) > 1
ORDER BY COUNT(*) DESC
LIMIT 10;
```

## Логика обработки

### 1. Связь таблиц

```sql
FROM nomenclature_tyres t
INNER JOIN tyres_prices_stock p ON t.cae = p.cae  -- Связь по артикулу
```

- **INNER JOIN** (по умолчанию): только товары с ценами
- **LEFT JOIN** (с флагом `-include-no-price`): все товары

### 2. Агрегация для товаров на нескольких складах

**Проблема:** 4,502 товара находятся на 2-13 складах с разными ценами

**Решение:**
```sql
MIN(p.price) as price_kopeks,  -- Минимальная цена (конкурентность)
SUM(p.stock) as total_stock    -- Суммарный остаток
GROUP BY t.cae, t.brand, t.model, ...
```

#### Пример:
```
Товар: YSTX5R1516
Склад 1: 32.94₽, 20 шт
Склад 2: 32.94₽, 18 шт
Склад 3: 32.94₽, 15 шт
...
Склад 13: 32.94₽, 17 шт
───────────────────────────
Итого: 32.94₽, 232 шт  ← В CSV
```

### 3. Нормализация данных

| Поле БД | Исходное | Нормализованное | Примечание |
|---------|----------|-----------------|------------|
| diameter | `"R16,00"` | `"16"` | Убрать R и ,00 |
| diameter | `"R17,00"` | `"17"` | |
| is_studded | `"шип"` | `"Шипованные"` | Преобразование |
| is_studded | `"не шип"` | `"Нешипованные"` | |
| price | `443600` (копейки) | `"4436.00"` | Делить на 100 |
| width | `185.0` | `"185"` | Без дробной части |
| height | `60.0` | `"60"` | |

### 4. Формирование названия

```go
"{brand} {model} {width}/{height} R{diameter} {load_index}{speed_index}"
```

#### Примеры:
- `Bridgestone Turanza GR80 185/60 R14 82H`
- `Michelin Pilot Alpin 5 225/40 R19 112T`
- `Viatti Brina Nordico V-522 185/60 R14 82T`

## Маппинг на Bitrix CSV

### Заполненные поля

| Bitrix Колонка | Источник | Пример | Описание |
|----------------|----------|--------|----------|
| IE_XML_ID | `cae` | `OTR_037` | Уникальный ID |
| IE_NAME | `generated` | `TopTrust SH215 12/80 R—1800` | Форматированное название |
| IP_PROP140 | `width` | `12` | Ширина (мм) |
| IP_PROP141 | `height` | `80` | Высота профиля (%) |
| IP_PROP142 | `diameter` | `1800` | Диаметр (нормализован) |
| IP_PROP145 | `season` | `Всесезонная` | Сезон |
| IP_PROP146 | `is_studded` | `Шипованные` | Шипы |
| IP_PROP149 | `load_index` | `82` | Индекс нагрузки |
| IP_PROP150 | `speed_index` | `H` | Индекс скорости |
| IP_PROP151 | `cae` | `OTR_037` | Артикул (дубль) |
| IP_PROP163 | `runflat` | пусто | RunFlat |
| IC_GROUP0 | `1` | `1` | Категория: шины |
| IC_GROUP1 | `brand` | `TopTrust` | Бренд |
| IC_GROUP2 | `model` | `SH215` | Модель |
| CP_QUANTITY | `SUM(stock)` | `6` | Остаток (агрегат) |
| CP_WEIGHT | `0` | `0` | Вес (нет данных) |
| CV_PRICE_1 | `MIN(price)/100` | `142.20` | Цена (мин, рубли) |
| CV_CURRENCY_1 | `RUB` | `RUB` | Валюта |

### Пустые поля

Следующие поля **пустые** (нет данных):
- `IE_PREVIEW_TEXT`, `IE_DETAIL_TEXT` - описания
- `IP_PROP138`, `IP_PROP139` - неизвестные свойства
- `IP_PROP143`, `IP_PROP144` - неизвестные свойства
- `IP_PROP147`, `IP_PROP148` - неизвестные свойства
- `IP_PROP152-160` - доп. свойства (разболтовка, ET, DIA для дисков)
- `IP_PROP161-162` - неизвестные свойства
- `IP_PROP164-188` - расширенные свойства
- `IP_PROP481-482` - доп. свойства
- `CP_WIDTH`, `CP_HEIGHT`, `CP_LENGTH` - габариты
- `CV_QUANTITY_FROM`, `CV_QUANTITY_TO` - диапазоны

## SQL запрос для экспорта

### Основной запрос (с агрегацией)

```sql
SELECT
    t.cae,
    COALESCE(t.brand, '') as brand,
    COALESCE(t.model, '') as model,
    t.width,
    t.height,
    t.diameter,
    COALESCE(t.season, '') as season,
    CASE
        WHEN t.is_studded = 'шип' THEN 'Шипованные'
        WHEN t.is_studded = 'не шип' THEN 'Нешипованные'
        ELSE ''
    END as studded,
    COALESCE(t.load_index, '') as load_index,
    COALESCE(t.speed_index, '') as speed_index,
    COALESCE(t.runflat, '') as runflat,
    COALESCE(MIN(p.price), 0) as price_kopeks,
    COALESCE(SUM(p.stock), 0) as total_stock
FROM nomenclature_tyres t
INNER JOIN tyres_prices_stock p ON t.cae = p.cae
GROUP BY t.id, t.cae, t.brand, t.model, t.width, t.height, t.diameter,
         t.season, t.is_studded, t.load_index, t.speed_index, t.runflat
ORDER BY t.id;
```

### Вариант с товарами без цен

```sql
SELECT ... FROM nomenclature_tyres t
LEFT JOIN tyres_prices_stock p ON t.cae = p.cae  -- ← LEFT JOIN
GROUP BY ...
```

Товары без цен получат:
- `price_kopeks = 0` → `CV_PRICE_1 = "0.00"`
- `total_stock = 0` → `CP_QUANTITY = "0"`

## Примеры результатов

### Образец из БД:

| CAE | Brand | Model | Warehouse | Price | Stock |
|-----|-------|-------|-----------|-------|-------|
| OTR_037 | TopTrust | SH215 | Новосибирск | 14220₽ | 2 |
| OTR_037 | TopTrust | SH215 | Домодедово | 14220₽ | 2 |
| OTR_037 | TopTrust | SH215 | Давыдово | 14220₽ | 2 |

### Результат в CSV:

```csv
IE_XML_ID,IE_NAME,...,CV_PRICE_1,CV_CURRENCY_1
OTR_037,"TopTrust SH215 12/80 R1800",...,142.20,RUB
```

- Цена: MIN(14220) = 14220 копеек = **142.20₽**
- Остаток: SUM(2+2+2) = **6 шт**

## Статистика экспорта

| Параметр | Значение |
|----------|----------|
| Всего товаров в БД | 60,637 |
| Товаров с ценами | 8,272 (13.6%) |
| Товаров без цен | 52,365 (86.4%) |
| Товаров на нескольких складах | 4,502 (54.4% от товаров с ценами) |
| Макс. складов на товар | 13 |
| Средняя цена | 65.42₽ |

## Импорт в Bitrix

### Шаг 1: Создать инфоблок

В админке Bitrix:
1. **Контент** → **Инфоблоки** → **Типы инфоблоков**
2. Создать тип: "Каталог"
3. Создать инфоблок: "Шины"
4. Добавить свойства (IP_PROP140-150, IC_GROUP0-2)

### Шаг 2: Импорт CSV

1. **Контент** → **Инфоблоки** → **Импорт** → **CSV**
2. Выбрать файл `bitrix_catalog.csv`
3. Указать разделитель: **запятая**
4. Кодировка: **UTF-8**
5. Сопоставить колонки
6. Запустить импорт

### Шаг 3: Проверка

```sql
-- В Bitrix БД
SELECT COUNT(*) FROM b_iblock_element WHERE IBLOCK_ID = X;
```

## Troubleshooting

### Проблема: "Duplicate entry for key 'XML_ID'"

**Причина:** Дубли по `cae` в БД

**Решение:**
```sql
SELECT cae, COUNT(*) FROM nomenclature_tyres GROUP BY cae HAVING COUNT(*) > 1;
```

### Проблема: "Invalid price format"

**Причина:** Цена не в формате `"1234.56"`

**Проверка:**
```bash
grep -v '^\d+\.\d\d$' bitrix_catalog.csv
```

### Проблема: Неправильное количество колонок

**Причина:** CSV содержит не 70 колонок

**Проверка:**
```bash
head -1 bitrix_catalog.csv | tr ',' '\n' | wc -l  # Должно быть 70
```

### Проблема: Кириллица в кракозябрах

**Решение:** Открыть в редакторе с UTF-8:
- Numbers (macOS) - автоопределение
- LibreOffice - выбрать UTF-8 при открытии
- Excel - File → Import → выбрать UTF-8

## Дополнительные возможности

### Фильтр по бренду

```go
// В buildQuery() добавить:
query += ` WHERE t.brand = 'Bridgestone'`
```

### Фильтр по сезону

```go
query += ` WHERE t.season = 'Зимняя'`
```

### Сортировка по популярности (остаткам)

```go
query += ` ORDER BY total_stock DESC`
```

### Экспорт только со скидкой

```go
query += ` HAVING MIN(p.price) < AVG(p.price) * 0.9`
```

## Производительность

| Количество | Время экспорта | Размер файла |
|------------|----------------|--------------|
| 10 | ~0.5 сек | ~2 KB |
| 100 | ~1 сек | ~20 KB |
| 1,000 | ~5 сек | ~200 KB |
| 8,272 (все с ценами) | ~40 сек | ~1.6 MB |
| 60,637 (все товары) | ~5 мин | ~12 MB |

## Автоматизация

### Cron job (ежедневный экспорт)

```bash
# /etc/crontab
0 3 * * * cd /path/to/project && go run cmd/export-bitrix-csv/main.go -limit=0
```

### Docker контейнер

```dockerfile
FROM golang:alpine
WORKDIR /app
COPY . .
RUN go build -o export cmd/export-bitrix-csv/main.go
CMD ["./export", "-limit=0", "-output=/data/bitrix_catalog.csv"]
```

## Контрольный список перед импортом

- [ ] Проверить количество записей в CSV
- [ ] Проверить формат цен (`1234.56`)
- [ ] Проверить наличие артикулов (IE_XML_ID)
- [ ] Проверить кодировку (UTF-8)
- [ ] Проверить количество колонок (70)
- [ ] Убедиться, что нет пустых строк
- [ ] Проверить дубли по XML_ID
- [ ] Проверить соответствие валюты (RUB)
