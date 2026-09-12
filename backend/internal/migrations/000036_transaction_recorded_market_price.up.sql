-- What a holding was worth when the app first saw it.
--
-- The growth series nets every transaction out of the change in value, so what
-- is left is return. A purchase made while the series watches is netted at what
-- it cost (transaction_cash_flow, 000027): the owner paid that much, and
-- whatever the position is worth by the next snapshot is the market's doing.
--
-- A position loaded with history is not that. Its trades are dated long before
-- the series had a point to measure them against, and the holding walks in at
-- its market value while its flow said cost. The gain it had already made before
-- anyone here saw it became the return of the one day it was typed in — and
-- since most of what an account brings into the app already has gains, a
-- portfolio worth 15% over its cost reported +42% of "real" return.
--
-- The honest flow for such a trade is what the holding was worth when it came
-- in, the way an in-kind transfer is valued. The app keeps no price history, so
-- that value has to be written down when the row is recorded or it is gone:
-- recorded_market_price is the asset's price per unit at that moment, in the
-- portfolio's base currency. It is the valuation portfolio_summary applies
-- (000025) — same price fallback, same rate — so it is in the units of the
-- snapshot the holding lands in.
--
-- Every row gets one, not only the ones that turn out to be history: which kind
-- a trade is depends on where it lands in the series, and that is decided when
-- the series is read.
ALTER TABLE transactions
  ADD COLUMN IF NOT EXISTS recorded_market_price NUMERIC(20, 8);

-- A BEFORE trigger for the reason 000030 gives: the rule belongs to the schema,
-- not to each writer. A single transaction, a new position, a file import and a
-- fixture all get the same price without having to know the column exists. A
-- writer that names it keeps its own value.
--
-- INSERT only. The price is a fact about the moment the row was recorded, and
-- editing the row later does not move that moment.
CREATE OR REPLACE FUNCTION transactions_record_market_price()
RETURNS TRIGGER AS $$
BEGIN
  IF NEW.recorded_market_price IS NULL THEN
    SELECT COALESCE(uap.price, a.current_price, pe.price)
         * COALESCE(fx_rate(p.user_id, COALESCE(a.currency, pe.cost_currency), p.base_currency), 1)
      INTO NEW.recorded_market_price
      FROM portfolio_entries pe
      JOIN portfolios p ON p.id = pe.portfolio_id
      JOIN assets a     ON a.id = pe.asset_id
      LEFT JOIN user_asset_prices uap
        ON uap.asset_id = pe.asset_id AND uap.user_id = p.user_id
     WHERE pe.id = NEW.entry_id;
  END IF;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- The rows already here were recorded without it, and the price they had then
-- was never kept. Today's is the nearest thing there is: exact for whatever still
-- lands on the live point, and for older history off by however much the asset
-- moved since it was loaded — a far smaller error than the whole gain it carried
-- in, which is what netting it at cost booked as return.
--
-- User triggers are off for the backfill. recalculate_avg_cost fires on every
-- UPDATE and would rewrite each position's cost, and its updated_at, to the same
-- numbers it already has.
ALTER TABLE transactions DISABLE TRIGGER USER;

UPDATE transactions t
   SET recorded_market_price = COALESCE(uap.price, a.current_price, pe.price)
       * COALESCE(fx_rate(p.user_id, COALESCE(a.currency, pe.cost_currency), p.base_currency), 1)
  FROM portfolio_entries pe
  JOIN portfolios p ON p.id = pe.portfolio_id
  JOIN assets a     ON a.id = pe.asset_id
  LEFT JOIN user_asset_prices uap
    ON uap.asset_id = pe.asset_id AND uap.user_id = p.user_id
 WHERE pe.id = t.entry_id
   AND t.recorded_market_price IS NULL;

ALTER TABLE transactions ENABLE TRIGGER USER;

ALTER TABLE transactions ALTER COLUMN recorded_market_price SET NOT NULL;

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_class WHERE relkind = 'r' AND relname = 'transactions') THEN
    EXECUTE 'DROP TRIGGER IF EXISTS trg_transactions_record_market_price ON transactions';
    EXECUTE 'CREATE TRIGGER trg_transactions_record_market_price
             BEFORE INSERT ON transactions
             FOR EACH ROW EXECUTE FUNCTION transactions_record_market_price()';
  END IF;
END$$;

-- The sign convention for loaded history, the counterpart of
-- transaction_cash_flow for trades made under watch. Both growth series need it,
-- and two copies of it would be two chances to disagree.
--
--   buy, transfer_in       the holding comes in at its recorded value.
--   sell, transfer_out     it goes out at its recorded value, so a history that
--                          bought ten and sold four brings in the six still held.
--   dividend, interest,    nothing. Income paid and fees charged before the
--   fee                    series existed happened to none of its points.
--   split                  nothing, as in transaction_cash_flow.
CREATE OR REPLACE FUNCTION transaction_holding_flow(
  p_type       transaction_type,
  p_quantity   NUMERIC,
  p_unit_value NUMERIC
)
RETURNS NUMERIC
LANGUAGE sql
IMMUTABLE
AS $$
  SELECT CASE p_type
    WHEN 'buy'          THEN  COALESCE(p_quantity, 0) * COALESCE(p_unit_value, 0)
    WHEN 'transfer_in'  THEN  COALESCE(p_quantity, 0) * COALESCE(p_unit_value, 0)
    WHEN 'sell'         THEN -(COALESCE(p_quantity, 0) * COALESCE(p_unit_value, 0))
    WHEN 'transfer_out' THEN -(COALESCE(p_quantity, 0) * COALESCE(p_unit_value, 0))
    ELSE 0
  END;
$$;
