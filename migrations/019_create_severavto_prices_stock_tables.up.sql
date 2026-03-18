-- Migration 019: Create Severavto prices and stock tables
-- Создание таблиц цен и остатков для поставщика Северавто

-- Остатки и цены шин
CREATE TABLE IF NOT EXISTS tyres_prices_stock_severavto (
    id BIGSERIAL PRIMARY KEY,
    commodity_id VARCHAR(50) NOT NULL,         -- NNCOMMODIF (ID товара)
    territory_name VARCHAR(255) NOT NULL,      -- TERRITORY_NAME (название территории/склада)
    price INTEGER NOT NULL,                    -- NPRICE_PREPAYMENT (цена по предоплате, в рублях!)
    price_delayed INTEGER,                     -- NPRICE_DELAYED (цена по отсрочке, в рублях)
    price_rrp INTEGER,                         -- NPRICE_RRP (рекомендованная розничная цена)
    stock INTEGER NOT NULL DEFAULT 0,          -- NREST (остаток)
    place_type VARCHAR(50),                    -- SPLACE (Склад / Пункт выдачи)
    is_sellout BOOLEAN DEFAULT FALSE,          -- SSELLOUT (распродажа: Да/Нет)
    provider VARCHAR(100) DEFAULT 'Северавто', -- провайдер (для совместимости с общей схемой)
    updated_at TIMESTAMP DEFAULT NOW(),

    -- Уникальность: один товар на одной территории от одного провайдера
    CONSTRAINT tyres_prices_stock_severavto_unique UNIQUE (commodity_id, territory_name, provider),

    -- Связь с номенклатурой
    CONSTRAINT fk_tyres_prices_stock_severavto_commodity
        FOREIGN KEY (commodity_id)
        REFERENCES nomenclature_tyres_severavto(commodity_id)
        ON DELETE CASCADE
);

-- Индексы для быстрого поиска
CREATE INDEX idx_severavto_tyres_prices_commodity ON tyres_prices_stock_severavto(commodity_id);
CREATE INDEX idx_severavto_tyres_prices_territory ON tyres_prices_stock_severavto(territory_name);
CREATE INDEX idx_severavto_tyres_prices_stock ON tyres_prices_stock_severavto(stock);
CREATE INDEX idx_severavto_tyres_prices_updated ON tyres_prices_stock_severavto(updated_at);

-- Остатки и цены дисков
CREATE TABLE IF NOT EXISTS rims_prices_stock_severavto (
    id BIGSERIAL PRIMARY KEY,
    commodity_id VARCHAR(50) NOT NULL,         -- NNCOMMODIF (ID товара)
    territory_name VARCHAR(255) NOT NULL,      -- TERRITORY_NAME (название территории/склада)
    price INTEGER NOT NULL,                    -- NPRICE_PREPAYMENT (цена по предоплате, в рублях!)
    price_delayed INTEGER,                     -- NPRICE_DELAYED (цена по отсрочке, в рублях)
    price_rrp INTEGER,                         -- NPRICE_RRP (рекомендованная розничная цена)
    stock INTEGER NOT NULL DEFAULT 0,          -- NREST (остаток)
    place_type VARCHAR(50),                    -- SPLACE (Склад / Пункт выдачи)
    is_sellout BOOLEAN DEFAULT FALSE,          -- SSELLOUT (распродажа: Да/Нет)
    provider VARCHAR(100) DEFAULT 'Северавто', -- провайдер
    updated_at TIMESTAMP DEFAULT NOW(),

    -- Уникальность: один товар на одной территории от одного провайдера
    CONSTRAINT rims_prices_stock_severavto_unique UNIQUE (commodity_id, territory_name, provider),

    -- Связь с номенклатурой
    CONSTRAINT fk_rims_prices_stock_severavto_commodity
        FOREIGN KEY (commodity_id)
        REFERENCES nomenclature_rims_severavto(commodity_id)
        ON DELETE CASCADE
);

-- Индексы для быстрого поиска
CREATE INDEX idx_severavto_rims_prices_commodity ON rims_prices_stock_severavto(commodity_id);
CREATE INDEX idx_severavto_rims_prices_territory ON rims_prices_stock_severavto(territory_name);
CREATE INDEX idx_severavto_rims_prices_stock ON rims_prices_stock_severavto(stock);
CREATE INDEX idx_severavto_rims_prices_updated ON rims_prices_stock_severavto(updated_at);
