-- Pockets: a subaccount of a cash account, not a platform of its own.
--
-- One entity pays differently depending on where the money sits. The account
-- pays 8 % and the "cajita" pays 10 %, and both are the same bank. A pocket
-- belongs to a platform and a currency, and its money counts inside that
-- platform: every figure by platform — the detail, the reports, the allocation,
-- MCP — keeps adding the pocket in, because it never left.
--
-- The main account is pocket_id NULL. Everything that exists today already is
-- the main account, so there is no backfill, and a platform without pockets
-- behaves exactly as it did.
--
-- What changes is the key of an account. Where a rate, its tiers (000046) and
-- the ledger (000044) looked up a platform and a currency, they now look up a
-- platform, a currency and a pocket. A pocket earns its own rate, on what its
-- own balances hold, and the main account earns its own on the rest.
--
-- kind is 'flexible' here. 'fixed' — a deposit that keeps the rate of the day
-- it was opened, with a maturity — is opened in the phase after this one; the
-- value exists now so the rows written today already say what they mean, and
-- so matures_on can be checked against it.
--
-- Deleting a platform already refuses to take one with positions, so the
-- cascade only ever carries empty pockets away with it. A pocket with balances
-- is held by portfolio_entries.pocket_id, which does not cascade: it has to be
-- emptied before it can go.

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'cash_pocket_kind') THEN
    CREATE TYPE cash_pocket_kind AS ENUM ('flexible', 'fixed');
  END IF;
END$$;

CREATE TABLE IF NOT EXISTS cash_pockets (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  source_id  UUID NOT NULL REFERENCES investment_sources(id) ON DELETE CASCADE,
  currency   CHAR(3) NOT NULL,
  name       VARCHAR(100) NOT NULL,
  kind       cash_pocket_kind NOT NULL DEFAULT 'flexible',
  -- The day the pocket started holding money. A fixed deposit earns from it.
  opened_on  DATE NOT NULL,
  -- Only a fixed deposit matures, and never on the day it opened.
  matures_on DATE CHECK (matures_on IS NULL OR (kind = 'fixed' AND matures_on > opened_on)),
  -- The day it stopped: cancelled, or matured and emptied.
  closed_on  DATE CHECK (closed_on IS NULL OR closed_on >= opened_on),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  -- Two pockets of one account cannot share a name; two accounts can.
  CONSTRAINT uk_cash_pockets_name UNIQUE (source_id, currency, name)
);

DROP TRIGGER IF EXISTS trg_cash_pockets_updated_at ON cash_pockets;
CREATE TRIGGER trg_cash_pockets_updated_at
  BEFORE UPDATE ON cash_pockets
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE INDEX IF NOT EXISTS idx_cash_pockets_account ON cash_pockets(source_id, currency);

-- A balance belongs to a pocket, or to the main account when it is NULL. No
-- cascade: a pocket that still holds money is not deleted out from under it.
ALTER TABLE portfolio_entries ADD COLUMN IF NOT EXISTS pocket_id UUID REFERENCES cash_pockets(id);
-- A rate belongs to one too, and goes with it.
ALTER TABLE cash_yield_rates  ADD COLUMN IF NOT EXISTS pocket_id UUID REFERENCES cash_pockets(id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_entries_pocket ON portfolio_entries(pocket_id) WHERE pocket_id IS NOT NULL;

-- The main account keeps the key it always had, now only over the rows that
-- are the main account. A pocket is one balance per portfolio instead: its
-- asset and its platform follow from the pocket, so the pocket is the key.
--
-- Careful with a partial index: Postgres only uses it for an ON CONFLICT that
-- repeats its predicate. The three writers that open a position — CreateEntry,
-- the transaction import and CreateCashMovement — say WHERE pocket_id IS NULL,
-- and a write that left it out would fail with "no unique or exclusion
-- constraint matching the ON CONFLICT specification".
DROP INDEX IF EXISTS idx_entries_portfolio_asset_source;
CREATE UNIQUE INDEX idx_entries_portfolio_asset_source
  ON portfolio_entries(portfolio_id, asset_id, COALESCE(source_id::TEXT, ''))
  WHERE pocket_id IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_entries_portfolio_pocket
  ON portfolio_entries(portfolio_id, pocket_id)
  WHERE pocket_id IS NOT NULL;

-- A version of a rate is one per account and day, and the pocket is part of the
-- account. COALESCE keeps the main account — pocket_id NULL — to one version a
-- day too, which a plain unique constraint over a nullable column would not.
ALTER TABLE cash_yield_rates DROP CONSTRAINT IF EXISTS uk_cash_yield_rates_version;
CREATE UNIQUE INDEX IF NOT EXISTS uk_cash_yield_rates_version
  ON cash_yield_rates(source_id, currency, COALESCE(pocket_id::TEXT, ''), effective_from);
