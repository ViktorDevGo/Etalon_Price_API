-- Migration 019 Down: Drop Severavto prices/stock tables

DROP INDEX IF EXISTS idx_severavto_rims_prices_updated;
DROP INDEX IF EXISTS idx_severavto_rims_prices_stock;
DROP INDEX IF EXISTS idx_severavto_rims_prices_territory;
DROP INDEX IF EXISTS idx_severavto_rims_prices_commodity;

DROP TABLE IF EXISTS rims_prices_stock_severavto;

DROP INDEX IF EXISTS idx_severavto_tyres_prices_updated;
DROP INDEX IF EXISTS idx_severavto_tyres_prices_stock;
DROP INDEX IF EXISTS idx_severavto_tyres_prices_territory;
DROP INDEX IF EXISTS idx_severavto_tyres_prices_commodity;

DROP TABLE IF EXISTS tyres_prices_stock_severavto;
