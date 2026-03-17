# ✅ Deployment Checklist

## Перед запуском

### 1. Получение доступа к API 4tochki

- [ ] Связались с поставщиком (https://www.4tochki.ru/)
- [ ] Получили логин и пароль для B2B API
- [ ] Получили тестовые коды товаров
- [ ] Проверили доступность WSDL: http://api-b2b.4tochki.ru/WCF/ClientService.svc?wsdl

### 2. Подготовка окружения

- [ ] Docker Desktop установлен и запущен
- [ ] Go 1.24+ установлен (для локальной разработки)
- [ ] PostgreSQL client установлен (опционально, для ручной проверки БД)

### 3. Конфигурация

- [ ] Скопирован `.env.example` в `.env`
- [ ] В `.env` указаны реальные credentials:
  ```env
  FOURTOCHKI_LOGIN=ваш_логин
  FOURTOCHKI_PASSWORD=ваш_пароль
  ```
- [ ] Проверены остальные настройки в `.env`

---

## Запуск (выберите один вариант)

### Вариант A: Автоматический запуск (рекомендуется)

```bash
# Запустить setup скрипт
./setup.sh
```

Скрипт выполнит:
- ✅ Создание .env (если не существует)
- ✅ Запуск PostgreSQL
- ✅ Применение миграций
- ✅ Сборку приложения
- ✅ Запуск HTTP сервера
- ✅ Проверку health endpoint

### Вариант B: Через Makefile

```bash
# Запустить Docker stack
make docker-up

# Подождать ~10 секунд для инициализации
sleep 10

# Проверить статус
make curl-health
```

### Вариант C: Вручную через Docker Compose

```bash
# 1. Запустить PostgreSQL
docker-compose up -d postgres

# 2. Применить миграции
docker-compose up migrate

# 3. Запустить приложение
docker-compose up -d app

# 4. Проверить логи
docker-compose logs -f app
```

---

## Проверка работоспособности

### 1. Health Check

```bash
curl http://localhost:8080/healthz
```

Ожидаемый ответ:
```json
{"status":"healthy"}
```

- [ ] Health endpoint возвращает 200 OK
- [ ] Status = "healthy"

### 2. Readiness Check

```bash
curl http://localhost:8080/readyz
```

Ожидаемый ответ:
```json
{"ready":true,"providers":["4tochki"]}
```

- [ ] Readiness endpoint возвращает 200 OK
- [ ] Providers содержит "4tochki"

### 3. Проверка базы данных

```bash
# Подключиться к PostgreSQL
docker-compose exec postgres psql -U etalon -d etalon_price

# Проверить таблицы
\dt
```

Должны быть созданы таблицы:
- [ ] tyres_api
- [ ] rims_api
- [ ] product_offers_api
- [ ] sync_runs_api

---

## Первая синхронизация

### Вариант 1: HTTP API

```bash
curl -X POST http://localhost:8080/sync/4tochki \
  -H "Content-Type: application/json" \
  -d '{
    "codes": ["2329500", "WHS063930"]
  }'
```

### Вариант 2: CLI

```bash
docker-compose run --rm app /app/sync \
  --supplier=4tochki \
  --codes=2329500,WHS063930
```

### Вариант 3: Интеграционный тест

```bash
./test_integration.sh
```

---

## Проверка результатов

### 1. Проверить ответ синхронизации

Успешный ответ должен содержать:
```json
{
  "supplier": "4tochki",
  "sync_run_id": 1,
  "total_codes": 2,
  "tyres_saved": 1,
  "rims_saved": 1,
  "offers_saved": 2,
  "duration_ms": 1234,
  "errors": []
}
```

- [ ] HTTP status = 200
- [ ] sync_run_id > 0
- [ ] tyres_saved + rims_saved > 0
- [ ] errors = []

### 2. Проверить данные в БД

```bash
docker-compose exec postgres psql -U etalon -d etalon_price
```

```sql
-- Проверить шины
SELECT code, brand, model, width, height, diameter
FROM tyres_api
LIMIT 5;

-- Проверить диски
SELECT code, brand, model, width, diameter
FROM rims_api
LIMIT 5;

-- Проверить цены
SELECT code, price, stock, warehouse_name
FROM product_offers_api
LIMIT 5;

-- Проверить синхронизацию
SELECT * FROM sync_runs_api
ORDER BY created_at DESC
LIMIT 1;
```

Проверьте:
- [ ] Есть записи в tyres_api или rims_api
- [ ] Есть записи в product_offers_api
- [ ] sync_runs_api показывает status = "completed"
- [ ] success_products > 0

### 3. Проверить логи

```bash
docker-compose logs app | tail -50
```

Логи должны содержать:
- [ ] "HTTP server starting"
- [ ] "Provider registered" provider="4tochki"
- [ ] "Sync request received" (при синхронизации)
- [ ] "Sync completed" (при успехе)
- [ ] НЕТ ERROR или FATAL сообщений

---

## Troubleshooting

### Проблема: SOAP Fault / Authentication Failed

**Симптомы:**
```
SOAP Fault: Authentication failed
```

**Решение:**
1. Проверьте credentials в .env:
   ```bash
   cat .env | grep FOURTOCHKI
   ```
2. Убедитесь, что логин и пароль верные
3. Перезапустите приложение:
   ```bash
   docker-compose restart app
   ```

---

### Проблема: "No data in database"

**Симптомы:**
- Синхронизация завершается успешно
- Но БД пустая

**Решение:**
1. Проверьте коды товаров - они должны существовать у поставщика
2. Проверьте логи на уровне debug:
   ```bash
   # В .env
   LOG_LEVEL=debug
   # Перезапустить
   docker-compose restart app
   ```
3. Проверьте таблицу sync_runs_api:
   ```sql
   SELECT error_message FROM sync_runs_api
   ORDER BY created_at DESC LIMIT 1;
   ```

---

### Проблема: Connection timeout

**Симптомы:**
```
context deadline exceeded
connection timeout
```

**Решение:**
Увеличьте таймауты в .env:
```env
FOURTOCHKI_TIMEOUT=120s
FOURTOCHKI_RETRY_COUNT=5
FOURTOCHKI_RETRY_DELAY=10s
```

Перезапустите:
```bash
docker-compose restart app
```

---

### Проблема: "migrate: no change"

**Симптомы:**
Миграции не применяются

**Решение:**
```bash
# Проверить версию миграций
docker-compose exec postgres psql -U etalon -d etalon_price -c "SELECT * FROM schema_migrations;"

# Если нужно, применить принудительно
docker-compose run --rm migrate -path /migrations \
  -database "postgres://etalon:etalon_pass@postgres:5432/etalon_price?sslmode=disable" \
  up
```

---

## Production Deployment

Перед деплоем в production:

- [ ] Сменить `APP_ENV=production` в .env
- [ ] Сменить `LOG_LEVEL=info` или `warn`
- [ ] Использовать безопасный DB_DSN с SSL
- [ ] Настроить backup БД
- [ ] Настроить мониторинг (health checks)
- [ ] Настроить alerts на ошибки
- [ ] Настроить cron для периодической синхронизации
- [ ] Убедиться, что используются production credentials

---

## Полезные команды

```bash
# Просмотр логов
docker-compose logs -f app

# Перезапуск приложения
docker-compose restart app

# Остановка всего
docker-compose down

# Полная очистка (БД будет удалена!)
docker-compose down -v

# Статус контейнеров
docker-compose ps

# Проверка health
curl http://localhost:8080/healthz

# Проверка stats
curl http://localhost:8080/stats
```

---

## Финальный чеклист

### Готово к работе, если:

- ✅ Docker containers running (postgres, app)
- ✅ /healthz returns {"status":"healthy"}
- ✅ /readyz returns {"ready":true}
- ✅ Таблицы созданы в БД
- ✅ Тестовая синхронизация прошла успешно
- ✅ Данные появились в БД
- ✅ Нет ошибок в логах

### Готово к production, если:

- ✅ Используются production credentials
- ✅ APP_ENV=production
- ✅ SSL для БД настроен
- ✅ Backup настроен
- ✅ Мониторинг настроен
- ✅ Проведено нагрузочное тестирование

---

**Дата:** 2025-03-12
**Версия:** 1.0.0
