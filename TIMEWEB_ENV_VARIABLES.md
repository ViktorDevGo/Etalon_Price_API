# Переменные окружения для Timeweb Cloud

## Общие настройки (для всех сервисов)

### База данных
```bash
DATABASE_DSN=postgresql://gen_user:Poison-79@c37e696087932476c61fd621.twc1.net:5432/default_db?sslmode=require
```
**⚠️ ВАЖНО:** Используйте `DATABASE_DSN`, НЕ `DB_DSN`!

### Приложение
```bash
APP_ENV=production
LOG_LEVEL=info
```

### Email уведомления
```bash
EMAIL_ENABLED=true
EMAIL_SMTP_HOST=mail.hosting.reg.ru
EMAIL_SMTP_PORT=587
EMAIL_USERNAME=admin@etalon-shina.ru
EMAIL_PASSWORD=S69Y1ypojVLCZHO8
EMAIL_FROM=admin@etalon-shina.ru
EMAIL_NOTIFICATION_TO=v.boyarkin@etalon-shina.ru
```

---

## Сервис 1: API Server (Dockerfile.production)

### Порт
```bash
HTTP_PORT=8080
```

### Форточки API
```bash
FOURTOCHKI_WSDL_URL=http://api-b2b.4tochki.ru/WCF/ClientService.svc?wsdl
FOURTOCHKI_BASE_URL=http://api-b2b.4tochki.ru/WCF/ClientService.svc?wsdl
FOURTOCHKI_LOGIN=sa69263
FOURTOCHKI_PASSWORD=jkHP4Nj3)z
FOURTOCHKI_BATCH_SIZE=50
FOURTOCHKI_TIMEOUT=60s
FOURTOCHKI_RETRY_COUNT=3
FOURTOCHKI_RETRY_DELAY=5s
FOURTOCHKI_BATCH_DELAY=2s
```

### Северавто API
```bash
SEVERAVTO_BASE_URL=http://webmim.svrauto.ru
SEVERAVTO_API_KEY=TpU0K90z0X0LeNWyWiBdWS0kP3mvdzOf
SEVERAVTO_TIMEOUT=120s
```

### Timezone
```bash
TZ=Europe/Moscow
```

**Порт в Timeweb:** 8080

---

## Сервис 2: Nomenclature Scheduler (Dockerfile.nomenclature-scheduler)

Запускается каждый день в **2:00 AM MSK**.

### Переменные окружения:
```bash
APP_ENV=production
LOG_LEVEL=info
DATABASE_DSN=postgresql://gen_user:Poison-79@c37e696087932476c61fd621.twc1.net:5432/default_db?sslmode=require

FOURTOCHKI_WSDL_URL=http://api-b2b.4tochki.ru/WCF/ClientService.svc?wsdl
FOURTOCHKI_BASE_URL=http://api-b2b.4tochki.ru/WCF/ClientService.svc?wsdl
FOURTOCHKI_LOGIN=sa69263
FOURTOCHKI_PASSWORD=jkHP4Nj3)z
FOURTOCHKI_BATCH_SIZE=50
FOURTOCHKI_TIMEOUT=60s
FOURTOCHKI_RETRY_COUNT=3
FOURTOCHKI_RETRY_DELAY=5s
FOURTOCHKI_BATCH_DELAY=2s

EMAIL_ENABLED=true
EMAIL_SMTP_HOST=mail.hosting.reg.ru
EMAIL_SMTP_PORT=587
EMAIL_USERNAME=admin@etalon-shina.ru
EMAIL_PASSWORD=S69Y1ypojVLCZHO8
EMAIL_FROM=admin@etalon-shina.ru
EMAIL_NOTIFICATION_TO=v.boyarkin@etalon-shina.ru

TZ=Europe/Moscow
```

**Порт:** Не требуется (фоновый процесс)

---

## Сервис 3: Prices Scheduler (Dockerfile.prices-scheduler)

Запускается **каждые 3 часа** (00:00, 03:00, 06:00, 09:00, 12:00, 15:00, 18:00, 21:00 MSK).

### Переменные окружения:
```bash
APP_ENV=production
LOG_LEVEL=info
DATABASE_DSN=postgresql://gen_user:Poison-79@c37e696087932476c61fd621.twc1.net:5432/default_db?sslmode=require

FOURTOCHKI_BASE_URL=http://api-b2b.4tochki.ru/WCF/ClientService.svc?wsdl
FOURTOCHKI_LOGIN=sa69263
FOURTOCHKI_PASSWORD=jkHP4Nj3)z
FOURTOCHKI_BATCH_SIZE=50
FOURTOCHKI_TIMEOUT=60s
FOURTOCHKI_RETRY_COUNT=3
FOURTOCHKI_RETRY_DELAY=5s
FOURTOCHKI_BATCH_DELAY=2s

PRICES_CRON_SCHEDULE=0 */3 * * *

EMAIL_ENABLED=true
EMAIL_SMTP_HOST=mail.hosting.reg.ru
EMAIL_SMTP_PORT=587
EMAIL_USERNAME=admin@etalon-shina.ru
EMAIL_PASSWORD=S69Y1ypojVLCZHO8
EMAIL_FROM=admin@etalon-shina.ru
EMAIL_NOTIFICATION_TO=v.boyarkin@etalon-shina.ru

TZ=Europe/Moscow
```

**Порт:** Не требуется (фоновый процесс)

---

## Сервис 4: Prices Upload (Dockerfile.prices-upload)

Запускается каждый день в **10:00 IRKT** (05:00 MSK).
Выгружает каталог и переоценку на 1C-Битрикс.

### Переменные окружения:
```bash
APP_ENV=production
LOG_LEVEL=info
DATABASE_DSN=postgresql://gen_user:Poison-79@c37e696087932476c61fd621.twc1.net:5432/default_db?sslmode=require
WORK_DIR=/app
TZ=Asia/Irkutsk
```

**Порт:** Не требуется (фоновый процесс)

---

## Сервис 5: Severavto Scheduler (Dockerfile.severavto-scheduler)

Запускается **каждые 3 часа**:
- Шины: 00:00, 03:00, 06:00, 09:00, 12:00, 15:00, 18:00, 21:00 MSK
- Диски: 00:15, 03:15, 06:15, 09:15, 12:15, 15:15, 18:15, 21:15 MSK

### Переменные окружения:
```bash
APP_ENV=production
LOG_LEVEL=info
DATABASE_DSN=postgresql://gen_user:Poison-79@c37e696087932476c61fd621.twc1.net:5432/default_db?sslmode=require

SEVERAVTO_BASE_URL=http://webmim.svrauto.ru
SEVERAVTO_API_KEY=TpU0K90z0X0LeNWyWiBdWS0kP3mvdzOf
SEVERAVTO_TIMEOUT=120s

EMAIL_ENABLED=true
EMAIL_SMTP_HOST=mail.hosting.reg.ru
EMAIL_SMTP_PORT=587
EMAIL_USERNAME=admin@etalon-shina.ru
EMAIL_PASSWORD=S69Y1ypojVLCZHO8
EMAIL_FROM=admin@etalon-shina.ru
EMAIL_NOTIFICATION_TO=v.boyarkin@etalon-shina.ru

TZ=Europe/Moscow
```

**Порт:** Не требуется (фоновый процесс)

---

## Инструкция по деплою на Timeweb Cloud

### 1. Создание Cloud Apps

Создайте **5 отдельных Cloud App** в Timeweb:

1. **etalon-api** (Dockerfile.production)
2. **etalon-nomenclature-scheduler** (Dockerfile.nomenclature-scheduler)
3. **etalon-prices-scheduler** (Dockerfile.prices-scheduler)
4. **etalon-prices-upload** (Dockerfile.prices-upload)
5. **etalon-severavto-scheduler** (Dockerfile.severavto-scheduler)

### 2. Настройка для каждого сервиса

Для каждого Cloud App:

1. **Подключите GitHub репозиторий**
2. **Выберите нужный Dockerfile** (см. выше)
3. **Добавьте переменные окружения** (скопируйте из соответствующего раздела выше)
4. **Укажите порт** (только для API: 8080)
5. **Запустите деплой**

### 3. Проверка работы

После деплоя проверьте логи каждого сервиса:

```bash
# API сервер - должен быть доступен
curl https://ваш-домен.twc1.net/health

# Планировщики - проверьте логи в Timeweb панели
# Должны видеть:
# - "Starting ... scheduler"
# - "Next sync at ..."
```

### 4. Мониторинг

Email уведомления будут приходить на `v.boyarkin@etalon-shina.ru` после каждой синхронизации:
- ✅ Успешно: список загруженных товаров
- ❌ Ошибка: детали ошибки

---

## Важные замечания

⚠️ **Северавто API**
- Сервис запустится, но синхронизация не будет работать до получения доступа к API
- После подтверждения доступа от техподдержки, синхронизация начнется автоматически
- Письмо в поддержку отправлено, ждем ответа

⚠️ **Секреты**
- Не коммитьте файл `.env` в Git!
- Все пароли указывайте только в переменных окружения Timeweb

⚠️ **База данных**
- Используется управляемая PostgreSQL от Timeweb
- SSL обязателен: `?sslmode=require`
- Строка подключения одна для всех сервисов

---

## Контакты для поддержки

- **Timeweb Support:** через панель управления
- **Email:** v.boyarkin@etalon-shina.ru
- **GitHub Issues:** для багов и вопросов по коду
