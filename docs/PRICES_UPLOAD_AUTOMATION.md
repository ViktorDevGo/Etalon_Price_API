# Автоматическая выгрузка цен на сервер 1C-Битрикс

## Описание

Автоматическая система ежедневной выгрузки файла "Переоценка шин" на сервер 1C-Битрикс.

## 📋 Конфигурация

### Расписание
- **Время:** Каждый день в **10:00** по иркутскому времени (UTC+8)
- **Timezone:** Asia/Irkutsk

### Сервер назначения
- **Хост:** 147.45.215.76
- **Пользователь:** root
- **Путь:** /home/bitrix/www/upload/1c_catalog

### Файл
- **Формат:** CSV
- **Имя:** Переоценка_шины_YYYYMMDD.csv
- **Пример:** Переоценка_шины_20260317.csv

## 🚀 Запуск

### Вариант 1: Docker Compose (рекомендуется)

```bash
# Запуск сервиса выгрузки
docker-compose up -d prices-upload

# Проверка логов
docker-compose logs -f prices-upload

# Проверка расписания
docker-compose exec prices-upload cat /etc/crontabs/root
```

### Вариант 2: Ручной запуск скрипта

```bash
# Локальный запуск
./scripts/upload_prices_to_server.sh

# Или через make
make upload-prices
```

## 🔧 Процесс работы

### Автоматический процесс (ежедневно в 10:00):

1. **Генерация файла**
   - Запускается `cmd/export-bitrix-prices/main.go`
   - Выбираются товары с `isimport = 0` и `stock > 4`
   - Применяется логика расчета цен и сроков доставки
   - Файл сохраняется во временную директорию

2. **Загрузка на сервер**
   - Файл копируется на сервер через SCP
   - Используется автоматическая аутентификация через sshpass

3. **Обновление базы данных**
   - Всем выгруженным товарам проставляется `isimport = 1`
   - Товары с `isimport = 1` не будут включены в следующую выгрузку

4. **Очистка**
   - Временные файлы удаляются

## 📊 Логика выгрузки

### Фильтрация товаров
```sql
-- Условия включения в файл:
- MIN(isimport) = 0  -- хотя бы одна строка с isimport=0
- SUM(stock) > 4     -- суммарный остаток > 4
```

### Расчет цены
```
CV_PRICE_1 = CEIL(MIN(price) * 1.12)
-- Минимальная цена + 12% наценка, округление вверх
```

### Срок доставки (IP_PROP171)
- **"Доставим курьером Сегодня и позже"**
  Если есть хотя бы один товар от: Запаска, Форточки, Бигмашин

- **"Доставим курьером Завтра и позже"**
  Если все товары только от: Группа Бринекс, Северавто

### Колонки в файле (5 шт)
1. **IE_XML_ID** - CAE (артикул)
2. **IP_PROP171** - срок доставки
3. **QUANTITY** - суммарный остаток
4. **CV_PRICE_1** - цена с наценкой (целое число)
5. **CV_CURRENCY_1** - валюта (RUB)

## 🔍 Мониторинг

### Проверка статуса контейнера
```bash
docker-compose ps prices-upload
```

### Просмотр логов
```bash
# Все логи
docker-compose logs prices-upload

# Последние 100 строк
docker-compose logs --tail=100 prices-upload

# Real-time
docker-compose logs -f prices-upload
```

### Проверка времени следующего запуска
```bash
docker-compose exec prices-upload date
# Должно показать время по Asia/Irkutsk
```

### Проверка файла на сервере
```bash
ssh root@147.45.215.76 "ls -lh /home/bitrix/www/upload/1c_catalog/Переоценка_шины_*.csv"
```

## 🐛 Troubleshooting

### Проблема: Файл не загружается на сервер

**Проверка 1: Доступность сервера**
```bash
ping 147.45.215.76
```

**Проверка 2: SSH подключение**
```bash
sshpass -p 'k7MF4xi99Ty^^T' ssh root@147.45.215.76 "echo 'Connection OK'"
```

**Проверка 3: Права на директорию**
```bash
ssh root@147.45.215.76 "ls -ld /home/bitrix/www/upload/1c_catalog"
```

### Проблема: Неправильное время запуска

**Проверка timezone:**
```bash
docker-compose exec prices-upload cat /etc/timezone
# Должно быть: Asia/Irkutsk

docker-compose exec prices-upload date
# Должно показать правильное иркутское время
```

### Проблема: Пустой файл или мало товаров

**Проверка данных в БД:**
```sql
SELECT COUNT(*)
FROM tyres_prices_stock
WHERE isimport = 0 AND stock > 0;

-- Если товаров нет, значит все уже были выгружены
-- Нужно сбросить isimport обратно в 0
```

**Сброс isimport для повторной выгрузки:**
```sql
UPDATE tyres_prices_stock SET isimport = 0;
```

## ⚙️ Конфигурация

### Изменение времени запуска

Отредактируйте `Dockerfile.prices-upload`:
```dockerfile
# Изменить время в cron (формат: минута час день месяц день_недели)
RUN echo "0 10 * * * /app/upload_prices.sh >> /var/log/cron.log 2>&1" > /etc/crontabs/root
#         ^^
#         Время запуска (10:00)
```

После изменения:
```bash
docker-compose build prices-upload
docker-compose up -d prices-upload
```

### Изменение сервера назначения

Отредактируйте `scripts/upload_prices_to_server.sh`:
```bash
REMOTE_HOST="новый_хост"
REMOTE_USER="новый_пользователь"
REMOTE_PASSWORD="новый_пароль"
REMOTE_PATH="/новый/путь"
```

## 📝 Тестирование

### Тестовый запуск (без обновления isimport)
```bash
# Локально
go run cmd/export-bitrix-prices/main.go --limit=10

# Полный процесс (генерация + загрузка)
./scripts/upload_prices_to_server.sh
```

### Проверка содержимого файла
```bash
# На локальной машине
cat /tmp/bitrix_export/Переоценка_шины_*.csv | head -20

# На сервере
ssh root@147.45.215.76 "head -20 /home/bitrix/www/upload/1c_catalog/Переоценка_шины_*.csv"
```

## 🔐 Безопасность

⚠️ **ВАЖНО:**
- Пароль хранится в скрипте в открытом виде
- Для production рекомендуется использовать SSH ключи
- Файл `upload_prices_to_server.sh` должен иметь права 700
- Docker контейнер изолирован от хост-системы

### Переход на SSH ключи (рекомендуется)

1. Генерация ключа:
```bash
ssh-keygen -t rsa -b 4096 -f ~/.ssh/bitrix_upload
```

2. Копирование на сервер:
```bash
ssh-copy-id -i ~/.ssh/bitrix_upload.pub root@147.45.215.76
```

3. Обновить скрипт (убрать sshpass, добавить ключ):
```bash
scp -i ~/.ssh/bitrix_upload "$TEMP_DIR/$FILENAME" "$REMOTE_USER@$REMOTE_HOST:$REMOTE_PATH/"
```

## 📊 Статистика и отчеты

### Email уведомления

Для получения email уведомлений о результатах выгрузки, добавьте в конец скрипта:
```bash
# В scripts/upload_prices_to_server.sh
# После успешной выгрузки
echo "Выгрузка завершена: $FILENAME" | mail -s "Переоценка шин выгружена" admin@etalon-shina.ru
```

## 🔄 Обновление системы

```bash
# Получить последние изменения
git pull

# Пересобрать контейнер
docker-compose build prices-upload

# Перезапустить
docker-compose up -d prices-upload
```

## 📞 Поддержка

При возникновении проблем:
1. Проверьте логи контейнера
2. Проверьте доступность сервера
3. Проверьте количество товаров с isimport=0
4. Проверьте настройки timezone
