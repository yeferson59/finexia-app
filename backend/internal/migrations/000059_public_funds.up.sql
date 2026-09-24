-- Unit values the Superintendencia Financiera publishes (docs/PLAN_FONDOS_INVERSION.md, §12).
--
-- Every FIC reports the value of its unit each day, and the SFC publishes them
-- as open data (datos.gov.co, dataset qhpu-8ixx). An owner whose fund is there
-- no longer has to copy the unit value from the statement: they link their fund
-- to it, and the published values become its marks.
--
-- The values are not kept in a table of their own. They are written as marks of
-- every fund linked to them (fund_marks, 000057), with source = 'public', which
-- is what lets everything the marks already drive — the price copied to
-- user_asset_prices, the snapshots a late mark revalues, the returns by period —
-- work for them unchanged. A published value arrives two days after its date,
-- and that is exactly the late mark 000057 was built to absorb.

-- The catalog: every fund and type of participation of the latest day the SFC
-- published, refreshed by a job. It is shared, like assets.current_price
-- (000018): public data, fetched with no one's key.
--
--   id              the five codes that identify a type of participation,
--                   "5-31-3644-1-501": entity type, entity, fund, compartment,
--                   participation. An entity's code is unique only within its
--                   type, and a fund's only within its entity.
--   participation   each type has its own unit value; the owner's statement
--                   says which one is theirs, or its unit value gives it away.
--   search_text     entity and fund name, lower case and without accents, so a
--                   search for "renta" finds "RENTA" and "fiduciaria" finds
--                   "Fiduciaria". Written by the job.
--   unit_value,     the latest value published and its day. A fund the SFC
--   value_date      stopped publishing keeps its last one.
CREATE TABLE IF NOT EXISTS public_funds (
  id            VARCHAR(40)    PRIMARY KEY,
  entity_type   INTEGER        NOT NULL,
  entity_code   INTEGER        NOT NULL,
  fund_code     INTEGER        NOT NULL,
  compartment   INTEGER        NOT NULL,
  participation INTEGER        NOT NULL,
  entity_name   VARCHAR(255)   NOT NULL,
  fund_name     VARCHAR(255)   NOT NULL,
  fund_kind     VARCHAR(255)   NOT NULL DEFAULT '',
  search_text   TEXT           NOT NULL,
  unit_value    NUMERIC(20, 8) NOT NULL CHECK (unit_value > 0),
  value_date    DATE           NOT NULL,
  investors     INTEGER        NOT NULL DEFAULT 0,
  updated_at    TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

-- The fund an owner's fund is linked to. Only a fund followed by units can be:
-- one followed by balance has synthetic units, and a published unit value means
-- nothing against them.
ALTER TABLE user_funds
  ADD COLUMN IF NOT EXISTS public_fund_id VARCHAR(40) REFERENCES public_funds(id) ON DELETE SET NULL;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'user_funds_public_units') THEN
    ALTER TABLE user_funds
      ADD CONSTRAINT user_funds_public_units CHECK (public_fund_id IS NULL OR tracking = 'units');
  END IF;
END$$;

-- The job walks the funds linked to each published one.
CREATE INDEX IF NOT EXISTS idx_user_funds_public_fund
  ON user_funds(public_fund_id) WHERE public_fund_id IS NOT NULL;

-- Who said what a fund was worth on a day:
--
--   user    the owner, from the statement. It always wins: a published value
--           never overwrites it.
--   public  the SFC, through the link. Unlinking the fund takes these back.
ALTER TABLE fund_marks
  ADD COLUMN IF NOT EXISTS source VARCHAR(10) NOT NULL DEFAULT 'user';

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fund_marks_source') THEN
    ALTER TABLE fund_marks
      ADD CONSTRAINT fund_marks_source CHECK (source IN ('user', 'public'));
  END IF;
END$$;
