-- Migration: Create Severavto mapping table for CAE to commodity_id mapping
-- This table stores the relationship between Fourtochki CAE and Severavto commodity_id
-- Used for importing prices and stock from Severavto into tyres_prices_stock

CREATE TABLE IF NOT EXISTS severavto_tyres_mapping (
    id BIGSERIAL PRIMARY KEY,
    cae VARCHAR(50) NOT NULL UNIQUE,
    commodity_id VARCHAR(50) NOT NULL,
    match_priority INTEGER NOT NULL,  -- 1=exact, 2=brand+model+width+indices, 3=brand+width+indices, 4=brand+sizes
    match_details JSONB,              -- Store matching details for debugging
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Index for fast lookup by CAE
CREATE INDEX IF NOT EXISTS idx_severavto_mapping_cae ON severavto_tyres_mapping(cae);

-- Index for fast lookup by commodity_id
CREATE INDEX IF NOT EXISTS idx_severavto_mapping_commodity_id ON severavto_tyres_mapping(commodity_id);

-- Index for priority-based queries
CREATE INDEX IF NOT EXISTS idx_severavto_mapping_priority ON severavto_tyres_mapping(match_priority);

COMMENT ON TABLE severavto_tyres_mapping IS 'Mapping between Fourtochki CAE and Severavto commodity_id for price/stock synchronization';
COMMENT ON COLUMN severavto_tyres_mapping.match_priority IS '1=exact match (brand+model+sizes+indices), 2=brand+model+width+indices, 3=brand+width+indices, 4=brand+sizes';
COMMENT ON COLUMN severavto_tyres_mapping.match_details IS 'JSON with matching parameters: brand, model, width, height, diameter, load_index, speed_index';
