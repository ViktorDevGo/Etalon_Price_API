-- 014_add_isimport_to_nomenclature.sql
-- Добавление столбца isimport для отслеживания использования данных номенклатуры

-- Добавить столбец isimport в nomenclature_tyres
ALTER TABLE nomenclature_tyres 
ADD COLUMN IF NOT EXISTS isimport INTEGER NOT NULL DEFAULT 0;

-- Добавить столбец isimport в nomenclature_rims
ALTER TABLE nomenclature_rims 
ADD COLUMN IF NOT EXISTS isimport INTEGER NOT NULL DEFAULT 0;

-- Создать индекс для быстрого поиска неиспользованных записей
CREATE INDEX IF NOT EXISTS idx_nomenclature_tyres_isimport 
ON nomenclature_tyres(isimport) WHERE isimport = 0;

CREATE INDEX IF NOT EXISTS idx_nomenclature_rims_isimport 
ON nomenclature_rims(isimport) WHERE isimport = 0;

-- Комментарии
COMMENT ON COLUMN nomenclature_tyres.isimport IS '0 - данные загружены, 1 - данные использованы/экспортированы';
COMMENT ON COLUMN nomenclature_rims.isimport IS '0 - данные загружены, 1 - данные использованы/экспортированы';
