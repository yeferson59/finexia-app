DROP INDEX IF EXISTS idx_assets_created_by;

DROP TABLE IF EXISTS user_catalog_assets;

ALTER TABLE assets
  DROP COLUMN IF EXISTS created_by,
  DROP COLUMN IF EXISTS is_curated;
