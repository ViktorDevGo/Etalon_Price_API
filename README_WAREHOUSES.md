## Склады 4tochki

## Описание

Справочник складов (распределительных центров) 4tochki с полной информацией о доставке, самовывозе и логистике.

## Таблица warehouses

Содержит информацию о всех складах (54 склада):

### Поля
- `id` - ID склада (PRIMARY KEY)
- `key` - Уникальный ключ склада
- `name` - Полное название
- `short_name` - Короткое название для отображения
- `sto_id` - ID в системе STO
- `background_color` - Цвет для UI (#9ACD32, #9696F8)
- `have_delivery` - Есть доставка (boolean)
- `have_pickup` - Есть самовывоз (boolean)
- `is_paid_delivery` - Платная доставка (boolean)
- `logistic_days` - Дни логистики (int)

### Связи
- `tyres_prices_stock.warehouse_id` → `warehouses.id` (FK)
- `rims_prices_stock.warehouse_id` → `warehouses.id` (FK)

## Синхронизация

### API метод
`GetWarehouses` - возвращает список всех складов с актуальной информацией

### Запуск
```bash
# Синхронизация складов
go run cmd/sync-warehouses/main.go

# Проверка данных
go run cmd/check-warehouses/main.go
```

### Автоматизация
Склады редко меняются, рекомендуется синхронизировать:
- При первой настройке
- Раз в неделю/месяц
- При получении уведомления об изменениях от 4tochki

## Примеры складов

**Крупные региональные склады:**
- [1423] Новосибирск (НВСБ3) - доставка, самовывоз
- [232] Мос.обл. Ям (ЯМК) - доставка, самовывоз
- [1] Москва КрС (МСК) - доставка, самовывоз

**Пункты самовывоза (ОХ - Охранные Хранилища):**
- [364] ОХ г. Новосибирск Чехова - самовывоз
- [871] ОХ г.Новосибирск Петухова, 67 - самовывоз

## SQL запросы

### Список всех складов с доставкой
```sql
SELECT id, short_name, name
FROM warehouses
WHERE have_delivery = true
ORDER BY name;
```

### Склады с товарами в наличии
```sql
SELECT
    w.id,
    w.short_name,
    w.name,
    COUNT(DISTINCT p.cae) AS products_count,
    SUM(p.stock) AS total_stock
FROM warehouses w
JOIN tyres_prices_stock p ON w.id = p.warehouse_id
WHERE p.stock > 0
GROUP BY w.id, w.short_name, w.name
ORDER BY products_count DESC;
```

### Товары доступные на конкретном складе
```sql
SELECT
    n.cae,
    n.name,
    n.brand,
    p.price / 100.0 AS price_rub,
    p.stock
FROM tyres_prices_stock p
JOIN nomenclature_tyres n ON p.cae = n.cae
WHERE p.warehouse_id = 1423  -- Новосибирск
  AND p.stock > 0
ORDER BY p.price
LIMIT 100;
```

### Поиск товара с ценами по всем складам
```sql
SELECT
    n.name,
    w.short_name AS warehouse,
    p.price / 100.0 AS price_rub,
    p.stock,
    w.have_delivery,
    w.have_pickup,
    w.logistic_days
FROM nomenclature_tyres n
JOIN tyres_prices_stock p ON n.cae = p.cae
JOIN warehouses w ON p.warehouse_id = w.id
WHERE n.cae = '3151013'
  AND p.stock > 0
ORDER BY p.price;
```

## Интеграция с API

При работе с API методы `GetFindTyre` и `GetFindDisk` возвращают данные с `warehouse_id`, которые теперь связаны со справочником складов:

```go
// Получить информацию о складе
warehouse, err := warehouseRepo.GetWarehouseByID(ctx, warehouseID)

// Проверить доступность доставки
if warehouse.HaveDelivery {
    // Можно оформить доставку
}

// Рассчитать срок доставки
deliveryDays := warehouse.LogisticDays + processingDays
```

## Миграции

- **008_create_warehouses_table.up.sql** - создание таблицы warehouses
- **009_add_warehouse_foreign_keys.up.sql** - добавление FK constraints

## Статистика

После синхронизации:
- **Всего складов:** 54
- **С доставкой:** ~25
- **Только самовывоз:** ~29
- **Регионы:** Москва, СПб, Новосибирск, Екатеринбург, и др.

## Топ-10 складов по товарам (шины)

1. [1423] Новосибирск - 4,681 товар
2. [232] Мос.обл. Ям - 3,374 товара
3. [2043] Новокузнецк 3 - 2,985 товаров
4. [2184] Домодедово (Кучино) - 2,719 товаров
5. [1492] Томск - 2,006 товаров
6. [1222] Давыдово - 1,434 товара
7. [9] Склад 2 - 255 товаров
8. [3] Склад 3 - 193 товара
9. [1998] ОХ Бердское - 192 товара
10. [2167] ОХ Бердск - 153 товара

## Цвета складов

API возвращает цвета для визуального различия складов в UI:
- `#9ACD32` - основные склады с доставкой
- `#9696F8` - пункты самовывоза (ОХ)

## Примечания

- Таблица `warehouses` является справочником и редко изменяется
- Все цены и остатки связаны со складами через `warehouse_id`
- При удалении склада автоматически удаляются связанные цены (ON DELETE CASCADE)
- Рекомендуется синхронизировать склады перед первой загрузкой цен
