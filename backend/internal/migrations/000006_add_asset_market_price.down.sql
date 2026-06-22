-- 000006_add_asset_market_price.down.sql
-- Reverts the market price columns added to assets.

ALTER TABLE assets DROP COLUMN IF EXISTS price_updated_at;
ALTER TABLE assets DROP COLUMN IF EXISTS current_price;
