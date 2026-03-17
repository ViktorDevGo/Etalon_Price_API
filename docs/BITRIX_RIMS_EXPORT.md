# Экспорт каталога дисков для 1C-Битрикс

## Описание

Скрипт `cmd/export-bitrix-rims/main.go` формирует каталог дисков для импорта в 1C-Битрикс.

**Особенности:**
- Только разрешенные бренды (68 брендов)
- Умная логика остатков (≤4 → 0)
- Все диски из nomenclature_rims (независимо от наличия цен)

## Использование

### Полный каталог (production)
```bash
go run cmd/export-bitrix-rims/main.go -output-dir=/Users/viktor/Desktop
```

**Результат:**
- Файл: `Каталог_диски_YYYYMMDD.csv`
- Товаров: ~92,386 дисков (из 68 разрешенных брендов)

### Тестовый экспорт
```bash
go run cmd/export-bitrix-rims/main.go -output-dir=/Users/viktor/Desktop -limit=10
```

**Результат:**
- Файл: `Каталог_диски_YYYYMMDD_TEST_10.csv`
- Товаров: 10 (для тестирования)

## Логика экспорта

### Источники данных
1. **nomenclature_rims** - основная таблица номенклатуры дисков
2. **rims_prices_stock** - цены и остатки (LEFT JOIN)

### Фильтрация по брендам
Экспортируются только диски следующих брендов (68 брендов):
- Advan, Advanti, Aero, Alcasta, Alutec, Asterro, BBS, Better, Black Rhino, Buffalo
- COX, Carwel, CrossStreet, Enkei, FF, FM, FR design, FR replica, Fuel, Inforget
- K&K, Khomen Wheels, KoKo, Konig, Kosei, Kronprinz/Accuride
- LS, LS FlowForming, LS Forged, LegeArtis, LegeArtis Concept, Lenso, Lizardo
- MAK, Magnetto, Mefro, Megami, Momo, N2O, NZ, Neo, Next
- OZ, Off-Road Wheels, PDW, Premium Series, RST, Race Ready, Remain, Replay
- SKAD Original, SRW, Sakura, Tech Line, Trebl, Venti, Vossen
- X-Race, X`trike, YST, Yamato, Yamato Samurai
- iFree, iFree Original, Евродиск, СКАД, ТЗСК

**Статистика:** Из 99,995 дисков в БД → 92,386 пройдут фильтр (92%)

### SQL запрос
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
WHERE n.brand IN ('Advan', 'Advanti', 'Aero', ... -- 68 брендов
ORDER BY n.cae
```

### Обработка остатков и цен
- **LEFT JOIN** гарантирует, что ВСЕ диски из nomenclature_rims попадут в экспорт
- **Умная логика остатков:**
  - Если SUM(stock) ≤ 4 → CP_QUANTITY = 0 (скрываем малые остатки)
  - Если SUM(stock) > 4 → CP_QUANTITY = реальное значение
  - Если нет записи в rims_prices_stock → CP_QUANTITY = 0
- **COALESCE(p.min_price, 0)** → если нет записи в rims_prices_stock, price = 0

## Структура CSV файла

### Колонки (64 штуки)

| № | Колонка | Источник | Описание |
|---|---------|----------|----------|
| 1 | IE_XML_ID | cae | Уникальный код товара |
| 2 | IE_NAME | name | Название диска |
| 3-4 | IE_PREVIEW_TEXT, IE_DETAIL_TEXT | - | Пустые |
| 5 | IP_PROP281 | - | Метка (пусто) |
| 6 | IP_PROP282 | width | Ширина диска |
| 7 | IP_PROP283 | diameter | Диаметр |
| 8 | IP_PROP284 | bolts_count | Количество болтов |
| 9 | IP_PROP285 | bolts_spacing | Разболтовка |
| 10 | IP_PROP286 | et | Вылет (ET) |
| 11 | IP_PROP287 | dia | DIA (диаметр ступицы) |
| 12-28 | IP_PROP288-304 | - | Пустые колонки |
| 29 | IP_PROP302 | cae | CAE (дубль) |
| 30-50 | IP_PROP303-324 | - | Пустые колонки |
| 51 | IP_PROP325 | brand | Бренд |
| 52 | IP_PROP326 | color | Цвет |
| 53 | IC_GROUP0 | brand | Бренд |
| 54 | IC_GROUP1 | model | Модель |
| 55 | IC_GROUP2 | - | Пустое |
| 56 | CP_QUANTITY | SUM(stock) | Суммарный остаток (0 если нет данных) |
| 57-60 | CP_WEIGHT, CP_WIDTH, CP_HEIGHT, CP_LENGTH | - | Не заполняем |
| 61-62 | CV_QUANTITY_FROM, CV_QUANTITY_TO | - | Пустые |
| 63 | CV_PRICE_1 | MIN(price) | Минимальная цена (0 если нет данных) |
| 64 | CV_CURRENCY_1 | "RUB" | Валюта |

## Примеры данных

### Диск с большим остатком (> 4)
```csv
WHS021361,"5,5x14/4x98 ET38 D58,6 Ягуар (КЛ147) Селена",,,,5.5,14,4,98.0,38,58.6,...,WHS021361,...,СКАД,Селена,СКАД,Ягуар (КЛ147),,53,,,,,,,6388,RUB
```
- CAE: WHS021361
- IP_PROP325 (Бренд): СКАД
- IP_PROP326 (Цвет): Селена
- IC_GROUP0 (Бренд): СКАД
- IC_GROUP1 (Модель): Ягуар (КЛ147)
- **Реальный остаток: 53 шт** → отображается 53 ✓
- Цена: 6,388₽

### Диск с малым остатком (≤ 4)
```csv
1016,FF 395 9.0x22.0 6*139.7 ET15 D106.1 BMF,,,,9.0,22,6,139.7,15,106.1,...,1016,...,FF,BMF,FF,395,,0,,,,,,,26651,RUB
```
- CAE: 1016
- IP_PROP325 (Бренд): FF
- IP_PROP326 (Цвет): BMF
- IC_GROUP0 (Бренд): FF
- IC_GROUP1 (Модель): 395
- **Реальный остаток: 1 шт** → отображается 0 ✓
- Цена: 26,651₽

### Диск БЕЗ цен/остатков
```csv
190.108.719,"7x20/10x335 ET153,5 D281 RAL 7004",,,,7.0,20,10,335.0,154,281.0,...,190.108.719,...,Jantsa,RAL 7004,Jantsa,"10/335/281/153,5",,0,,,,,,,0,RUB
```
- CAE: 190.108.719
- IP_PROP325 (Бренд): Jantsa
- IP_PROP326 (Цвет): RAL 7004
- IC_GROUP0 (Бренд): Jantsa
- IC_GROUP1 (Модель): 10/335/281/153,5
- **Нет в rims_prices_stock** → отображается 0 ✓
- Цена: 0₽

## Важные замечания

1. **Фильтр по брендам** - экспортируются только 68 разрешенных брендов (92,386 дисков из 99,995)
2. **Цены в рублях** - значения уже в рублях, деление на 100 НЕ нужно
3. **LEFT JOIN** - гарантирует полноту каталога для выбранных брендов
4. **Умная логика остатков:**
   - Остаток ≤ 4 → CP_QUANTITY = 0 (не показываем малые остатки)
   - Остаток > 4 → CP_QUANTITY = реальное значение
   - Нет записи → CP_QUANTITY = 0
5. **COALESCE** - автоматически обрабатывает NULL значения
6. **Автоматическое имя файла** - с текущей датой в формате YYYYMMDD

## Интеграция с 1C-Битрикс

После формирования файл можно импортировать в 1C-Битрикс через стандартный механизм импорта товаров из CSV.

### Рекомендуемый процесс:
1. Сформировать файл локально или на сервере
2. Загрузить в `/home/bitrix/www/upload/1c_catalog/`
3. Запустить импорт через административную панель 1C-Битрикс
4. Проверить корректность импорта

## Автоматизация

Для автоматической выгрузки можно настроить cron:
```bash
# Ежедневно в 3:00 по московскому времени
0 3 * * * cd /app && go run cmd/export-bitrix-rims/main.go -output-dir=/home/bitrix/www/upload/1c_catalog
```
