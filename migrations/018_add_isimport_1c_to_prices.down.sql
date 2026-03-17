-- Откат: удаление столбца isimport_1C

DROP INDEX IF EXISTS idx_tyres_prices_stock_isimport_1c;
ALTER TABLE tyres_prices_stock DROP COLUMN IF EXISTS "isimport_1C";
