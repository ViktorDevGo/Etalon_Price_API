-- Откат: удаление столбца isimport_1С

DROP INDEX IF EXISTS idx_nomenclature_tyres_isimport_1c;
ALTER TABLE nomenclature_tyres DROP COLUMN IF EXISTS "isimport_1С";
