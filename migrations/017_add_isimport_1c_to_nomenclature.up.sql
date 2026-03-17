-- Добавление столбца isimport_1С в таблицу nomenclature_tyres

ALTER TABLE nomenclature_tyres
ADD COLUMN IF NOT EXISTS "isimport_1С" INTEGER DEFAULT 0;

-- Проставить значения для всех существующих записей
UPDATE nomenclature_tyres SET isimport = 1, "isimport_1С" = 0;

-- Добавить индекс для быстрого поиска
CREATE INDEX IF NOT EXISTS idx_nomenclature_tyres_isimport_1c ON nomenclature_tyres("isimport_1С");
