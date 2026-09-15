-- The interest a cash balance earns, day by day.
--
-- Every day a rate is in effect (000043), the balance it applies to earns
-- interest on what it held at the close of that day. This table is that ledger:
-- one row per balance and day, with what the interest was computed on and at
-- what rate, and the cash_interest transaction (000040) that credited it.
--
-- The row is written before the transaction, and the unique key on the balance
-- and the day is what makes computing a day twice impossible: a second run, a
-- second replica, or a catch-up after the job was down all find the day taken
-- and credit nothing. It also outlives the transaction. An owner who deletes a
-- credited interest leaves the row with no transaction behind it, and the day
-- stays computed, so the interest is not credited again the next morning.
--
-- The amounts are copies, not references. The rate a day was earned at stays
-- in the row even if the rate is later ended or its platform removed.
--
--   balance_basis    what the balance held at the close of the day.
--   gross_amount     balance_basis × the daily equivalent of the annual rate.
--   withholding      the share of it withheld as tax.
--   net_amount       gross minus withholding, at full precision.
--   rounding_carry   what was left after crediting the net amount, plus the
--                    carry of the day before, rounded to the currency's minor
--                    unit. It is added to the next day, so rounding every day
--                    loses nothing over a year.
--
--   status  posted   a transaction credited it; transaction_id names it, or is
--                    null if the owner deleted that transaction.
--           carried  the rounded amount was zero, and all of it went to the
--                    carry.
--           pending  computed and not yet credited, for rates posted monthly.

CREATE TABLE IF NOT EXISTS cash_interest_accruals (
  id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  entry_id       UUID NOT NULL REFERENCES portfolio_entries(id) ON DELETE CASCADE,
  rate_id        UUID REFERENCES cash_yield_rates(id) ON DELETE SET NULL,
  accrual_date   DATE NOT NULL,
  annual_rate    NUMERIC(9, 6)  NOT NULL,
  balance_basis  NUMERIC(20, 8) NOT NULL,
  gross_amount   NUMERIC(20, 8) NOT NULL,
  withholding    NUMERIC(20, 8) NOT NULL,
  net_amount     NUMERIC(20, 8) NOT NULL,
  rounding_carry NUMERIC(20, 8) NOT NULL DEFAULT 0,
  status         TEXT NOT NULL CHECK (status IN ('pending', 'posted', 'carried')),
  transaction_id UUID REFERENCES transactions(id) ON DELETE SET NULL,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT uk_cash_accrual_day UNIQUE (entry_id, accrual_date)
);

-- A movement is read as automatic by looking its transaction up here.
CREATE INDEX IF NOT EXISTS idx_cash_accruals_transaction
  ON cash_interest_accruals(transaction_id) WHERE transaction_id IS NOT NULL;

-- A rate that has earned interest can no longer be corrected or deleted, and
-- the writes ask that of this column.
CREATE INDEX IF NOT EXISTS idx_cash_accruals_rate
  ON cash_interest_accruals(rate_id, accrual_date);
