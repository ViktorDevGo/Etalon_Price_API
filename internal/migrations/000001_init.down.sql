-- Drop triggers
DROP TRIGGER IF EXISTS update_tyres_api_updated_at ON tyres_api;
DROP TRIGGER IF EXISTS update_rims_api_updated_at ON rims_api;
DROP TRIGGER IF EXISTS update_product_offers_api_updated_at ON product_offers_api;

-- Drop trigger function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables
DROP TABLE IF EXISTS sync_runs_api;
DROP TABLE IF EXISTS product_offers_api;
DROP TABLE IF EXISTS rims_api;
DROP TABLE IF EXISTS tyres_api;
