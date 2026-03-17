-- Добавление столбца isimport_1C в таблицу tyres_prices_stock

ALTER TABLE tyres_prices_stock
ADD COLUMN IF NOT EXISTS "isimport_1C" INTEGER DEFAULT 0;

-- Проставить значение 0 для всех существующих записей
UPDATE tyres_prices_stock SET "isimport_1C" = 0;

-- Добавить индекс для быстрого поиска
CREATE INDEX IF NOT EXISTS idx_tyres_prices_stock_isimport_1c ON tyres_prices_stock("isimport_1C");
