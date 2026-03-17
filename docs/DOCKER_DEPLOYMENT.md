# Docker Deployment Guide

## Обзор

Проект Etalon Price API полностью поддерживает развёртывание в Docker Compose с автоматическими обновлениями данных.

## Архитектура

```
┌─────────────────────────────────────────────────────────────┐
│                     Docker Compose Stack                     │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌──────────────┐  ┌──────────────────────────────────┐    │
│  │   API Server │  │  Nomenclature Scheduler          │    │
│  │   (port 8080)│  │  (cron: daily at 2:00 AM)        │    │
│  └──────────────┘  └──────────────────────────────────┘    │
│                                                               │
│  ┌──────────────────────────────────────┐                   │
│  │  Prices Scheduler                    │                   │
│  │  (cron: every 3 hours by default)    │                   │
│  └──────────────────────────────────────┘                   │
│                                                               │
│  ┌──────────────────────────────────────┐                   │
│  │  PostgreSQL 16                       │                   │
│  │  (optional - can use external DB)    │                   │
│  └──────────────────────────────────────┘                   │
│                                                               │
└─────────────────────────────────────────────────────────────┘
```

## Быстрый старт

### 1. Подготовка окружения

Создайте файл `.env` в корне проекта:

```bash
# Database
DATABASE_DSN=postgres://user:password@postgres:5432/etalon_price?sslmode=disable
POSTGRES_DB=etalon_price
POSTGRES_USER=postgres
POSTGRES_PASSWORD=your_secure_password
POSTGRES_PORT=5432

# Fourtochki API
FOURTOCHKI_BASE_URL=https://api.4tochki.ru
FOURTOCHKI_LOGIN=your_login
FOURTOCHKI_PASSWORD=your_password

# API Server
API_PORT=8080

# Prices sync schedule (cron format)
# Default: every 3 hours
PRICES_CRON_SCHEDULE=0 */3 * * *
```

### 2. Запуск всех сервисов

```bash
docker-compose up -d
```

### 3. Проверка статуса

```bash
# Проверить все контейнеры
docker-compose ps

# Логи всех сервисов
docker-compose logs -f

# Логи конкретного сервиса
docker-compose logs -f nomenclature-scheduler
docker-compose logs -f prices-scheduler
```

## Конфигурация

### Расписание синхронизации

#### Nomenclature Scheduler
Жёстко зашит в код: **каждый день в 2:00 AM**

Реализован через `robfig/cron/v3` в Go коде (`cmd/nomenclature-scheduler/main.go`).

#### Prices Scheduler
Настраивается через переменную окружения `PRICES_CRON_SCHEDULE`.

Формат: стандартный cron (5 полей)

Примеры:
```bash
# Каждые 3 часа (по умолчанию)
PRICES_CRON_SCHEDULE=0 */3 * * *

# Каждые 6 часов
PRICES_CRON_SCHEDULE=0 */6 * * *

# Каждый час
PRICES_CRON_SCHEDULE=0 * * * *

# Каждый день в 3:00 AM
PRICES_CRON_SCHEDULE=0 3 * * *

# Каждый день в 3:00 и 15:00
PRICES_CRON_SCHEDULE=0 3,15 * * *
```

### Часовой пояс

Все сервисы используют `TZ=Europe/Moscow`.

Для изменения установите переменную `TZ` в `docker-compose.yml` или `.env`.

## Использование внешней БД

Если вы используете внешний PostgreSQL (например, на хостинге):

1. Отредактируйте `docker-compose.yml` - закомментируйте или удалите сервис `postgres`
2. Обновите `DATABASE_DSN` в `.env`:
   ```bash
   DATABASE_DSN=postgres://user:password@external-host:5432/dbname?sslmode=disable
   ```
3. Удалите `depends_on: - postgres` из всех сервисов

## Ручной запуск синхронизации

### Nomenclature (Fourtochki)

```bash
# Синхронизация номенклатуры шин
docker-compose run --rm nomenclature-scheduler ./sync-nomenclature -type=tyres

# Синхронизация номенклатуры дисков
docker-compose run --rm nomenclature-scheduler ./sync-nomenclature -type=rims

# Синхронизация всей номенклатуры
docker-compose run --rm nomenclature-scheduler ./sync-nomenclature -type=all
```

### Prices (Fourtochki)

```bash
# Синхронизация цен и остатков шин
docker-compose run --rm prices-scheduler ./sync-prices -type=tyres

# Синхронизация цен и остатков дисков
docker-compose run --rm prices-scheduler ./sync-prices -type=rims

# Синхронизация всех цен
docker-compose run --rm prices-scheduler ./sync-prices -type=all
```

## Мониторинг и логи

### Просмотр логов

```bash
# Все логи (real-time)
docker-compose logs -f

# Логи nomenclature-scheduler
docker-compose logs -f nomenclature-scheduler

# Логи prices-scheduler
docker-compose logs -f prices-scheduler

# Последние 100 строк
docker-compose logs --tail=100 nomenclature-scheduler
```

### Проверка расписания cron

Для prices-scheduler (Alpine crond):
```bash
docker-compose exec prices-scheduler cat /etc/crontabs/root
```

### Health Check

API сервер имеет встроенный health check endpoint:
```bash
curl http://localhost:8080/health
```

Docker автоматически проверяет health каждые 30 секунд.

### Ротация логов

Логи автоматически ротируются Docker:
- Максимальный размер файла: 10MB
- Количество файлов: 3
- Итого: до 30MB логов на сервис

Конфигурация в `docker-compose.yml`:
```yaml
logging:
  driver: "json-file"
  options:
    max-size: "10m"
    max-file: "3"
```

## Обновление приложения

### Пересборка после изменений кода

```bash
# Остановить контейнеры
docker-compose down

# Пересобрать образы
docker-compose build

# Запустить заново
docker-compose up -d
```

### Быстрое обновление конкретного сервиса

```bash
# Пересобрать и перезапустить только API
docker-compose up -d --build api

# Пересобрать и перезапустить scheduler'ы
docker-compose up -d --build nomenclature-scheduler prices-scheduler
```

## Применение миграций

### Автоматическое применение

При использовании встроенного PostgreSQL миграции применяются автоматически при первом запуске.

### Ручное применение

```bash
# Подключиться к контейнеру postgres
docker-compose exec postgres psql -U postgres -d etalon_price

# Или выполнить миграцию из файла
docker-compose exec postgres psql -U postgres -d etalon_price -f /docker-entrypoint-initdb.d/001_initial_schema.up.sql
```

Для внешней БД используйте стандартные инструменты (psql, DBeaver, etc).

## Бэкап и восстановление

### Бэкап базы данных

```bash
# Создать бэкап
docker-compose exec postgres pg_dump -U postgres etalon_price > backup_$(date +%Y%m%d_%H%M%S).sql

# С сжатием
docker-compose exec postgres pg_dump -U postgres etalon_price | gzip > backup_$(date +%Y%m%d_%H%M%S).sql.gz
```

### Восстановление из бэкапа

```bash
# Из несжатого файла
cat backup_20260315_120000.sql | docker-compose exec -T postgres psql -U postgres etalon_price

# Из сжатого файла
gunzip -c backup_20260315_120000.sql.gz | docker-compose exec -T postgres psql -U postgres etalon_price
```

## Troubleshooting

### Scheduler не запускается

**Проверка 1: Контейнер работает?**
```bash
docker-compose ps
```

**Проверка 2: Логи**
```bash
docker-compose logs nomenclature-scheduler
docker-compose logs prices-scheduler
```

**Проверка 3: Переменные окружения**
```bash
docker-compose exec nomenclature-scheduler env | grep DATABASE
```

### Cron не выполняется

**Для nomenclature-scheduler:**
Проверьте логи - вы должны видеть сообщение "Starting nomenclature scheduler..."

**Для prices-scheduler:**
```bash
# Проверить cron настройки
docker-compose exec prices-scheduler cat /etc/crontabs/root

# Проверить логи cron
docker-compose exec prices-scheduler tail -f /var/log/cron.log
```

### Ошибка подключения к БД

**Симптом:** `Failed to connect to database`

**Решение:**
1. Проверьте `DATABASE_DSN` в `.env`
2. Убедитесь, что PostgreSQL запущен: `docker-compose ps postgres`
3. Проверьте логи PostgreSQL: `docker-compose logs postgres`
4. Для внешней БД проверьте доступность хоста и порта

### Timezone проблемы

**Симптом:** Cron выполняется не в то время

**Решение:**
```bash
# Проверить timezone в контейнере
docker-compose exec nomenclature-scheduler date
docker-compose exec prices-scheduler date

# Должно показать время по Moscow (MSK)
```

Если время неправильное, проверьте `TZ=Europe/Moscow` в `docker-compose.yml`.

### Недостаточно места на диске

**Симптом:** `no space left on device`

**Решение:**
```bash
# Очистить неиспользуемые образы
docker system prune -a

# Очистить volumes (ВНИМАНИЕ: удалит данные!)
docker system prune -a --volumes

# Посмотреть использование
docker system df
```

## Production рекомендации

### 1. Используйте конкретные версии образов

В `docker-compose.yml` замените `alpine:latest` и `postgres:16-alpine` на конкретные версии:
```yaml
postgres:
  image: postgres:16.2-alpine  # вместо postgres:16-alpine
```

### 2. Настройте мониторинг

Интегрируйте с системами мониторинга:
- Prometheus + Grafana для метрик
- ELK Stack или Loki для централизованных логов
- Uptime проверки для API endpoint

### 3. Secrets management

Не храните пароли в `.env` файле в production. Используйте:
- Docker Secrets (Swarm mode)
- HashiCorp Vault
- Облачные secrets managers (AWS Secrets Manager, etc.)

### 4. Reverse proxy

Добавьте nginx или traefik перед API:
```yaml
nginx:
  image: nginx:alpine
  ports:
    - "80:80"
    - "443:443"
  volumes:
    - ./nginx.conf:/etc/nginx/nginx.conf:ro
    - ./certs:/etc/nginx/certs:ro
```

### 5. Healthchecks

Настройте внешние healthchecks (UptimeRobot, Pingdom) для критических сервисов.

### 6. Backup автоматизация

Создайте отдельный сервис для автоматических бэкапов:
```yaml
backup:
  image: postgres:16-alpine
  environment:
    - PGPASSWORD=${POSTGRES_PASSWORD}
  volumes:
    - ./backups:/backups
  entrypoint: |
    sh -c 'while true; do
      pg_dump -h postgres -U postgres etalon_price | gzip > /backups/backup_$$(date +%Y%m%d_%H%M%S).sql.gz
      sleep 86400
    done'
```

## Дополнительные ресурсы

- [Docker Compose документация](https://docs.docker.com/compose/)
- [PostgreSQL Docker Hub](https://hub.docker.com/_/postgres)
- [Alpine Linux документация](https://wiki.alpinelinux.org/)
- [Cron формат](https://crontab.guru/)

## Поддержка

При возникновении проблем:
1. Проверьте логи: `docker-compose logs -f`
2. Проверьте статус: `docker-compose ps`
3. Проверьте документацию проекта
4. Создайте issue в репозитории
