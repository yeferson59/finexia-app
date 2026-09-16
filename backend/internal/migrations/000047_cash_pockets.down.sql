-- The account key goes back to the platform and the currency alone, over every
-- row rather than over the main account only. A pocket's balances and rates are
-- gone with it, so nothing is left to collide with the main account's.
--
-- What a pocket held is lost: its balances, its movements and its rates are
-- deleted, not moved back into the main account, because merging two positions
-- would rewrite their average cost and their history.

DELETE FROM transactions t
USING portfolio_entries pe
WHERE pe.id = t.entry_id AND pe.pocket_id IS NOT NULL;

DELETE FROM portfolio_entries WHERE pocket_id IS NOT NULL;

-- A pocket's rates go before the column that says which pocket they are on.
-- Without them the restored key — one version per platform, currency and day —
-- would find the account's own version and its pocket's on the same day and
-- refuse to be created.
DELETE FROM cash_yield_rates WHERE pocket_id IS NOT NULL;

DROP INDEX IF EXISTS uk_cash_yield_rates_version;
ALTER TABLE cash_yield_rates DROP COLUMN IF EXISTS pocket_id;
ALTER TABLE cash_yield_rates
  ADD CONSTRAINT uk_cash_yield_rates_version UNIQUE (source_id, currency, effective_from);

DROP INDEX IF EXISTS idx_entries_portfolio_pocket;
DROP INDEX IF EXISTS idx_entries_pocket;
ALTER TABLE portfolio_entries DROP COLUMN IF EXISTS pocket_id;

DROP INDEX IF EXISTS idx_entries_portfolio_asset_source;
CREATE UNIQUE INDEX idx_entries_portfolio_asset_source
  ON portfolio_entries(portfolio_id, asset_id, COALESCE(source_id::TEXT, ''));

DROP TABLE IF EXISTS cash_pockets;
DROP TYPE IF EXISTS cash_pocket_kind;
