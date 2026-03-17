# Деплой в Timeweb Cloud

⚠️ **ВАЖНО:** Timeweb Cloud использует Docker для деплоя, НЕ docker-compose!

## 📋 Переменные окружения

### Формат для копирования в Timeweb Cloud:

```env
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

## 🚀 Инструкция по деплою

### 1. Подготовка репозитория

```bash
# Инициализация git (если еще не сделано)
git init
git add .
git commit -m "Add production Dockerfile for Timeweb Cloud"

# Добавление удаленного репозитория
git remote add origin <your-git-repo-url>
git push -u origin master
```

### 2. Создание приложения в Timeweb Cloud

1. Войдите в панель Timeweb Cloud: https://timeweb.cloud/
2. Перейдите в раздел **"Серверы и хостинг"** → **"Cloud Apps"**
3. Нажмите **"Создать приложение"**
4. Выберите:
   - **Способ деплоя**: Docker
   - **Репозиторий**: подключите ваш Git репозиторий (GitHub/GitLab)
   - **Ветка**: master (или main)

### 3. Настройка Docker деплоя

В настройках приложения укажите:

**Dockerfile:**
```
Dockerfile.production
```

**Port (внутренний порт приложения):**
```
8080
```

**Публичный порт** (будет назначен автоматически Timeweb)

### 4. Добавление переменных окружения

В разделе **"Переменные окружения"** (Environment Variables) добавьте все переменные из списка выше.

**Важно:**
- Копируйте каждую переменную отдельно
- НЕ используйте символ `=` в имени переменной
- Формат: `Имя → Значение`

Пример:
```
Имя: DATABASE_DSN
Значение: postgresql://gen_user:Poison-79@c37e696087932476c61fd621.twc1.net:5432/default_db?sslmode=require
```

### 5. Настройка базы данных

База данных уже развернута на Timeweb Cloud:
- **Host:** c37e696087932476c61fd621.twc1.net
- **Port:** 5432
- **Database:** default_db
- **User:** gen_user
- **Password:** Poison-79
- **SSL Mode:** require (обязательно!)

⚠️ **ВАЖНО:** Используйте переменную `DATABASE_DSN` (не `DB_DSN`!)

### 6. Применение миграций

После первого деплоя выполните миграции **локально** через psql:

```bash
# Из вашей локальной машины
cd /Users/viktor/Pro_Koleso/Etalon_Price_API

# Выполните все миграции по порядку
for file in migrations/*.up.sql; do
    echo "Applying: $file"
    psql "postgresql://gen_user:Poison-79@c37e696087932476c61fd621.twc1.net:5432/default_db?sslmode=require" -f "$file"
done
```

Или вручную через DBeaver/pgAdmin подключитесь к БД и выполните миграции.

### 7. Настройка планировщиков (отдельные приложения)

⚠️ **ВАЖНО:** Каждый планировщик = отдельное Cloud App на Timeweb!

#### 7.1. Nomenclature Scheduler (ежедневная синхронизация номенклатуры)

Создайте **второе приложение** на Timeweb:

1. Название: `etalon-nomenclature-scheduler`
2. Dockerfile: `Dockerfile.production` (тот же)
3. Command override: `./nomenclature-scheduler`
4. Переменные окружения: те же самые
5. Port: не требуется (фоновый процесс)

**Альтернатива:** Используйте Timeweb Cron Jobs (если доступно в вашем плане):
```bash
Команда: curl -X POST https://ваш-домен.com/api/sync/nomenclature
Расписание: 0 2 * * *
```

#### 7.2. Prices Scheduler (синхронизация цен каждые 3 часа)

Создайте **третье приложение** на Timeweb:

1. Название: `etalon-prices-scheduler`
2. Dockerfile: `Dockerfile.prices-scheduler`
3. Переменные:
   - `PRICES_CRON_SCHEDULE=0 */3 * * *`
   - Все остальные как в основном приложении
4. Port: не требуется

#### 7.3. Prices Upload (выгрузка на 1C-Битрикс ежедневно в 10:00 Иркутск)

Создайте **четвертое приложение** на Timeweb:

1. Название: `etalon-prices-upload`
2. Dockerfile: `Dockerfile.prices-upload`
3. Переменные:
   - `TZ=Asia/Irkutsk`
   - `DATABASE_DSN=...` (как в основном)
4. Port: не требуется

**Итого:** 4 приложения на Timeweb:
- ✅ `etalon-api` - основное API (порт 8080)
- ✅ `etalon-nomenclature-scheduler` - синхронизация номенклатуры (2:00 MSK)
- ✅ `etalon-prices-scheduler` - синхронизация цен (каждые 3 часа)
- ✅ `etalon-prices-upload` - выгрузка на Битрикс (10:00 IRKT)

### 8. Настройка логирования

В настройках приложения:
- **Log Level:** info (для production)
- **Log Retention:** 7 дней (или по необходимости)

### 9. Health Check

Настройте health check endpoint:
```
HTTP GET /health
Port: 8080
Interval: 30s
Timeout: 10s
```

### 10. Масштабирование

Рекомендуемые настройки для production:
- **CPU:** 1-2 vCPU
- **RAM:** 512 MB - 1 GB
- **Instances:** 1-2 (для высокой доступности)

## 🔧 Troubleshooting

### Проблема: "Error: Docker compose file not found"

**Причина:** Timeweb ищет docker-compose.yml вместо Dockerfile

**Решение:**
1. В настройках приложения выберите **"Docker"**, НЕ "Docker Compose"
2. Укажите `Dockerfile.production` в поле "Dockerfile path"

### Проблема: Приложение не запускается

**Проверка 1: Логи в Timeweb**
```
Панель Timeweb Cloud → Ваше приложение → Логи
```

Смотрите на ошибки при запуске:
- `connection refused` → проблема с БД
- `bind: address already in use` → порт занят
- `no such file` → не хватает файлов в образе

**Проверка 2: Переменные окружения**
Убедитесь, что:
- ✅ `DATABASE_DSN` (НЕ `DB_DSN`!)
- ✅ Все пароли без спецсимволов в имени переменной
- ✅ Нет лишних пробелов в начале/конце значений

**Проверка 3: База данных**
```bash
# Локально проверьте подключение
psql "postgresql://gen_user:Poison-79@c37e696087932476c61fd621.twc1.net:5432/default_db?sslmode=require" -c "SELECT 1;"
```

### Проблема: API не отвечает (502 Bad Gateway)

**Возможные причины:**

1. **Приложение не слушает на правильном порту**
   - Проверьте: `HTTP_PORT=8080` в переменных
   - Проверьте: в Dockerfile.production порт 8080 открыт

2. **Health check падает**
   - Добавьте эндпоинт `/health` в API
   - Или отключите health check в Timeweb

3. **Приложение крашится при старте**
   - Смотрите логи в панели Timeweb
   - Проверьте подключение к БД

### Проблема: "Cannot connect to database"

**Проверьте:**
```bash
# 1. Правильная переменная?
echo $DATABASE_DSN

# 2. БД доступна?
psql "postgresql://gen_user:Poison-79@c37e696087932476c61fd621.twc1.net:5432/default_db?sslmode=require" -c "\dt"

# 3. Миграции применены?
psql "postgresql://gen_user:Poison-79@c37e696087932476c61fd621.twc1.net:5432/default_db?sslmode=require" -c "SELECT * FROM schema_migrations;"
```

### Проблема: Email не отправляются

**Проверка SMTP:**
```bash
# Проверьте доступность SMTP сервера
telnet mail.hosting.reg.ru 587

# Проверьте переменные
echo $EMAIL_SMTP_HOST
echo $EMAIL_SMTP_PORT
echo $EMAIL_USERNAME
```

## 📊 Мониторинг

### Метрики для отслеживания:
- CPU usage
- Memory usage
- HTTP response time
- Database connections
- Email send rate
- Error rate

### Алерты:
- CPU > 80%
- Memory > 90%
- HTTP errors > 5%
- Database connection errors

## 🔐 Безопасность

### Рекомендации:
1. ✅ Используйте HTTPS для всех API endpoints
2. ✅ Регулярно обновляйте зависимости
3. ✅ Ротация паролей раз в 3 месяца
4. ✅ Мониторинг логов на подозрительную активность
5. ✅ Бэкапы базы данных (ежедневно)

## 📝 Checklist деплоя

### Подготовка
- [ ] Dockerfile.production создан
- [ ] Git репозиторий подключен к Timeweb
- [ ] Последний код залит в master/main

### Приложение 1: API (etalon-api)
- [ ] Создано на Timeweb Cloud
- [ ] Способ деплоя: Docker
- [ ] Dockerfile: `Dockerfile.production`
- [ ] Port: 8080
- [ ] Переменные окружения добавлены (DATABASE_DSN!)
- [ ] Приложение успешно задеплоено
- [ ] Health check работает: `https://ваш-домен.com/health`
- [ ] Логи без ошибок

### База данных
- [ ] Миграции применены локально
- [ ] Таблицы созданы (проверить через psql)
- [ ] Тестовое подключение работает

### Приложение 2: Nomenclature Scheduler
- [ ] Создано на Timeweb
- [ ] Dockerfile: `Dockerfile.production`
- [ ] Command override: `./nomenclature-scheduler`
- [ ] Переменные окружения добавлены
- [ ] Запускается без ошибок (проверить логи)

### Приложение 3: Prices Scheduler
- [ ] Создано на Timeweb
- [ ] Dockerfile: `Dockerfile.prices-scheduler`
- [ ] Переменная: `PRICES_CRON_SCHEDULE=0 */3 * * *`
- [ ] Запускается без ошибок

### Приложение 4: Prices Upload
- [ ] Создано на Timeweb
- [ ] Dockerfile: `Dockerfile.prices-upload`
- [ ] Переменная: `TZ=Asia/Irkutsk`
- [ ] Скрипт `/app/upload_prices.sh` работает

### Тестирование
- [ ] API отвечает на `/health`
- [ ] Email уведомления приходят
- [ ] Синхронизация с 4tochki работает
- [ ] Выгрузка на 1C-Битрикс работает
- [ ] Cron jobs запускаются по расписанию

### Мониторинг
- [ ] Настроены алерты на ошибки
- [ ] Проверяются логи ежедневно
- [ ] CPU/Memory в пределах нормы

## 🔗 Полезные ссылки

- [Timeweb Cloud Docs](https://timeweb.cloud/docs)
- [Go on Timeweb Cloud](https://timeweb.cloud/docs/cloud-apps/go)
- [Environment Variables](https://timeweb.cloud/docs/cloud-apps/env)
- [Cron Jobs](https://timeweb.cloud/docs/cloud-apps/cron)
