# Ведение истории изменений номенклатуры

## Обзор

Номенклатура шин и дисков обновляется **раз в сутки** из XML файлов Форточек.

**Важно:** Вместо обновления существующих записей, система **добавляет новые строки** только если данные изменились. Это позволяет отслеживать историю изменений.

## Логика работы

### До изменений (старая логика)
```sql
-- Было: UPSERT с обновлением
INSERT INTO nomenclature_tyres (cae, name, ...)
VALUES (...)
ON CONFLICT (cae) DO UPDATE SET
    name = EXCLUDED.name,
    ... -- обновляли все поля
```

**Проблема:** Терялась история изменений. Если цена или характеристики товара менялись, старые данные перезаписывались.

### После изменений (новая логика)
```sql
-- Стало: INSERT только если данные отличаются
INSERT INTO nomenclature_tyres (cae, name, ...)
SELECT cae, name, ...
FROM temp_tyres t
WHERE NOT EXISTS (
    SELECT 1 FROM nomenclature_tyres nt
    WHERE nt.cae = t.cae
      AND nt.name = t.name
      AND ... -- проверяем ВСЕ поля
)
```

**Результат:**
- ✅ **Полный дубль** (все поля совпадают) → запись НЕ добавляется (skipped)
- ✅ **Есть изменения** → добавляется НОВАЯ строка с новыми данными
- ✅ **История сохраняется** → можно видеть когда и что изменилось

## Структура таблиц

### До миграции 013
```sql
CREATE TABLE nomenclature_tyres (
    id BIGSERIAL PRIMARY KEY,
    cae VARCHAR(50) UNIQUE NOT NULL,  -- UNIQUE constraint
    ...
);
```

### После миграции 013
```sql
CREATE TABLE nomenclature_tyres (
    id BIGSERIAL PRIMARY KEY,
    cae VARCHAR(50) NOT NULL,  -- НЕ UNIQUE!
    ...
    created_at TIMESTAMP,  -- когда добавлена версия
    updated_at TIMESTAMP   -- не используется
);

-- Индекс для проверки дублей по всем полям
CREATE INDEX idx_nomenclature_tyres_duplicate_check ON nomenclature_tyres(
    cae, name, width, height, diameter, ...
);

-- Индекс для быстрого поиска по cae
CREATE INDEX idx_nomenclature_tyres_cae_lookup ON nomenclature_tyres(cae);
```

## Примеры

### Пример 1: Первая загрузка
```sql
-- XML содержит:
-- CAE: TIRE001, Name: "Michelin Pilot Sport 225/45R17", Width: 225, ...

-- Результат:
INSERT INTO nomenclature_tyres ...
-- Добавлена 1 новая запись
```

### Пример 2: Повторная загрузка (данные не изменились)
```sql
-- XML содержит те же данные
-- CAE: TIRE001, Name: "Michelin Pilot Sport 225/45R17", Width: 225, ...

-- Результат:
-- Запись НЕ добавлена (полный дубль)
-- skipped_duplicates: 1
```

### Пример 3: Данные изменились
```sql
-- XML содержит измененные данные:
-- CAE: TIRE001, Name: "Michelin Pilot Sport 225/45R17 NEW", Width: 225, ...
--                                                    ^^^^

-- Результат:
INSERT INTO nomenclature_tyres ...
-- Добавлена новая запись с обновленным названием
-- В таблице теперь 2 записи с CAE = TIRE001:
-- 1. id=1, name="...225/45R17",     created_at=2026-03-14
-- 2. id=2, name="...225/45R17 NEW", created_at=2026-03-15
```

## Запросы

### Получить актуальную версию товара
```sql
-- Последняя добавленная запись = актуальная версия
SELECT *
FROM nomenclature_tyres
WHERE cae = 'TIRE001'
ORDER BY created_at DESC
LIMIT 1;
```

### Получить историю изменений товара
```sql
SELECT id, cae, name, width, height, created_at
FROM nomenclature_tyres
WHERE cae = 'TIRE001'
ORDER BY created_at DESC;
```

### Найти товары с изменениями за последний месяц
```sql
SELECT cae, COUNT(*) as versions, MIN(created_at) as first_seen, MAX(created_at) as last_updated
FROM nomenclature_tyres
WHERE created_at > NOW() - INTERVAL '1 month'
GROUP BY cae
HAVING COUNT(*) > 1
ORDER BY last_updated DESC;
```

### Статистика дублей при синхронизации
```sql
SELECT
    sync_type,
    started_at,
    total_items,
    new_items,
    updated_items as skipped_duplicates,  -- в поле updated_items хранятся skipped
    ROUND(100.0 * updated_items / total_items, 2) as duplicate_percent
FROM nomenclature_sync_runs
WHERE status = 'completed'
ORDER BY started_at DESC
LIMIT 10;
```

## Ежедневная синхронизация

### Вариант 1: Cron планировщик (рекомендуется)

Запустите планировщик как daemon:

```bash
# Запуск планировщика (синхронизация каждый день в 2:00 AM)
go run cmd/nomenclature-scheduler/main.go

# Или через systemd (см. ниже)
```

**Расписание:** Каждый день в 2:00 AM (по серверному времени)

### Вариант 2: Системный cron

Добавьте в crontab:

```bash
# Редактировать crontab
crontab -e

# Добавить строку (каждый день в 2:00 AM)
0 2 * * * cd /path/to/Etalon_Price_API && /usr/local/go/bin/go run cmd/sync-nomenclature/main.go -type=all >> /var/log/nomenclature-sync.log 2>&1
```

### Вариант 3: Systemd service + timer

**1. Создайте service файл:**

```ini
# /etc/systemd/system/nomenclature-sync.service
[Unit]
Description=Nomenclature Sync Service
After=network.target

[Service]
Type=oneshot
User=your-user
WorkingDirectory=/path/to/Etalon_Price_API
ExecStart=/usr/local/go/bin/go run cmd/sync-nomenclature/main.go -type=all
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

**2. Создайте timer файл:**

```ini
# /etc/systemd/system/nomenclature-sync.timer
[Unit]
Description=Daily Nomenclature Sync Timer
Requires=nomenclature-sync.service

[Timer]
OnCalendar=daily
OnCalendar=*-*-* 02:00:00
Persistent=true

[Install]
WantedBy=timers.target
```

**3. Активируйте:**

```bash
sudo systemctl daemon-reload
sudo systemctl enable nomenclature-sync.timer
sudo systemctl start nomenclature-sync.timer

# Проверить статус
sudo systemctl status nomenclature-sync.timer
sudo systemctl list-timers
```

### Вариант 4: Docker + cron

```dockerfile
# Dockerfile
FROM golang:1.21-alpine

WORKDIR /app
COPY . .
RUN go build -o scheduler cmd/nomenclature-scheduler/main.go

# Установить supercronic для cron в Docker
RUN apk add --no-cache ca-certificates
RUN wget -O /usr/local/bin/supercronic https://github.com/aptible/supercronic/releases/download/v0.2.1/supercronic-linux-amd64
RUN chmod +x /usr/local/bin/supercronic

# Создать crontab файл
RUN echo "0 2 * * * /app/scheduler" > /etc/crontabs/root

CMD ["supercronic", "/etc/crontabs/root"]
```

## Мониторинг

### Проверить последний запуск
```sql
SELECT *
FROM nomenclature_sync_runs
ORDER BY started_at DESC
LIMIT 1;
```

### Проверить количество версий товара
```sql
-- Товары с наибольшим количеством версий
SELECT cae, COUNT(*) as versions
FROM nomenclature_tyres
GROUP BY cae
ORDER BY versions DESC
LIMIT 10;
```

### Размер таблиц
```sql
SELECT
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
FROM pg_tables
WHERE tablename IN ('nomenclature_tyres', 'nomenclature_rims')
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
```

## Очистка старых версий

Если таблица растёт слишком быстро, можно удалять старые версии:

```sql
-- Удалить версии старше 6 месяцев (оставить только последнюю)
DELETE FROM nomenclature_tyres
WHERE id NOT IN (
    SELECT DISTINCT ON (cae) id
    FROM nomenclature_tyres
    ORDER BY cae, created_at DESC
)
AND created_at < NOW() - INTERVAL '6 months';

-- Проверить сколько будет удалено (без удаления)
SELECT COUNT(*)
FROM nomenclature_tyres
WHERE id NOT IN (
    SELECT DISTINCT ON (cae) id
    FROM nomenclature_tyres
    ORDER BY cae, created_at DESC
)
AND created_at < NOW() - INTERVAL '6 months';
```

## Миграция

### Применить изменения

```bash
# Применить миграцию
psql "$DATABASE_DSN" -f migrations/013_modify_nomenclature_for_history.up.sql

# Проверить
psql "$DATABASE_DSN" -c "\d nomenclature_tyres" | grep -i index
```

### Откатить (если нужно)

```bash
# ВНИМАНИЕ: Откат удалит индексы и попытается создать UNIQUE constraint
# Может не сработать если уже есть дубликаты по CAE
psql "$DATABASE_DSN" -f migrations/013_modify_nomenclature_for_history.down.sql
```

## Troubleshooting

### Все записи skipped (0 new items)

Это **нормально** если данные не изменились. Система работает корректно - не добавляет дубли.

### Слишком много новых записей

Проверьте, не изменился ли формат XML или логика парсинга:

```sql
-- Посмотреть недавно добавленные записи
SELECT cae, name, created_at
FROM nomenclature_tyres
WHERE created_at > NOW() - INTERVAL '1 day'
ORDER BY created_at DESC
LIMIT 100;
```

### Ошибка "duplicate key value violates unique constraint"

Если вы видите эту ошибку после миграции, значит миграция 013 не была применена:

```bash
# Проверить наличие UNIQUE constraint
psql "$DATABASE_DSN" -c "\d nomenclature_tyres" | grep UNIQUE

# Применить миграцию если её нет
psql "$DATABASE_DSN" -f migrations/013_modify_nomenclature_for_history.up.sql
```
