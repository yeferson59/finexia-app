-- Cash balances: what cash_interest (000040) means to the four functions that
-- decide what a transaction type does.
--
-- A cash balance is a position like any other — an asset of type cash, worth one
-- unit of its own currency, held on a platform inside a portfolio — so most of it
-- needed nothing new. A deposit is a transfer_in and a withdrawal a transfer_out:
-- both already move the quantity, and both already count as money the owner put
-- in or took out, which is what keeps a deposit from reading as return. The one
-- thing without a representation was the interest the balance earns, and each
-- rule below is where that interest has to be told apart from the 'interest' a
-- holding pays out:
--
--   recalculate_avg_cost        adds it to the balance. Not to the average cost,
--                               which is built from purchases alone: interest
--                               cost nothing, and a balance stays at one unit of
--                               its currency per unit held.
--   transaction_cash_flow       no flow. The money was earned inside the pool, so
--                               the value it adds is return. A paid-out interest
--                               reaches the same credit by leaving the pool; this
--                               one reaches it by staying.
--   transaction_holding_flow    in at its recorded value, like a transfer_in. An
--                               interest loaded with history is part of the
--                               balance that walks in, and leaving it out would
--                               book it as the return of the day it was typed —
--                               the mistake 000036 exists to prevent.
--   retire_transaction_version  kept once a snapshot has seen it, like every row
--                               that moves a quantity. Deleting it afterwards
--                               would otherwise read as a loss on that day.

CREATE OR REPLACE FUNCTION recalculate_avg_cost()
RETURNS TRIGGER AS $$
DECLARE
  v_entry_ids UUID[];
  v_entry_id  UUID;
  v_net_qty   NUMERIC;
  v_buy_qty   NUMERIC;
  v_buy_cost  NUMERIC;
BEGIN
  IF TG_OP = 'DELETE' THEN
    v_entry_ids := ARRAY[OLD.entry_id];
  ELSIF TG_OP = 'UPDATE' AND OLD.entry_id IS DISTINCT FROM NEW.entry_id THEN
    v_entry_ids := ARRAY[NEW.entry_id, OLD.entry_id];
  ELSE
    v_entry_ids := ARRAY[NEW.entry_id];
  END IF;

  FOREACH v_entry_id IN ARRAY v_entry_ids LOOP
    SELECT
      COALESCE(SUM(CASE
        WHEN type IN ('buy', 'transfer_in', 'cash_interest') THEN quantity
        WHEN type IN ('sell', 'transfer_out')                THEN -quantity
        ELSE 0 END), 0),
      COALESCE(SUM(CASE WHEN type IN ('buy', 'transfer_in') THEN quantity ELSE 0 END), 0),
      -- What a buy cost in the position's currency, at its own rate (000029).
      COALESCE(SUM(CASE
        WHEN type IN ('buy', 'transfer_in') THEN quantity * price * COALESCE(fx_rate, 1)
        ELSE 0 END), 0)
    INTO v_net_qty, v_buy_qty, v_buy_cost
    FROM transactions
    WHERE entry_id = v_entry_id;

    UPDATE portfolio_entries
    SET
      quantity   = GREATEST(v_net_qty, 0),
      price      = CASE WHEN v_buy_qty > 0 THEN v_buy_cost / v_buy_qty ELSE price END,
      updated_at = NOW()
    WHERE id = v_entry_id;
  END LOOP;

  IF TG_OP = 'DELETE' THEN
    RETURN OLD;
  END IF;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- The sign convention of 000027, with the one new arm. See that migration for
-- the rest.
CREATE OR REPLACE FUNCTION transaction_cash_flow(
  p_type     transaction_type,
  p_quantity NUMERIC,
  p_price    NUMERIC,
  p_fees     NUMERIC
)
RETURNS NUMERIC
LANGUAGE sql
IMMUTABLE
AS $$
  SELECT CASE p_type
    WHEN 'buy'           THEN  COALESCE(p_quantity, 0) * COALESCE(p_price, 0) + COALESCE(p_fees, 0)
    WHEN 'transfer_in'   THEN  COALESCE(p_quantity, 0) * COALESCE(p_price, 0) + COALESCE(p_fees, 0)
    WHEN 'fee'           THEN  COALESCE(p_fees, 0)
    WHEN 'sell'          THEN -(COALESCE(p_quantity, 0) * COALESCE(p_price, 0) - COALESCE(p_fees, 0))
    WHEN 'transfer_out'  THEN -(COALESCE(p_quantity, 0) * COALESCE(p_price, 0) - COALESCE(p_fees, 0))
    WHEN 'dividend'      THEN -(COALESCE(p_quantity, 0) * COALESCE(p_price, 0) - COALESCE(p_fees, 0))
    WHEN 'interest'      THEN -(COALESCE(p_quantity, 0) * COALESCE(p_price, 0) - COALESCE(p_fees, 0))
    WHEN 'cash_interest' THEN 0
    ELSE 0
  END;
$$;

-- The loaded-history convention of 000036, with the one new arm.
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
    WHEN 'buy'           THEN  COALESCE(p_quantity, 0) * COALESCE(p_unit_value, 0)
    WHEN 'transfer_in'   THEN  COALESCE(p_quantity, 0) * COALESCE(p_unit_value, 0)
    WHEN 'cash_interest' THEN  COALESCE(p_quantity, 0) * COALESCE(p_unit_value, 0)
    WHEN 'sell'          THEN -(COALESCE(p_quantity, 0) * COALESCE(p_unit_value, 0))
    WHEN 'transfer_out'  THEN -(COALESCE(p_quantity, 0) * COALESCE(p_unit_value, 0))
    ELSE 0
  END;
$$;

-- 000038's retirement, with cash_interest among the types that move a quantity.
CREATE OR REPLACE FUNCTION retire_transaction_version(
  p_old           transactions,
  p_portfolio_id  UUID,
  p_cost_currency CHAR(3)
)
RETURNS VOID
LANGUAGE plpgsql
AS $$
BEGIN
  IF p_old.type NOT IN ('buy', 'sell', 'transfer_in', 'transfer_out', 'cash_interest') THEN
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
