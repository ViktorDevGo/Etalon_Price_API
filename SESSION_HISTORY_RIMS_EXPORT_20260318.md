# История сессии: Экспорт каталога дисков для 1C-Битрикс
**Дата:** 18 марта 2026
**Задача:** Создание и настройка экспорта каталога дисков с разбивкой на файлы

---

## 🎯 Основная задача

Создать экспорт каталога дисков из `nomenclature_rims` в формате CSV для импорта в 1C-Битрикс с разбивкой на файлы по 30,000 строк.

---

## 📋 Выполненные работы

### 1. Создание базового экспорта дисков

**Файл:** `cmd/export-bitrix-rims/main.go`

**Основные характеристики:**
- 64 колонки в CSV
- LEFT JOIN с `rims_prices_stock` для цен и остатков
- Умная логика остатков: ≤4 → 0, >4 → реальное значение
- Маппинг полей из БД в CSV колонки

### 2. Добавление фильтра по брендам

**Задача:** Экспортировать только 68 разрешенных брендов

**Список брендов:**
```
Advan, Advanti, Aero, Alcasta, Alutec, Asterro, BBS, Better, Black Rhino,
Buffalo, COX, Carwel, CrossStreet, Enkei, FF, FM, FR design, FR replica,
Fuel, Inforget, K&K, Khomen Wheels, KoKo, Konig, Kosei, Kronprinz/Accuride,
LS, LS FlowForming, LS Forged, LegeArtis, LegeArtis Concept, Lenso, Lizardo,
MAK, Magnetto, Mefro, Megami, Momo, N2O, NZ, Neo, Next, OZ, Off-Road Wheels,
PDW, Premium Series, RST, Race Ready, Remain, Replay, SKAD Original, SRW,
Sakura, Tech Line, Trebl, Venti, Vossen, X-Race, X`trike, YST, Yamato,
Yamato Samurai, iFree, iFree Original, Евродиск, СКАД, ТЗСК
```

**Результат:**
- Всего дисков в БД: 99,995
- Пройдут фильтр: 92,386 (92%)
- Отфильтруются: 7,609 (8%)

**SQL изменение:**
```sql
WHERE n.brand IN ('Advan', 'Advanti', ... -- 68 брендов)
```

### 3. Изменение структуры колонок IC_GROUP

**Было:**
- IC_GROUP0 (колонка 53) = model
- IC_GROUP1 (колонка 54) = пусто

**Стало:**
- IC_GROUP0 (колонка 53) = brand
- IC_GROUP1 (колонка 54) = model

**Код:**
```go
// Было:
record[52] = *model  // IC_GROUP0
record[53] = ""      // IC_GROUP1

// Стало:
record[52] = *brand  // IC_GROUP0
record[53] = *model  // IC_GROUP1
```

### 4. Исправление формата размеров в nomenclature_rims

**Проблема:** Размеры в формате `9.0x22.0` вместо `9x22`

**Решение:**
```sql
-- Убираем .0 из формата 9.0x22.0 → 9x22
UPDATE nomenclature_rims
SET name = regexp_replace(name, '(\d+)\.0x(\d+)\.0', '\1x\2', 'g')
WHERE name ~ '\d+\.\d+x\d+\.\d+';

-- Убираем .0 в конце диаметра (8.5x20.0 → 8.5x20)
UPDATE nomenclature_rims
SET name = regexp_replace(name, 'x(\d+)\.0', 'x\1', 'g')
WHERE name ~ 'x\d+\.0';
```

**Результат:**
- Обновлено записей (первый этап): 73
- Обновлено записей (второй этап): 44
- **Итого: 117 записей исправлено**

### 5. Установка бренда PDW для записей с (PDW)

**Задача:** Где в `name` есть "(PDW)", установить `brand = 'PDW'`

**SQL:**
```sql
UPDATE nomenclature_rims
SET brand = 'PDW'
WHERE name LIKE '%(PDW)%' AND (brand IS NULL OR brand != 'PDW');
```

**Результат:**
- Обновлено: 27 записей
- Всего с (PDW): 139 записей (112 уже были с правильным brand)

### 6. Работа с колонкой IP_PROP302

#### Этап 1: Добавление апострофа для строкового типа
**Код:**
```go
record[28] = "'" + cae  // IP_PROP302 = '1016
```

**Результат:** IP_PROP302 = '1016 (строковый тип в Excel)

#### Этап 2: Удаление апострофа
**Требование пользователя:** Убрать апостроф

**Код:**
```go
record[28] = cae  // IP_PROP302 = 1016
```

#### Этап 3: Исправление позиции колонки
**Проблема:** CAE находился в IP_PROP305 (колонка 29) вместо IP_PROP302 (колонка 26)

**Решение:**
```go
// Было (НЕПРАВИЛЬНО):
record[28] = cae  // Колонка 29 = IP_PROP305

// Стало (ПРАВИЛЬНО):
record[25] = cae  // Колонка 26 = IP_PROP302
```

**Обновление циклов:**
```go
// Было:
for i := 11; i < 28; i++ { record[i] = "" }

// Стало:
for i := 11; i < 25; i++ { record[i] = "" }
for i := 26; i < 50; i++ { record[i] = "" }
```

### 7. Создание полного каталога с разбивкой на части

#### Этап 1: Генерация полного файла
**Команда:**
```bash
go run cmd/export-bitrix-rims/main.go -output-dir=/Users/viktor/Desktop
```

**Результат:**
- Файл: `Каталог_диски_20260318.csv`
- Строк: 92,414 (92,413 данных + 1 заголовок)
- Размер: 16 MB

#### Этап 2: Разбивка на части по 30,000 строк

**Проблема 1:** Команда `split` не понимает CSV формат (не учитывает кавычки)

**Решение:** Создана утилита на Go для правильной разбивки CSV

**Утилита:** `/tmp/split_csv.go`
```go
// Использует encoding/csv для правильного парсинга
reader := csv.NewReader(file)
writer := csv.NewWriter(outputFile)
```

**Результат:**
- **Part 1:** 30,000 строк данных, 5.1 MB
- **Part 2:** 30,000 строк данных, 5.2 MB
- **Part 3:** 30,000 строк данных, 5.7 MB
- **Part 4:** 2,413 строк данных, 451 KB

**Итого:** 92,413 дисков, 16.5 MB

---

## ✅ Итоговые файлы

### Расположение
```
/Users/viktor/Desktop/Каталог_диски_20260318_part1.csv
/Users/viktor/Desktop/Каталог_диски_20260318_part2.csv
/Users/viktor/Desktop/Каталог_диски_20260318_part3.csv
/Users/viktor/Desktop/Каталог_диски_20260318_part4.csv
```

### Проверка целостности (через CSV parser)
```
Part 1:
  Первая:     IE_XML_ID=1016,      IP_PROP302=1016 ✓
  Последняя:  IE_XML_ID=WHS199559, IP_PROP302=WHS199559 ✓

Part 2:
  Первая:     IE_XML_ID=WHS199560, IP_PROP302=WHS199560 ✓
  Последняя:  IE_XML_ID=WHS498828, IP_PROP302=WHS498828 ✓

Part 3:
  Первая:     IE_XML_ID=WHS498829, IP_PROP302=WHS498829 ✓
  Последняя:  IE_XML_ID=WHS531617, IP_PROP302=WHS531617 ✓

Part 4:
  Первая:     IE_XML_ID=WHS531618, IP_PROP302=WHS531618 ✓
  Последняя:  IE_XML_ID=К2-236902,  IP_PROP302=К2-236902 ✓
```

---

## 📊 Структура CSV файла

### Колонки (64 штуки)

| № | Колонка | Источник | Описание |
|---|---------|----------|----------|
| 1 | IE_XML_ID | cae | Уникальный код товара |
| 2 | IE_NAME | name | Название диска |
| 3-4 | IE_PREVIEW_TEXT, IE_DETAIL_TEXT | - | Пустые |
| 5 | IP_PROP281 | - | Пустое |
| 6 | IP_PROP282 | width | Ширина |
| 7 | IP_PROP283 | diameter | Диаметр |
| 8 | IP_PROP284 | bolts_count | Количество болтов |
| 9 | IP_PROP285 | bolts_spacing | Разболтовка |
| 10 | IP_PROP286 | et | Вылет (ET) |
| 11 | IP_PROP287 | dia | DIA |
| 12-25 | IP_PROP288-301 | - | Пустые |
| **26** | **IP_PROP302** | **cae** | **CAE (дубль)** ✓ |
| 27-50 | IP_PROP303-324 | - | Пустые |
| 51 | IP_PROP325 | brand | Бренд |
| 52 | IP_PROP326 | color | Цвет |
| 53 | IC_GROUP0 | brand | Бренд |
| 54 | IC_GROUP1 | model | Модель |
| 55 | IC_GROUP2 | - | Пустое |
| 56 | CP_QUANTITY | display_stock | Остаток (≤4→0, >4→реальный) |
| 57-60 | CP_WEIGHT, CP_WIDTH, CP_HEIGHT, CP_LENGTH | - | Пустые |
| 61-62 | CV_QUANTITY_FROM, CV_QUANTITY_TO | - | Пустые |
| 63 | CV_PRICE_1 | min_price | Минимальная цена (руб.) |
| 64 | CV_CURRENCY_1 | "RUB" | Валюта |

---

## 🔧 Технические детали

### SQL запрос для экспорта
```sql
WITH rim_aggregates AS (
    SELECT
        cae,
        SUM(stock) as total_stock,
        MIN(price) as min_price
    FROM rims_prices_stock
    GROUP BY cae
)
SELECT
    n.cae,
    n.name,
    n.width,
    n.diameter,
    n.bolts_count,
    n.bolts_spacing,
    n.bolts_spacing2,
    n.et,
    n.dia,
    n.model,
    n.brand,
    n.color,
    CASE
        WHEN COALESCE(p.total_stock, 0) <= 4 THEN 0
        ELSE COALESCE(p.total_stock, 0)
    END as display_stock,
    COALESCE(p.min_price, 0) as min_price
FROM nomenclature_rims n
LEFT JOIN rim_aggregates p ON n.cae = p.cae
WHERE n.brand IN ('Advan', 'Advanti', ... -- 68 брендов)
ORDER BY n.cae
```

### Логика остатков
- Остаток ≤ 4 → CP_QUANTITY = 0 (не показываем малые остатки)
- Остаток > 4 → CP_QUANTITY = реальное значение
- Нет записи → CP_QUANTITY = 0

### Цены
- Хранятся в рублях (INTEGER)
- CV_PRICE_1 = MIN(price) по CAE
- БЕЗ наценки (в отличие от шин, где × 1.12)

### CSV формат
- Кодировка: UTF-8
- Разделитель: запятая
- Экранирование: двойные кавычки для полей с запятыми
- Пример: `"6,5x17/5x112 ET38 D66,6 Concept-MR519 BKF"`

---

## 📝 Важные замечания

1. **Фильтр по брендам** обязателен (68 брендов)
2. **LEFT JOIN** гарантирует полноту каталога
3. **CSV parser** необходим для правильной проверки файлов (awk не понимает CSV кавычки)
4. **IP_PROP302** должен быть в колонке 26 (record[25])
5. **IC_GROUP0** = brand, **IC_GROUP1** = model
6. **Формат размеров** без .0 (9x22 вместо 9.0x22.0)

---

## 🎯 Результат

✅ Создано 4 файла с каталогом дисков
✅ Всего 92,413 дисков
✅ Все поля корректно заполнены
✅ CSV формат правильно экранирован
✅ Готово к импорту в 1C-Битрикс

---

## 📁 Связанные файлы

- `cmd/export-bitrix-rims/main.go` - основной скрипт экспорта
- `docs/BITRIX_RIMS_EXPORT.md` - документация по экспорту
- `migrations/` - изменения в БД (формат размеров, бренд PDW)

---

**Дата завершения:** 18 марта 2026
**Статус:** ✅ Завершено
