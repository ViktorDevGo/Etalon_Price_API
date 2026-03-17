# 🚀 START HERE - Быстрый старт

## Ваши credentials уже настроены!

✅ **4tochki API:** sa69263
✅ **Cloud Database:** TWC PostgreSQL
✅ **.env файл:** создан и настроен

---

## Вариант 1: Автоматический запуск (рекомендуется)

```bash
# Настроить окружение и применить миграции
./run_cloud.sh

# Запустить HTTP сервер
./start_server.sh
```

В другом терминале:
```bash
# Тест синхронизации
curl -X POST http://localhost:8080/sync/4tochki \
  -H "Content-Type: application/json" \
  -d '{"codes":["2329500","WHS063930"]}'
```

---

## Вариант 2: CLI синхронизация

```bash
# Запустить синхронизацию напрямую
./start_sync.sh

# Или с кастомными кодами
./start_sync.sh --codes=ВАШ_КОД_1,ВАШ_КОД_2
```

---

## Вариант 3: Ручной запуск

### 1. Установить SSL сертификат

```bash
./setup_ssl.sh
```

### 2. Применить миграции

```bash
export PGSSLROOTCERT=$HOME/.cloud-certs/root.crt
export DB_DSN="postgresql://gen_user:Poison-79@c37e696087932476c61fd621.twc1.net:5432/default_db?sslmode=verify-full"

# Если установлен golang-migrate
migrate -path internal/migrations -database "$DB_DSN" up
```

### 3. Собрать проект

```bash
go build -o bin/app ./cmd/app
go build -o bin/sync ./cmd/sync
```

### 4. Запустить HTTP сервер

```bash
export PGSSLROOTCERT=$HOME/.cloud-certs/root.crt
./bin/app
```

### 5. Или запустить CLI

```bash
export PGSSLROOTCERT=$HOME/.cloud-certs/root.crt
./bin/sync --supplier=4tochki --codes=2329500,WHS063930
```

---

## Проверка работы

### Health Check

```bash
curl http://localhost:8080/healthz
```

Ожидаемый ответ:
```json
{"status":"healthy"}
```

### Синхронизация

```bash
curl -X POST http://localhost:8080/sync/4tochki \
  -H "Content-Type: application/json" \
  -d '{
    "codes": ["2329500", "WHS063930"]
  }'
```

### Проверка БД

```bash
export PGSSLROOTCERT=$HOME/.cloud-certs/root.crt
psql 'postgresql://gen_user:Poison-79@c37e696087932476c61fd621.twc1.net:5432/default_db?sslmode=verify-full'
```

SQL запросы:
```sql
-- Проверить таблицы
\dt

-- Шины
SELECT code, brand, model, width, height, diameter
FROM tyres_api
LIMIT 10;

-- Диски
SELECT code, brand, model, width, diameter
FROM rims_api
LIMIT 10;

-- Цены
SELECT code, price, stock, warehouse_name
FROM product_offers_api
LIMIT 10;

-- История синхронизаций
SELECT * FROM sync_runs_api
ORDER BY created_at DESC
LIMIT 5;
```

---

## Полезные команды

```bash
# Запустить сервер
./start_server.sh

# Запустить синхронизацию
./start_sync.sh

# С кастомными кодами
./start_sync.sh --codes=CODE1,CODE2

# Из файла
./start_sync.sh --codes-file=test_codes.txt

# Проверить health
curl http://localhost:8080/healthz

# Проверить БД
export PGSSLROOTCERT=$HOME/.cloud-certs/root.crt
psql 'postgresql://gen_user:Poison-79@c37e696087932476c61fd621.twc1.net:5432/default_db?sslmode=verify-full'
```

---

## Если что-то не работает

### 1. SSL ошибка

```bash
# Установить сертификат
./setup_ssl.sh

# Проверить
ls -la ~/.cloud-certs/root.crt
```

### 2. Database connection error

```bash
# Проверить доступность БД
export PGSSLROOTCERT=$HOME/.cloud-certs/root.crt
psql 'postgresql://gen_user:Poison-79@c37e696087932476c61fd621.twc1.net:5432/default_db?sslmode=verify-full' -c "SELECT version();"
```

### 3. SOAP Fault

- Проверьте credentials в .env (уже настроены)
- Убедитесь что коды товаров существуют
- Включите debug: `LOG_LEVEL=debug` в .env

### 4. Пустая БД после синхронизации

```sql
-- Проверить последнюю синхронизацию
SELECT * FROM sync_runs_api
ORDER BY created_at DESC
LIMIT 1;

-- Если есть ошибки
SELECT error_message FROM sync_runs_api
WHERE status = 'failed'
ORDER BY created_at DESC
LIMIT 1;
```

---

## Что дальше?

После успешного запуска:

1. **Получите реальные коды товаров** от поставщика 4tochki
2. **Добавьте их в test_codes.txt**
3. **Запустите полную синхронизацию:**
   ```bash
   ./start_sync.sh --codes-file=test_codes.txt
   ```

4. **Интегрируйте с вашим приложением:**
   - Используйте HTTP API для синхронизации
   - Читайте данные напрямую из БД
   - Настройте cron для периодической синхронизации

---

## Production Deployment

Для production:

1. Смените `APP_ENV=production` в .env
2. Уменьшите `LOG_LEVEL=info`
3. Настройте backup БД
4. Добавьте мониторинг
5. Настройте автоматическую синхронизацию (cron)

---

**Все готово к работе!** 🎉

Просто запустите:
```bash
./run_cloud.sh  # Настройка
./start_server.sh  # HTTP сервер
```

Или:
```bash
./start_sync.sh  # Прямая синхронизация
```
