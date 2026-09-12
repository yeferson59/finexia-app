-- Keep what a deleted or re-sized transaction put into the growth series.
--
-- A snapshot freezes a portfolio's value on its day, and the growth series nets
-- flows out of those values by reading the transactions as they are now. The
-- two stay in step only while a transaction is never taken back. Delete a
-- position that has been in the series for two weeks and its value is still in
-- those two weeks of snapshots while its transactions are gone: the day it came
-- in reads as a gain nobody made, and the day it left as a loss nobody took. A
-- quantity edit does the same with the difference. One account carried +$30.61
-- on 2026-08-22 against −$29.58 on 2026-09-03, and +$19.44 on 2026-08-14 against
-- −$19.71 on 2026-09-04: about two points of "real" return that never happened.
--
-- retired_transactions keeps the version that left, and the series reads it
-- twice: as it was, on the point where it came in, so the past keeps the flow
-- its value needs; and reversed, at what the holding was worth when it went, on
-- the point where it left.
--
-- Only what moves a quantity is kept — buy, sell, transfer_in, transfer_out —
-- and only once a snapshot has seen it. Correcting a dividend, a fee, or just
-- the price or date of a trade never changed any snapshot's value, so those stay
-- retroactive, which is what fixing a typo should be.
CREATE TABLE IF NOT EXISTS retired_transactions (
  id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  portfolio_id          UUID NOT NULL REFERENCES portfolios(id) ON DELETE CASCADE,
  cost_currency         CHAR(3) NOT NULL,
  type                  transaction_type NOT NULL,
  quantity              NUMERIC(20, 8) NOT NULL,
  price                 NUMERIC(20, 8) NOT NULL,
  currency              CHAR(3) NOT NULL,
  fx_rate               NUMERIC(20, 8) NOT NULL,
  fees                  NUMERIC(20, 8) NOT NULL,
  fees_currency         CHAR(3) NOT NULL,
  transaction_date      DATE NOT NULL,
  -- When the retired version entered the series, and what a unit was worth then.
  recorded_at           TIMESTAMPTZ NOT NULL,
  recorded_market_price NUMERIC(20, 8) NOT NULL,
  -- When it left, and what a unit was worth then. Both prices are in the
  -- portfolio's base currency, like the snapshots.
  retired_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  retired_market_price  NUMERIC(20, 8) NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_retired_transactions_portfolio
  ON retired_transactions(portfolio_id);

-- When a transaction's current version entered the series. It starts as
-- created_at and moves only when an edit retires the previous version, so the
-- new one comes in on the day of the edit — the first day a snapshot holds it —
-- and not on the day the original was typed.
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS flow_recorded_at TIMESTAMPTZ;

-- recalculate_avg_cost fires on every UPDATE and would rewrite each position to
-- the numbers it already has; nothing about the backfill concerns it.
ALTER TABLE transactions DISABLE TRIGGER USER;
UPDATE transactions SET flow_recorded_at = created_at WHERE flow_recorded_at IS NULL;
ALTER TABLE transactions ENABLE TRIGGER USER;

ALTER TABLE transactions ALTER COLUMN flow_recorded_at SET NOT NULL;

-- A default that can read the row, for the reason 000030 gives.
CREATE OR REPLACE FUNCTION transactions_default_flow_recorded_at()
RETURNS TRIGGER AS $$
BEGIN
  IF NEW.flow_recorded_at IS NULL THEN
    NEW.flow_recorded_at := NEW.created_at;
  END IF;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- What one unit of a position's asset is worth now, in its portfolio's base
-- currency: portfolio_summary's valuation (000025) for a single position, and
-- the one transactions_record_market_price (000036) applies on insert.
CREATE OR REPLACE FUNCTION entry_unit_market_value(p_entry_id UUID)
RETURNS NUMERIC
LANGUAGE sql
STABLE
AS $$
  SELECT COALESCE(uap.price, a.current_price, pe.price)
       * COALESCE(fx_rate(p.user_id, COALESCE(a.currency, pe.cost_currency), p.base_currency), 1)
  FROM portfolio_entries pe
  JOIN portfolios p ON p.id = pe.portfolio_id
  JOIN assets a     ON a.id = pe.asset_id
  LEFT JOIN user_asset_prices uap
    ON uap.asset_id = pe.asset_id AND uap.user_id = p.user_id
  WHERE pe.id = p_entry_id;
$$;

-- Whether a version recorded at p_recorded_at is already in some snapshot: the
-- same test the growth series uses to decide where a transaction lands.
CREATE OR REPLACE FUNCTION transaction_version_seen(p_portfolio_id UUID, p_recorded_at TIMESTAMPTZ)
RETURNS BOOLEAN
LANGUAGE sql
STABLE
AS $$
  SELECT EXISTS (
    SELECT 1
    FROM portfolio_snapshots ps
    WHERE ps.portfolio_id = p_portfolio_id
      AND ps.created_at >= p_recorded_at
  );
$$;

-- Keeps a version that is leaving the series. Callers have checked that the
-- series saw it; this only declines the types that never moved a quantity.
CREATE OR REPLACE FUNCTION retire_transaction_version(
  p_old           transactions,
  p_portfolio_id  UUID,
  p_cost_currency CHAR(3)
)
RETURNS VOID
LANGUAGE plpgsql
AS $$
BEGIN
  IF p_old.type NOT IN ('buy', 'sell', 'transfer_in', 'transfer_out') THEN
    RETURN;
  END IF;

  INSERT INTO retired_transactions (
    portfolio_id, cost_currency, type, quantity, price, currency, fx_rate,
    fees, fees_currency, transaction_date,
    recorded_at, recorded_market_price, retired_market_price
  ) VALUES (
    p_portfolio_id, p_cost_currency, p_old.type, p_old.quantity, p_old.price,
    p_old.currency, p_old.fx_rate, p_old.fees, p_old.fees_currency,
    p_old.transaction_date, p_old.flow_recorded_at, p_old.recorded_market_price,
    COALESCE(entry_unit_market_value(p_old.entry_id), p_old.recorded_market_price)
  );
END;
$$;

-- A transaction deleted on its own. One deleted with its position arrives here
-- with the position already gone, and the position's trigger below has kept it.
CREATE OR REPLACE FUNCTION transactions_retire_on_delete()
RETURNS TRIGGER AS $$
DECLARE
  v_portfolio_id  UUID;
  v_cost_currency CHAR(3);
BEGIN
  SELECT pe.portfolio_id, pe.cost_currency
    INTO v_portfolio_id, v_cost_currency
    FROM portfolio_entries pe
   WHERE pe.id = OLD.entry_id;

  IF v_portfolio_id IS NOT NULL
     AND transaction_version_seen(v_portfolio_id, OLD.flow_recorded_at) THEN
    PERFORM retire_transaction_version(OLD, v_portfolio_id, v_cost_currency);
  END IF;

  RETURN OLD;
END;
$$ LANGUAGE plpgsql;

-- A position deleted, with every transaction it cascades to. A position
-- deleted along with its portfolio takes the portfolio's snapshots too, so there
-- is nothing left to keep in step.
CREATE OR REPLACE FUNCTION entries_retire_transactions()
RETURNS TRIGGER AS $$
DECLARE
  v_tx transactions%ROWTYPE;
BEGIN
  IF NOT EXISTS (SELECT 1 FROM portfolios WHERE id = OLD.portfolio_id) THEN
    RETURN OLD;
  END IF;

  FOR v_tx IN SELECT * FROM transactions WHERE entry_id = OLD.id LOOP
    IF transaction_version_seen(OLD.portfolio_id, v_tx.flow_recorded_at) THEN
      PERFORM retire_transaction_version(v_tx, OLD.portfolio_id, OLD.cost_currency);
    END IF;
  END LOOP;

  RETURN OLD;
END;
$$ LANGUAGE plpgsql;

-- An edit that changes how much of the asset the transaction moves. The old
-- version is kept and the new one enters the series now, at today's price; an
-- edit of anything else leaves the row where it was.
CREATE OR REPLACE FUNCTION transactions_retire_on_resize()
RETURNS TRIGGER AS $$
DECLARE
  v_portfolio_id  UUID;
  v_cost_currency CHAR(3);
BEGIN
  IF NEW.quantity IS NOT DISTINCT FROM OLD.quantity
     AND NEW.type IS NOT DISTINCT FROM OLD.type
     AND NEW.entry_id IS NOT DISTINCT FROM OLD.entry_id THEN
    RETURN NEW;
  END IF;

  SELECT pe.portfolio_id, pe.cost_currency
    INTO v_portfolio_id, v_cost_currency
    FROM portfolio_entries pe
   WHERE pe.id = OLD.entry_id;

  IF v_portfolio_id IS NULL
     OR NOT transaction_version_seen(v_portfolio_id, OLD.flow_recorded_at) THEN
    RETURN NEW;
  END IF;

  PERFORM retire_transaction_version(OLD, v_portfolio_id, v_cost_currency);

  NEW.flow_recorded_at := NOW();
  NEW.recorded_market_price := COALESCE(entry_unit_market_value(NEW.entry_id), NEW.recorded_market_price);

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_transactions_default_flow_recorded_at ON transactions;
CREATE TRIGGER trg_transactions_default_flow_recorded_at
  BEFORE INSERT ON transactions
  FOR EACH ROW EXECUTE FUNCTION transactions_default_flow_recorded_at();

DROP TRIGGER IF EXISTS trg_transactions_retire_on_delete ON transactions;
CREATE TRIGGER trg_transactions_retire_on_delete
  BEFORE DELETE ON transactions
  FOR EACH ROW EXECUTE FUNCTION transactions_retire_on_delete();

DROP TRIGGER IF EXISTS trg_transactions_retire_on_resize ON transactions;
CREATE TRIGGER trg_transactions_retire_on_resize
  BEFORE UPDATE ON transactions
  FOR EACH ROW EXECUTE FUNCTION transactions_retire_on_resize();

DROP TRIGGER IF EXISTS trg_entries_retire_transactions ON portfolio_entries;
CREATE TRIGGER trg_entries_retire_transactions
  BEFORE DELETE ON portfolio_entries
  FOR EACH ROW EXECUTE FUNCTION entries_retire_transactions();

-- Every flow the growth series nets, in one place for the account-wide and the
-- per-portfolio query: the transactions as they are, and each retired version
-- twice — as it came in, and reversed as it went. unit_value is the price a
-- holding flow uses, in the portfolio's base currency.
CREATE OR REPLACE VIEW growth_flow_events AS
SELECT
  pe.portfolio_id, pe.cost_currency, t.type, t.quantity, t.price, t.currency,
  t.fx_rate, t.fees, t.fees_currency, t.transaction_date,
  t.flow_recorded_at      AS recorded_at,
  t.recorded_market_price AS unit_value,
  FALSE                   AS reversal
FROM transactions t
JOIN portfolio_entries pe ON pe.id = t.entry_id
UNION ALL
SELECT
  r.portfolio_id, r.cost_currency, r.type, r.quantity, r.price, r.currency,
  r.fx_rate, r.fees, r.fees_currency, r.transaction_date,
  r.recorded_at, r.recorded_market_price, FALSE
FROM retired_transactions r
UNION ALL
SELECT
  r.portfolio_id, r.cost_currency, r.type, r.quantity, r.price, r.currency,
  r.fx_rate, r.fees, r.fees_currency, r.transaction_date,
  r.retired_at, r.retired_market_price, TRUE
FROM retired_transactions r;
