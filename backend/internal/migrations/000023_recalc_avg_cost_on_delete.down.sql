-- Restore the 000003 behaviour: recompute on INSERT and UPDATE only, from
-- NEW.entry_id alone. Positions repaired by the up migration keep their
-- corrected values — the drift only reappears on the next deletion.

CREATE OR REPLACE FUNCTION recalculate_avg_cost()
RETURNS TRIGGER AS $$
DECLARE
  v_net_qty   NUMERIC;
  v_buy_qty   NUMERIC;
  v_buy_cost  NUMERIC;
BEGIN
  SELECT
    COALESCE(SUM(CASE
      WHEN type IN ('buy', 'transfer_in')   THEN quantity
      WHEN type IN ('sell', 'transfer_out') THEN -quantity
      ELSE 0 END), 0),
    COALESCE(SUM(CASE WHEN type IN ('buy', 'transfer_in') THEN quantity ELSE 0 END), 0),
    COALESCE(SUM(CASE WHEN type IN ('buy', 'transfer_in') THEN quantity * price ELSE 0 END), 0)
  INTO v_net_qty, v_buy_qty, v_buy_cost
  FROM transactions
  WHERE entry_id = NEW.entry_id;

  UPDATE portfolio_entries
  SET
    quantity   = GREATEST(v_net_qty, 0),
    price      = CASE WHEN v_buy_qty > 0 THEN v_buy_cost / v_buy_qty ELSE price END,
    updated_at = NOW()
  WHERE id = NEW.entry_id;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace WHERE c.relkind = 'r' AND c.relname = 'transactions') THEN
    EXECUTE 'DROP TRIGGER IF EXISTS trg_recalculate_avg_cost ON transactions';
    EXECUTE 'CREATE TRIGGER trg_recalculate_avg_cost AFTER INSERT OR UPDATE ON transactions FOR EACH ROW EXECUTE FUNCTION recalculate_avg_cost()';
  END IF;
END$$;
