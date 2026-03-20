-- Откат: удаление столбца isimport_1C

DROP INDEX IF EXISTS idx_nomenclature_rims_isimport_1c;
ALTER TABLE nomenclature_rims DROP COLUMN IF EXISTS "isimport_1C";
