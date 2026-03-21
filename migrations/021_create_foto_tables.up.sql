-- Migration 021: Create tables for product photos from Severavto API
-- Таблицы для хранения URL фотографий товаров

-- Таблица фотографий шин
CREATE TABLE IF NOT EXISTS foto_tyres (
    id BIGSERIAL PRIMARY KEY,
    commodity_id VARCHAR(50) UNIQUE NOT NULL,  -- ID товара из Северавто API
    brand VARCHAR(255),                        -- Бренд
    model VARCHAR(255),                        -- Модель
    size VARCHAR(50),                          -- Размер (195/65 R15)
    season VARCHAR(50),                        -- Сезон
    picture_url TEXT,                          -- URL фотографии
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Индексы для быстрого поиска
CREATE INDEX idx_foto_tyres_commodity_id ON foto_tyres(commodity_id);
CREATE INDEX idx_foto_tyres_brand ON foto_tyres(brand);
CREATE INDEX idx_foto_tyres_model ON foto_tyres(model);
CREATE INDEX idx_foto_tyres_picture_url ON foto_tyres(picture_url) WHERE picture_url IS NOT NULL;

-- Таблица фотографий дисков
CREATE TABLE IF NOT EXISTS foto_rims (
    id BIGSERIAL PRIMARY KEY,
    commodity_id VARCHAR(50) UNIQUE NOT NULL,  -- ID товара из Северавто API
    brand VARCHAR(255),                        -- Бренд
    model VARCHAR(255),                        -- Модель
    size VARCHAR(50),                          -- Размер (7.0x17)
    picture_url TEXT,                          -- URL фотографии
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Индексы для быстрого поиска
CREATE INDEX idx_foto_rims_commodity_id ON foto_rims(commodity_id);
CREATE INDEX idx_foto_rims_brand ON foto_rims(brand);
CREATE INDEX idx_foto_rims_model ON foto_rims(model);
CREATE INDEX idx_foto_rims_picture_url ON foto_rims(picture_url) WHERE picture_url IS NOT NULL;

-- Комментарии к таблицам
COMMENT ON TABLE foto_tyres IS 'URL фотографий шин из Северавто API';
COMMENT ON TABLE foto_rims IS 'URL фотографий дисков из Северавто API';
COMMENT ON COLUMN foto_tyres.picture_url IS 'URL формата: https://webmim.svrauto.ru/images/{ID}_1.jpg';
COMMENT ON COLUMN foto_rims.picture_url IS 'URL формата: https://webmim.svrauto.ru/images/{ID}_1.jpg';
