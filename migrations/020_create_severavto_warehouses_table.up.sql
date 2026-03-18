-- Migration 020: Create Severavto warehouses table
-- Создание таблицы территорий/складов для поставщика Северавто

CREATE TABLE IF NOT EXISTS warehouses_severavto (
    id BIGSERIAL PRIMARY KEY,
    territory_name VARCHAR(255) UNIQUE NOT NULL, -- TERRITORY_NAME (уникальное название территории)
    region VARCHAR(255),                         -- регион (Новосибирский, Московский и т.д.)
    place_type VARCHAR(50),                      -- SPLACE (Склад / Пункт выдачи)
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Индексы для быстрого поиска
CREATE INDEX idx_severavto_warehouses_territory ON warehouses_severavto(territory_name);
CREATE INDEX idx_severavto_warehouses_region ON warehouses_severavto(region);
CREATE INDEX idx_severavto_warehouses_place_type ON warehouses_severavto(place_type);

-- Комментарии
COMMENT ON TABLE warehouses_severavto IS 'Справочник территорий/складов Северавто (177 территорий по всей России)';
COMMENT ON COLUMN warehouses_severavto.territory_name IS 'Название территории (уникальное)';
COMMENT ON COLUMN warehouses_severavto.region IS 'Регион (напр. Новосибирский регион, Московский регион)';
COMMENT ON COLUMN warehouses_severavto.place_type IS 'Тип места: Склад или Пункт выдачи';
