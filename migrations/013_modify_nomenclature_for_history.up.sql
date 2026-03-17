-- Изменение таблиц номенклатуры для ведения истории изменений
-- Убираем UNIQUE constraint на cae, чтобы можно было хранить историю изменений

-- Для шин (используем CASCADE для удаления зависимостей)
ALTER TABLE nomenclature_tyres DROP CONSTRAINT IF EXISTS nomenclature_tyres_cae_key CASCADE;

-- Создаём индекс для быстрой проверки дублей (все значимые поля)
CREATE INDEX IF NOT EXISTS idx_nomenclature_tyres_duplicate_check ON nomenclature_tyres(
    cae, name, width, height, diameter, diameter_out,
    load_index, speed_index, model, brand, season,
    is_studded, tiretype, runflat, reinforced
);

-- Для дисков (используем CASCADE для удаления зависимостей)
ALTER TABLE nomenclature_rims DROP CONSTRAINT IF EXISTS nomenclature_rims_cae_key CASCADE;

-- Создаём индекс для быстрой проверки дублей (все значимые поля)
CREATE INDEX IF NOT EXISTS idx_nomenclature_rims_duplicate_check ON nomenclature_rims(
    cae, name, width, diameter, bolts_count,
    bolts_spacing, bolts_spacing2, et, dia,
    model, brand, color, rim_type
);

-- Добавляем индекс на cae для быстрого поиска (не уникальный)
CREATE INDEX IF NOT EXISTS idx_nomenclature_tyres_cae_lookup ON nomenclature_tyres(cae);
CREATE INDEX IF NOT EXISTS idx_nomenclature_rims_cae_lookup ON nomenclature_rims(cae);

-- Комментарии
COMMENT ON INDEX idx_nomenclature_tyres_duplicate_check IS 'Индекс для проверки полных дублей при обновлении';
COMMENT ON INDEX idx_nomenclature_rims_duplicate_check IS 'Индекс для проверки полных дублей при обновлении';
