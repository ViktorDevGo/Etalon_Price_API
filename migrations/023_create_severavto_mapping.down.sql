-- Rollback: Drop Severavto mapping table

DROP INDEX IF EXISTS idx_severavto_mapping_priority;
DROP INDEX IF EXISTS idx_severavto_mapping_commodity_id;
DROP INDEX IF EXISTS idx_severavto_mapping_cae;
DROP TABLE IF EXISTS severavto_tyres_mapping;
