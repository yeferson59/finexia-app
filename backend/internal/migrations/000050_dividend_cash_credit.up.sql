-- Dividends credited to cash: the link from a credit to the dividend it pays,
-- and what cash_dividend (000049) means to the functions that decide what a
-- transaction type does.
--
-- The credit is a row on the cash balance rather than a flag on the dividend,
-- because every rule that reads a balance reads its rows: its quantity, the walk
-- that prices it, the interest it earns, the versions a snapshot saw. A row is
-- the one thing all of them already understand.
--
-- dividend_id ties the credit to its dividend. A dividend has at most one, only
-- a credit has one, and it cascades: deleting a dividend, or the position that
-- holds it, takes its credit along. The app checks first that the balance can
-- spare it, the way it checks a withdrawal.
--
-- What each rule makes of it:
--
--   recalculate_avg_cost         adds it to the balance.
--   cash_entry_avg_cost          units that cost nothing, like interest (000042).
--                                The dividend is gain the balance holds, not
--                                money the owner put in, so a figure that
--                                subtracts the cost sees it.
--   cash_entry_at_par            one unit per unit, like every row the cash
--                                writers write, so the balance keeps taking
--                                deposits after one.
--   transaction_cash_flow        money in, at its amount. It is what keeps the
--                                dividend from counting twice: the dividend
--                                flows out of the holding (000027) and its
--                                credit back into the balance, on the same day
--                                and in the same currency, so the two cancel and
--                                what is left is the value the balance gained —
--                                return, once. The portfolio earned exactly
--                                what it earned before; the money is just
--                                somewhere now.
--   transaction_holding_flow     in at its recorded value, like cash_interest. A
--                                dividend loaded with history carries no flow
--                                (000036), and the cash it left walks in with
--                                the balance.
--   retire_transaction_version   kept once a snapshot has seen it, like every
--                                row that moves a quantity.
--   cash_entry_balance_at_close  counts from its date, like a deposit: the money
--                                earns interest from the day it lands.

ALTER TABLE transactions
  ADD COLUMN IF NOT EXISTS dividend_id UUID REFERENCES transactions(id) ON DELETE CASCADE;

CREATE UNIQUE INDEX IF NOT EXISTS uk_transactions_dividend_credit
  ON transactions(dividend_id) WHERE dividend_id IS NOT NULL;

ALTER TABLE transactions DROP CONSTRAINT IF EXISTS chk_transactions_dividend_credit;
ALTER TABLE transactions
  ADD CONSTRAINT chk_transactions_dividend_credit
  CHECK ((type = 'cash_dividend') = (dividend_id IS NOT NULL));

-- 000042's recalculation, with cash_dividend among the rows that move the
-- quantity.
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
        WHEN type IN ('buy', 'transfer_in', 'cash_interest', 'cash_dividend') THEN quantity
        WHEN type IN ('sell', 'transfer_out')                                 THEN -quantity
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
      price      = COALESCE(
                     cash_entry_avg_cost(v_entry_id),
                     CASE WHEN v_buy_qty > 0 THEN v_buy_cost / v_buy_qty ELSE price END
                   ),
      updated_at = NOW()
    WHERE id = v_entry_id;
  END LOOP;

  IF TG_OP = 'DELETE' THEN
    RETURN OLD;
  END IF;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 000042's walk, taken by a balance that holds either kind of income. Both add
-- units and nothing to the cost; a withdrawal takes its share of both.
CREATE OR REPLACE FUNCTION cash_entry_avg_cost(p_entry_id UUID)
RETURNS NUMERIC
LANGUAGE plpgsql
STABLE
AS $$
DECLARE
  v_units NUMERIC := 0;
  v_cost  NUMERIC := 0;
  t       RECORD;
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM transactions tx
    JOIN portfolio_entries pe ON pe.id = tx.entry_id
    JOIN assets a             ON a.id = pe.asset_id
    WHERE tx.entry_id = p_entry_id
      AND tx.type IN ('cash_interest', 'cash_dividend')
      AND a.asset_type = 'cash'
  ) THEN
    RETURN NULL;
  END IF;

  FOR t IN
    SELECT type, quantity, price, COALESCE(fx_rate, 1) AS fx_rate
    FROM transactions
    WHERE entry_id = p_entry_id
    ORDER BY transaction_date, created_at, id
  LOOP
    IF t.type IN ('buy', 'transfer_in') THEN
      v_units := v_units + t.quantity;
      v_cost  := v_cost + t.quantity * t.price * t.fx_rate;
    ELSIF t.type IN ('cash_interest', 'cash_dividend') THEN
      v_units := v_units + t.quantity;
    ELSIF t.type IN ('sell', 'transfer_out') THEN
      IF t.quantity >= v_units THEN
        v_units := 0;
        v_cost  := 0;
      ELSE
        v_cost  := v_cost * (v_units - t.quantity) / v_units;
        v_units := v_units - t.quantity;
      END IF;
    END IF;
  END LOOP;

  IF v_units > 0 THEN
    RETURN v_cost / v_units;
  END IF;

  RETURN NULL;
END;
$$;

-- 000042's test, with a credit among the rows that make a balance the app's
-- own: one holding only dividends sits at a price of zero, like one holding
-- only interest.
CREATE OR REPLACE FUNCTION cash_entry_at_par(p_entry_id UUID)
RETURNS BOOLEAN
LANGUAGE sql
STABLE
AS $$
  SELECT EXISTS (
    SELECT 1
    FROM portfolio_entries pe
    WHERE pe.id = p_entry_id
      AND NOT EXISTS (
        SELECT 1
        FROM transactions t
        WHERE t.entry_id = pe.id
          AND t.type IN ('buy', 'transfer_in', 'sell', 'transfer_out')
          AND (t.price <> 1 OR COALESCE(t.fx_rate, 1) <> 1 OR t.currency <> pe.cost_currency)
      )
      AND (
        pe.price = 1
        OR EXISTS (
          SELECT 1
          FROM transactions t
          WHERE t.entry_id = pe.id
            AND t.type IN ('buy', 'transfer_in', 'sell', 'transfer_out', 'cash_interest', 'cash_dividend')
        )
      )
  );
$$;

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
    WHEN 'cash_dividend' THEN  COALESCE(p_quantity, 0) * COALESCE(p_price, 0)
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
    WHEN 'cash_dividend' THEN  COALESCE(p_quantity, 0) * COALESCE(p_unit_value, 0)
    WHEN 'sell'          THEN -(COALESCE(p_quantity, 0) * COALESCE(p_unit_value, 0))
    WHEN 'transfer_out'  THEN -(COALESCE(p_quantity, 0) * COALESCE(p_unit_value, 0))
    ELSE 0
  END;
$$;

-- 000038's retirement, with cash_dividend among the types that move a quantity.
CREATE OR REPLACE FUNCTION retire_transaction_version(
  p_old           transactions,
  p_portfolio_id  UUID,
  p_cost_currency CHAR(3)
)
RETURNS VOID
LANGUAGE plpgsql
AS $$
BEGIN
  IF p_old.type NOT IN ('buy', 'sell', 'transfer_in', 'transfer_out', 'cash_interest', 'cash_dividend') THEN
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

-- 000045's balance at the close of a day, with a credit counted from its date.
-- It is not the ledger's, so no accrual links to it and it reads like a
-- deposit.
CREATE OR REPLACE FUNCTION cash_entry_balance_at_close(p_entry_id UUID, p_day DATE)
RETURNS NUMERIC
LANGUAGE sql
STABLE
AS $$
  SELECT
    COALESCE((
      SELECT SUM(CASE
        WHEN t.type IN ('buy', 'transfer_in', 'cash_interest', 'cash_dividend') THEN t.quantity
        WHEN t.type IN ('sell', 'transfer_out')                                 THEN -t.quantity
        ELSE 0 END)
      FROM transactions t
      LEFT JOIN LATERAL (
        SELECT MAX(ac.accrual_date) AS last_day
        FROM cash_interest_accruals ac
        WHERE ac.transaction_id = t.id
      ) credit ON TRUE
      WHERE t.entry_id = p_entry_id
        AND COALESCE(credit.last_day + 1, t.transaction_date) <= p_day
    ), 0)
    +
    COALESCE((
      SELECT SUM(ac.net_amount)
      FROM cash_interest_accruals ac
      LEFT JOIN LATERAL (
        SELECT MAX(paid.accrual_date) AS last_day
        FROM cash_interest_accruals paid
        WHERE paid.transaction_id = ac.transaction_id
      ) credit ON TRUE
      WHERE ac.entry_id = p_entry_id
        AND ac.accrual_date < p_day
        AND (ac.status = 'pending' OR credit.last_day >= p_day)
    ), 0);
$$;
