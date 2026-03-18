# Переменные окружения для Timeweb VM (Docker Compose)

## 📋 Полный список переменных для настройки в панели Timeweb

Скопируйте эти переменные в раздел **"Переменные окружения"** при настройке деплоя.

### Основные настройки приложения

```env
APP_ENV=production
HTTP_PORT=8080
LOG_LEVEL=info
```

### База данных (управляемая PostgreSQL от Timeweb)

```env
DATABASE_DSN=postgresql://gen_user:Poison-79@c37e696087932476c61fd621.twc1.net:5432/default_db?sslmode=require
```

### API Форточки (4tochki)

```env
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

### Планировщики

```env
PRICES_CRON_SCHEDULE=0 */3 * * *
```

### Email уведомления

```env
EMAIL_ENABLED=true
EMAIL_SMTP_HOST=mail.hosting.reg.ru
EMAIL_SMTP_PORT=587
EMAIL_USERNAME=admin@etalon-shina.ru
EMAIL_PASSWORD=S69Y1ypojVLCZHO8
EMAIL_FROM=admin@etalon-shina.ru
EMAIL_NOTIFICATION_TO=v.boyarkin@etalon-shina.ru
```

### Timezone

```env
TZ=Europe/Moscow
```

---

## 🚀 Инструкция по деплою на Timeweb VM

### 1. Создайте VM на Timeweb Cloud

1. Откройте панель Timeweb Cloud
2. Перейдите в **Cloud Apps** → **Создать приложение**
3. Выберите способ деплоя: **Docker Compose**

### 2. Настройте репозиторий

- **Репозиторий:** ваш GitHub репозиторий
- **Ветка:** `master`
- **Файл docker-compose:** `docker-compose.yml` (по умолчанию)

### 3. Укажите порт

- **Порт:** `8080`

### 4. Добавьте переменные окружения

Скопируйте все переменные из раздела выше в поле "Переменные окружения" в панели Timeweb.

**⚠️ ВАЖНО:** Вставляйте по одной переменной за раз или весь блок целиком (зависит от интерфейса Timeweb).

### 5. Запустите деплой

Нажмите **"Создать"** и дождитесь завершения сборки.

---

## ✅ Проверка после деплоя

### Проверьте API

```bash
curl https://ваш-vm-url.timeweb.cloud/health
```

Ожидаемый ответ:
```json
{"status":"healthy"}
```

### Проверьте логи

В панели Timeweb перейдите в **Логи** для каждого сервиса:
- **api** - должен слушать порт 8080
- **nomenclature-scheduler** - должен показывать расписание (2:00 AM MSK)
- **prices-scheduler** - должен показывать расписание (каждые 3 часа)
- **prices-upload** - должен показывать расписание (10:00 AM IRKT)

---

## 🔍 Troubleshooting

### 1. Ошибка: "Cannot connect to database"

**Проверьте:**
- Переменная называется `DATABASE_DSN` (не `DB_DSN`)
- В строке подключения есть `?sslmode=require`
- Данные для подключения к БД правильные

### 2. API возвращает 502 Bad Gateway

**Проверьте:**
- Логи контейнера `api` в панели Timeweb
- Переменная `HTTP_PORT=8080` установлена
- В настройках указан порт 8080

### 3. Планировщики не запускаются

**Проверьте:**
- Логи соответствующего контейнера
- Переменная `PRICES_CRON_SCHEDULE` установлена (для prices-scheduler)
- Все переменные FOURTOCHKI_* установлены

### 4. Email уведомления не приходят

**Проверьте:**
- `EMAIL_ENABLED=true`
- Все переменные EMAIL_* установлены правильно
- SMTP хост и порт доступны с VM

---

## 📚 Дополнительная информация

### Структура сервисов

После успешного деплоя у вас будет работать **4 контейнера**:

1. **api** - REST API сервер (порт 8080)
   - Health check: `/health`
   - Использует `Dockerfile.production`

2. **nomenclature-scheduler** - планировщик синхронизации номенклатуры
   - Запускается ежедневно в 2:00 AM (Москва)
   - Использует `Dockerfile.nomenclature-scheduler`

3. **prices-scheduler** - планировщик синхронизации цен
   - Запускается каждые 3 часа (настраивается через `PRICES_CRON_SCHEDULE`)
   - Использует `Dockerfile.prices-scheduler`

4. **prices-upload** - выгрузка данных на 1C-Битрикс
   - Запускается ежедневно в 10:00 AM (Иркутск)
   - Использует `Dockerfile.prices-upload`

### Обновление приложения

Для обновления:
1. Запушьте изменения в репозиторий: `git push origin master`
2. В панели Timeweb нажмите **"Пересобрать"** или **"Rebuild"**

---

**Статус:** ✅ Готово к деплою
**Последнее обновление:** 18.03.2026
