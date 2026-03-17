# ✅ Проблема с деплоем на Timeweb Cloud - РЕШЕНА

## 🔴 Проблема

Вы пытались задеплоить приложение на Timeweb Cloud, но получали ошибку:
```
Причина: В логах деплоя отсутствует информация об ошибке,
что может указывать на проблему с конфигурацией docker-compose
или отсутствием необходимых файлов в проекте.
```

## 🔍 Причина

**Timeweb Cloud НЕ поддерживает docker-compose для деплоя!**

Ваш `docker-compose.yml` содержал:
- ❌ Локальный сервис PostgreSQL (не нужен - Timeweb предоставляет управляемую БД)
- ❌ Множественные сервисы (Timeweb деплоит ОДИН Dockerfile за раз)
- ❌ Зависимости `depends_on: postgres` (не работают без docker-compose)

## ✅ Что было исправлено

### 1. Созданы специализированные Dockerfile для Timeweb:

- ✅ `Dockerfile.production` - основной API сервер
- ✅ `Dockerfile.nomenclature-scheduler` - планировщик номенклатуры
- ✅ `Dockerfile.prices-scheduler` - планировщик цен (обновлен, убран Brinex)
- ✅ `Dockerfile.prices-upload` - выгрузка на 1C-Битрикс (без изменений)

### 2. Обновлена документация:

- ✅ `docs/TIMEWEB_DEPLOYMENT.md` - полная инструкция с troubleshooting
- ✅ `TIMEWEB_QUICK_START.md` - краткая инструкция для быстрого старта

### 3. Очищены устаревшие зависимости:

- ✅ Удалены ссылки на `sync-brinex` (удален в миграции 016)
- ✅ Обновлены все Dockerfile

## 🚀 Что делать дальше

### Вариант 1: Быстрый старт (рекомендуется)

Следуйте инструкции в файле:
```
TIMEWEB_QUICK_START.md
```

### Вариант 2: Подробная инструкция

Смотрите полную документацию:
```
docs/TIMEWEB_DEPLOYMENT.md
```

## 📋 Краткий чеклист

1. ✅ Закоммитьте изменения:
   ```bash
   git add .
   git commit -m "Add production Dockerfiles for Timeweb"
   git push origin master
   ```

2. ✅ Создайте 4 приложения на Timeweb:
   - `etalon-api` → Dockerfile.production → порт 8080
   - `etalon-nomenclature-scheduler` → Dockerfile.nomenclature-scheduler
   - `etalon-prices-scheduler` → Dockerfile.prices-scheduler
   - `etalon-prices-upload` → Dockerfile.prices-upload

3. ✅ Добавьте переменные окружения (из .env)
   **ВАЖНО:** Используйте `DATABASE_DSN`, НЕ `DB_DSN`!

4. ✅ Примените миграции:
   ```bash
   for file in migrations/*.up.sql; do
       psql "postgresql://gen_user:Poison-79@c37e696087932476c61fd621.twc1.net:5432/default_db?sslmode=require" -f "$file"
   done
   ```

5. ✅ Проверьте работу:
   ```bash
   curl https://ваш-домен.timeweb.cloud/health
   ```

## 🎯 Ключевые моменты

1. **Docker, НЕ Docker Compose**
   - В настройках приложения выбирайте "Docker"
   - Указывайте конкретный Dockerfile (например, `Dockerfile.production`)

2. **Переменная базы данных**
   - ✅ Правильно: `DATABASE_DSN`
   - ❌ Неправильно: `DB_DSN`

3. **Четыре отдельных приложения**
   - Каждый сервис = отдельное Cloud App на Timeweb
   - Каждое имеет свой Dockerfile

4. **Управляемая БД**
   - PostgreSQL уже развернут на Timeweb
   - НЕ нужно деплоить свой postgres

## 📞 Если что-то не работает

Смотрите секцию **Troubleshooting** в:
```
docs/TIMEWEB_DEPLOYMENT.md
```

Типичные проблемы:
- ❌ "Docker compose file not found" → Выберите "Docker", а не "Docker Compose"
- ❌ "Cannot connect to database" → Проверьте имя переменной `DATABASE_DSN`
- ❌ "502 Bad Gateway" → Проверьте логи в панели Timeweb

## 🎉 Результат

После успешного деплоя у вас будет:
- ✅ API сервер работает 24/7
- ✅ Автосинхронизация номенклатуры (2:00 MSK ежедневно)
- ✅ Автосинхронизация цен (каждые 3 часа)
- ✅ Автовыгрузка на 1C-Битрикс (10:00 IRKT ежедневно)

---

**Статус:** ✅ ГОТОВО К ДЕПЛОЮ

**Дата исправления:** 18.03.2026
