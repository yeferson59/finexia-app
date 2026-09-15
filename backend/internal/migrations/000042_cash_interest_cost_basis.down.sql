-- Restore recalculate_avg_cost as 000041 left it, put the balances that hold
-- interest back on its formula, and drop the two functions.

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

-- A balance that holds only interest has no purchase to average, and was opened
-- at one.
UPDATE portfolio_entries pe
SET price = COALESCE((
  SELECT SUM(t.quantity * t.price * COALESCE(t.fx_rate, 1)) / NULLIF(SUM(t.quantity), 0)
  FROM transactions t
  WHERE t.entry_id = pe.id
    AND t.type IN ('buy', 'transfer_in')
), 1)
FROM assets a
WHERE a.id = pe.asset_id
  AND a.asset_type = 'cash'
  AND EXISTS (
    SELECT 1 FROM transactions t WHERE t.entry_id = pe.id AND t.type = 'cash_interest'
  );

DROP FUNCTION IF EXISTS cash_entry_at_par(UUID);
DROP FUNCTION IF EXISTS cash_entry_avg_cost(UUID);
