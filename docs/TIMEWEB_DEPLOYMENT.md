# Деплой в Timeweb Cloud

## 📋 Переменные окружения

### Формат для копирования в Timeweb Cloud:

```env
APP_ENV=production
HTTP_PORT=8080
LOG_LEVEL=info
DB_DSN=postgresql://gen_user:Poison-79@c37e696087932476c61fd621.twc1.net:5432/default_db?sslmode=require
FOURTOCHKI_WSDL_URL=http://api-b2b.4tochki.ru/WCF/ClientService.svc?wsdl
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

## 🚀 Инструкция по деплою

### 1. Подготовка репозитория

```bash
# Инициализация git (если еще не сделано)
git init
git add .
git commit -m "Initial commit for Timeweb Cloud deployment"

# Добавление удаленного репозитория
git remote add origin <your-git-repo-url>
git push -u origin master
```

### 2. Создание приложения в Timeweb Cloud

1. Войдите в панель Timeweb Cloud
2. Перейдите в раздел "Cloud Apps" или "Приложения"
3. Нажмите "Создать приложение"
4. Выберите:
   - **Фреймворк**: Go
   - **Версия Go**: 1.23 или новее
   - **Репозиторий**: подключите ваш Git репозиторий

### 3. Настройка сборки

В настройках приложения укажите:

**Build Command:**
```bash
go mod download && go build -o main cmd/api/main.go
```

**Start Command:**
```bash
./main
```

**Port:**
```
8080
```

### 4. Добавление переменных окружения

В разделе "Environment Variables" добавьте все переменные из списка выше.

**Важно:** Копируйте каждую переменную отдельно в формате:
```
Имя: APP_ENV
Значение: production
```

### 5. Настройка базы данных

База данных уже развернута на Timeweb Cloud:
- **Host:** c37e696087932476c61fd621.twc1.net
- **Port:** 5432
- **Database:** default_db
- **User:** gen_user
- **Password:** Poison-79

Переменная `DB_DSN` уже содержит правильную строку подключения.

### 6. Применение миграций

После первого деплоя выполните миграции вручную:

```bash
# Подключитесь к контейнеру приложения через SSH
# Или используйте локальный psql

psql "postgresql://gen_user:Poison-79@c37e696087932476c61fd621.twc1.net:5432/default_db?sslmode=require" -f migrations/001_initial_schema.up.sql

# Или все миграции по порядку
for file in migrations/*.up.sql; do
    psql "postgresql://gen_user:Poison-79@c37e696087932476c61fd621.twc1.net:5432/default_db?sslmode=require" -f "$file"
done
```

### 7. Настройка планировщиков (Cron Jobs)

#### Nomenclature Scheduler (ежедневно в 2:00 AM)

Создайте отдельное приложение или используйте Timeweb Cron:

**Команда:**
```bash
go run cmd/nomenclature-scheduler/main.go
```

**Расписание:**
```
0 2 * * *
```

#### Prices Scheduler (каждые 3 часа)

**Команда:**
```bash
go run cmd/sync-prices/main.go -type=all
```

**Расписание:**
```
0 */3 * * *
```

#### Prices Upload (ежедневно в 10:00 Иркутск)

**Команда:**
```bash
/app/scripts/upload_prices_to_server.sh
```

**Расписание:**
```
0 10 * * *
```

**Timezone:**
```
Asia/Irkutsk
```

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

### Проблема: Приложение не запускается

**Проверка 1: Логи**
```bash
# В панели Timeweb Cloud → Logs
# Или через CLI
timeweb-cloud logs <app-id>
```

**Проверка 2: Переменные окружения**
Убедитесь, что все переменные добавлены правильно.

**Проверка 3: База данных**
```bash
# Проверьте подключение
psql "postgresql://gen_user:Poison-79@c37e696087932476c61fd621.twc1.net:5432/default_db?sslmode=require" -c "SELECT 1;"
```

### Проблема: API не отвечает

**Проверка порта:**
- Убедитесь, что `HTTP_PORT=8080` установлен
- Проверьте, что приложение слушает на `0.0.0.0:8080`, а не на `localhost:8080`

### Проблема: Email не отправляются

**Проверка SMTP:**
```bash
# Проверьте настройки email
telnet mail.hosting.reg.ru 587
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

- [ ] Переменные окружения добавлены
- [ ] Build и Start команды настроены
- [ ] Порт 8080 открыт
- [ ] Миграции применены
- [ ] Планировщики настроены
- [ ] Health check работает
- [ ] Логи проверены
- [ ] Email уведомления работают
- [ ] Синхронизация с 4tochki работает
- [ ] Выгрузка на 1C-Битрикс работает

## 🔗 Полезные ссылки

- [Timeweb Cloud Docs](https://timeweb.cloud/docs)
- [Go on Timeweb Cloud](https://timeweb.cloud/docs/cloud-apps/go)
- [Environment Variables](https://timeweb.cloud/docs/cloud-apps/env)
- [Cron Jobs](https://timeweb.cloud/docs/cloud-apps/cron)
