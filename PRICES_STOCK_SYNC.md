## Синхронизация цен и остатков 4tochki

## Описание

Система загрузки актуальных цен и остатков по всем товарам через SOAP API 4tochki.

## Источник данных

- **API методы:** GetFindTyre, GetFindDisk
- **Данные:** Цены и остатки по складам в реальном времени
- **Обновление:** По требованию (рекомендуется ежедневно)

## Таблицы в БД

### tyres_prices_stock
Цены и остатки шин по складам (18,599 записей):
- `cae` - код товара (FK → nomenclature_tyres)
- `warehouse_id` - ID склада
- `price` - цена в копейках
- `stock` - остаток в штуках
- `updated_at` - дата обновления

### rims_prices_stock
Цены и остатки дисков по складам (18,793 записей):
- `cae` - код товара (FK → nomenclature_rims)
- `warehouse_id` - ID склада
- `price` - цена в копейках
- `stock` - остаток в штуках
- `updated_at` - дата обновления

### prices_stock_sync_runs
История синхронизаций:
- Время старта и завершения
- Количество товаров и складских позиций
- Статус и ошибки

## Запуск синхронизации

### Вручную

```bash
# Синхронизация всего (шины + диски)
go run cmd/sync-prices/main.go -type=all

# Только шины
go run cmd/sync-prices/main.go -type=tyres

# Только диски
go run cmd/sync-prices/main.go -type=rims
```

### Через скрипт

```bash
chmod +x sync_prices.sh
./sync_prices.sh
```

### Автоматически (cron)

Добавить в crontab для ежедневного запуска в 07:00:

```bash
crontab -e
```

```cron
# Ежедневная синхронизация цен и остатков 4tochki в 07:00
0 7 * * * cd /path/to/Etalon_Price_API && ./sync_prices.sh >> logs/prices_sync.log 2>&1
```

## Производительность

- **Загрузка через API:** ~2 сек на страницу (2000 товаров)
- **Сохранение в БД:** ~1 сек на 5000 записей
- **Общее время:** ~45 секунд для полной синхронизации

Оптимизация:
- Пагинация по 2000 товаров
- Bulk upsert через временные таблицы
- Один запрос для проверки существующих записей
- ON CONFLICT DO UPDATE для upsert

## Структура данных

### Пример записи шины
```
CAE: 3151013
Warehouse ID: 871
Price: 4731 (47.31 руб)
Stock: 1 шт
Updated: 2026-03-13 21:44:00
```

### Связь с номенклатурой
```sql
-- Полная информация о шине с ценами
SELECT
    n.cae,
    n.name,
    n.brand,
    n.model,
    p.warehouse_id,
    p.price / 100.0 AS price_rub,
    p.stock,
    p.updated_at
FROM nomenclature_tyres n
JOIN tyres_prices_stock p ON n.cae = p.cae
WHERE n.cae = '3151013';
```

## Примеры SQL запросов

### Товары в наличии
```sql
-- Шины в наличии (stock > 0)
SELECT
    n.cae,
    n.name,
    n.brand,
    COUNT(p.warehouse_id) AS warehouses_count,
    SUM(p.stock) AS total_stock,
    MIN(p.price) / 100.0 AS min_price_rub,
    MAX(p.price) / 100.0 AS max_price_rub
FROM nomenclature_tyres n
JOIN tyres_prices_stock p ON n.cae = p.cae
WHERE p.stock > 0
GROUP BY n.cae, n.name, n.brand
ORDER BY total_stock DESC
LIMIT 20;
```

### Цены по конкретному товару
```sql
-- Все склады с ценами для товара
SELECT
    cae,
    warehouse_id,
    price / 100.0 AS price_rub,
    stock,
    updated_at
FROM tyres_prices_stock
WHERE cae = '3151013'
ORDER BY price;
```

### Минимальная цена среди складов
```sql
-- Товары с минимальной ценой
SELECT
    n.cae,
    n.name,
    n.brand,
    n.model,
    MIN(p.price) / 100.0 AS min_price_rub,
    SUM(p.stock) AS total_stock
FROM nomenclature_tyres n
JOIN tyres_prices_stock p ON n.cae = p.cae
WHERE p.stock > 0
GROUP BY n.cae, n.name, n.brand, n.model
HAVING SUM(p.stock) > 0
ORDER BY min_price_rub
LIMIT 100;
```

### Статистика по складам
```sql
-- Количество позиций на каждом складе
SELECT
    warehouse_id,
    COUNT(*) AS products_count,
    SUM(stock) AS total_stock,
    AVG(price) / 100.0 AS avg_price_rub
FROM tyres_prices_stock
WHERE stock > 0
GROUP BY warehouse_id
ORDER BY products_count DESC;
```

### История синхронизаций
```sql
SELECT
    sync_type,
    started_at,
    completed_at,
    total_items,
    total_warehouses,
    new_items,
    updated_items,
    status,
    EXTRACT(EPOCH FROM (completed_at - started_at)) AS duration_seconds
FROM prices_stock_sync_runs
ORDER BY started_at DESC
LIMIT 10;
```

## Проверка данных

```bash
# Создать утилиту для проверки
cat > cmd/check-prices/main.go << 'EOF'
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/prokoleso/etalon-price-api/internal/config"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	pool, err := pgxpool.New(context.Background(), cfg.Database.DSN)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	ctx := context.Background()

	fmt.Println("Проверка цен и остатков:")
	fmt.Println("================================================================")

	// Tyres stats
	var tyresProducts, tyresWarehouses int
	pool.QueryRow(ctx, "SELECT COUNT(DISTINCT cae) FROM tyres_prices_stock").Scan(&tyresProducts)
	pool.QueryRow(ctx, "SELECT COUNT(*) FROM tyres_prices_stock").Scan(&tyresWarehouses)
	fmt.Printf("Шины:\n")
	fmt.Printf("  - Уникальных товаров: %d\n", tyresProducts)
	fmt.Printf("  - Складских позиций: %d\n", tyresWarehouses)

	// Rims stats
	var rimsProducts, rimsWarehouses int
	pool.QueryRow(ctx, "SELECT COUNT(DISTINCT cae) FROM rims_prices_stock").Scan(&rimsProducts)
	pool.QueryRow(ctx, "SELECT COUNT(*) FROM rims_prices_stock").Scan(&rimsWarehouses)
	fmt.Printf("Диски:\n")
	fmt.Printf("  - Уникальных товаров: %d\n", rimsProducts)
	fmt.Printf("  - Складских позиций: %d\n", rimsWarehouses)

	fmt.Println("================================================================")
}
EOF

# Запустить проверку
go run cmd/check-prices/main.go
```

## Мониторинг

Рекомендуется настроить:
1. Проверку успешности ежедневной синхронизации
2. Оповещения при ошибках API
3. Контроль времени выполнения (норма: < 1 мин)
4. Мониторинг количества товаров в наличии

## Интеграция с номенклатурой

Цены и остатки связаны с номенклатурой через CAE код:

```sql
-- Полная информация: номенклатура + цены + остатки
SELECT
    n.cae,
    n.name,
    n.brand,
    n.model,
    n.width,
    n.height,
    n.diameter,
    n.season,
    p.warehouse_id,
    p.price / 100.0 AS price_rub,
    p.stock
FROM nomenclature_tyres n
LEFT JOIN tyres_prices_stock p ON n.cae = p.cae
WHERE n.season = 'Зимняя'
  AND p.stock > 0
ORDER BY p.price
LIMIT 100;
```

## Рекомендации

1. **Частота обновления:** Ежедневно в 07:00 (после обновления номенклатуры)
2. **Порядок синхронизации:**
   - 06:00 - Синхронизация номенклатуры (XML)
   - 07:00 - Синхронизация цен и остатков (API)
3. **Мониторинг:** Отслеживать изменения в количестве товаров в наличии
4. **Backup:** Регулярное резервное копирование таблиц

## Troubleshooting

### Проблема: Таблицы не созданы
```bash
go run cmd/run-migration/main.go migrations/007_create_prices_stock_tables.up.sql
```

### Проблема: Медленная синхронизация
- Проверить индексы на таблицах
- Убедиться что используется bulk upsert
- Проверить производительность API

### Проблема: Ошибка API
- Проверить доступность API
- Проверить credentials в .env
- Увеличить timeout в config

### Проблема: Нет связи с номенклатурой
- Убедиться что CAE коды совпадают
- Проверить что номенклатура синхронизирована
- Проверить внешние ключи
