-- The external cash flow of one transaction, signed from the portfolio's side.
--
-- The growth series measures rentability by comparing what a portfolio was
-- worth on two dates, and that comparison is only a return if the money the
-- owner put in or took out during the period is netted out first. Until now the
-- only proxy available was the change in cost base, which gets a purchase right
-- and a sale wrong: a sale leaves the series at market value but only lowers
-- the cost base by what the shares originally cost, so the realised gain was
-- being counted as a loss on the day it was taken.
--
-- This function is the sign convention for that netting, in one place, so the
-- account-wide series and the per-portfolio one cannot disagree about it.
--
--   buy, transfer_in       money in, at the traded amount plus the commission —
--                          the fee left the owner's pocket and bought nothing,
--                          so it belongs in the flow and reads as a drag on the
--                          return, which is what it is.
--   sell, transfer_out     money out, at the proceeds actually received, which
--                          is the traded amount minus the commission.
--   dividend, interest     money out too: the payment is income the holdings
--                          produced, but the app tracks no cash position, so it
--                          leaves the measured pool the moment it is paid.
--                          Treating it as a withdrawal is what credits it to
--                          the return instead of losing it.
--   fee                    money in that bought nothing: a standalone cost.
--   split                  no flow. It re-cuts the same holding, and this app's
--                          average-cost trigger does not even move the quantity
--                          on it.
--
-- The amount is in the transaction's own currency; converting it is the
-- caller's job, the same way it already converts the snapshots.
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
    WHEN 'buy'          THEN  COALESCE(p_quantity, 0) * COALESCE(p_price, 0) + COALESCE(p_fees, 0)
    WHEN 'transfer_in'  THEN  COALESCE(p_quantity, 0) * COALESCE(p_price, 0) + COALESCE(p_fees, 0)
    WHEN 'fee'          THEN  COALESCE(p_fees, 0)
    WHEN 'sell'         THEN -(COALESCE(p_quantity, 0) * COALESCE(p_price, 0) - COALESCE(p_fees, 0))
    WHEN 'transfer_out' THEN -(COALESCE(p_quantity, 0) * COALESCE(p_price, 0) - COALESCE(p_fees, 0))
    WHEN 'dividend'     THEN -(COALESCE(p_quantity, 0) * COALESCE(p_price, 0) - COALESCE(p_fees, 0))
    WHEN 'interest'     THEN -(COALESCE(p_quantity, 0) * COALESCE(p_price, 0) - COALESCE(p_fees, 0))
    ELSE 0
  END;
$$;

-- The growth queries look up, for every transaction, the first snapshot on or
-- after its date. Without this the lookup scans the whole table per snapshot.
CREATE INDEX IF NOT EXISTS idx_transactions_entry_date ON transactions(entry_id, transaction_date);
