-- Rollback изменений номенклатуры

-- Удаляем индексы
DROP INDEX IF EXISTS idx_nomenclature_tyres_duplicate_check;
DROP INDEX IF EXISTS idx_nomenclature_rims_duplicate_check;
DROP INDEX IF EXISTS idx_nomenclature_tyres_cae_lookup;
DROP INDEX IF EXISTS idx_nomenclature_rims_cae_lookup;

-- Восстанавливаем UNIQUE constraint (внимание: может не сработать если есть дубли)
ALTER TABLE nomenclature_tyres ADD CONSTRAINT nomenclature_tyres_cae_key UNIQUE (cae);
ALTER TABLE nomenclature_rims ADD CONSTRAINT nomenclature_rims_cae_key UNIQUE (cae);
