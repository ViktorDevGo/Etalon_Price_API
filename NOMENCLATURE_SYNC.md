# Синхронизация номенклатуры 4tochki

## Описание

Система ежедневной загрузки номенклатуры (каталога товаров) из XML файлов 4tochki.

## Источники данных

- **Шины:** https://b2b.4tochki.ru/export_data/Tires.xml
- **Диски:** https://b2b.4tochki.ru/export_data/Rims.xml

Файлы обновляются ежедневно в ~05:00 по Москве.

## Таблицы в БД

### nomenclature_tyres
Полный каталог шин (60,637 записей):
- `cae` - код товара (уникальный)
- `name` - полное название
- `width`, `height`, `diameter` - размеры
- `brand`, `model` - бренд и модель
- `season` - сезон (Зимняя, Летняя, Всесезонная)
- `is_studded` - наличие шипов (шип / не шип)
- `tiretype` - тип шины (Легковая, и т.д.)
- `load_index`, `speed_index` - индексы нагрузки и скорости
- `runflat`, `reinforced` - специальные характеристики

### nomenclature_rims
Полный каталог дисков (99,995 записей):
- `cae` - код товара (уникальный)
- `name` - полное название
- `width`, `diameter` - ширина и диаметр
- `brand`, `model` - бренд и модель
- `bolts_count`, `bolts_spacing` - разболтовка
- `et` - вылет
- `dia` - центральное отверстие
- `color`, `rim_type` - цвет и тип

### nomenclature_sync_runs
История синхронизаций:
- Время старта и завершения
- Количество новых и обновленных записей
- Статус и ошибки

## Запуск синхронизации

### Вручную

```bash
# Синхронизация всего (шины + диски)
go run cmd/sync-nomenclature/main.go -type=all

# Только шины
go run cmd/sync-nomenclature/main.go -type=tyres

# Только диски
go run cmd/sync-nomenclature/main.go -type=rims
```

### Через скрипт

```bash
chmod +x sync_nomenclature.sh
./sync_nomenclature.sh
```

### Автоматически (cron)

Добавить в crontab для ежедневного запуска в 06:00:

```bash
crontab -e
```

```cron
# Ежедневная синхронизация номенклатуры 4tochki в 06:00
0 6 * * * cd /path/to/Etalon_Price_API && ./sync_nomenclature.sh >> logs/nomenclature_sync.log 2>&1
```

## Производительность

- **Загрузка XML:** ~4-5 секунд на файл
- **Парсинг XML:** ~1 секунда
- **Сохранение в БД:** ~1 секунда на 1000 записей
- **Общее время:** ~3 минуты для полной синхронизации

Оптимизация:
- Использование PostgreSQL COPY для массовой загрузки
- Bulk upsert через временные таблицы
- ON CONFLICT DO UPDATE для обновления существующих записей

## Проверка данных

```bash
# Проверить количество записей
go run cmd/check-nomenclature/main.go
```

## Структура файлов

```
cmd/
  └── sync-nomenclature/    - CLI для синхронизации
  └── check-nomenclature/   - Проверка данных в БД

internal/
  ├── domain/
  │   └── nomenclature.go   - Модели данных
  ├── providers/4tochki/
  │   └── xml_parser.go     - Парсинг XML файлов
  ├── repository/
  │   └── nomenclature_repository.go - Работа с БД
  └── service/
      └── nomenclature_service.go    - Бизнес-логика

migrations/
  └── 006_create_nomenclature_tables.up.sql - Миграция таблиц

sync_nomenclature.sh       - Скрипт для cron
NOMENCLATURE_SYNC.md       - Эта документация
```

## Примеры использования

### Поиск шин по размеру
```sql
SELECT cae, name, brand, model, season
FROM nomenclature_tyres
WHERE width = 205
  AND height = 55
  AND diameter = 'R16,00'
  AND season = 'Зимняя'
LIMIT 10;
```

### Поиск дисков по разболтовке
```sql
SELECT cae, name, brand, model
FROM nomenclature_rims
WHERE diameter = 17
  AND bolts_count = 5
  AND bolts_spacing = 114.3
LIMIT 10;
```

### История синхронизаций
```sql
SELECT sync_type,
       started_at,
       completed_at,
       total_items,
       new_items,
       updated_items,
       status
FROM nomenclature_sync_runs
ORDER BY started_at DESC
LIMIT 10;
```

## Мониторинг

Рекомендуется настроить мониторинг:
1. Проверка успешности ежедневной синхронизации
2. Оповещения при ошибках
3. Контроль количества записей (не должно резко меняться)

## Troubleshooting

### Проблема: Таблицы не созданы
```bash
go run cmd/run-migration/main.go migrations/006_create_nomenclature_tables.up.sql
```

### Проблема: Ошибка скачивания XML
- Проверить доступность URL
- Проверить интернет-соединение
- Увеличить timeout в nomenclature_service.go

### Проблема: Медленная синхронизация
- Проверить производительность БД
- Убедиться, что используется bulk upsert
- Проверить индексы на таблицах
