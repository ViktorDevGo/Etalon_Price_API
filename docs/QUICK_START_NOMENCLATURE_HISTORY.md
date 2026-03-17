# Быстрый старт: История изменений номенклатуры

## ⚡ Что изменилось?

**До:** При повторной загрузке XML обновлялись существующие записи (терялась история)
**Теперь:** Новая запись добавляется только если данные изменились (сохраняется история)

## 🚀 Применение изменений

### 1. Применить миграцию
```bash
psql "$DATABASE_DSN" -f migrations/013_modify_nomenclature_for_history.up.sql
```

### 2. Проверить структуру
```bash
# Проверить что UNIQUE constraint на cae удалён
psql "$DATABASE_DSN" -c "\d nomenclature_tyres" | grep -i unique

# Должно показать только:
# "tyres_prices_stock_cae_warehouse_unique" UNIQUE CONSTRAINT
# НЕ должно быть: "nomenclature_tyres_cae_key"
```

### 3. Тестовый запуск синхронизации
```bash
go run cmd/sync-nomenclature/main.go -type=tyres
```

**Ожидаемый результат:**
- При первом запуске: много новых записей (new: N, skipped_duplicates: 0)
- При повторном запуске: мало новых или 0 (new: 0, skipped_duplicates: N)

## 📊 Проверка результата

```sql
-- Посмотреть последнюю синхронизацию
SELECT
    sync_type,
    started_at,
    total_items,
    new_items,
    updated_items as skipped_duplicates,
    ROUND(100.0 * updated_items / total_items, 2) as duplicate_percent
FROM nomenclature_sync_runs
ORDER BY started_at DESC
LIMIT 1;

-- Пример ответа после повторной синхронизации:
-- sync_type | started_at | total_items | new_items | skipped_duplicates | duplicate_percent
-- tyres     | 2026-03-15 | 60637       | 15        | 60622              | 99.98%
--
-- Это значит: из 60637 товаров только 15 изменились, остальные 60622 - дубли (пропущены)
```

## ⏰ Настройка ежедневной синхронизации

### Вариант 1: Встроенный планировщик (рекомендуется)

```bash
# Запустить daemon
go run cmd/nomenclature-scheduler/main.go

# Или собрать и запустить бинарник
go build -o nomenclature-scheduler cmd/nomenclature-scheduler/main.go
./nomenclature-scheduler
```

**Расписание:** Каждый день в 2:00 AM

### Вариант 2: Системный cron

```bash
# Добавить в crontab
crontab -e

# Вставить строку:
0 2 * * * cd /path/to/Etalon_Price_API && go run cmd/sync-nomenclature/main.go -type=all >> /var/log/nomenclature.log 2>&1
```

## 🔍 Полезные запросы

### Получить актуальную версию товара
```sql
SELECT *
FROM nomenclature_tyres
WHERE cae = 'YOUR_CAE'
ORDER BY created_at DESC
LIMIT 1;
```

### История изменений товара
```sql
SELECT id, cae, name, brand, model, created_at
FROM nomenclature_tyres
WHERE cae = 'YOUR_CAE'
ORDER BY created_at DESC;
```

### Товары изменившиеся за последнюю неделю
```sql
SELECT cae, COUNT(*) as versions, MAX(created_at) as last_updated
FROM nomenclature_tyres
WHERE created_at > NOW() - INTERVAL '7 days'
GROUP BY cae
HAVING COUNT(*) > 1
ORDER BY last_updated DESC;
```

## 📚 Подробная документация

См. [NOMENCLATURE_HISTORY.md](NOMENCLATURE_HISTORY.md)

## ❓ FAQ

### Q: Почему в логах "skipped_duplicates: 60000"?
**A:** Это нормально! Если данные в XML не изменились, записи не добавляются. Система работает корректно.

### Q: Будет ли таблица расти бесконечно?
**A:** Да, но медленно. Рост происходит только при реальных изменениях товаров. Можно настроить очистку старых версий (см. документацию).

### Q: Как получить только актуальные товары без истории?
**A:** Используйте `DISTINCT ON`:
```sql
SELECT DISTINCT ON (cae) *
FROM nomenclature_tyres
ORDER BY cae, created_at DESC;
```

### Q: Ошибка "duplicate key value violates unique constraint"
**A:** Миграция 013 не применена. Выполните:
```bash
psql "$DATABASE_DSN" -f migrations/013_modify_nomenclature_for_history.up.sql
```
