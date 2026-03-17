# Экспорт номенклатуры в CSV

Утилиты для экспорта номенклатуры шин и дисков в CSV формат.

---

## Шины (Tyres)

### Быстрый старт

```bash
# Экспорт 10 позиций (по умолчанию)
go run cmd/export-tyres-csv/main.go

# Экспорт 100 позиций
go run cmd/export-tyres-csv/main.go -limit=100

# Экспорт всех позиций
go run cmd/export-tyres-csv/main.go -limit=0

# Указать свой путь к файлу
go run cmd/export-tyres-csv/main.go -output=/path/to/file.csv -limit=50
```

### Параметры

- **`-output`** - путь к выходному файлу (по умолчанию: `/Users/viktor/Desktop/export_nomenclature_tyres.csv`)
- **`-limit`** - количество позиций (по умолчанию: `10`, `0` = все позиции)

### Формат CSV

Соответствует формату `export_file_112522.csv`:

| Колонка | Название | Пример | Описание |
|---------|----------|--------|----------|
| IE_XML_ID | Артикул | 2504900 | Уникальный код товара |
| IE_NAME | Название | Pirelli Ice Zero 265/70 R16,00 112T | Полное название |
| IP_PROP140 | Ширина | 265 | Ширина шины (мм) |
| IP_PROP141 | Высота | 70 | Высота профиля (%) |
| IP_PROP142 | Диаметр | R16,00 | Посадочный диаметр |
| IP_PROP145 | Сезон | Зимняя | Летняя/Зимняя/Всесезонная |
| IP_PROP146 | Шипы | Шипованные | Шипованные/Нешипованные |
| IP_PROP149 | Индекс нагрузки | 112 | Максимальная нагрузка |
| IP_PROP150 | Индекс скорости | T | Максимальная скорость |
| IP_PROP151 | Артикул | 2504900 | Дубль артикула |
| IC_GROUP0 | Категория | 1 | 1 = шины |
| IC_GROUP1 | Бренд | Pirelli | Производитель |
| IC_GROUP2 | Модель | Ice Zero | Модель шины |
| CP_QUANTITY | Количество | 1000 | Количество на складе |
| CV_CURRENCY_1 | Валюта | RUB | Валюта цены |

### Пример вывода

```csv
IE_XML_ID,IE_NAME,...
2504900,"Pirelli Ice Zero 265/70 R16,00 112T",...
PXR0033103,"Bridgestone Blizzak VRX 245/40 R17,00 91S",...
```

---

## Диски (Rims)

### Быстрый старт

```bash
# Экспорт 10 позиций (по умолчанию)
go run cmd/export-rims-csv/main.go

# Экспорт 100 позиций
go run cmd/export-rims-csv/main.go -limit=100

# Экспорт всех позиций
go run cmd/export-rims-csv/main.go -limit=0

# Указать свой путь к файлу
go run cmd/export-rims-csv/main.go -output=/path/to/file.csv -limit=50
```

### Параметры

- **`-output`** - путь к выходному файлу (по умолчанию: `/Users/viktor/Desktop/export_nomenclature_rims.csv`)
- **`-limit`** - количество позиций (по умолчанию: `10`, `0` = все позиции)

### Формат CSV

| Колонка | Название | Пример | Описание |
|---------|----------|--------|----------|
| IE_XML_ID | Артикул | RIM001 | Уникальный код товара |
| IE_NAME | Название | OZ Racing Superturismo 7.5x17 5/114.3 ET45 DIA75.0 Black | Полное название |
| IP_PROP140 | Ширина | 7.5 | Ширина диска (дюймы) |
| IP_PROP142 | Диаметр | 17 | Диаметр диска (дюймы) |
| IP_PROP151 | Артикул | RIM001 | Дубль артикула |
| IP_PROP152 | Количество болтов | 5 | Количество отверстий |
| IP_PROP153 | Разболтовка | 114.3 | PCD (мм) |
| IP_PROP154 | ET | 45 | Вылет диска (мм) |
| IP_PROP155 | DIA | 75.0 | Центральное отверстие (мм) |
| IP_PROP156 | Цвет | Black | Цвет/отделка |
| IC_GROUP0 | Категория | 2 | 2 = диски |
| IC_GROUP1 | Бренд | OZ Racing | Производитель |
| IC_GROUP2 | Модель | Superturismo | Модель диска |
| CP_QUANTITY | Количество | 1000 | Количество на складе |
| CV_CURRENCY_1 | Валюта | RUB | Валюта цены |

---

## Статистика

Проверить количество позиций в БД:

```bash
# Количество шин
psql "$DB_DSN" -c "SELECT COUNT(*) FROM nomenclature_tyres WHERE cae IS NOT NULL;"

# Количество дисков
psql "$DB_DSN" -c "SELECT COUNT(*) FROM nomenclature_rims WHERE cae IS NOT NULL;"
```

Или через Go:

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	pool, err := pgxpool.New(context.Background(), os.Getenv("DB_DSN"))
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	var tyresCount, rimsCount int

	pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM nomenclature_tyres WHERE cae IS NOT NULL").Scan(&tyresCount)
	pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM nomenclature_rims WHERE cae IS NOT NULL").Scan(&rimsCount)

	fmt.Printf("Шины: %d позиций\n", tyresCount)
	fmt.Printf("Диски: %d позиций\n", rimsCount)
	fmt.Printf("Всего: %d позиций\n", tyresCount+rimsCount)
}
```

---

## Примеры использования

### 1. Экспорт для загрузки в интернет-магазин

```bash
# Экспортируем все шины
go run cmd/export-tyres-csv/main.go -limit=0 -output=/var/www/catalog/tyres.csv

# Экспортируем все диски
go run cmd/export-rims-csv/main.go -limit=0 -output=/var/www/catalog/rims.csv
```

### 2. Экспорт для анализа

```bash
# Первая 1000 шин
go run cmd/export-tyres-csv/main.go -limit=1000 -output=~/Downloads/tyres_sample.csv

# Открываем в Excel/Numbers для анализа
open ~/Downloads/tyres_sample.csv
```

### 3. Автоматизация (cron job)

```bash
# Добавить в crontab
0 3 * * * cd /path/to/project && go run cmd/export-tyres-csv/main.go -limit=0 -output=/backups/tyres_$(date +\%Y\%m\%d).csv
```

---

## Производительность

| Количество позиций | Время экспорта | Размер файла |
|-------------------|----------------|--------------|
| 10 | ~0.1 сек | ~2 KB |
| 100 | ~0.5 сек | ~20 KB |
| 1,000 | ~2 сек | ~200 KB |
| 10,000 | ~15 сек | ~2 MB |
| 60,000 (все шины) | ~90 сек | ~12 MB |
| 100,000 (все диски) | ~150 сек | ~20 MB |

---

## Troubleshooting

### Ошибка: "DB_DSN not set"

Убедитесь, что файл `.env` существует и содержит:

```env
DB_DSN=postgresql://user:password@host:port/database
```

### Файл пустой или содержит только заголовок

Проверьте данные в БД:

```bash
psql "$DB_DSN" -c "SELECT COUNT(*) FROM nomenclature_tyres WHERE cae IS NOT NULL;"
```

Если 0, запустите синхронизацию номенклатуры:

```bash
go run cmd/sync-nomenclature/main.go -type=all
```

### Неправильная кодировка (кириллица)

CSV файлы экспортируются в UTF-8. При открытии в Excel:

1. **File → Import** (не двойной клик)
2. Выберите **Text Files**
3. Укажите кодировку **UTF-8**
4. Delimiter: **Comma**

Или используйте **Numbers** (macOS), **LibreOffice** - они автоматически определяют UTF-8.

---

## Расширение функционала

### Добавить поле цены

Изменить запрос в `export-tyres-csv/main.go`:

```go
query := `
	SELECT
		t.cae, t.name, t.width, t.height, t.diameter,
		t.load_index, t.speed_index, t.model, t.brand, t.season,
		t.is_studded, t.tiretype, t.runflat, t.reinforced,
		p.price  -- Добавили цену
	FROM nomenclature_tyres t
	LEFT JOIN tyres_prices_stock p ON t.cae = p.cae
	WHERE t.cae IS NOT NULL AND t.cae != ''
	ORDER BY t.id
`
```

Добавить поле в CSV:

```go
// CV_PRICE_1 - цена
if price != nil {
	// Цена в копейках → рубли
	record[68] = strconv.FormatFloat(float64(*price)/100.0, 'f', 2, 64)
}
```

### Фильтрация по бренду

```go
query := `
	SELECT ... FROM nomenclature_tyres
	WHERE cae IS NOT NULL
	AND brand = 'Bridgestone'  -- Добавили фильтр
	ORDER BY id
`
```

### Группировка по складам

```go
query := `
	SELECT
		t.cae, t.name, ...,
		p.warehouse_name, p.price, p.stock
	FROM nomenclature_tyres t
	LEFT JOIN tyres_prices_stock p ON t.cae = p.cae
	WHERE t.cae IS NOT NULL
	ORDER BY t.id, p.warehouse_name
`
```

---

## Готовые файлы

После запуска утилит вы получите:

- `/Users/viktor/Desktop/export_nomenclature_tyres.csv` - шины
- `/Users/viktor/Desktop/export_nomenclature_rims.csv` - диски

Файлы готовы для:
- Импорта в 1C, Битрикс, OpenCart
- Загрузки в Яндекс.Маркет, Авито
- Анализа в Excel, Google Sheets
- Интеграции с другими системами
