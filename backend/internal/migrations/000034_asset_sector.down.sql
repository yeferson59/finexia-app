DROP INDEX IF EXISTS idx_assets_sector;

ALTER TABLE assets
  DROP COLUMN IF EXISTS sector;
