-- Tiers: a rate that pays differently on different parts of the balance.
--
-- "12 % up to five million, 8 % on what is above it." A version of a rate
-- (000043) keeps the rate it pays from zero in annual_rate, which is still the
-- figure the platform quotes and the one every reader already shows. Each tier
-- is a step above that: from what the account holds, the rate it earns on what
-- is above that point, up to the next step.
--
-- The steps belong to the account, as the cap did, and not to any one of its
-- balances. A day computes the interest of the whole account on what all its
-- balances held at the close, and each balance takes its share, in proportion
-- to what it holds.
--
-- A cap is a step at 0 %: from there the account earns nothing. It was a column
-- of its own (max_balance); it becomes that step, and the column goes, so there
-- is one way to say it. Under the cap and over it, a balance earns exactly what
-- it did.
--
-- The ledger (000044) is unchanged, and two of its columns now read as their
-- comments always said:
--
--   balance_basis   what the balance held at the close of the day. A capped
--                   day kept its share of the cap there; with steps there is no
--                   single part that earns, so it keeps what was held.
--   annual_rate     the rate the day was earned at: with steps, what they come
--                   to together on what the account held, (1 + gross/held)^365
--                   − 1. Without them, the version's own rate, as before.

CREATE TABLE IF NOT EXISTS cash_yield_rate_tiers (
  rate_id      UUID NOT NULL REFERENCES cash_yield_rates(id) ON DELETE CASCADE,
  -- The step from zero is the version's annual_rate, so a tier starts above it.
  from_balance NUMERIC(20, 8) NOT NULL CHECK (from_balance > 0),
  -- A fraction, as annual_rate is. Zero is a cap.
  annual_rate  NUMERIC(9, 6)  NOT NULL CHECK (annual_rate >= 0 AND annual_rate <= 1),
  PRIMARY KEY (rate_id, from_balance)
);

DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name = 'cash_yield_rates' AND column_name = 'max_balance'
  ) THEN
    INSERT INTO cash_yield_rate_tiers (rate_id, from_balance, annual_rate)
    SELECT id, max_balance, 0 FROM cash_yield_rates WHERE max_balance IS NOT NULL
    ON CONFLICT DO NOTHING;

    ALTER TABLE cash_yield_rates DROP COLUMN max_balance;
  END IF;
END$$;
