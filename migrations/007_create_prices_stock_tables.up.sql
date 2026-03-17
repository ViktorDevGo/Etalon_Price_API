-- Таблица цен и остатков для шин
CREATE TABLE IF NOT EXISTS tyres_prices_stock (
    id BIGSERIAL PRIMARY KEY,
    cae VARCHAR(50) NOT NULL, -- Код товара (связь с nomenclature_tyres)
    warehouse_id INTEGER NOT NULL, -- ID склада
    price INTEGER NOT NULL, -- Цена в копейках
    stock INTEGER NOT NULL, -- Остаток в штуках

    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- Уникальность: один товар на одном складе
    CONSTRAINT tyres_prices_stock_cae_warehouse_unique UNIQUE (cae, warehouse_id),

    -- Внешний ключ на номенклатуру
    CONSTRAINT tyres_prices_stock_cae_fkey FOREIGN KEY (cae)
        REFERENCES nomenclature_tyres(cae) ON DELETE CASCADE
);

-- Индексы для быстрого поиска
CREATE INDEX idx_tyres_prices_stock_cae ON tyres_prices_stock(cae);
CREATE INDEX idx_tyres_prices_stock_warehouse ON tyres_prices_stock(warehouse_id);
CREATE INDEX idx_tyres_prices_stock_updated ON tyres_prices_stock(updated_at DESC);

-- Таблица цен и остатков для дисков
CREATE TABLE IF NOT EXISTS rims_prices_stock (
    id BIGSERIAL PRIMARY KEY,
    cae VARCHAR(50) NOT NULL, -- Код товара (связь с nomenclature_rims)
    warehouse_id INTEGER NOT NULL, -- ID склада
    price INTEGER NOT NULL, -- Цена в копейках
    stock INTEGER NOT NULL, -- Остаток в штуках

    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- Уникальность: один товар на одном складе
    CONSTRAINT rims_prices_stock_cae_warehouse_unique UNIQUE (cae, warehouse_id),

    -- Внешний ключ на номенклатуру
    CONSTRAINT rims_prices_stock_cae_fkey FOREIGN KEY (cae)
        REFERENCES nomenclature_rims(cae) ON DELETE CASCADE
);

-- Индексы для быстрого поиска
CREATE INDEX idx_rims_prices_stock_cae ON rims_prices_stock(cae);
CREATE INDEX idx_rims_prices_stock_warehouse ON rims_prices_stock(warehouse_id);
CREATE INDEX idx_rims_prices_stock_updated ON rims_prices_stock(updated_at DESC);

-- Таблица истории обновлений цен и остатков
CREATE TABLE IF NOT EXISTS prices_stock_sync_runs (
    id BIGSERIAL PRIMARY KEY,
    sync_type VARCHAR(20) NOT NULL, -- 'tyres' или 'rims'
    started_at TIMESTAMP NOT NULL,
    completed_at TIMESTAMP,
    status VARCHAR(20) NOT NULL, -- 'running', 'completed', 'failed'
    total_items INTEGER DEFAULT 0, -- Всего товаров обработано
    total_warehouses INTEGER DEFAULT 0, -- Всего складских позиций
    new_items INTEGER DEFAULT 0, -- Новых записей
    updated_items INTEGER DEFAULT 0, -- Обновленных записей
    error_message TEXT,

    CONSTRAINT prices_stock_sync_runs_sync_type_check CHECK (sync_type IN ('tyres', 'rims')),
    CONSTRAINT prices_stock_sync_runs_status_check CHECK (status IN ('running', 'completed', 'failed'))
);

CREATE INDEX idx_prices_stock_sync_runs_type_started ON prices_stock_sync_runs(sync_type, started_at DESC);
