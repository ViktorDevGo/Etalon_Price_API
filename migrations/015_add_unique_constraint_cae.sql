-- Add UNIQUE constraint on CAE column to prevent duplicates
-- Migration 015: Simplify duplicate logic - check only by CAE

-- First, ensure no duplicates exist (keep only the latest record for each CAE)
DELETE FROM nomenclature_tyres
WHERE id IN (
    SELECT id
    FROM (
        SELECT id, cae,
               ROW_NUMBER() OVER (PARTITION BY cae ORDER BY id DESC) as rn
        FROM nomenclature_tyres
    ) t
    WHERE rn > 1
);

DELETE FROM nomenclature_rims
WHERE id IN (
    SELECT id
    FROM (
        SELECT id, cae,
               ROW_NUMBER() OVER (PARTITION BY cae ORDER BY id DESC) as rn
        FROM nomenclature_rims
    ) t
    WHERE rn > 1
);

-- Add UNIQUE constraint on cae column for nomenclature_tyres
ALTER TABLE nomenclature_tyres
ADD CONSTRAINT nomenclature_tyres_cae_unique UNIQUE (cae);

-- Add UNIQUE constraint on cae column for nomenclature_rims
ALTER TABLE nomenclature_rims
ADD CONSTRAINT nomenclature_rims_cae_unique UNIQUE (cae);

-- Add comments
COMMENT ON CONSTRAINT nomenclature_tyres_cae_unique ON nomenclature_tyres IS 'Ensures each article code (CAE) is unique in tyres nomenclature';
COMMENT ON CONSTRAINT nomenclature_rims_cae_unique ON nomenclature_rims IS 'Ensures each article code (CAE) is unique in rims nomenclature';
