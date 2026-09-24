-- Investment funds: how each owner follows one, what it was worth by date, and
-- what every snapshot valued it at.
--
-- A fund is a position of units (000056). What it needs that a share does not
-- is a price nobody else can supply: no market-data provider lists a FIC, so
-- the owner writes the unit value down from the statement — daily, weekly, with
-- the monthly extract, often days after the date it belongs to.
--
-- The latest of those values is copied to user_asset_prices, where every
-- valuation already looks first (000018). That is what keeps portfolio_summary,
-- the holdings, the snapshot job and the recorded market price of 000036
-- unchanged: to all of them a fund is a position priced by its owner. The
-- history stays here, because the return a fund publishes — "9.8 % E.A. over
-- 30 days" — is a ratio of two of those values.

-- How the owner follows the fund, chosen when it is created and never changed:
--
--   units    the statement shows units and the unit value, and both are written
--            as they appear there.
--   balance  the app only shows a balance. Units are synthetic — the first
--            contribution buys at 1 and each balance fixes the unit value — so
--            the return is still time-weighted. Opened by a later phase.
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'fund_tracking') THEN
    CREATE TYPE fund_tracking AS ENUM ('units', 'balance');
  END IF;
END$$;

-- One row per fund an owner follows. The key is the owner and the asset, the
-- same key user_asset_prices has: a fund held in two portfolios has one unit
-- value, the one on the statement.
CREATE TABLE IF NOT EXISTS user_funds (
  user_id    UUID NOT NULL REFERENCES users(id)  ON DELETE CASCADE,
  asset_id   UUID NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
  tracking   fund_tracking NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (user_id, asset_id)
);

-- What the fund was worth on a day. In 'units' the fact is unit_value; in
-- 'balance' the fact is balance, and unit_value is derived from it and the
-- units held that day.
--
-- The primary key serves the read every valuation makes — the latest mark on or
-- before a day — as a backward scan.
CREATE TABLE IF NOT EXISTS fund_marks (
  user_id    UUID NOT NULL,
  asset_id   UUID NOT NULL,
  mark_date  DATE NOT NULL,
  unit_value NUMERIC(20, 8) NOT NULL CHECK (unit_value > 0),
  -- A balance of zero defines no unit value: emptying a fund is a withdrawal.
  balance    NUMERIC(20, 8) CHECK (balance IS NULL OR balance > 0),
  notes      VARCHAR(500),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (user_id, asset_id, mark_date),
  FOREIGN KEY (user_id, asset_id) REFERENCES user_funds(user_id, asset_id) ON DELETE CASCADE
);

-- What each snapshot valued each fund position at.
--
-- A mark usually arrives after its date: the statement of the 30th is typed in
-- on the 5th. Without this the gain would land on the 5th — the day it was
-- recorded — and the chart would draw a flat month and a jump. With it, writing
-- a mark revalues every snapshot from its date on: the new unit value against
-- the one stored here, times the units and the rate stored here, and nothing
-- has to be guessed about what the snapshot used.
--
-- The snapshot job writes these rows from the same statement that reads the
-- totals, so a row and the total beside it describe the same instant.
--
--   unit_value  what the snapshot used, or what a later mark restated it to.
--   unit_cost   the position's average cost that day: its value when no mark
--               on or before the date is left.
--   fx_rate     asset currency → the portfolio's base, as the snapshot applied it.
CREATE TABLE IF NOT EXISTS fund_snapshot_values (
  entry_id      UUID NOT NULL REFERENCES portfolio_entries(id) ON DELETE CASCADE,
  snapshot_date DATE NOT NULL,
  portfolio_id  UUID NOT NULL REFERENCES portfolios(id) ON DELETE CASCADE,
  asset_id      UUID NOT NULL REFERENCES assets(id)     ON DELETE CASCADE,
  units         NUMERIC(20, 8)  NOT NULL,
  unit_value    NUMERIC(20, 8)  NOT NULL,
  unit_cost     NUMERIC(20, 8)  NOT NULL,
  fx_rate       NUMERIC(24, 10) NOT NULL,
  PRIMARY KEY (entry_id, snapshot_date)
);

-- A mark revalues one fund's rows from a date on.
CREATE INDEX IF NOT EXISTS idx_fund_snapshot_values_asset
  ON fund_snapshot_values(asset_id, snapshot_date);

-- The job replaces one portfolio's rows for one day on every run.
CREATE INDEX IF NOT EXISTS idx_fund_snapshot_values_portfolio
  ON fund_snapshot_values(portfolio_id, snapshot_date);
