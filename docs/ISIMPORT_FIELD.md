# Поле isimport в таблицах номенклатуры

## Назначение

Поле `isimport` используется для отслеживания статуса использования данных номенклатуры.

### Значения:
- **0** - данные загружены из источника, но еще не использованы
- **1** - данные использованы (экспортированы, переданы в другую систему и т.д.)

## Таблицы

Поле добавлено в следующие таблицы:
- `nomenclature_tyres`
- `nomenclature_rims`

## Логика работы

### При загрузке данных

1. **Проверка на полный дубль**
   - Сравниваются ВСЕ поля записи (кроме id, created_at, updated_at)
   - Если найден полный дубль → запись НЕ добавляется (пропускается)
   - Статистика: increment `skipped_duplicates`

2. **Частичный дубль (изменение данных)**
   - Если CAE совпадает, но другие поля отличаются
   - Добавляется НОВАЯ строка с `isimport = 0`
   - Это позволяет вести историю изменений номенклатуры
   - Статистика: increment `new`

3. **Новая запись**
   - Если CAE не найден в таблице
   - Добавляется с `isimport = 0`
   - Статистика: increment `new`

### При использовании данных

Когда данные экспортируются или используются:
```sql
-- Обновить статус на "использовано"
UPDATE nomenclature_tyres 
SET isimport = 1 
WHERE cae = 'SOME_CAE' AND isimport = 0;
```

### Поиск неиспользованных данных

```sql
-- Получить все неиспользованные записи
SELECT * FROM nomenclature_tyres WHERE isimport = 0;

-- Подсчет неиспользованных
SELECT COUNT(*) FROM nomenclature_tyres WHERE isimport = 0;
```

## Индексы

Для оптимизации поиска созданы частичные индексы:
```sql
CREATE INDEX idx_nomenclature_tyres_isimport 
ON nomenclature_tyres(isimport) WHERE isimport = 0;

CREATE INDEX idx_nomenclature_rims_isimport 
ON nomenclature_rims(isimport) WHERE isimport = 0;
```

## Примеры использования

### 1. Экспорт только новых данных

```go
// Получить только неиспользованные записи
rows, err := pool.Query(ctx, `
    SELECT cae, name, brand, model 
    FROM nomenclature_tyres 
    WHERE isimport = 0
    ORDER BY created_at DESC
`)

// После экспорта пометить как использованные
_, err = pool.Exec(ctx, `
    UPDATE nomenclature_tyres 
    SET isimport = 1 
    WHERE isimport = 0
`)
```

### 2. Отслеживание изменений

```sql
-- Найти все версии конкретного товара
SELECT cae, name, brand, model, isimport, created_at
FROM nomenclature_tyres
WHERE cae = '632133'
ORDER BY created_at DESC;
```

### 3. Статистика

```sql
-- Распределение по статусу
SELECT 
    isimport,
    COUNT(*) as count,
    CASE 
        WHEN isimport = 0 THEN 'Загружено'
        WHEN isimport = 1 THEN 'Использовано'
    END as status
FROM nomenclature_tyres
GROUP BY isimport;
```

## Миграция

Поле добавлено миграцией `014_add_isimport_to_nomenclature.sql`:
- Столбец `isimport INTEGER NOT NULL DEFAULT 0`
- Индексы для быстрого поиска
- Комментарии к столбцам

## Важные замечания

1. **Ведение истории**: Один CAE может иметь несколько записей с разными значениями полей
2. **Полные дубли пропускаются**: Если все поля совпадают, новая запись не создается
3. **Автоматическая установка**: При загрузке `isimport` всегда устанавливается в 0
4. **Ручное управление**: Изменение статуса на 1 происходит вручную при использовании данных

## См. также

- `docs/NOMENCLATURE_HISTORY.md` - ведение истории изменений номенклатуры
- `internal/repository/nomenclature_repository.go` - реализация логики
