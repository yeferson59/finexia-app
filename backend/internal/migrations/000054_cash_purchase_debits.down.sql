-- Restore the seven functions as 000052 left them, and the link as it was
-- there. The purchases' debits stay, as 000052's down migration leaves the
-- sales' credits: with no rule giving cash_purchase a meaning, each reads as a
-- row that moves no quantity and no flow, and loses its link first so the
-- restored constraint — only a cash_dividend or a cash_sale has one — holds.
--
-- The money stays out of the balance it left, which is the honest half of an
-- undo: the shares it bought are still there. A balance that reads high after
-- this held a purchase whose debit no longer counts, and the row is still in
-- its ledger to say so.

UPDATE transactions SET credited_from = NULL WHERE type = 'cash_purchase';

ALTER TABLE transactions DROP CONSTRAINT IF EXISTS chk_transactions_cash_link;
ALTER TABLE transactions
  ADD CONSTRAINT chk_transactions_cash_credit
  CHECK (
    (type IN ('cash_dividend', 'cash_sale')) = (credited_from IS NOT NULL)
    AND (type = 'cash_sale') = (cost_basis IS NOT NULL)
  );

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
        WHEN type IN ('buy', 'transfer_in', 'cash_interest', 'cash_dividend', 'cash_sale') THEN quantity
        WHEN type IN ('sell', 'transfer_out')                                              THEN -quantity
        ELSE 0 END), 0),
      COALESCE(SUM(CASE WHEN type IN ('buy', 'transfer_in') THEN quantity ELSE 0 END), 0),
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
                     CASE
                       WHEN v_buy_qty > 0 THEN v_buy_cost / v_buy_qty
                       WHEN EXISTS (
                         SELECT 1 FROM assets a
                         WHERE a.id = portfolio_entries.asset_id
                           AND a.asset_type = 'cash'
                           AND a.currency = portfolio_entries.cost_currency
                       ) THEN 1
                       ELSE price
                     END
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
      AND tx.type IN ('cash_interest', 'cash_dividend', 'cash_sale')
      AND a.asset_type = 'cash'
  ) THEN
    RETURN NULL;
  END IF;

  FOR t IN
    SELECT type, quantity, price, COALESCE(fx_rate, 1) AS fx_rate, cost_basis
    FROM transactions
    WHERE entry_id = p_entry_id
    ORDER BY transaction_date, created_at, id
  LOOP
    IF t.type IN ('buy', 'transfer_in') THEN
      v_units := v_units + t.quantity;
      v_cost  := v_cost + t.quantity * t.price * t.fx_rate;
    ELSIF t.type IN ('cash_interest', 'cash_dividend') THEN
      v_units := v_units + t.quantity;
    ELSIF t.type = 'cash_sale' THEN
      v_units := v_units + t.quantity;
      v_cost  := v_cost + COALESCE(t.cost_basis, t.quantity);
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
            AND t.type IN ('buy', 'transfer_in', 'sell', 'transfer_out',
                           'cash_interest', 'cash_dividend', 'cash_sale')
        )
      )
  );
$$;

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
    WHEN 'cash_sale'     THEN  COALESCE(p_quantity, 0) * COALESCE(p_price, 0)
    ELSE 0
  END;
$$;

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
    WHEN 'cash_sale'     THEN  COALESCE(p_quantity, 0) * COALESCE(p_unit_value, 0)
    WHEN 'sell'          THEN -(COALESCE(p_quantity, 0) * COALESCE(p_unit_value, 0))
    WHEN 'transfer_out'  THEN -(COALESCE(p_quantity, 0) * COALESCE(p_unit_value, 0))
    ELSE 0
  END;
$$;

CREATE OR REPLACE FUNCTION retire_transaction_version(
  p_old           transactions,
  p_portfolio_id  UUID,
  p_cost_currency CHAR(3)
)
RETURNS VOID
LANGUAGE plpgsql
AS $$
BEGIN
  IF p_old.type NOT IN ('buy', 'sell', 'transfer_in', 'transfer_out',
                        'cash_interest', 'cash_dividend', 'cash_sale') THEN
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

CREATE OR REPLACE FUNCTION cash_entry_balance_at_close(p_entry_id UUID, p_day DATE)
RETURNS NUMERIC
LANGUAGE sql
STABLE
AS $$
  SELECT
    COALESCE((
      SELECT SUM(CASE
        WHEN t.type IN ('buy', 'transfer_in', 'cash_interest', 'cash_dividend', 'cash_sale') THEN t.quantity
        WHEN t.type IN ('sell', 'transfer_out')                                              THEN -t.quantity
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
