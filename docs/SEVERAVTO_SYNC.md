# Синхронизация цен и остатков Северавто

## Обзор

Система автоматической синхронизации цен и остатков шин от поставщика Северавто в единую таблицу `tyres_prices_stock`.

**Интервал синхронизации:** каждые 3 часа

## Архитектура

```
┌──────────────────────────────────┐
│ tyres_prices_stock_severavto     │  ← Загружается из API Северавто
│ (55,528 записей)                 │
│ - commodity_id                   │
│ - territory_name (склад)         │
│ - price (рубли)                  │
│ - stock (штуки)                  │
└────────────┬─────────────────────┘
             │ commodity_id
             ▼
┌──────────────────────────────────┐
│ severavto_tyres_mapping          │  ← Маппинг CAE ↔ commodity_id
│ (22,319 связей)                  │
│ - commodity_id → cae             │
│ - match_priority (1-4)           │
└────────────┬─────────────────────┘
             │ cae
             ▼
┌──────────────────────────────────┐
│ tyres_prices_stock               │  ← Единая таблица всех цен
│ + cae                            │
│ + warehouse_name = territory     │
│ + price                          │
│ + stock                          │
│ + provider = 'Северавто'         │
│ + isimport = 0                   │
└──────────────────────────────────┘
```

## Компоненты

### 1. Маппинг CAE ↔ commodity_id

**Таблица:** `severavto_tyres_mapping`

Создается с помощью интеллектуальной логики сопоставления:

**Приоритеты совпадения:**
1. **Приоритет 1** (24.5%): Точное совпадение (бренд + модель + размеры + индексы)
2. **Приоритет 2** (2.4%): Бренд + модель + ширина + индексы
3. **Приоритет 3** (38.8%): Бренд + ширина + индексы
4. **Приоритет 4** (34.3%): Бренд + размеры

**Скрипты:**
- `cmd/build-severavto-mapping/` - построение маппинга
- `cmd/view-mapping/` - просмотр статистики

### 2. Синхронизация цен/остатков

**Источник:** `tyres_prices_stock_severavto` (загружается из API Северавто)

**Данные:**
- `commodity_id` - ID товара Северавто
- `territory_name` - название склада (22 склада)
- `price` - цена в рублях
- `stock` - остаток в штуках
- `price_delayed` - цена с задержкой
- `price_rrp` - РРЦ

**Процесс синхронизации:**
1. Загрузка всех записей из `tyres_prices_stock_severavto`
2. JOIN с `severavto_tyres_mapping` для получения CAE
3. Фильтрация: только stock > 0 и price > 0
4. Bulk upsert в `tyres_prices_stock` (батчи по 5000)
5. ON CONFLICT: обновление price, stock, isimport=0

**Статистика:**
- Загружается: ~55,528 записей
- Уникальных товаров: ~22,317 CAE
- Складов: 22
- Время выполнения: ~20 секунд

### 3. Автоматический запуск

**Метод:** Cron (встроен в основной Dockerfile)

**Интервал:** каждые 3 часа (0 */3 * * *)

**Команда:**
```bash
/app/sync-severavto-prices >> /var/log/sync-severavto-prices.log 2>&1
```

**Логика:**
- Запускается через cron каждые 3 часа
- Логи в `/var/log/sync-severavto-prices.log`
- Автоматический перезапуск при падении контейнера

## Использование

### Ручной запуск синхронизации

```bash
go run cmd/sync-severavto-prices/main.go
```

### Деплой на Timeweb Cloud

Синхронизация встроена в основной Dockerfile и запускается автоматически через cron.

1. Собрать образ:
```bash
docker build -t etalon-price-api .
```

2. Переменные окружения:
```
DATABASE_DSN=postgres://user:pass@host/db?sslmode=require
```

3. Запуск контейнера:
```bash
docker run -e DATABASE_DSN="..." etalon-price-api
```

Cron автоматически запустит `sync-severavto-prices` каждые 3 часа.

## Маппинг данных

| Источник (Severavto) | Целевая таблица | Пример |
|---------------------|-----------------|--------|
| commodity_id → cae (через маппинг) | cae | "278596" |
| territory_name | warehouse_name | "Шушары ЦС" |
| price | price | 12190 ₽ |
| stock | stock | 4 шт |
| "Северавто" | provider | "Северавто" |
| - | isimport | 0 (загружено) |

## Проверка работы

### Проверить количество записей

```sql
SELECT COUNT(*)
FROM tyres_prices_stock
WHERE provider = 'Северавто';
-- Ожидается: ~55,528

SELECT COUNT(DISTINCT cae)
FROM tyres_prices_stock
WHERE provider = 'Северавто';
-- Ожидается: ~22,317
```

### Проверить склады

```sql
SELECT warehouse_name, COUNT(*) as cnt
FROM tyres_prices_stock
WHERE provider = 'Северавто'
GROUP BY warehouse_name
ORDER BY cnt DESC;
-- Ожидается: 22 склада
```

### Проверить свежесть данных

```sql
SELECT
    MAX(updated_at) as last_sync,
    COUNT(*) as records
FROM tyres_prices_stock
WHERE provider = 'Северавто';
```

## Логи и мониторинг

Планировщик выводит логи:
```
🔄 Запуск синхронизации: 2026-03-21 13:42:56
📥 Загрузка данных...
   Загружено: 55528 записей за 3.2 сек
💾 Сохранение...
   Обработано: 55528/55528
📊 Результат:
   ✅ Новые записи: 55528
   🔄 Обновлено: 0
✅ Синхронизация завершена!
⏰ Следующий запуск: 2026-03-21 16:42:56
```

## Особенности

1. **Bulk upsert** для высокой производительности (батчи по 5000)
2. **ON CONFLICT** автоматически обновляет существующие записи
3. **Транзакции** для целостности данных
4. **Фильтрация:** только товары с остатком > 0 и ценой > 0
5. **Маппинг через приоритеты:** максимальное покрытие товаров
6. **Автоматический перезапуск** каждые 3 часа

## Troubleshooting

### Нет новых записей (все обновлены)

Это нормально при повторных запусках. Система обновляет существующие записи.

### Меньше записей чем ожидалось

Проверьте фильтры:
- Только stock > 0
- Только price > 0
- Только товары с маппингом (22,317 CAE)

### Ошибки подключения к БД

Проверьте `DATABASE_DSN` и SSL режим (`?sslmode=require` для Timeweb).

## См. также

- [NOMENCLATURE_SIMPLE_LOGIC.md](NOMENCLATURE_SIMPLE_LOGIC.md) - логика маппинга
- [TIMEWEB_DEPLOYMENT.md](TIMEWEB_DEPLOYMENT.md) - деплой на Timeweb Cloud
- [EMAIL_NOTIFICATIONS.md](EMAIL_NOTIFICATIONS.md) - уведомления о синхронизации
