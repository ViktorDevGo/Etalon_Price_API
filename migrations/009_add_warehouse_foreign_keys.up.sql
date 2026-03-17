-- Добавляем внешние ключи на таблицу warehouses

-- Для шин
ALTER TABLE tyres_prices_stock
    ADD CONSTRAINT tyres_prices_stock_warehouse_fkey
    FOREIGN KEY (warehouse_id) REFERENCES warehouses(id) ON DELETE CASCADE;

-- Для дисков
ALTER TABLE rims_prices_stock
    ADD CONSTRAINT rims_prices_stock_warehouse_fkey
    FOREIGN KEY (warehouse_id) REFERENCES warehouses(id) ON DELETE CASCADE;
