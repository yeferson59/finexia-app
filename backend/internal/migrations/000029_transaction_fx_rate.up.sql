-- The exchange rate a transaction was settled at.
--
-- A trade has two currencies whenever the asset is quoted in one and the
-- account is funded in another: LVMH prints in EUR, the broker debits USD, and
-- the rate between them on that morning is part of what the position cost. The
-- app had nowhere to put that rate, so a position could be recorded one of two
-- ways and neither was right.
--
-- Recorded in the quote currency, the cost basis was re-translated at today's
-- rate on every read — exchange_rates has held one row per pair since 000014,
-- so fx_rate() only ever answers "now". Cost and market value then moved with
-- the same multiplier and cancelled it out: the number on the screen was the
-- return the asset had in its own currency, wearing the account's currency
-- symbol. Recorded in the account currency instead, the figure came out right
-- but only because the owner had multiplied price by rate by hand before
-- typing it, and the two factors were unrecoverable from the row afterwards.
--
-- fx_rate stores that multiplier: how much of portfolio_entries.cost_currency
-- one unit of transactions.currency bought, on the day of the trade. price
-- stays in the currency the trade was quoted in — 606.60 EUR stays 606.60 EUR,
-- which is what the broker's confirmation says and what makes the row
-- checkable — and price * fx_rate is what the position actually cost, in the
-- currency it was paid in. The historical rate is thereby frozen where it
-- belongs, on the trade, while the market side keeps floating at today's rate.
-- That difference between the two is the FX result, and it now survives to the
-- screen instead of being divided out.
--
-- DEFAULT 1 is not a placeholder: it is the true rate for every row already
-- here, all of which were written with currency = cost_currency because that is
-- the only thing the application could produce. So the backfill is exact and no
-- recomputation of existing positions is needed — the migration below rewrites
-- the average-cost function, and feeding it fx_rate = 1 reproduces the numbers
-- it produced before, to the digit.
--
-- The one rule this column has cannot be a CHECK, because it spans two tables:
-- when currency = cost_currency the rate must be 1, since a currency does not
-- convert into itself at anything else. portfolio.TransactionInput.Validate
-- enforces it on the way in, and CreateTransaction/UpdateTransaction read the
-- entry's cost_currency inside the same transaction as the insert so the check
-- cannot race the row it is checking.
ALTER TABLE transactions
  ADD COLUMN IF NOT EXISTS fx_rate NUMERIC(20, 8) NOT NULL DEFAULT 1;

-- A zero or negative rate is not a conversion, and money.Convert rejects one
-- anyway; refusing it at the column keeps a corrupt row from ever reaching the
-- trigger below, where it would silently zero out a position's cost basis.
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'chk_transactions_fx_rate_positive'
  ) THEN
    ALTER TABLE transactions
      ADD CONSTRAINT chk_transactions_fx_rate_positive CHECK (fx_rate > 0);
  END IF;
END$$;

-- Which of the two currencies the commission was charged in.
--
-- The rate above settles the price; it does not settle the fee, because a fee
-- is not part of the instrument's price and brokers do not consistently charge
-- it on the same side. The confirmation that prompted this column quotes the
-- fill at 606.60 EUR and the commission at 0.00 USD: one line, two currencies,
-- and multiplying both by the same rate is right for exactly one of them.
--
-- The column is deliberately not free: a fee is either in the currency the
-- trade was quoted in or in the one the account settled in, and nothing else is
-- a fee on this transaction. That is a two-table rule, so it lives in
-- TransactionInput.Validate rather than in a CHECK — but it is what makes the
-- conversion below a choice between fx_rate and 1, with no third rate to store.
--
-- Backfilled from currency rather than defaulted to a constant: every existing
-- row has fees in the currency it also priced in, which is the same as its cost
-- currency, so the two candidate answers coincide and this one is exact. The
-- column is filled before it is made NOT NULL so no row is ever unlabelled.
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS fees_currency CHAR(3);
UPDATE transactions SET fees_currency = currency WHERE fees_currency IS NULL;
ALTER TABLE transactions ALTER COLUMN fees_currency SET NOT NULL;

-- The fee, in the position's cost currency.
--
-- It is a function for the reason transaction_cash_flow (000027) is one: the
-- account-wide growth series and the per-portfolio one both need this rule, and
-- two copies of it are two chances to disagree about whose rate a commission
-- takes.
--
-- The rule is small because the column above is constrained. A fee charged in
-- the currency of the trade rode the same conversion the trade did; a fee
-- charged in the currency of the account is already there. There is no third
-- case, so there is no third rate to look up and nothing here can fail.
CREATE OR REPLACE FUNCTION transaction_fees_in_cost(
  p_fees          NUMERIC,
  p_fees_currency CHAR(3),
  p_currency      CHAR(3),
  p_fx_rate       NUMERIC
)
RETURNS NUMERIC
LANGUAGE sql
IMMUTABLE
AS $$
  SELECT COALESCE(p_fees, 0) *
    CASE WHEN p_fees_currency IS NOT DISTINCT FROM p_currency
      THEN COALESCE(p_fx_rate, 1)
      ELSE 1
    END;
$$;

-- The average cost becomes an average of converted costs.
--
-- This is the whole reason the column exists. portfolio_entries.price is read
-- everywhere as being in cost_currency — postgres_entry.go tags it that way,
-- portfolio_summary converts it with fx_rate(pe.cost_currency, ...), and the
-- holdings endpoint reports it under CostCurrency — but it was computed as a
-- plain average of transactions.price, which is in the *trade* currency. While
-- the two currencies were forced to be equal that distinction was invisible.
-- Multiplying by fx_rate is what makes the column's value match the label it
-- has always carried.
--
-- Everything else here is unchanged from 000023: the same set of affected
-- entries on INSERT/UPDATE/DELETE, the same net-quantity rule, the same
-- fallback to the existing price when no buy remains to derive one from.
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
      -- The only changed line: the cost of a buy is what it cost in the
      -- position's currency, not what its price said in another one.
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
