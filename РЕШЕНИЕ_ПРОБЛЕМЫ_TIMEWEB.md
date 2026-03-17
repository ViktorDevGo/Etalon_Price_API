# ✅ Проблема с деплоем на Timeweb Cloud - РЕШЕНА

## 🔴 Проблема

При деплое на Timeweb Cloud появлялась ошибка:
```
ERROR | Sanitizer check error volumes is not allowed in docker-compose.yml
```

И предупреждения о пустых переменных:
```
WARNING | The "DATABASE_DSN" variable is not set. Defaulting to a blank string.
WARNING | The "FOURTOCHKI_WSDL_URL" variable is not set. Defaulting to a blank string.
```

## 🔍 Причина

**Проблема #1: Volumes запрещены**
- Timeweb Cloud НЕ поддерживает `volumes` в docker-compose.yml
- Ваш `docker-compose.yml` содержал volumes для PostgreSQL

**Проблема #2: Локальный PostgreSQL не нужен**
- Timeweb предоставляет управляемую БД
- Локальный контейнер postgres из docker-compose не нужен

**Проблема #3: Переменные окружения**
- Вы выбрали "Docker Compose" в настройках, но не указали переменные окружения

## ✅ Что было исправлено

### 1. Созданы специализированные Dockerfile для Timeweb:

- ✅ `Dockerfile.production` - основной API сервер
- ✅ `Dockerfile.nomenclature-scheduler` - планировщик номенклатуры
- ✅ `Dockerfile.prices-scheduler` - планировщик цен (обновлен, убран Brinex)
- ✅ `Dockerfile.prices-upload` - выгрузка на 1C-Битрикс (без изменений)

### 2. Создан docker-compose БЕЗ volumes (опционально):

- ✅ `docker-compose.timeweb.yml` - БЕЗ volumes, БЕЗ postgres контейнера
- ✅ Использует управляемую БД Timeweb
- ✅ Переменная `DATABASE_DSN` вместо `DB_DSN`

### 3. Обновлена документация:

- ✅ `docs/TIMEWEB_DEPLOYMENT.md` - полная инструкция с troubleshooting
- ✅ `TIMEWEB_QUICK_START.md` - краткая инструкция для быстрого старта
- ✅ `РЕШЕНИЕ_ПРОБЛЕМЫ_TIMEWEB.md` - решение ошибки volumes

### 4. Очищены устаревшие зависимости:

- ✅ Удалены ссылки на `sync-brinex` (удален в миграции 016)
- ✅ Обновлены все Dockerfile

## 🚀 Что делать дальше

### ⭐ Вариант 1: Docker (рекомендуется)

**Деплойте 4 отдельных приложения**, используя отдельные Dockerfile:

1. Следуйте инструкции: `TIMEWEB_QUICK_START.md`
2. В настройках Timeweb выбирайте **"Docker"** (НЕ "Docker Compose")
3. Для каждого приложения указывайте свой Dockerfile:
   - `etalon-api` → `Dockerfile.production`
   - `etalon-nomenclature-scheduler` → `Dockerfile.nomenclature-scheduler`
   - `etalon-prices-scheduler` → `Dockerfile.prices-scheduler`
   - `etalon-prices-upload` → `Dockerfile.prices-upload`

**Плюсы:**
- ✅ Каждый сервис деплоится независимо
- ✅ Легче управлять и масштабировать
- ✅ Рекомендуется Timeweb

---

### Вариант 2: Docker Compose (альтернатива)

**Деплойте через docker-compose БЕЗ volumes:**

1. В настройках Timeweb выберите **"Docker Compose"**
2. Укажите файл: `docker-compose.timeweb.yml`
3. Добавьте ВСЕ переменные окружения в настройках Timeweb

**Минусы:**
- ❌ Все сервисы в одном приложении (сложнее управлять)
- ❌ Нужно вручную добавлять переменные в настройках

---

### Вариант 3: Подробная документация

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

### Типичные проблемы:

**1. `volumes is not allowed in docker-compose.yml`**
- **Решение А:** Используйте `docker-compose.timeweb.yml` вместо `docker-compose.yml`
- **Решение Б:** Выберите "Docker" вместо "Docker Compose" (рекомендуется)

**2. `The "DATABASE_DSN" variable is not set`**
- **Причина:** Переменные окружения не добавлены в настройках приложения
- **Решение:** Добавьте все переменные из `.env` в панели Timeweb

**3. `Docker compose file not found`**
- **Решение:** Выберите "Docker", а не "Docker Compose"

**4. `Cannot connect to database`**
- **Проверьте:** Переменная называется `DATABASE_DSN`, НЕ `DB_DSN`
- **Проверьте:** В строке есть `?sslmode=require`

**5. `502 Bad Gateway`**
- **Проверьте:** Логи приложения в панели Timeweb
- **Проверьте:** `HTTP_PORT=8080` установлен
- **Проверьте:** Health check endpoint `/health` работает

## 🎉 Результат

После успешного деплоя у вас будет:
- ✅ API сервер работает 24/7
- ✅ Автосинхронизация номенклатуры (2:00 MSK ежедневно)
- ✅ Автосинхронизация цен (каждые 3 часа)
- ✅ Автовыгрузка на 1C-Битрикс (10:00 IRKT ежедневно)

---

**Статус:** ✅ ГОТОВО К ДЕПЛОЮ

**Дата исправления:** 18.03.2026
