# Логика экспорта для 1C-Битрикс

## Обзор

Экспорт номенклатуры (шин и дисков) в формат 1C-Битрикс с применением бизнес-правил по ценам и остаткам.

## Бизнес-правила

### 1. Остатки (CP_QUANTITY / CATALOG_QUANTITY)

**Источник:** Таблица `tyres_prices_stock` или `rims_prices_stock`, колонка `stock`

**Логика:**
1. Суммируем все остатки по одному `cae` (артикулу) со всех складов
2. **Если сумма > 4** → берем товар в экспорт
3. **Если сумма ≤ 4** → пропускаем (skip)

**SQL:**
```sql
SUM(p.stock) as total_stock
...
HAVING SUM(p.stock) > 4
```

**Пример:**
- Склад 1: 3 шт
- Склад 2: 2 шт
- Склад 3: 1 шт
- **Итого: 6 шт > 4** → ✅ БЕРЕМ

- Склад 1: 2 шт
- Склад 2: 1 шт
- **Итого: 3 шт ≤ 4** → ❌ SKIP

### 2. Цены (CV_PRICE_1)

**Источник:** Таблица `tyres_prices_stock` или `rims_prices_stock`, колонка `price`

**Логика:**
1. Берем все цены по одному `cae` (артикулу) со всех складов
2. Выбираем **минимальную** цену
3. **Увеличиваем на 12%** (наценка)
4. Конвертируем из копеек в рубли

**SQL:**
```sql
CAST(MIN(p.price) * 1.12 AS INTEGER) as price_with_markup
```

**Пример:**
- Склад 1: 10000 коп (100₽)
- Склад 2: 12000 коп (120₽)
- Склад 3: 9000 коп (90₽)
- **MIN = 9000 коп (90₽)**
- **90₽ * 1.12 = 100.80₽** → результат

## Использование

### Экспорт шин

```bash
# Все записи
go run cmd/export-bitrix-updated/main.go -type=tyres -output=/Users/viktor/Desktop/tyres_export.csv

# Тестовые 10 записей
go run cmd/export-bitrix-updated/main.go -type=tyres -output=/Users/viktor/Desktop/tyres_test.csv -limit=10
```

### Экспорт дисков

```bash
# Все записи
go run cmd/export-bitrix-updated/main.go -type=rims -output=/Users/viktor/Desktop/rims_export.csv

# Тестовые 10 записей
go run cmd/export-bitrix-updated/main.go -type=rims -output=/Users/viktor/Desktop/rims_test.csv -limit=10
```

## Формат вывода

**Колонок:** 74
**Формат:** CSV

**Основные поля:**
- `IE_XML_ID` - Артикул (CAE)
- `IE_NAME` - Название товара
- `IP_PROP140-142` - Ширина, Высота, Диаметр
- `IP_PROP145` - Сезон
- `IC_GROUP0-2` - Категории
- `CP_QUANTITY` - Остаток (SUM с фильтром > 4)
- `CATALOG_QUANTITY` - Дубль остатка
- `CV_PRICE_1` - Цена (MIN * 1.12)
- `CV_CURRENCY_1` - RUB

## Примеры расчетов

### Пример 1: Товар берется

**Артикул:** 024701 (Michelin)

**Остатки:**
- Склад A: 5 шт
- Склад B: 4 шт
- **Сумма: 9 шт > 4** → ✅ БЕРЕМ

**Цены:**
- Склад A: 382.19₽
- Склад B: 400.00₽
- **MIN: 382.19₽**
- **382.19₽ * 1.12 = 428.05₽** → результат

### Пример 2: Товар пропускается

**Артикул:** 000219 (Michelin)

**Остатки:**
- Склад A: 1 шт
- **Сумма: 1 шт ≤ 4** → ❌ SKIP

Цена не рассчитывается, товар не попадает в экспорт.

### Пример 3: Товар с несколькими складами

**Артикул:** F7640 (Yokohama)

**Остатки:**
- Склад 1: 20 шт
- Склад 2: 15 шт
- Склад 3: 10 шт
- Склад 4: 10 шт
- **Сумма: 55 шт > 4** → ✅ БЕРЕМ

**Цены:**
- Склад 1: 9546 коп (95.46₽)
- Склад 2: 9800 коп (98.00₽)
- Склад 3: 9600 коп (96.00₽)
- Склад 4: 9700 коп (97.00₽)
- **MIN: 9546 коп (95.46₽)**
- **95.46₽ * 1.12 = 106.92₽** → результат

## SQL запрос

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
    CAST(MIN(p.price) * 1.12 AS INTEGER) as price_with_markup,
    SUM(p.stock) as total_stock
FROM nomenclature_tyres t
INNER JOIN tyres_prices_stock p ON t.cae = p.cae
GROUP BY t.id, t.cae, t.brand, t.model, t.width, t.height, t.diameter,
         t.season, t.is_studded, t.load_index, t.speed_index, t.runflat
HAVING SUM(p.stock) > 4
ORDER BY t.id
```

## Сравнение старой и новой логики

| Параметр | Старая логика | Новая логика |
|----------|---------------|--------------|
| **Остаток** | SUM(stock) | SUM(stock) > 4 |
| **Фильтр остатков** | Нет | Есть (> 4) |
| **Цена** | MIN(price) | MIN(price) * 1.12 |
| **Наценка** | 0% | +12% |
| **Результат** | Все товары | Только > 4 шт |

## Статистика экспорта

При экспорте выводится:
- ✅ Экспортировано: количество записей
- ⚠️ Пропущено: количество записей с остатком ≤ 4
- 📁 Путь к файлу

**Пример вывода:**
```
✅ Экспортировано: 1234 позиций
⚠️ Пропущено (остаток ≤ 4): 567 позиций
📁 Файл: /Users/viktor/Desktop/export.csv
```

## Тестирование

Создать тестовый файл с 10 записями:

```bash
go run cmd/export-bitrix-updated/main.go \
  -type=tyres \
  -output=/Users/viktor/Desktop/bitrix_test_10.csv \
  -limit=10
```

## Технические детали

**Файл:** `cmd/export-bitrix-updated/main.go`

**Ключевые функции:**
- `buildQuery()` - построение SQL с фильтрацией
- `formatProductName()` - форматирование названия
- `normalizeDiameter()` - нормализация диаметров

**Зависимости:**
- `nomenclature_tyres` / `nomenclature_rims` - номенклатура
- `tyres_prices_stock` / `rims_prices_stock` - цены и остатки

## Обновление данных

Перед экспортом убедитесь что данные актуальны:

```bash
# Синхронизация номенклатуры
go run cmd/sync-nomenclature/main.go -type=all

# Синхронизация цен и остатков
go run cmd/sync-prices/main.go -type=all
```

## Импорт в 1C-Битрикс

1. Сгенерировать CSV файл
2. Войти в админ-панель 1C-Битрикс
3. Каталог → Импорт → Загрузить CSV
4. Выбрать файл
5. Настроить маппинг колонок (если нужно)
6. Запустить импорт

## Troubleshooting

### Проблема: Мало записей в экспорте

**Причина:** Фильтр остатков > 4 убирает товары

**Решение:** Проверить остатки в базе:
```sql
SELECT SUM(stock), COUNT(*)
FROM tyres_prices_stock
GROUP BY cae
HAVING SUM(stock) <= 4
```

### Проблема: Цены слишком высокие

**Причина:** Наценка 12%

**Решение:** Это ожидаемое поведение. Если нужно изменить наценку, измените множитель в SQL:
```sql
CAST(MIN(p.price) * 1.15 AS INTEGER)  -- 15%
CAST(MIN(p.price) * 1.10 AS INTEGER)  -- 10%
```

### Проблема: Неправильные остатки

**Причина:** Данные не синхронизированы

**Решение:**
```bash
go run cmd/sync-prices/main.go -type=all
```
