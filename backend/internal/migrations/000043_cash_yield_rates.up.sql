-- The rate a cash account earns.
--
-- A savings account, a pocket, a dollar wallet with an APY: the platform quotes
-- a rate, and the balance grows by it. The rate belongs to the account — a
-- platform and a currency — rather than to a portfolio, because that is what
-- the bank quotes and what its statement shows. When the account's money counts
-- in several portfolios, each balance earns the rate on its own share.
--
-- It is stored as an effective annual rate (E.A.), the convention savings
-- accounts quote in Colombia and the same figure as the APY of a dollar
-- account, as a fraction: 0.092500 is 9.25 % E.A.
--
-- A change of rate is a new version from a date, never an edit of the one
-- before, so every day keeps the rate it was earned at. The version that applies
-- on a day is the one with the latest effective_from on or before it, unless it
-- ended earlier: ended_on is the last day a version earns. Pausing a rate sets
-- it; resuming is a new version.
--
-- posting and max_balance describe how the interest is credited and up to what
-- balance. Every rate is posted daily for now; both columns are here so the
-- rows written today already say what they mean.

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'cash_interest_posting') THEN
    CREATE TYPE cash_interest_posting AS ENUM ('daily', 'monthly');
  END IF;
END$$;

CREATE TABLE IF NOT EXISTS cash_yield_rates (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  source_id        UUID NOT NULL REFERENCES investment_sources(id) ON DELETE CASCADE,
  currency         CHAR(3) NOT NULL,
  annual_rate      NUMERIC(9, 6) NOT NULL CHECK (annual_rate > 0 AND annual_rate <= 1),
  withholding_rate NUMERIC(5, 4) NOT NULL DEFAULT 0 CHECK (withholding_rate >= 0 AND withholding_rate < 1),
  max_balance      NUMERIC(20, 8) CHECK (max_balance IS NULL OR max_balance > 0),
  posting          cash_interest_posting NOT NULL DEFAULT 'daily',
  effective_from   DATE NOT NULL,
  ended_on         DATE CHECK (ended_on IS NULL OR ended_on >= effective_from),
  created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  -- Also the index every lookup of an account's versions reads.
  CONSTRAINT uk_cash_yield_rates_version UNIQUE (source_id, currency, effective_from)
);

DROP TRIGGER IF EXISTS trg_cash_yield_rates_updated_at ON cash_yield_rates;
CREATE TRIGGER trg_cash_yield_rates_updated_at
  BEFORE UPDATE ON cash_yield_rates
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();
