-- 000006_add_asset_market_price.up.sql
-- Adds market price tracking to assets so portfolios can be valued at market
-- instead of only at cost. Idempotent: safe to run multiple times.

ALTER TABLE assets ADD COLUMN IF NOT EXISTS current_price NUMERIC(20, 8);
ALTER TABLE assets ADD COLUMN IF NOT EXISTS price_updated_at TIMESTAMPTZ;

-- Seed reference market prices for the assets created in 000004 so the
-- holdings view shows market valuation out of the box. Only sets the price
-- when it hasn't been set yet to avoid overwriting live updates.
UPDATE assets SET current_price = 195.00,    price_updated_at = NOW() WHERE id = 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa' AND current_price IS NULL;
UPDATE assets SET current_price = 420.00,    price_updated_at = NOW() WHERE id = 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb' AND current_price IS NULL;
UPDATE assets SET current_price = 540.00,    price_updated_at = NOW() WHERE id = 'cccccccc-cccc-cccc-cccc-cccccccccccc' AND current_price IS NULL;
UPDATE assets SET current_price = 65000.00,  price_updated_at = NOW() WHERE id = 'dddddddd-dddd-dddd-dddd-dddddddddddd' AND current_price IS NULL;
UPDATE assets SET current_price = 3500.00,   price_updated_at = NOW() WHERE id = 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee' AND current_price IS NULL;
UPDATE assets SET current_price = 72.50,     price_updated_at = NOW() WHERE id = 'ffffffff-ffff-ffff-ffff-ffffffffffff' AND current_price IS NULL;
