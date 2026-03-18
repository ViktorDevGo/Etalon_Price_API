-- Migration 018 Down: Drop Severavto nomenclature tables

DROP INDEX IF EXISTS idx_severavto_rims_pcd;
DROP INDEX IF EXISTS idx_severavto_rims_diameter;
DROP INDEX IF EXISTS idx_severavto_rims_article;
DROP INDEX IF EXISTS idx_severavto_rims_model;
DROP INDEX IF EXISTS idx_severavto_rims_brand;
DROP INDEX IF EXISTS idx_severavto_rims_commodity_id;

DROP TABLE IF EXISTS nomenclature_rims_severavto;

DROP INDEX IF EXISTS idx_severavto_tyres_dims;
DROP INDEX IF EXISTS idx_severavto_tyres_season;
DROP INDEX IF EXISTS idx_severavto_tyres_article;
DROP INDEX IF EXISTS idx_severavto_tyres_model;
DROP INDEX IF EXISTS idx_severavto_tyres_brand;
DROP INDEX IF EXISTS idx_severavto_tyres_commodity_id;

DROP TABLE IF EXISTS nomenclature_tyres_severavto;
