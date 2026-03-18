# Автоматическая выгрузка цен на сервер 1C-Битрикс

## Описание

Автоматическая ежедневная выгрузка файлов переоценки на сервер 1C-Битрикс в **10:00 по иркутскому времени (UTC+8)**.

## Выгружаемые файлы (в порядке генерации)

### ВАЖНО: Порядок генерации файлов критичен!

1️⃣ **Переоценка_МРЦ_YYYYMMDD.csv** - переоценка с МРЦ ценами (генерируется ПЕРВЫМ!)
   - Источник: `tyres_prices_stock` + `nomenclature_tyres` + `mrc_etalon`
   - Фильтр: `tiretype = 'Легковая'`
   - Формат: 5 колонок (IE_XML_ID, IP_PROP171, QUANTITY, CV_PRICE_1, CV_CURRENCY_1)
   - **Особенность:** НЕ обновляет `isimport` (файл для информации)
   - **Цена:** CEIL(mrc) если есть в mrc_etalon.article, иначе CEIL(MIN(price) × 1.12)

2️⃣ **Переоценка_шины_YYYYMMDD.csv** - переоценка легковых шин
   - Источник: `tyres_prices_stock` + `nomenclature_tyres`
   - Фильтр: `tiretype = 'Легковая'`
   - Формат: 5 колонок (IE_XML_ID, IP_PROP171, QUANTITY, CV_PRICE_1, CV_CURRENCY_1)
   - **Особенность:** Обновляет `isimport = 1`

3️⃣ **Переоценка_мотошины_YYYYMMDD.csv** - переоценка мотошин
   - Источник: `tyres_prices_stock` + `nomenclature_tyres`
   - Фильтр: `tiretype = 'Мотошина'`
   - Формат: 5 колонок (IE_XML_ID, IP_PROP171, QUANTITY, CV_PRICE_1, CV_CURRENCY_1)
   - **Особенность:** Обновляет `isimport = 1`

4️⃣ **Переоценка_диски_YYYYMMDD.csv** - переоценка дисков
   - Источник: `rims_prices_stock` + `nomenclature_rims`
   - Фильтр: БЕЗ фильтра по типу (все диски)
   - Формат: 5 колонок (IE_XML_ID, IP_PROP171, QUANTITY, CV_PRICE_1, CV_CURRENCY_1)
   - **Особенность:** Обновляет `isimport = 1`

## Логика генерации

### Общие правила

- **Фильтры:**
  - `SUM(stock) > 4` - только товары с суммарным остатком больше 4
  - `MIN(isimport) = 0` - только новые данные (не экспортированные ранее)

- **Колонки:**
  1. `IE_XML_ID` - CAE товара
  2. `IP_PROP171` - срок доставки:
     - "Доставим курьером Сегодня и позже" - если есть Запаска/Форточки/Бигмашин
     - "Доставим курьером Завтра и позже" - только Группа Бринекс/Северавто
  3. `QUANTITY` - `SUM(stock)` по CAE
  4. `CV_PRICE_1` - `CEIL(MIN(price) × 1.12)` - минимальная цена + 12% с округлением вверх
  5. `CV_CURRENCY_1` - "RUB"

- **После экспорта:** автоматическое обновление `isimport = 1` для всех экспортированных товаров

## Развертывание на сервере

### Docker контейнер

**Dockerfile:** `Dockerfile.prices-upload`

**Расписание:** Каждый день в 10:00 по Иркутску (Asia/Irkutsk, UTC+8)

**Переменные окружения:**
```env
DB_DSN=postgresql://user:pass@host:5432/db?sslmode=require
TZ=Asia/Irkutsk
WORK_DIR=/app
```

**Запуск на Timeweb Cloud:**
```bash
# 1. Создать Cloud App "prices-upload"
# 2. Dockerfile: Dockerfile.prices-upload
# 3. Установить переменные окружения (см. TIMEWEB_ENV_VARIABLES.md)
# 4. Deploy
```

### Скрипт выгрузки

**Файл:** `scripts/upload_prices_to_server.sh`

**Сервер назначения:**
- **Host:** 147.45.215.76
- **User:** root
- **Path:** /home/bitrix/www/upload/1c_catalog

**Алгоритм работы:**

1. **Генерация файлов:**
   - Запуск `cmd/export-bitrix-prices/main.go` (легковые шины)
   - Запуск `cmd/export-bitrix-prices-moto/main.go` (мотошины)
   - Сохранение в `/tmp/bitrix_export/`

2. **Загрузка на сервер:**
   - SCP через `sshpass` для автоматической передачи пароля
   - Загрузка обоих файлов в `/home/bitrix/www/upload/1c_catalog/`

3. **Очистка:**
   - Удаление временных файлов

## Ручной запуск

### Тестовый режим (без загрузки на сервер)

```bash
# Генерация файлов и копирование на Desktop
./scripts/test_upload_prices.sh
```

**Что делает:**
- Генерирует оба файла переоценки
- Показывает первые 5 строк каждого файла
- Копирует файлы на Desktop для проверки
- **НЕ загружает** на сервер

### Полная выгрузка на сервер

```bash
# ВНИМАНИЕ: загрузит файлы на production сервер!
export DB_DSN="postgresql://..."
./scripts/upload_prices_to_server.sh
```

## Мониторинг

### Проверка работы контейнера

```bash
# Логи контейнера
docker logs prices-upload

# Логи cron
docker exec prices-upload tail -f /var/log/cron.log
```

### Проверка файлов на сервере

```bash
ssh root@147.45.215.76
ls -lh /home/bitrix/www/upload/1c_catalog/Переоценка_*
```

## Локальная разработка и тестирование

### Генерация файлов вручную

```bash
# МРЦ
go run cmd/export-bitrix-prices-mrc/main.go -output-dir=/Users/viktor/Desktop

# Легковые шины
go run cmd/export-bitrix-prices/main.go -output-dir=/Users/viktor/Desktop

# Мотошины
go run cmd/export-bitrix-prices-moto/main.go -output-dir=/Users/viktor/Desktop

# Диски
go run cmd/export-bitrix-prices-rims/main.go -output-dir=/Users/viktor/Desktop

# Тестовый режим (limit 10)
go run cmd/export-bitrix-prices-mrc/main.go -output-dir=/Users/viktor/Desktop -limit=10
go run cmd/export-bitrix-prices/main.go -output-dir=/Users/viktor/Desktop -limit=10
go run cmd/export-bitrix-prices-moto/main.go -output-dir=/Users/viktor/Desktop -limit=10
go run cmd/export-bitrix-prices-rims/main.go -output-dir=/Users/viktor/Desktop -limit=10
```

### Проверка расписания

Cron выражение: `0 10 * * *`
- Каждый день в 10:00 по локальному времени контейнера (Asia/Irkutsk)

## Статистика (по состоянию на 19.03.2026)

### МРЦ файл
- Всего товаров: ~5,783 (все легковые с stock > 4, isimport = 0)
- С МРЦ ценами: ~3,443 (совпадение CAE с mrc_etalon.article)
- С обычными ценами: ~2,340 (нет в mrc_etalon)
- **НЕ обновляет isimport**

### Легковые шины
- Всего в номенклатуре: 53,690
- С остатком > 0: ~6,500
- Экспортируется (stock > 4, isimport = 0): ~5,783
- **Обновляет isimport = 1**

### Мотошины
- Всего в номенклатуре: 279
- С остатком > 0: 80
- Экспортируется (stock > 4, isimport = 0): ~46
- **Обновляет isimport = 1**

### Диски
- Всего в номенклатуре: 99,995
- С остатком > 0: ~20,000
- Экспортируется (stock > 4, isimport = 0): ~15,000
- **Обновляет isimport = 1**

## Troubleshooting

### Файлы не генерируются

**Проблема:** Нет данных с `isimport = 0`

**Решение:**
```sql
-- Сбросить isimport для повторной выгрузки
UPDATE tyres_prices_stock SET isimport = 0 WHERE provider = 'Форточки';
```

### Ошибка подключения к БД

**Проблема:** `password authentication failed`

**Решение:** Проверить `DB_DSN` в переменных окружения контейнера

### Файлы не загружаются на сервер

**Проблема:** SSH ошибка или неправильные credentials

**Решение:**
1. Проверить доступность сервера: `ping 147.45.215.76`
2. Проверить SSH подключение: `ssh root@147.45.215.76`
3. Проверить пароль в скрипте `upload_prices_to_server.sh`

## Обновление скриптов

После изменения `scripts/upload_prices_to_server.sh`:

1. Закоммитить изменения в git
2. Push в GitHub
3. Пересобрать Docker образ на Timeweb Cloud:
   ```bash
   # В настройках Cloud App "prices-upload"
   # Нажать "Rebuild" или "Redeploy"
   ```

## См. также

- [BITRIX_PRICES_EXPORT.md](BITRIX_PRICES_EXPORT.md) - переоценка легковых шин
- [TIMEWEB_ENV_VARIABLES.md](../TIMEWEB_ENV_VARIABLES.md) - переменные окружения для Timeweb Cloud
- [TIMEWEB_DEPLOYMENT.md](TIMEWEB_DEPLOYMENT.md) - общая информация о деплое
