-- Keep a position in step with its transactions when one of them is deleted.
--
-- trg_recalculate_avg_cost has fired AFTER INSERT OR UPDATE since 000003, never
-- on DELETE. Removing a transaction therefore left portfolio_entries.quantity
-- and .price holding the totals of rows that no longer exist: the position
-- summary (cantidad total, precio promedio, costo total) stays larger than the
-- transactions listed underneath it, and the ROI derived from that stale cost
-- basis is pure fiction.
--
-- Deleting a transaction can only be done by hand in the database today — there
-- is no endpoint for it — so nothing in the application layer ever had a chance
-- to compensate. The recomputation has to live with the data.
--
-- The function had a second, quieter hole on UPDATE: moving a transaction to a
-- different entry recomputed the destination and left the origin stale, because
-- only NEW.entry_id was ever considered. Both holes close the same way — every
-- entry the row touched gets recomputed.

CREATE OR REPLACE FUNCTION recalculate_avg_cost()
RETURNS TRIGGER AS $$
DECLARE
  v_entry_ids UUID[];
  v_entry_id  UUID;
  v_net_qty   NUMERIC;
  v_buy_qty   NUMERIC;
  v_buy_cost  NUMERIC;
BEGIN
  -- NEW is unassigned on DELETE and OLD is unassigned on INSERT, so the set of
  -- affected entries is built from whichever records this operation has.
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

-- Recreate the trigger with DELETE included.
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace WHERE c.relkind = 'r' AND c.relname = 'transactions') THEN
    EXECUTE 'DROP TRIGGER IF EXISTS trg_recalculate_avg_cost ON transactions';
    EXECUTE 'CREATE TRIGGER trg_recalculate_avg_cost AFTER INSERT OR UPDATE OR DELETE ON transactions FOR EACH ROW EXECUTE FUNCTION recalculate_avg_cost()';
  END IF;
END$$;

-- Repair the positions that already drifted while DELETE went unnoticed.
--
-- Every entry is inserted with quantity 0 and gets its numbers exclusively from
-- transactions (CreatePortfolioEntry and the CSV import both insert the entry
-- and its first transaction in one transaction), so recomputing from an empty
-- transaction set to zero is the correct answer, not a data loss.
--
-- price keeps its previous value when no buy remains: that column doubles as
-- the entry's own reference price and there is nothing left to derive it from.
--
-- The average is rounded to the column's own scale before being compared, so
-- entries that are already correct are left untouched rather than having their
-- updated_at bumped by a difference that exists only below the 8th decimal.
WITH computed AS (
  SELECT
    pe.id,
    GREATEST(COALESCE(agg.net_qty, 0), 0) AS quantity,
    CASE
      WHEN COALESCE(agg.buy_qty, 0) > 0 THEN ROUND(agg.buy_cost / agg.buy_qty, 8)
      ELSE pe.price
    END                                   AS price
  FROM portfolio_entries pe
  LEFT JOIN LATERAL (
    SELECT
      SUM(CASE
        WHEN t.type IN ('buy', 'transfer_in')   THEN t.quantity
        WHEN t.type IN ('sell', 'transfer_out') THEN -t.quantity
        ELSE 0 END)                                                        AS net_qty,
      SUM(CASE WHEN t.type IN ('buy', 'transfer_in') THEN t.quantity ELSE 0 END)           AS buy_qty,
      SUM(CASE WHEN t.type IN ('buy', 'transfer_in') THEN t.quantity * t.price ELSE 0 END) AS buy_cost
    FROM transactions t
    WHERE t.entry_id = pe.id
  ) agg ON TRUE
)
UPDATE portfolio_entries pe
SET
  quantity   = c.quantity,
  price      = c.price,
  updated_at = NOW()
FROM computed c
WHERE c.id = pe.id
  AND (pe.quantity IS DISTINCT FROM c.quantity OR pe.price IS DISTINCT FROM c.price);
