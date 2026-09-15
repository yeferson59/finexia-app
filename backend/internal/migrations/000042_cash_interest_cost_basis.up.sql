-- Interest a cash balance earns is gain.
--
-- portfolio_summary reads a position's gain as its market value minus its cost
-- base, and the cost base as quantity × price, where price is the average cost
-- recalculate_avg_cost keeps. 000041 built that average from purchases alone
-- and let cash_interest add to the quantity, so a balance stayed at one unit of
-- its currency per unit: the right value, and the wrong cost. A thousand dollars
-- deposited and ten earned were worth 1010 against a cost of 1010, and every
-- figure that subtracts the two — the portfolio's gain, the platform's, the
-- total_gain_loss each snapshot keeps — reported the interest as nothing. Only
-- the growth series, which nets flows out instead of subtracting a cost, ever
-- counted it.
--
-- Interest is units that cost nothing, and averaged in as that, the thousand
-- and ten cost 1000 and the ten reads as gain wherever a gain is read, without
-- any of those queries changing.
--
-- The average has to be walked in order, which the formula of 000041 does not
-- do: it averages every purchase the position ever had. On a balance that is
-- emptied and filled again — the ordinary life of an account — the interest of
-- the first round would sit in the cost of the second and read as gain nobody
-- made. Walking the rows by date:
--
--   buy, transfer_in     add units, and what they cost.
--   cash_interest        adds units, and nothing to the cost.
--   sell, transfer_out   take units out at the average, so the cost keeps the
--                        share the remaining units carry. A withdrawal takes its
--                        share of the interest with it: what stays as gain is
--                        the interest still in the balance, the same way a sale
--                        takes a share's gain with it.
--   an emptied balance   starts over, so the next deposit costs what it cost.
--
-- Only cash positions that hold interest take the walk. Every other position —
-- cash without interest included — keeps the formula of 000041 exactly, so no
-- other cost moves; and a balance whose last interest is deleted goes back to
-- it, and to its price of one.

-- The average cost of a cash position that holds interest, walked in order.
-- NULL when the rule does not apply: not a cash position, no interest on it, or
-- nothing left in it — the caller keeps its own price for those.
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
      AND tx.type = 'cash_interest'
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
    ELSIF t.type = 'cash_interest' THEN
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

-- The cash writers told a balance they keep from one opened by hand by its
-- price: one unit of its currency per unit. Interest now lowers that price, so
-- the test moves to what the price stood in for — every row that moved the
-- quantity at one unit per unit, in the position's own currency.
--
-- A position with no such rows yet is judged by its price, which is what the
-- writer seeds it with.
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
            AND t.type IN ('buy', 'transfer_in', 'sell', 'transfer_out', 'cash_interest')
        )
      )
  );
$$;

-- 000041's recalculation, with the price of a cash position that holds interest
-- taken from the walk above. The quantity is untouched: it stays the unordered
-- sum the Go writers check a balance against before they write.
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

-- The balances that already hold interest. portfolio_entries is written
-- directly, so none of the triggers on transactions fire for it.
UPDATE portfolio_entries pe
SET price = cash_entry_avg_cost(pe.id)
WHERE cash_entry_avg_cost(pe.id) IS NOT NULL;
