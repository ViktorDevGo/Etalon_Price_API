-- Добавление столбца isimport_1C в таблицу nomenclature_rims

ALTER TABLE nomenclature_rims
ADD COLUMN IF NOT EXISTS "isimport_1C" INTEGER DEFAULT 0;

-- Проставить значение 0 для всех существующих записей
UPDATE nomenclature_rims SET "isimport_1C" = 0;

-- Добавить индекс для быстрого поиска
CREATE INDEX IF NOT EXISTS idx_nomenclature_rims_isimport_1c ON nomenclature_rims("isimport_1C");
