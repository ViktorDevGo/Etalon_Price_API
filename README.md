# Etalon Price API

Production-ready сервис интеграции с поставщиками автомобильных шин и дисков.

## 🚀 Возможности

- 🔄 Синхронизация данных о товарах (шины, диски) от поставщиков
- 💰 Получение актуальных цен и остатков
- 🗄️ Хранение данных в PostgreSQL с автоматическим upsert
- 🔌 Расширяемая архитектура с поддержкой нескольких поставщиков
- 📊 Структурированное логирование в JSON формате
- 🐳 Docker support с multi-stage builds
- 🔁 Retry/backoff для надежности
- ⚡ Batch processing для производительности
- 🧪 Unit тесты

## 📦 Поддерживаемые поставщики

- ✅ **4tochki (Форточки)** - SOAP API интеграция
- 🔜 Другие поставщики (расширяемая архитектура)

## 🛠 Требования

- **Go** 1.24+
- **PostgreSQL** 14+
- **Docker & Docker Compose** (опционально)
- **golang-migrate** (для миграций)

## 📥 Быстрый старт

### 1. Клонирование репозитория

```bash
git clone <repository-url>
cd Etalon_Price_API
```

### 2. Настройка окружения

```bash
cp .env.example .env
# Отредактируйте .env и укажите реальные учетные данные
```

Обязательные переменные:
```env
FOURTOCHKI_LOGIN=your_login
FOURTOCHKI_PASSWORD=your_password
```

### 3. Запуск через Docker (рекомендуется)

```bash
# Полный стек (PostgreSQL + миграции + API)
make docker-up

# Или вручную
docker-compose up -d
```

API будет доступен на `http://localhost:8080`

### 4. Локальная разработка

```bash
# Установить зависимости
make deps

# Запустить только БД
make db-setup

# Запустить HTTP сервер
make run

# Запустить синхронизацию
make sync
```

## 📁 Структура проекта

```
.
├── cmd/
│   ├── app/              # HTTP API сервер
│   └── sync/             # CLI для синхронизации
├── internal/
│   ├── config/           # Конфигурация
│   ├── logger/           # Структурированное логирование
│   ├── domain/           # Domain модели
│   ├── providers/        # Интерфейсы и реализации поставщиков
│   │   └── 4tochki/      # SOAP клиент для 4tochki
│   ├── service/          # Бизнес-логика синхронизации
│   ├── repository/       # PostgreSQL репозитории
│   │   └── postgres/     # pgx реализация
│   ├── httpserver/       # HTTP handlers
│   └── migrations/       # SQL миграции
├── tests/                # Unit тесты
├── Dockerfile            # Multi-stage build
├── docker-compose.yml    # Development stack
├── Makefile              # Команды для разработки
└── README.md             # Этот файл
```

## ⚙️ Конфигурация

### Переменные окружения

#### Основные настройки

| Переменная | Описание | По умолчанию |
|-----------|----------|--------------|
| `APP_ENV` | Окружение (development/production/staging) | development |
| `HTTP_PORT` | Порт HTTP сервера | 8080 |
| `LOG_LEVEL` | Уровень логирования (debug/info/warn/error) | info |

#### База данных

| Переменная | Описание | Обязательно |
|-----------|----------|-------------|
| `DB_DSN` | PostgreSQL connection string | ✅ |

Пример:
```env
DB_DSN=postgres://etalon:etalon_pass@localhost:5432/etalon_price?sslmode=disable
```

#### Провайдер 4tochki

| Переменная | Описание | По умолчанию |
|-----------|----------|--------------|
| `FOURTOCHKI_WSDL_URL` | URL WSDL сервиса | http://api-b2b.4tochki.ru/WCF/ClientService.svc?wsdl |
| `FOURTOCHKI_LOGIN` | Логин | ✅ Обязательно |
| `FOURTOCHKI_PASSWORD` | Пароль | ✅ Обязательно |
| `FOURTOCHKI_BATCH_SIZE` | Размер батча | 100 |
| `FOURTOCHKI_TIMEOUT` | Таймаут запроса | 30s |
| `FOURTOCHKI_RETRY_COUNT` | Количество повторов | 3 |
| `FOURTOCHKI_RETRY_DELAY` | Задержка между повторами | 5s |
| `FOURTOCHKI_BATCH_DELAY` | Задержка между батчами | 1s |

## 🗄️ База данных

### Таблицы

- **tyres_api** - Шины (бренд, модель, размеры, сезон, характеристики)
- **rims_api** - Диски (бренд, модель, размеры, параметры)
- **product_offers_api** - Цены и остатки (склады, количество, цены)
- **sync_runs_api** - История синхронизаций (статус, метрики)

### Миграции

```bash
# Применить миграции
make migrate-up

# Откатить последнюю миграцию
make migrate-down

# Создать новую миграцию
make migrate-create name=add_supplier_table

# Форсировать версию (при проблемах)
make migrate-force version=1
```

## 🔄 Использование

### HTTP API

#### Запуск сервера

```bash
# Локально
make run

# В Docker
make docker-up
```

#### Endpoints

**Health Check**
```bash
curl http://localhost:8080/healthz
```

**Readiness Check**
```bash
curl http://localhost:8080/readyz
```

**Синхронизация товаров**
```bash
curl -X POST http://localhost:8080/sync/4tochki \
  -H "Content-Type: application/json" \
  -d '{
    "codes": ["2329500", "WHS063930"]
  }'
```

Ответ:
```json
{
  "supplier": "4tochki",
  "sync_run_id": 123,
  "total_codes": 2,
  "tyres_saved": 1,
  "rims_saved": 1,
  "offers_saved": 2,
  "duration_ms": 1234,
  "errors": []
}
```

**Статистика**
```bash
curl http://localhost:8080/stats
```

### CLI Синхронизация

#### Синхронизация по кодам (через флаг)

```bash
./bin/sync --supplier=4tochki --codes=2329500,WHS063930
```

#### Синхронизация из файла

Создайте файл `codes.txt`:
```
2329500
WHS063930
# Комментарии игнорируются
```

Запустите:
```bash
./bin/sync --supplier=4tochki --codes-file=codes.txt
```

#### Опции CLI

| Флаг | Описание | По умолчанию |
|------|----------|--------------|
| `--supplier` | Имя поставщика | 4tochki |
| `--codes` | Коды через запятую | - |
| `--codes-file` | Путь к файлу с кодами | - |
| `--batch-size` | Размер батча | Из конфига |

## 🧪 Тестирование

```bash
# Запустить все тесты
make test

# Только unit тесты
make test-unit

# С покрытием
make test-coverage
```

## 📋 Makefile команды

```bash
make help           # Показать все команды
make build          # Собрать бинарники
make run            # Запустить HTTP сервер
make sync           # Запустить синхронизацию
make test           # Запустить тесты

# Docker
make docker-build   # Собрать Docker образы
make docker-up      # Запустить контейнеры
make docker-down    # Остановить контейнеры
make docker-logs    # Показать логи
make docker-clean   # Полная очистка

# База данных
make db-setup       # Настроить БД
make db-reset       # Сбросить БД
make migrate-up     # Применить миграции
make migrate-down   # Откатить миграции

# Разработка
make deps           # Загрузить зависимости
make fmt            # Форматировать код
make vet            # Проверить код

# Тестовые curl запросы
make curl-health    # Проверить health
make curl-sync      # Тестовая синхронизация
```

## 🏗️ Архитектура

### Компоненты

1. **HTTP Server** (`cmd/app`) - REST API для синхронизации
2. **CLI** (`cmd/sync`) - Командная строка для batch операций
3. **SOAP Client** (`providers/4tochki`) - Интеграция с 4tochki API
4. **Sync Service** (`service`) - Оркестрация синхронизации
5. **Repositories** (`repository/postgres`) - Работа с БД

### Поток данных

```
CLI/HTTP → Sync Service → Provider (SOAP) → Mapper → Repository → PostgreSQL
                ↓
          Sync Runs Tracking
```

### Ключевые паттерны

- **Repository Pattern** - изоляция БД логики
- **Provider Pattern** - расширяемость для новых поставщиков
- **Batch Processing** - эффективная обработка больших объемов
- **Graceful Shutdown** - безопасная остановка
- **Structured Logging** - машинно-читаемые логи

## 🚀 Production Deployment

### Docker

```bash
# Собрать образ
docker build --target app -t etalon-price-api:latest .

# Запустить
docker run -d \
  -p 8080:8080 \
  -e DB_DSN="postgres://..." \
  -e FOURTOCHKI_LOGIN="..." \
  -e FOURTOCHKI_PASSWORD="..." \
  etalon-price-api:latest
```

### Environment Variables

Для production обязательно установите:
- `APP_ENV=production`
- `LOG_LEVEL=info` или `warn`
- Безопасный `DB_DSN` с SSL
- Реальные credentials для поставщиков

### Мониторинг

- Health check: `GET /healthz`
- Readiness probe: `GET /readyz`
- Metrics: `GET /stats`
- Logs: JSON structured (stdout)

## 🔧 Разработка

### Добавление нового поставщика

1. Создайте папку `internal/providers/newprovider/`
2. Реализуйте интерфейс `SupplierProvider`
3. Зарегистрируйте в `main.go`

```go
provider := newprovider.NewProvider(config)
registry.Register(provider)
```

### Code Style

```bash
# Форматирование
make fmt

# Проверка кода
make vet
```

## 📄 Лицензия

Proprietary

## 👥 Команда

ProKoleso Team

---

**Документация обновлена:** 2025-03-12
