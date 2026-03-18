-- Migration 018: Create Severavto nomenclature tables
-- Создание таблиц номенклатуры для поставщика Северавто

-- Номенклатура шин Северавто
CREATE TABLE IF NOT EXISTS nomenclature_tyres_severavto (
    id BIGSERIAL PRIMARY KEY,
    commodity_id VARCHAR(50) UNIQUE NOT NULL,  -- NNCOMMODIF (уникальный ID товара)
    article VARCHAR(100) NOT NULL,             -- SARTICLE (артикул)
    name TEXT NOT NULL,                        -- полное название
    brand VARCHAR(255),                        -- SBREND (бренд)
    model VARCHAR(255),                        -- SMODEL (модель)
    width DECIMAL(10,2),                       -- SWIDTH (ширина)
    height DECIMAL(10,2),                      -- SHEIGHT (профиль/высота)
    diameter VARCHAR(20),                      -- SRADIUS (радиус)
    load_index VARCHAR(20),                    -- SLOAD (индекс нагрузки)
    speed_index VARCHAR(10),                   -- SSPEED (индекс скорости)
    season VARCHAR(50),                        -- сезон (определяется из STHORNING + модели)
    is_studded VARCHAR(20),                    -- STHORNING (шипы: Да/Нет)
    tiretype VARCHAR(50) DEFAULT 'Легковая',   -- тип шины (по умолчанию Легковая)
    runflat VARCHAR(50),                       -- RunFlat (определяется из названия)
    manufacture_year INTEGER,                  -- SMNFYEAR (год производства)
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Индексы для быстрого поиска
CREATE INDEX idx_severavto_tyres_commodity_id ON nomenclature_tyres_severavto(commodity_id);
CREATE INDEX idx_severavto_tyres_brand ON nomenclature_tyres_severavto(brand);
CREATE INDEX idx_severavto_tyres_model ON nomenclature_tyres_severavto(model);
CREATE INDEX idx_severavto_tyres_article ON nomenclature_tyres_severavto(article);
CREATE INDEX idx_severavto_tyres_season ON nomenclature_tyres_severavto(season);
CREATE INDEX idx_severavto_tyres_dims ON nomenclature_tyres_severavto(width, height, diameter);

-- Номенклатура дисков Северавто
CREATE TABLE IF NOT EXISTS nomenclature_rims_severavto (
    id BIGSERIAL PRIMARY KEY,
    commodity_id VARCHAR(50) UNIQUE NOT NULL,  -- NNCOMMODIF (уникальный ID товара)
    article VARCHAR(100) NOT NULL,             -- SARTICLE (артикул)
    name TEXT NOT NULL,                        -- полное название
    brand VARCHAR(255),                        -- SBREND (бренд)
    model VARCHAR(255),                        -- SMODEL (модель)
    width DECIMAL(10,2),                       -- ширина диска
    diameter DECIMAL(10,2),                    -- диаметр диска
    bolts_count INTEGER,                       -- количество отверстий (PCD)
    bolts_spacing DECIMAL(10,2),               -- расстояние между отверстиями
    bolts_spacing2 DECIMAL(10,2),              -- доп. расстояние (для двойного PCD)
    et DECIMAL(10,2),                          -- ET (вылет диска)
    dia DECIMAL(10,2),                         -- DIA (диаметр центрального отверстия)
    color VARCHAR(100),                        -- цвет диска
    rim_type VARCHAR(10),                      -- тип (литой/кованый/штампованный)
    manufacture_year INTEGER,                  -- год производства
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Индексы для быстрого поиска
CREATE INDEX idx_severavto_rims_commodity_id ON nomenclature_rims_severavto(commodity_id);
CREATE INDEX idx_severavto_rims_brand ON nomenclature_rims_severavto(brand);
CREATE INDEX idx_severavto_rims_model ON nomenclature_rims_severavto(model);
CREATE INDEX idx_severavto_rims_article ON nomenclature_rims_severavto(article);
CREATE INDEX idx_severavto_rims_diameter ON nomenclature_rims_severavto(diameter);
CREATE INDEX idx_severavto_rims_pcd ON nomenclature_rims_severavto(bolts_count, bolts_spacing);
