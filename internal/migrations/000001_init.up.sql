-- Create tyres table
CREATE TABLE IF NOT EXISTS tyres_api (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(255) NOT NULL,
    supplier VARCHAR(100) NOT NULL,
    brand VARCHAR(255) NOT NULL,
    model VARCHAR(255) NOT NULL,
    width INTEGER NOT NULL,
    height INTEGER NOT NULL,
    diameter INTEGER NOT NULL,
    load_index VARCHAR(10) DEFAULT '',
    speed_index VARCHAR(10) DEFAULT '',
    season VARCHAR(50) DEFAULT '',
    run_flat BOOLEAN DEFAULT FALSE,
    studded BOOLEAN DEFAULT FALSE,
    description TEXT DEFAULT '',
    raw_payload JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT tyres_api_code_supplier_unique UNIQUE (code, supplier)
);

-- Create indexes for tyres
CREATE INDEX IF NOT EXISTS idx_tyres_api_supplier ON tyres_api(supplier);
CREATE INDEX IF NOT EXISTS idx_tyres_api_code ON tyres_api(code);
CREATE INDEX IF NOT EXISTS idx_tyres_api_brand ON tyres_api(brand);
CREATE INDEX IF NOT EXISTS idx_tyres_api_model ON tyres_api(model);
CREATE INDEX IF NOT EXISTS idx_tyres_api_size ON tyres_api(width, height, diameter);
CREATE INDEX IF NOT EXISTS idx_tyres_api_season ON tyres_api(season);
CREATE INDEX IF NOT EXISTS idx_tyres_api_created_at ON tyres_api(created_at);
CREATE INDEX IF NOT EXISTS idx_tyres_api_updated_at ON tyres_api(updated_at);
CREATE INDEX IF NOT EXISTS idx_tyres_api_raw_payload ON tyres_api USING GIN (raw_payload);

-- Create rims table
CREATE TABLE IF NOT EXISTS rims_api (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(255) NOT NULL,
    supplier VARCHAR(100) NOT NULL,
    brand VARCHAR(255) NOT NULL,
    model VARCHAR(255) NOT NULL,
    width DECIMAL(5,2) NOT NULL,
    diameter DECIMAL(5,2) NOT NULL,
    bolt_pattern VARCHAR(50) DEFAULT '',
    "offset" INTEGER DEFAULT 0,
    center_bore DECIMAL(5,2) DEFAULT 0,
    color VARCHAR(100) DEFAULT '',
    description TEXT DEFAULT '',
    raw_payload JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT rims_api_code_supplier_unique UNIQUE (code, supplier)
);

-- Create indexes for rims
CREATE INDEX IF NOT EXISTS idx_rims_api_supplier ON rims_api(supplier);
CREATE INDEX IF NOT EXISTS idx_rims_api_code ON rims_api(code);
CREATE INDEX IF NOT EXISTS idx_rims_api_brand ON rims_api(brand);
CREATE INDEX IF NOT EXISTS idx_rims_api_model ON rims_api(model);
CREATE INDEX IF NOT EXISTS idx_rims_api_size ON rims_api(width, diameter);
CREATE INDEX IF NOT EXISTS idx_rims_api_bolt_pattern ON rims_api(bolt_pattern);
CREATE INDEX IF NOT EXISTS idx_rims_api_created_at ON rims_api(created_at);
CREATE INDEX IF NOT EXISTS idx_rims_api_updated_at ON rims_api(updated_at);
CREATE INDEX IF NOT EXISTS idx_rims_api_raw_payload ON rims_api USING GIN (raw_payload);

-- Create product_offers table
CREATE TABLE IF NOT EXISTS product_offers_api (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(255) NOT NULL,
    supplier VARCHAR(100) NOT NULL,
    product_type VARCHAR(50) NOT NULL, -- 'tyre' or 'rim'
    price DECIMAL(12,2) NOT NULL DEFAULT 0,
    currency VARCHAR(10) DEFAULT 'RUB',
    stock INTEGER NOT NULL DEFAULT 0,
    warehouse_code VARCHAR(100) DEFAULT '',
    warehouse_name VARCHAR(255) DEFAULT '',
    delivery_days INTEGER DEFAULT 0,
    min_order_qty INTEGER DEFAULT 1,
    available_for_order BOOLEAN DEFAULT TRUE,
    raw_payload JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT product_offers_api_code_supplier_warehouse_unique UNIQUE (code, supplier, warehouse_code)
);

-- Create indexes for product_offers
CREATE INDEX IF NOT EXISTS idx_product_offers_api_supplier ON product_offers_api(supplier);
CREATE INDEX IF NOT EXISTS idx_product_offers_api_code ON product_offers_api(code);
CREATE INDEX IF NOT EXISTS idx_product_offers_api_product_type ON product_offers_api(product_type);
CREATE INDEX IF NOT EXISTS idx_product_offers_api_price ON product_offers_api(price);
CREATE INDEX IF NOT EXISTS idx_product_offers_api_stock ON product_offers_api(stock);
CREATE INDEX IF NOT EXISTS idx_product_offers_api_available ON product_offers_api(available_for_order);
CREATE INDEX IF NOT EXISTS idx_product_offers_api_warehouse_code ON product_offers_api(warehouse_code);
CREATE INDEX IF NOT EXISTS idx_product_offers_api_created_at ON product_offers_api(created_at);
CREATE INDEX IF NOT EXISTS idx_product_offers_api_updated_at ON product_offers_api(updated_at);
CREATE INDEX IF NOT EXISTS idx_product_offers_api_raw_payload ON product_offers_api USING GIN (raw_payload);

-- Create sync_runs table
CREATE TABLE IF NOT EXISTS sync_runs_api (
    id BIGSERIAL PRIMARY KEY,
    provider VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, running, completed, failed
    started_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP WITH TIME ZONE,
    total_products INTEGER DEFAULT 0,
    success_products INTEGER DEFAULT 0,
    failed_products INTEGER DEFAULT 0,
    error_message TEXT DEFAULT '',
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for sync_runs
CREATE INDEX IF NOT EXISTS idx_sync_runs_api_provider ON sync_runs_api(provider);
CREATE INDEX IF NOT EXISTS idx_sync_runs_api_status ON sync_runs_api(status);
CREATE INDEX IF NOT EXISTS idx_sync_runs_api_started_at ON sync_runs_api(started_at);
CREATE INDEX IF NOT EXISTS idx_sync_runs_api_completed_at ON sync_runs_api(completed_at);
CREATE INDEX IF NOT EXISTS idx_sync_runs_api_created_at ON sync_runs_api(created_at);
CREATE INDEX IF NOT EXISTS idx_sync_runs_api_metadata ON sync_runs_api USING GIN (metadata);

-- Create trigger function for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for tyres
CREATE TRIGGER update_tyres_api_updated_at
    BEFORE UPDATE ON tyres_api
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Create triggers for rims
CREATE TRIGGER update_rims_api_updated_at
    BEFORE UPDATE ON rims_api
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Create triggers for product_offers
CREATE TRIGGER update_product_offers_api_updated_at
    BEFORE UPDATE ON product_offers_api
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
