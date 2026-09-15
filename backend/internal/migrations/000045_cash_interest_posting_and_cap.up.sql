-- Interest posted monthly, and a cap on the balance that earns (000043, 000044).
--
-- A rate posted monthly computes every day as a daily one does, and credits the
-- month in one cash_interest on its last day. Until then the days wait in the
-- ledger as pending, and each day earns on the days held before it: interest
-- computed is interest owed, and it compounds whether or not the platform has
-- credited it yet. Posted monthly or daily, a year at a rate earns the same.
--
-- The month's credit is linked to every day it pays. A balance that stops
-- earning before its month closes — its rate is paused, or it is no longer one
-- the ledger keeps — has what it held credited on its last day, instead of
-- waiting for a month end that will not come.
--
-- A cap is the most an account earns on. It belongs to the account — a
-- platform and a currency — and so to all its balances together: when they hold
-- more than the cap, each earns on its share of it, in proportion to what it
-- holds.

-- What a cash balance held at the close of a day, for the interest it earns
-- that day.
--
-- Every transaction dated on or before the day counts, except the ledger's
-- credits, which count from the day after the last day they pay, whatever day
-- they are dated: a catch-up dates them later (000044), and reading them by that
-- date would leave each caught-up day compounding on less than the balance held.
--
-- Interest computed for an earlier day and not credited by the close of this
-- one counts too: a pending day, or a day whose credit pays a later one — a
-- month's credit, read from a day inside that month. That is what makes a
-- balance's days read the same no matter when the credit was written, and what
-- lets a cap read every balance of an account on the same terms.
CREATE OR REPLACE FUNCTION cash_entry_balance_at_close(p_entry_id UUID, p_day DATE)
RETURNS NUMERIC
LANGUAGE sql
STABLE
AS $$
  SELECT
    COALESCE((
      SELECT SUM(CASE
        WHEN t.type IN ('buy', 'transfer_in', 'cash_interest') THEN t.quantity
        WHEN t.type IN ('sell', 'transfer_out')                THEN -t.quantity
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

-- The days a balance holds: its basis sums them every day, and the job looks
-- for the balances that still hold some.
CREATE INDEX IF NOT EXISTS idx_cash_accruals_pending
  ON cash_interest_accruals(entry_id, accrual_date) WHERE status = 'pending';
