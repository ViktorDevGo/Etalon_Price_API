# 📦 Docker Compose файлы

## Два файла, две цели

| Файл | Назначение | Volumes | PostgreSQL |
|------|------------|---------|-----------|
| `docker-compose.yml` | ☁️ **Timeweb Cloud** | ❌ Нет | ❌ Управляемая БД |
| `docker-compose.local.yml` | 💻 **Локальная разработка** | ✅ Да | ✅ Локальный контейнер |

---

## ☁️ docker-compose.yml (для Timeweb Cloud)

**Используется:** Автоматически при деплое на Timeweb Cloud Apps

**Особенности:**
- ❌ БЕЗ volumes (Timeweb не поддерживает)
- ❌ БЕЗ локального PostgreSQL контейнера
- ✅ Использует управляемую БД Timeweb
- ✅ Переменная `DATABASE_DSN` для подключения

**Сервисы:**
- `api` - основной API сервер (порт 8080)
- `nomenclature-scheduler` - синхронизация номенклатуры (2:00 MSK)
- `prices-scheduler` - синхронизация цен (каждые 3 часа)
- `prices-upload` - выгрузка на 1C-Битрикс (10:00 IRKT)

---

## 💻 docker-compose.local.yml (для локальной разработки)

**Используется:** При локальной разработке

**Запуск:**
```bash
docker-compose -f docker-compose.local.yml up -d
```

**Особенности:**
- ✅ С volumes для PostgreSQL данных
- ✅ С локальным PostgreSQL контейнером
- ✅ Сеть между контейнерами
- ✅ Переменная `DB_DSN` для подключения к локальной БД

**Сервисы:**
- Все те же + `postgres` контейнер

---

## 🚀 Деплой на Timeweb Cloud

### Способ 1: Docker Compose (используется сейчас)

1. Timeweb автоматически использует `docker-compose.yml`
2. Добавьте переменные окружения в панели Timeweb
3. Redeploy

**Переменные окружения:** См. файл `TIMEWEB_ENV_COPY_PASTE.txt` на Desktop

---

### Способ 2: Отдельные Dockerfile (рекомендуется)

Для лучшей управляемости создайте 4 отдельных приложения:

- `etalon-api` → `Dockerfile.production`
- `etalon-nomenclature-scheduler` → `Dockerfile.nomenclature-scheduler`
- `etalon-prices-scheduler` → `Dockerfile.prices-scheduler`
- `etalon-prices-upload` → `Dockerfile.prices-upload`

**Инструкция:** `TIMEWEB_QUICK_START.md`

---

## 🔧 Локальная разработка

```bash
# Запуск всех сервисов
docker-compose -f docker-compose.local.yml up -d

# Остановка
docker-compose -f docker-compose.local.yml down

# Просмотр логов
docker-compose -f docker-compose.local.yml logs -f

# Остановка с удалением volumes
docker-compose -f docker-compose.local.yml down -v
```

---

## ⚠️ Важно

**НЕ используйте `docker-compose.local.yml` на Timeweb!**
- Он содержит volumes (запрещено)
- Он создаёт локальный PostgreSQL (не нужен)

**НЕ используйте `docker-compose.yml` локально!**
- Он не содержит PostgreSQL контейнер
- Он рассчитан на внешнюю БД

---

## 📝 История изменений

**18.03.2026:**
- Разделены файлы для Timeweb и локальной разработки
- `docker-compose.yml` → БЕЗ volumes (для Timeweb)
- `docker-compose.local.yml` → С volumes (для локальной разработки)

---

**Вопросы?** См. `TIMEWEB_README.md` или `РЕШЕНИЕ_ПРОБЛЕМЫ_TIMEWEB.md`
