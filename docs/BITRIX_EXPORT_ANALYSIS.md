# Анализ экспорта в 1C-Битрикс

## Шаг 1: Структура таблиц

### nomenclature_tyres (60,637 записей)

```sql
id              BIGSERIAL PRIMARY KEY
cae             VARCHAR(50) UNIQUE NOT NULL  -- Артикул (ключ связи)
name            TEXT NOT NULL
width           DECIMAL(10,2)
height          DECIMAL(10,2)
diameter        VARCHAR(20)                  -- Формат: "R16,00"
diameter_out    VARCHAR(20)
load_index      VARCHAR(20)                  -- "82", "102/100"
speed_index     VARCHAR(10)                  -- "H", "T", "V"
model           VARCHAR(255)
brand           VARCHAR(255)
season          VARCHAR(50)                  -- "Зимняя", "Летняя", "Всесезонная"
is_studded      VARCHAR(20)                  -- "шип", "не шип"
tiretype        VARCHAR(50)                  -- "Легковая" и т.д.
runflat         VARCHAR(50)
reinforced      VARCHAR(50)
```

### tyres_prices_stock (18,743 записей)

```sql
id              BIGSERIAL PRIMARY KEY
cae             VARCHAR(50) NOT NULL         -- Артикул (ключ связи)
warehouse_name  VARCHAR(255)                 -- Название склада
price           INTEGER NOT NULL             -- Цена в КОПЕЙКАХ
stock           INTEGER NOT NULL             -- Остаток (шт)
provider        VARCHAR(100)                 -- "Форточки"
isimport        INTEGER                      -- 0=загружено, 1=использовано
```

## Шаг 2: Ключ связи

**Поле связи:** `cae` (артикул товара)

- ✅ UNIQUE в nomenclature_tyres
- ✅ NOT NULL в обеих таблицах
- ✅ Индексирован
- ⚠️ Покрытие: только 13.6% товаров имеют цены (8,272 из 60,637)

## Шаг 3: Проблема дублей

**4,502 товара** находятся на **нескольких складах** (2-13 складов на товар)

### Примеры:
- `YSTX5R1516`: 13 складов, цена 32.94₽, всего 232 шт
- `CTS283246`: 13 складов, цена 29.34₽, всего 230 шт

### Стратегии обработки:

#### Вариант 1: **МИНИМАЛЬНАЯ ЦЕНА** (рекомендуется для конкурентности)
```sql
MIN(price) as price,
SUM(stock) as total_stock
```

#### Вариант 2: **МАКСИМАЛЬНАЯ ЦЕНА** (для максимизации прибыли)
```sql
MAX(price) as price,
SUM(stock) as total_stock
```

#### Вариант 3: **СРЕДНЯЯ ЦЕНА** (компромисс)
```sql
AVG(price) as price,
SUM(stock) as total_stock
```

#### Вариант 4: **СОЗДАТЬ НЕСКОЛЬКО ТОРГОВЫХ ПРЕДЛОЖЕНИЙ**
- Один товар → несколько SKU (по складам)
- Требует расширения CSV (столбцы для складов)

**Для Битрикс выбираем Вариант 1** (минимальная цена, суммарный остаток)

## Шаг 4: Маппинг на Bitrix CSV

| # | Bitrix Колонка | Источник | Пример | Примечание |
|---|----------------|----------|--------|------------|
| 1 | IE_XML_ID | `cae` | 2206 | Уникальный ID товара |
| 2 | IE_NAME | generated | "Bridgestone Turanza GR80 185/60 R14 82H" | Форматируем из полей |
| 3 | IE_PREVIEW_TEXT | - | пусто | Нет данных |
| 4 | IE_DETAIL_TEXT | - | пусто | Нет данных |
| 5 | IP_PROP138 | - | пусто | Неизвестно |
| 6 | IP_PROP139 | - | пусто | Неизвестно |
| 7 | IP_PROP140 | `width` | 185 | Ширина (мм) |
| 8 | IP_PROP141 | `height` | 60 | Высота профиля (%) |
| 9 | IP_PROP142 | `diameter` | 14 | Диаметр (нормализовать "R16,00" → "14") |
| 10 | IP_PROP143 | - | пусто | Неизвестно |
| 11 | IP_PROP144 | - | пусто | Неизвестно |
| 12 | IP_PROP145 | `season` | "Летняя" | Сезон |
| 13 | IP_PROP189 | - | пусто | Неизвестно |
| 14 | IP_PROP146 | `is_studded` | "Нешипованные" | Шипы (преобразовать) |
| 15 | IP_PROP147 | - | пусто | Неизвестно |
| 16 | IP_PROP148 | - | пусто | Неизвестно |
| 17 | IP_PROP149 | `load_index` | 82 | Индекс нагрузки |
| 18 | IP_PROP150 | `speed_index` | H | Индекс скорости |
| 19 | IP_PROP151 | `cae` | PITR300403 | Артикул (дубль) |
| 20-28 | IP_PROP152-160 | - | пусто | Нет данных |
| 29 | IP_PROP161 | - | пусто | Неизвестно |
| 30 | IP_PROP162 | - | пусто | Неизвестно |
| 31 | IP_PROP163 | `runflat` | пусто | RunFlat |
| 32-58 | IP_PROP164-188 | - | пусто | Нет данных |
| 59 | IC_GROUP0 | constant | 1 | Категория: 1=шины |
| 60 | IC_GROUP1 | `brand` | "Bridgestone" | Бренд |
| 61 | IC_GROUP2 | `model` | "Turanza GR80" | Модель |
| 62 | CP_QUANTITY | `SUM(stock)` | 1000 | Остаток (агрегат) |
| 63 | CP_WEIGHT | - | 0 | Вес (нет данных) |
| 64-66 | CP_WIDTH, CP_HEIGHT, CP_LENGTH | - | пусто | Габариты (нет данных) |
| 67-68 | CV_QUANTITY_FROM, CV_QUANTITY_TO | - | пусто | Диапазон цен |
| 69 | CV_PRICE_1 | `MIN(price)/100` | 4436.00 | Цена в рублях (мин) |
| 70 | CV_CURRENCY_1 | constant | RUB | Валюта |

## Шаг 5: Обработка данных

### Нормализация полей

1. **Диаметр**: `"R16,00"` → `"16"` (убрать "R" и ",00")
2. **Шипы**: `"шип"` → `"Шипованные"`, `"не шип"` → `"Нешипованные"`
3. **Цена**: копейки → рубли с 2 знаками (`price / 100.0` → `"4436.00"`)
4. **Название**: `"{brand} {model} {width}/{height} R{diameter} {load_index}{speed_index}"`

### Проверки качества

```sql
-- 1. Товары без артикула (критично)
SELECT COUNT(*) FROM nomenclature_tyres WHERE cae IS NULL OR cae = '';

-- 2. Товары без цены (показать в CSV с ценой 0.00)
SELECT COUNT(*) FROM nomenclature_tyres t
LEFT JOIN tyres_prices_stock p ON t.cae = p.cae
WHERE p.cae IS NULL;

-- 3. Товары без остатка (но с ценой)
SELECT COUNT(*) FROM tyres_prices_stock WHERE stock = 0;

-- 4. Товары без бренда/модели
SELECT COUNT(*) FROM nomenclature_tyres
WHERE brand IS NULL OR brand = '' OR model IS NULL OR model = '';

-- 5. Товары без размеров
SELECT COUNT(*) FROM nomenclature_tyres
WHERE width IS NULL OR height IS NULL OR diameter IS NULL;

-- 6. Дубли по cae (не должно быть)
SELECT cae, COUNT(*) FROM nomenclature_tyres GROUP BY cae HAVING COUNT(*) > 1;
```

## Шаг 6: SQL запрос для экспорта

### Основной запрос (с агрегацией)

```sql
SELECT
    t.cae as xml_id,
    -- Форматированное название
    CONCAT(
        COALESCE(t.brand, ''), ' ',
        COALESCE(t.model, ''), ' ',
        COALESCE(t.width::TEXT, ''), '/',
        COALESCE(t.height::TEXT, ''), ' ',
        'R', REGEXP_REPLACE(COALESCE(t.diameter, ''), '[R,]', '', 'g'), ' ',
        COALESCE(t.load_index, ''),
        COALESCE(t.speed_index, '')
    ) as name,

    -- Параметры
    t.width,
    t.height,
    REGEXP_REPLACE(COALESCE(t.diameter, ''), '[R,]', '', 'g') as diameter_clean,
    t.season,
    CASE
        WHEN t.is_studded = 'шип' THEN 'Шипованные'
        WHEN t.is_studded = 'не шип' THEN 'Нешипованные'
        ELSE ''
    END as studded,
    t.load_index,
    t.speed_index,
    t.brand,
    t.model,
    t.runflat,

    -- Агрегированные цена и остаток
    COALESCE(MIN(p.price), 0)::NUMERIC / 100.0 as price_rub,
    COALESCE(SUM(p.stock), 0) as total_stock

FROM nomenclature_tyres t
LEFT JOIN tyres_prices_stock p ON t.cae = p.cae
GROUP BY t.id, t.cae, t.name, t.width, t.height, t.diameter,
         t.season, t.is_studded, t.load_index, t.speed_index,
         t.brand, t.model, t.runflat
ORDER BY t.id
LIMIT 10;
```

### Запасной вариант (без агрегации, только первый склад)

```sql
SELECT DISTINCT ON (t.cae)
    t.cae, t.brand, t.model, t.width, t.height, t.diameter,
    t.season, t.is_studded, t.load_index, t.speed_index,
    p.price / 100.0 as price_rub, p.stock
FROM nomenclature_tyres t
LEFT JOIN tyres_prices_stock p ON t.cae = p.cae
ORDER BY t.cae, p.price ASC NULLS LAST  -- Берём минимальную цену
```

## Шаг 7: Статистика покрытия

| Параметр | Значение | % |
|----------|----------|---|
| Всего товаров | 60,637 | 100% |
| Товаров с ценами | 8,272 | 13.6% |
| Товаров без цен | 52,365 | 86.4% |
| Товаров на нескольких складах | 4,502 | 54.4% от товаров с ценами |
| Товаров с уникальной ценой | 3,770 | 45.6% от товаров с ценами |

## Шаг 8: Рекомендации

### ✅ Что делать с товарами без цен:

**Вариант A** (рекомендуется): Экспортировать с ценой 0.00 и остатком 0
- Плюсы: полный каталог, SEO
- Минусы: неактуальные товары в каталоге

**Вариант B**: Не экспортировать вообще
- Плюсы: только актуальные товары
- Минусы: маленький каталог (13.6%)

**Вариант C**: Экспортировать, но пометить "под заказ"
- Плюсы: полный каталог, потенциал заказов
- Минусы: требует доработки логики

### ✅ Что делать с полями, которых нет:

- IP_PROP152-160, IP_PROP164-188: **оставить пустыми**
- CP_WEIGHT, CP_WIDTH, CP_HEIGHT, CP_LENGTH: **заполнить "0"** или пустыми
- IE_PREVIEW_TEXT, IE_DETAIL_TEXT: **пустые** (можно генерировать из параметров)

### ✅ Приоритетные поля для заполнения:

1. ✅ IE_XML_ID (cae)
2. ✅ IE_NAME (generated)
3. ✅ IP_PROP140-142 (width, height, diameter)
4. ✅ IP_PROP145 (season)
5. ✅ IP_PROP146 (studded)
6. ✅ IP_PROP149-150 (load_index, speed_index)
7. ✅ IC_GROUP0-2 (категория, бренд, модель)
8. ✅ CP_QUANTITY (stock)
9. ✅ CV_PRICE_1 (price)
10. ✅ CV_CURRENCY_1 (RUB)
