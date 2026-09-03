-- Restore the 000023 average-cost function, then drop the column it reads.
--
-- Order matters: the function has to stop referencing fx_rate before the column
-- goes, or the DROP fails on the dependency. Rolling back is lossy by nature —
-- a position whose trades were settled across two currencies loses the rates
-- that reconciled them, and its average cost reverts to an average of prices in
-- the trade currency, which is the bug this migration fixed. Nothing here can
-- prevent that; the rollback simply puts the schema back.
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
        WHEN type IN ('buy', 'transfer_in')   THEN quantity
        WHEN type IN ('sell', 'transfer_out') THEN -quantity
        ELSE 0 END), 0),
      COALESCE(SUM(CASE WHEN type IN ('buy', 'transfer_in') THEN quantity ELSE 0 END), 0),
      COALESCE(SUM(CASE WHEN type IN ('buy', 'transfer_in') THEN quantity * price ELSE 0 END), 0)
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

DROP FUNCTION IF EXISTS transaction_fees_in_cost(NUMERIC, CHAR, CHAR, NUMERIC);

ALTER TABLE transactions DROP CONSTRAINT IF EXISTS chk_transactions_fx_rate_positive;
ALTER TABLE transactions DROP COLUMN IF EXISTS fees_currency;
ALTER TABLE transactions DROP COLUMN IF EXISTS fx_rate;
