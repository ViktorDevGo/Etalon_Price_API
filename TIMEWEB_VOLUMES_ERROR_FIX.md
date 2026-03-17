# 🚨 ОШИБКА: volumes is not allowed in docker-compose.yml

## ⚡ Быстрое решение

### Способ 1: Используйте отдельные Dockerfile (рекомендуется)

1. **В настройках Timeweb Cloud выберите "Docker"** (НЕ "Docker Compose")
2. **Создайте 4 отдельных приложения:**

   | Приложение | Dockerfile | Порт |
   |------------|-----------|------|
   | etalon-api | Dockerfile.production | 8080 |
   | etalon-nomenclature-scheduler | Dockerfile.nomenclature-scheduler | - |
   | etalon-prices-scheduler | Dockerfile.prices-scheduler | - |
   | etalon-prices-upload | Dockerfile.prices-upload | - |

3. **Для каждого приложения добавьте переменные окружения** (см. `.env`)

4. **Готово!** Следуйте инструкции: `TIMEWEB_QUICK_START.md`

---

### Способ 2: Используйте docker-compose БЕЗ volumes

1. **В настройках Timeweb Cloud:**
   - Способ деплоя: **"Docker Compose"**
   - Docker Compose файл: `docker-compose.timeweb.yml`

2. **Добавьте ВСЕ переменные окружения** в настройках приложения:
   ```
   DATABASE_DSN=postgresql://gen_user:password@host:5432/db?sslmode=require
   FOURTOCHKI_WSDL_URL=http://api-b2b.4tochki.ru/WCF/ClientService.svc?wsdl
   FOURTOCHKI_LOGIN=ваш_логин
   FOURTOCHKI_PASSWORD=ваш_пароль
   EMAIL_ENABLED=true
   EMAIL_SMTP_HOST=mail.hosting.reg.ru
   EMAIL_SMTP_PORT=587
   EMAIL_USERNAME=admin@etalon-shina.ru
   EMAIL_PASSWORD=ваш_пароль
   EMAIL_FROM=admin@etalon-shina.ru
   EMAIL_NOTIFICATION_TO=v.boyarkin@etalon-shina.ru
   ```

---

## 🔍 Что изменилось

| Файл | Что изменилось |
|------|----------------|
| `docker-compose.yml` | ❌ Содержит volumes (НЕ работает на Timeweb) |
| `docker-compose.timeweb.yml` | ✅ БЕЗ volumes (работает на Timeweb) |

---

## 📖 Дополнительная информация

- **Полное решение:** `РЕШЕНИЕ_ПРОБЛЕМЫ_TIMEWEB.md`
- **Быстрый старт:** `TIMEWEB_QUICK_START.md`
- **Подробная документация:** `docs/TIMEWEB_DEPLOYMENT.md`

---

**Статус:** ✅ ГОТОВО К ДЕПЛОЮ
**Рекомендация:** Используйте **Способ 1** (отдельные Dockerfile)
