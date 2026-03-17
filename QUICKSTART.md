# 🚀 Quick Start Guide

## Шаг 1: Получение доступа к API 4tochki

### Как получить credentials:

1. **Свяжитесь с поставщиком:**
   - Сайт: https://www.4tochki.ru/
   - Email: support@4tochki.ru (уточните актуальный)
   - Телефон: уточните на сайте

2. **Запросите B2B доступ:**
   - Укажите, что вам нужен доступ к SOAP API
   - WSDL URL: `http://api-b2b.4tochki.ru/WCF/ClientService.svc?wsdl`
   - Попросите предоставить:
     - Логин (username)
     - Пароль (password)
     - Тестовые коды товаров для проверки

3. **Получите тестовые коды:**
   - Попросите несколько артикулов шин и дисков
   - Это поможет проверить интеграцию

---

## Шаг 2: Настройка проекта

### Автоматическая настройка (рекомендуется):

```bash
# Запустите setup скрипт
./setup.sh
```

Скрипт автоматически:
- ✅ Создаст .env файл
- ✅ Запустит PostgreSQL
- ✅ Применит миграции
- ✅ Соберет и запустит приложение

### Ручная настройка:

```bash
# 1. Скопировать .env
cp .env.example .env

# 2. Отредактировать .env
nano .env

# Укажите реальные credentials:
FOURTOCHKI_LOGIN=ваш_логин
FOURTOCHKI_PASSWORD=ваш_пароль

# 3. Запустить Docker stack
make docker-up
```

---

## Шаг 3: Проверка работы

### 1. Проверить health:

```bash
curl http://localhost:8080/healthz
```

Ожидаемый ответ:
```json
{"status":"healthy"}
```

### 2. Проверить readiness:

```bash
curl http://localhost:8080/readyz
```

Ожидаемый ответ:
```json
{"ready":true,"providers":["4tochki"]}
```

---

## Шаг 4: Первая синхронизация

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
# С конкретными кодами
docker-compose run --rm app /app/sync \
  --supplier=4tochki \
  --codes=2329500,WHS063930

# Из файла
docker-compose run --rm app /app/sync \
  --supplier=4tochki \
  --codes-file=/app/test_codes.txt
```

### Вариант 3: Локальный бинарник

```bash
# Собрать
make build

# Запустить
./bin/sync --supplier=4tochki --codes=2329500,WHS063930
```

---

## Шаг 5: Проверка данных в БД

### Подключиться к PostgreSQL:

```bash
# Через Docker
docker-compose exec postgres psql -U etalon -d etalon_price

# Или напрямую
psql postgres://etalon:etalon_pass@localhost:5432/etalon_price
```

### Проверить синхронизированные данные:

```sql
-- Шины
SELECT code, brand, model, width, height, diameter, season
FROM tyres_api
ORDER BY created_at DESC
LIMIT 10;

-- Диски
SELECT code, brand, model, width, diameter, bolt_pattern
FROM rims_api
ORDER BY created_at DESC
LIMIT 10;

-- Цены и остатки
SELECT code, price, stock, warehouse_name
FROM product_offers_api
ORDER BY created_at DESC
LIMIT 10;

-- История синхронизаций
SELECT id, provider, status, total_products, success_products, started_at
FROM sync_runs_api
ORDER BY started_at DESC
LIMIT 10;
```

---

## Шаг 6: Мониторинг

### Просмотр логов:

```bash
# Все логи
docker-compose logs -f

# Только приложение
docker-compose logs -f app

# Последние 100 строк
docker-compose logs --tail=100 app
```

### Статистика БД:

```bash
curl http://localhost:8080/stats
```

---

## Troubleshooting

### Проблема: "SOAP Fault" ошибка

**Причина:** Неверные credentials или коды товаров

**Решение:**
1. Проверьте логин/пароль в .env
2. Убедитесь, что коды существуют в системе 4tochki
3. Проверьте логи: `docker-compose logs app`

### Проблема: "Connection refused" при подключении к БД

**Причина:** PostgreSQL не запущен или не готов

**Решение:**
```bash
# Перезапустить PostgreSQL
docker-compose restart postgres

# Подождать готовности
sleep 10

# Проверить статус
docker-compose ps postgres
```

### Проблема: Пустая БД после синхронизации

**Причина:** Возможно, коды не найдены или ошибка маппинга

**Решение:**
1. Проверьте ответ API в логах (LOG_LEVEL=debug)
2. Проверьте таблицу sync_runs_api на наличие ошибок:
```sql
SELECT * FROM sync_runs_api ORDER BY created_at DESC LIMIT 1;
```

### Проблема: Timeout при запросе к API

**Причина:** API медленно отвечает или большой батч

**Решение:**
Увеличьте timeout в .env:
```env
FOURTOCHKI_TIMEOUT=120s
FOURTOCHKI_BATCH_SIZE=20
```

---

## Debug режим

### Включить подробное логирование:

```env
# В .env
LOG_LEVEL=debug
```

Перезапустить:
```bash
docker-compose restart app
```

### Просмотр SOAP запросов/ответов:

Логи будут показывать:
- URL запроса
- Размер тела запроса/ответа
- Длительность выполнения
- Все ошибки

---

## Полезные команды

```bash
# Остановить все
docker-compose down

# Полная очистка (включая volumes)
docker-compose down -v

# Пересобрать образы
docker-compose build --no-cache

# Перезапустить приложение
docker-compose restart app

# Проверить статус
docker-compose ps

# Зайти в контейнер
docker-compose exec app sh

# Выполнить миграции вручную
docker-compose up migrate
```

---

## Готово к production?

После успешного тестирования:

1. **Смените окружение:**
   ```env
   APP_ENV=production
   LOG_LEVEL=info
   ```

2. **Настройте SSL для БД:**
   ```env
   DB_DSN=postgres://user:pass@host:5432/db?sslmode=require
   ```

3. **Добавьте мониторинг:**
   - Health checks: `/healthz`
   - Metrics: `/stats`
   - Logs: stdout (JSON format)

4. **Настройте backup БД**

5. **Настройте cron для периодической синхронизации**

---

## Контакты

При возникновении проблем:
- Проверьте логи: `docker-compose logs app`
- Проверьте README.md для подробной документации
- Создайте issue в репозитории

Успешной интеграции! 🚀
