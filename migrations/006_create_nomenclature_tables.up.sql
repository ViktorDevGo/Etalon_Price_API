-- Таблица номенклатуры шин
CREATE TABLE IF NOT EXISTS nomenclature_tyres (
    id BIGSERIAL PRIMARY KEY,
    cae VARCHAR(50) UNIQUE NOT NULL, -- Код товара
    name TEXT NOT NULL, -- Полное название
    width DECIMAL(10,2), -- Ширина
    height DECIMAL(10,2), -- Высота профиля
    diameter VARCHAR(20), -- Диаметр (R16,00)
    diameter_out VARCHAR(20), -- Внешний диаметр
    load_index VARCHAR(20), -- Индекс нагрузки
    speed_index VARCHAR(10), -- Индекс скорости
    model VARCHAR(255), -- Модель
    brand VARCHAR(255), -- Бренд
    season VARCHAR(50), -- Сезон (Зимняя, Летняя, Всесезонная)
    is_studded VARCHAR(20), -- Шипы (шип / не шип)
    tiretype VARCHAR(50), -- Тип шины (Легковая и т.д.)
    runflat VARCHAR(50), -- Ранфлет
    reinforced VARCHAR(50), -- Усиленная

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Индексы для быстрого поиска
CREATE INDEX idx_nomenclature_tyres_brand ON nomenclature_tyres(brand);
CREATE INDEX idx_nomenclature_tyres_model ON nomenclature_tyres(model);
CREATE INDEX idx_nomenclature_tyres_season ON nomenclature_tyres(season);
CREATE INDEX idx_nomenclature_tyres_width_height_diameter ON nomenclature_tyres(width, height, diameter);

-- Таблица номенклатуры дисков
CREATE TABLE IF NOT EXISTS nomenclature_rims (
    id BIGSERIAL PRIMARY KEY,
    cae VARCHAR(50) UNIQUE NOT NULL, -- Код товара
    name TEXT NOT NULL, -- Полное название
    width DECIMAL(10,2), -- Ширина
    diameter DECIMAL(10,2), -- Диаметр
    bolts_count INTEGER, -- Количество болтов
    bolts_spacing DECIMAL(10,2), -- Разболтовка
    bolts_spacing2 DECIMAL(10,2), -- Разболтовка 2
    et DECIMAL(10,2), -- Вылет (ET)
    dia DECIMAL(10,2), -- DIA (центральное отверстие)
    model VARCHAR(255), -- Модель
    brand VARCHAR(255), -- Бренд
    color VARCHAR(100), -- Цвет
    rim_type VARCHAR(10), -- Тип диска

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Индексы для быстрого поиска
CREATE INDEX idx_nomenclature_rims_brand ON nomenclature_rims(brand);
CREATE INDEX idx_nomenclature_rims_model ON nomenclature_rims(model);
CREATE INDEX idx_nomenclature_rims_diameter ON nomenclature_rims(diameter);
CREATE INDEX idx_nomenclature_rims_bolts ON nomenclature_rims(bolts_count, bolts_spacing);

-- Таблица истории загрузок номенклатуры
CREATE TABLE IF NOT EXISTS nomenclature_sync_runs (
    id BIGSERIAL PRIMARY KEY,
    sync_type VARCHAR(20) NOT NULL, -- 'tyres' или 'rims'
    started_at TIMESTAMP NOT NULL,
    completed_at TIMESTAMP,
    status VARCHAR(20) NOT NULL, -- 'running', 'completed', 'failed'
    total_items INTEGER DEFAULT 0,
    new_items INTEGER DEFAULT 0,
    updated_items INTEGER DEFAULT 0,
    error_message TEXT,

    CONSTRAINT nomenclature_sync_runs_sync_type_check CHECK (sync_type IN ('tyres', 'rims')),
    CONSTRAINT nomenclature_sync_runs_status_check CHECK (status IN ('running', 'completed', 'failed'))
);

CREATE INDEX idx_nomenclature_sync_runs_type_started ON nomenclature_sync_runs(sync_type, started_at DESC);
