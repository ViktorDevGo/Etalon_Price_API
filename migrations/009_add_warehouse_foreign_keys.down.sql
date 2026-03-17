-- Удаляем внешние ключи
ALTER TABLE tyres_prices_stock DROP CONSTRAINT IF EXISTS tyres_prices_stock_warehouse_fkey;
ALTER TABLE rims_prices_stock DROP CONSTRAINT IF EXISTS rims_prices_stock_warehouse_fkey;
