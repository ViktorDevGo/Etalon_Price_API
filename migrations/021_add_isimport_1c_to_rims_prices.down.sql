-- Откат: удаление столбца isimport_1C

DROP INDEX IF EXISTS idx_rims_prices_stock_isimport_1c;
ALTER TABLE rims_prices_stock DROP COLUMN IF EXISTS "isimport_1C";
