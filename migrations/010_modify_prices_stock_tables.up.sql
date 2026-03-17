-- Модификация таблиц цен и остатков

-- 1. Удаляем FK constraints
ALTER TABLE tyres_prices_stock DROP CONSTRAINT IF EXISTS tyres_prices_stock_warehouse_fkey;
ALTER TABLE rims_prices_stock DROP CONSTRAINT IF EXISTS rims_prices_stock_warehouse_fkey;

-- 2. Добавляем новые колонки в tyres_prices_stock
ALTER TABLE tyres_prices_stock ADD COLUMN IF NOT EXISTS warehouse_name VARCHAR(255);
ALTER TABLE tyres_prices_stock ADD COLUMN IF NOT EXISTS isimport INTEGER DEFAULT 0;
ALTER TABLE tyres_prices_stock ADD COLUMN IF NOT EXISTS provider VARCHAR(100) DEFAULT 'Форточки';

-- 3. Заполняем warehouse_name из таблицы warehouses
UPDATE tyres_prices_stock
SET warehouse_name = w.name
FROM warehouses w
WHERE tyres_prices_stock.warehouse_id = w.id
  AND tyres_prices_stock.warehouse_name IS NULL;

-- 4. Делаем warehouse_name NOT NULL после заполнения
ALTER TABLE tyres_prices_stock ALTER COLUMN warehouse_name SET NOT NULL;

-- 5. Удаляем старую колонку warehouse_id
ALTER TABLE tyres_prices_stock DROP COLUMN warehouse_id;

-- 6. Обновляем unique constraint
ALTER TABLE tyres_prices_stock DROP CONSTRAINT IF EXISTS tyres_prices_stock_cae_warehouse_unique;
ALTER TABLE tyres_prices_stock ADD CONSTRAINT tyres_prices_stock_cae_warehouse_unique
    UNIQUE (cae, warehouse_name, provider);

-- 7. Создаем индексы для tyres_prices_stock
CREATE INDEX IF NOT EXISTS idx_tyres_prices_stock_warehouse_name ON tyres_prices_stock(warehouse_name);
CREATE INDEX IF NOT EXISTS idx_tyres_prices_stock_isimport ON tyres_prices_stock(isimport);
CREATE INDEX IF NOT EXISTS idx_tyres_prices_stock_provider ON tyres_prices_stock(provider);
CREATE INDEX IF NOT EXISTS idx_tyres_prices_stock_cae_isimport ON tyres_prices_stock(cae, isimport);
CREATE INDEX IF NOT EXISTS idx_tyres_prices_stock_provider_isimport ON tyres_prices_stock(provider, isimport);

-- 8. Аналогично для rims_prices_stock
ALTER TABLE rims_prices_stock ADD COLUMN IF NOT EXISTS warehouse_name VARCHAR(255);
ALTER TABLE rims_prices_stock ADD COLUMN IF NOT EXISTS isimport INTEGER DEFAULT 0;
ALTER TABLE rims_prices_stock ADD COLUMN IF NOT EXISTS provider VARCHAR(100) DEFAULT 'Форточки';

-- 9. Заполняем warehouse_name из таблицы warehouses
UPDATE rims_prices_stock
SET warehouse_name = w.name
FROM warehouses w
WHERE rims_prices_stock.warehouse_id = w.id
  AND rims_prices_stock.warehouse_name IS NULL;

-- 10. Делаем warehouse_name NOT NULL после заполнения
ALTER TABLE rims_prices_stock ALTER COLUMN warehouse_name SET NOT NULL;

-- 11. Удаляем старую колонку warehouse_id
ALTER TABLE rims_prices_stock DROP COLUMN warehouse_id;

-- 12. Обновляем unique constraint
ALTER TABLE rims_prices_stock DROP CONSTRAINT IF EXISTS rims_prices_stock_cae_warehouse_unique;
ALTER TABLE rims_prices_stock ADD CONSTRAINT rims_prices_stock_cae_warehouse_unique
    UNIQUE (cae, warehouse_name, provider);

-- 13. Создаем индексы для rims_prices_stock
CREATE INDEX IF NOT EXISTS idx_rims_prices_stock_warehouse_name ON rims_prices_stock(warehouse_name);
CREATE INDEX IF NOT EXISTS idx_rims_prices_stock_isimport ON rims_prices_stock(isimport);
CREATE INDEX IF NOT EXISTS idx_rims_prices_stock_provider ON rims_prices_stock(provider);
CREATE INDEX IF NOT EXISTS idx_rims_prices_stock_cae_isimport ON rims_prices_stock(cae, isimport);
CREATE INDEX IF NOT EXISTS idx_rims_prices_stock_provider_isimport ON rims_prices_stock(provider, isimport);

-- 14. Обновляем старые индексы
DROP INDEX IF EXISTS idx_tyres_prices_stock_warehouse;
DROP INDEX IF EXISTS idx_rims_prices_stock_warehouse;
