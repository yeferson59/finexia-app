-- The cap comes back from the first step at 0 %. Every other step has no column
-- to go back to, and is lost with the table.
ALTER TABLE cash_yield_rates
  ADD COLUMN IF NOT EXISTS max_balance NUMERIC(20, 8) CHECK (max_balance IS NULL OR max_balance > 0);

UPDATE cash_yield_rates r
SET max_balance = (
  SELECT MIN(t.from_balance) FROM cash_yield_rate_tiers t
  WHERE t.rate_id = r.id AND t.annual_rate = 0
);

DROP TABLE IF EXISTS cash_yield_rate_tiers;
