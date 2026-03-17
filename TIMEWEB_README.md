# 🚀 Деплой на Timeweb Cloud

## 🚨 Если получаете ошибку: `volumes is not allowed`

**Быстрое решение:** `TIMEWEB_VOLUMES_ERROR_FIX.md`

---

## ⚡ Быстрый старт

### 1️⃣ Выберите способ деплоя

| Способ | Сложность | Рекомендация |
|--------|-----------|--------------|
| **Docker** (4 приложения) | Средняя | ⭐⭐⭐⭐⭐ Рекомендуется |
| **Docker Compose** | Простая | ⭐⭐⭐ Альтернатива |

---

### 2️⃣ Docker (рекомендуется)

**Следуйте инструкции:** `TIMEWEB_QUICK_START.md`

**Кратко:**
1. В Timeweb выберите **"Docker"** (НЕ "Docker Compose")
2. Создайте 4 приложения:
   - `etalon-api` → `Dockerfile.production` → порт 8080
   - `etalon-nomenclature-scheduler` → `Dockerfile.nomenclature-scheduler`
   - `etalon-prices-scheduler` → `Dockerfile.prices-scheduler`
   - `etalon-prices-upload` → `Dockerfile.prices-upload`
3. Добавьте переменные из `.env.example`

---

### 3️⃣ Docker Compose (альтернатива)

**Файл:** `docker-compose.timeweb.yml`

**Кратко:**
1. В Timeweb выберите **"Docker Compose"**
2. Укажите файл: `docker-compose.timeweb.yml`
3. Добавьте переменные окружения:
   ```bash
   DATABASE_DSN=postgresql://...
   FOURTOCHKI_WSDL_URL=...
   FOURTOCHKI_LOGIN=...
   FOURTOCHKI_PASSWORD=...
   EMAIL_ENABLED=true
   EMAIL_SMTP_HOST=...
   (и остальные из .env.example)
   ```

---

## 📋 Переменные окружения

**ВАЖНО:** Для Timeweb используйте `DATABASE_DSN`, НЕ `DB_DSN`!

Пример:
```bash
DATABASE_DSN=postgresql://gen_user:password@c37e696087932476c61fd621.twc1.net:5432/default_db?sslmode=require
```

**Полный список:** `.env.example`

---

## ✅ Проверка

После деплоя:

```bash
# API Health Check
curl https://ваш-домен.timeweb.cloud/health

# Должен вернуть:
{"status":"healthy"}
```

---

## 🔍 Troubleshooting

### Ошибка: `volumes is not allowed`
→ См. `TIMEWEB_VOLUMES_ERROR_FIX.md`

### Ошибка: `DATABASE_DSN variable is not set`
→ Добавьте переменные окружения в настройках Timeweb

### Ошибка: `Cannot connect to database`
→ Проверьте, что используете `DATABASE_DSN` (НЕ `DB_DSN`)
→ Проверьте `?sslmode=require` в строке подключения

### API возвращает 502
→ Проверьте логи в панели Timeweb
→ Убедитесь, что `HTTP_PORT=8080`

---

## 📚 Полная документация

- **Быстрый старт:** `TIMEWEB_QUICK_START.md`
- **Решение проблем:** `РЕШЕНИЕ_ПРОБЛЕМЫ_TIMEWEB.md`
- **Подробная документация:** `docs/TIMEWEB_DEPLOYMENT.md`
- **Ошибка volumes:** `TIMEWEB_VOLUMES_ERROR_FIX.md`

---

## 🎯 Что будет после деплоя

- ✅ API сервер (24/7, порт 8080)
- ✅ Синхронизация номенклатуры (ежедневно в 2:00 MSK)
- ✅ Синхронизация цен (каждые 3 часа)
- ✅ Выгрузка на 1C-Битрикс (ежедневно в 10:00 IRKT)

---

**Статус проекта:** ✅ ГОТОВО К ДЕПЛОЮ
**Последнее обновление:** 18.03.2026
