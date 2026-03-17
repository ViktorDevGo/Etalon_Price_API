-- Откат модификации таблиц цен и остатков

-- Tyres
ALTER TABLE tyres_prices_stock ADD COLUMN IF NOT EXISTS warehouse_id INTEGER;

UPDATE tyres_prices_stock
SET warehouse_id = w.id
FROM warehouses w
WHERE tyres_prices_stock.warehouse_name = w.name;

ALTER TABLE tyres_prices_stock DROP COLUMN IF EXISTS warehouse_name;
ALTER TABLE tyres_prices_stock DROP COLUMN IF EXISTS isimport;
ALTER TABLE tyres_prices_stock DROP COLUMN IF EXISTS provider;

ALTER TABLE tyres_prices_stock DROP CONSTRAINT IF EXISTS tyres_prices_stock_cae_warehouse_unique;
ALTER TABLE tyres_prices_stock ADD CONSTRAINT tyres_prices_stock_cae_warehouse_unique
    UNIQUE (cae, warehouse_id);

-- Rims
ALTER TABLE rims_prices_stock ADD COLUMN IF NOT EXISTS warehouse_id INTEGER;

UPDATE rims_prices_stock
SET warehouse_id = w.id
FROM warehouses w
WHERE rims_prices_stock.warehouse_name = w.name;

ALTER TABLE rims_prices_stock DROP COLUMN IF EXISTS warehouse_name;
ALTER TABLE rims_prices_stock DROP COLUMN IF EXISTS isimport;
ALTER TABLE rims_prices_stock DROP COLUMN IF EXISTS provider;

ALTER TABLE rims_prices_stock DROP CONSTRAINT IF EXISTS rims_prices_stock_cae_warehouse_unique;
ALTER TABLE rims_prices_stock ADD CONSTRAINT rims_prices_stock_cae_warehouse_unique
    UNIQUE (cae, warehouse_id);

-- Restore FK
ALTER TABLE tyres_prices_stock
    ADD CONSTRAINT tyres_prices_stock_warehouse_fkey
    FOREIGN KEY (warehouse_id) REFERENCES warehouses(id);

ALTER TABLE rims_prices_stock
    ADD CONSTRAINT rims_prices_stock_warehouse_fkey
    FOREIGN KEY (warehouse_id) REFERENCES warehouses(id);
