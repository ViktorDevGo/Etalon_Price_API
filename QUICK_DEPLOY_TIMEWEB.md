# Быстрый деплой на Timeweb Cloud

## Шаг 1: Создайте 5 Cloud Apps

| Название | Dockerfile | Порт | Описание |
|----------|-----------|------|----------|
| etalon-api | Dockerfile.production | 8080 | API сервер |
| etalon-nomenclature-scheduler | Dockerfile.nomenclature-scheduler | - | Синхронизация номенклатуры (2:00 MSK) |
| etalon-prices-scheduler | Dockerfile.prices-scheduler | - | Синхронизация цен (каждые 3 часа) |
| etalon-prices-upload | Dockerfile.prices-upload | - | Выгрузка на Битрикс (10:00 IRKT) |
| etalon-severavto-scheduler | Dockerfile.severavto-scheduler | - | Синхронизация Северавто (каждые 3 часа) |

## Шаг 2: Скопируйте переменные окружения

Для каждого сервиса скопируйте переменные из файла `TIMEWEB_ENV_VARIABLES.md` (соответствующий раздел).

### Основные переменные (для ВСЕХ сервисов):

```bash
DATABASE_DSN=postgresql://gen_user:Poison-79@c37e696087932476c61fd621.twc1.net:5432/default_db?sslmode=require
APP_ENV=production
LOG_LEVEL=info
TZ=Europe/Moscow
```

### Email (для сервисов с уведомлениями):

```bash
EMAIL_ENABLED=true
EMAIL_SMTP_HOST=mail.hosting.reg.ru
EMAIL_SMTP_PORT=587
EMAIL_USERNAME=admin@etalon-shina.ru
EMAIL_PASSWORD=S69Y1ypojVLCZHO8
EMAIL_FROM=admin@etalon-shina.ru
EMAIL_NOTIFICATION_TO=v.boyarkin@etalon-shina.ru
```

### Форточки (для API, nomenclature, prices):

```bash
FOURTOCHKI_WSDL_URL=http://api-b2b.4tochki.ru/WCF/ClientService.svc?wsdl
FOURTOCHKI_BASE_URL=http://api-b2b.4tochki.ru/WCF/ClientService.svc?wsdl
FOURTOCHKI_LOGIN=sa69263
FOURTOCHKI_PASSWORD=jkHP4Nj3)z
```

### Северавто (для severavto-scheduler):

```bash
SEVERAVTO_BASE_URL=http://webmim.svrauto.ru
SEVERAVTO_API_KEY=TpU0K90z0X0LeNWyWiBdWS0kP3mvdzOf
SEVERAVTO_TIMEOUT=120s
```

## Шаг 3: Запустите деплой

1. Подключите GitHub репозиторий к каждому Cloud App
2. Выберите нужный Dockerfile
3. Нажмите "Deploy"

## Шаг 4: Проверьте работу

### API сервер:
```bash
curl https://ваш-домен.twc1.net/health
```

### Логи планировщиков:
Откройте "Логи" в панели Timeweb для каждого сервиса.

Должны видеть:
```
✅ "Starting ... scheduler"
✅ "Next sync at ..."
```

## Готово! 🚀

Сервисы работают автоматически:
- **Форточки номенклатура:** каждый день в 2:00 MSK
- **Форточки цены:** каждые 3 часа
- **Битрикс выгрузка:** каждый день в 10:00 IRKT
- **Северавто:** каждые 3 часа (шины :00, диски :15)

Email уведомления приходят после каждой синхронизации.

---

**Подробности:** см. файл `TIMEWEB_ENV_VARIABLES.md`
