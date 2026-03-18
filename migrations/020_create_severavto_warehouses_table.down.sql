-- Migration 020 Down: Drop Severavto warehouses table

DROP INDEX IF EXISTS idx_severavto_warehouses_place_type;
DROP INDEX IF EXISTS idx_severavto_warehouses_region;
DROP INDEX IF EXISTS idx_severavto_warehouses_territory;

DROP TABLE IF EXISTS warehouses_severavto;
