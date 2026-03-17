-- Таблица складов
CREATE TABLE IF NOT EXISTS warehouses (
    id INTEGER PRIMARY KEY, -- ID склада из API
    key VARCHAR(100) UNIQUE NOT NULL, -- Ключ склада
    name TEXT NOT NULL, -- Полное название
    short_name VARCHAR(100), -- Короткое название
    sto_id INTEGER, -- ID в системе STO
    background_color VARCHAR(20), -- Цвет для UI
    have_delivery BOOLEAN DEFAULT false, -- Есть доставка
    have_pickup BOOLEAN DEFAULT false, -- Есть самовывоз
    is_paid_delivery BOOLEAN DEFAULT false, -- Платная доставка
    logistic_days INTEGER DEFAULT 0, -- Дни логистики

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Индексы
CREATE INDEX idx_warehouses_key ON warehouses(key);
CREATE INDEX idx_warehouses_name ON warehouses(name);

-- Примечание: FK на warehouses будут добавлены после загрузки данных о складах
-- См. миграцию 009_add_warehouse_foreign_keys.up.sql

-- Таблица истории синхронизации складов
CREATE TABLE IF NOT EXISTS warehouses_sync_runs (
    id BIGSERIAL PRIMARY KEY,
    started_at TIMESTAMP NOT NULL,
    completed_at TIMESTAMP,
    status VARCHAR(20) NOT NULL, -- 'running', 'completed', 'failed'
    total_warehouses INTEGER DEFAULT 0,
    new_warehouses INTEGER DEFAULT 0,
    updated_warehouses INTEGER DEFAULT 0,
    error_message TEXT,

    CONSTRAINT warehouses_sync_runs_status_check CHECK (status IN ('running', 'completed', 'failed'))
);

CREATE INDEX idx_warehouses_sync_runs_started ON warehouses_sync_runs(started_at DESC);
