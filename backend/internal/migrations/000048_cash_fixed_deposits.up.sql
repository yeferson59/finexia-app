-- Fixed deposits: a pocket that keeps the rate of the day it was opened.
--
-- A CDT, a term pocket, a promotional rate locked for ninety days. It is a
-- pocket of kind 'fixed' (000047) with one deposit — the one that opened it —
-- one version of its rate, and an optional maturity. Everything it needs is
-- already here: the daily accrual (000044), interest that costs nothing and so
-- reads as gain (000042), and the growth series.
--
-- What this migration adds is the third way of crediting what a rate earns.
-- 'daily' credits every day and 'monthly' on the last day of each month; a
-- deposit can also pay everything on the day it matures. The days are computed
-- the same way either way and wait in the ledger as pending, so a year at a
-- rate earns the same however it is credited; what changes is when the balance
-- shows it. Crediting at maturity is what makes the balance match a statement
-- that only shows the principal until the end, at the cost of the portfolio's
-- value jumping on that day. Daily stays the default.
--
-- A value added with ADD VALUE cannot be used in the transaction that adds it,
-- so this migration only adds it: nothing here writes an 'at_maturity' row.
ALTER TYPE cash_interest_posting ADD VALUE IF NOT EXISTS 'at_maturity';

-- The deposits the job has to look at each morning: the ones still open, by the
-- day they come due. Everything else in cash_pockets is read through its
-- account, so this is the only index the maturity sweep needs.
CREATE INDEX IF NOT EXISTS idx_cash_pockets_maturing
  ON cash_pockets(matures_on) WHERE kind = 'fixed' AND closed_on IS NULL;
