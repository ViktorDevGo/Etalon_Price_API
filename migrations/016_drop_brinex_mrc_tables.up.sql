-- Удаление всех таблиц Brinex и MRC API

-- Удаление таблиц Brinex (из миграций 011 и 012)
DROP TABLE IF EXISTS tyres_prices_stock_brinex CASCADE;
DROP TABLE IF EXISTS rims_prices_stock_brinex CASCADE;
DROP TABLE IF EXISTS nomenclature_tyres_brinex CASCADE;
DROP TABLE IF EXISTS nomenclature_rims_brinex CASCADE;
DROP TABLE IF EXISTS brinex_sync_runs CASCADE;
DROP TABLE IF EXISTS brinex_warehouses CASCADE;
DROP TABLE IF EXISTS brinex_product_images CASCADE;

-- Удаление таблиц MRC API (из миграции 014)
DROP TABLE IF EXISTS mrc_api CASCADE;
DROP TABLE IF EXISTS mrc_api_sync_runs CASCADE;
