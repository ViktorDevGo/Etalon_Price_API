# 🚀 Быстрый старт деплоя на Timeweb Cloud

## Проблема и решение

**Проблема:** Timeweb Cloud НЕ использует docker-compose для деплоя!

**Решение:** Используйте отдельные Dockerfile для каждого сервиса.

## 📦 Что было исправлено

1. ✅ Создан `Dockerfile.production` для основного API
2. ✅ Создан `Dockerfile.nomenclature-scheduler` для планировщика номенклатуры
3. ✅ Обновлен `Dockerfile.prices-scheduler` (убран Brinex)
4. ✅ Исправлена документация `docs/TIMEWEB_DEPLOYMENT.md`
5. ✅ Очищены устаревшие ссылки на sync-brinex

## 🎯 Пошаговая инструкция деплоя

### Шаг 1: Закоммитьте изменения

```bash
cd /Users/viktor/Pro_Koleso/Etalon_Price_API

git add .
git commit -m "Add production Dockerfiles for Timeweb Cloud deployment"
git push origin master
```

### Шаг 2: Создайте приложение #1 - API

1. Откройте https://timeweb.cloud/
2. Перейдите в **Cloud Apps** → **Создать приложение**
3. Настройки:
   - **Название:** `etalon-api`
   - **Способ деплоя:** Docker
   - **Репозиторий:** ваш Git репозиторий
   - **Ветка:** master
   - **Dockerfile:** `Dockerfile.production`
   - **Port:** 8080

4. **Переменные окружения** (скопируйте из .env):
   ```
   APP_ENV=production
   HTTP_PORT=8080
   LOG_LEVEL=info
   DATABASE_DSN=postgresql://gen_user:Poison-79@c37e696087932476c61fd621.twc1.net:5432/default_db?sslmode=require
   FOURTOCHKI_WSDL_URL=http://api-b2b.4tochki.ru/WCF/ClientService.svc?wsdl
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

5. Нажмите **Создать** и дождитесь деплоя

### Шаг 3: Примените миграции

```bash
# Из локальной машины
for file in migrations/*.up.sql; do
    echo "Applying: $file"
    psql "postgresql://gen_user:Poison-79@c37e696087932476c61fd621.twc1.net:5432/default_db?sslmode=require" -f "$file"
done
```

### Шаг 4: Проверьте API

```bash
# Получите URL приложения в панели Timeweb
curl https://ваш-домен.timeweb.cloud/health
```

Должен вернуть: `{"status":"healthy"}`

### Шаг 5: Создайте приложение #2 - Nomenclature Scheduler

1. **Cloud Apps** → **Создать приложение**
2. Настройки:
   - **Название:** `etalon-nomenclature-scheduler`
   - **Dockerfile:** `Dockerfile.nomenclature-scheduler`
   - **Port:** не требуется
   - **Переменные:** те же, что в API

### Шаг 6: Создайте приложение #3 - Prices Scheduler

1. **Cloud Apps** → **Создать приложение**
2. Настройки:
   - **Название:** `etalon-prices-scheduler`
   - **Dockerfile:** `Dockerfile.prices-scheduler`
   - **Port:** не требуется
   - **Переменные:** те же + добавить:
     ```
     PRICES_CRON_SCHEDULE=0 */3 * * *
     ```

### Шаг 7: Создайте приложение #4 - Prices Upload

1. **Cloud Apps** → **Создать приложение**
2. Настройки:
   - **Название:** `etalon-prices-upload`
   - **Dockerfile:** `Dockerfile.prices-upload`
   - **Port:** не требуется
   - **Переменные:** DATABASE_DSN + добавить:
     ```
     TZ=Asia/Irkutsk
     WORK_DIR=/app
     ```

## ✅ Проверка работы

### API
```bash
curl https://ваш-api.timeweb.cloud/health
```

### Логи планировщиков
```
Timeweb Cloud → Приложения → etalon-nomenclature-scheduler → Логи
```

Должны быть сообщения о запуске cron.

## 🔍 Troubleshooting

### Ошибка: "Docker compose file not found"
**Решение:** В настройках выберите "Docker", а НЕ "Docker Compose"

### Ошибка: "Cannot connect to database"
**Проверьте:**
1. Переменная называется `DATABASE_DSN`, а НЕ `DB_DSN`
2. В строке подключения есть `?sslmode=require`

### API возвращает 502
**Проверьте:**
1. Логи приложения в Timeweb
2. `HTTP_PORT=8080` установлен
3. Health check endpoint `/health` работает

## 📚 Подробная документация

Полная документация: `docs/TIMEWEB_DEPLOYMENT.md`

## 🎉 Готово!

После успешного деплоя у вас будет:
- ✅ API сервер (24/7)
- ✅ Автоматическая синхронизация номенклатуры (ежедневно в 2:00 MSK)
- ✅ Автоматическая синхронизация цен (каждые 3 часа)
- ✅ Автоматическая выгрузка на 1C-Битрикс (ежедневно в 10:00 IRKT)
