-- Sector classification for catalog assets.
--
-- The app could already answer "what do I own" (GET /portfolios/holdings) and
-- "what kind of thing is it" (GET /portfolios/allocation, grouped by
-- assets.asset_type). Neither answers the question an investor actually asks
-- about concentration: eight positions spread over three portfolios can be
-- eight different tickers and still be one bet on semiconductors.
--
-- The column lives on assets and not in a per-user table, unlike
-- user_asset_prices, because it is not provider data. 000021 kept the assets
-- row shared on exactly that condition — it carries no provider-licensed value
-- — and a sector written here is operator-curated catalog metadata, entered
-- through the admin edit, the asset spreadsheet import or the seed, the same
-- three doors that write name and asset_type. Nothing in this migration lets a
-- value fetched with somebody's personal API key reach this column; a BYO-key
-- backfill would have to land in a user-scoped table of its own, for the same
-- licensing reason prices did.
--
-- VARCHAR and not an ENUM like asset_type. The GICS sectors are stable but the
-- thing that makes them awkward as a type is that every provider spells them
-- differently ("Financial Services", "FINANCIALS", "Servicios financieros"),
-- so the value is normalised in Go before it ever gets here — market.Sector,
-- validated by NormalizeSector. An enum would make the database repeat a check
-- the application has to do anyway, and make each new spelling a migration.
-- The newer columns on this schema (market_credentials.provider,
-- exchange_rates.source) took the same shape for the same reason.
--
-- NULL is a real state and the default: it means "nobody has classified this
-- yet", which the allocation reports as its own bucket rather than hiding. An
-- asset whose type has no sector to speak of (crypto, cash, a flat) is also
-- NULL here — the two are told apart by the asset's type at read time, not by
-- two different spellings of nothing.
ALTER TABLE assets
  ADD COLUMN IF NOT EXISTS sector VARCHAR(32);

-- Serves the sector allocation, which groups the user's held assets by this
-- column. Partial because most of the catalog is unclassified for a while
-- after this ships, and those rows are found by the IS NULL branch anyway.
CREATE INDEX IF NOT EXISTS idx_assets_sector
  ON assets(sector)
  WHERE sector IS NOT NULL;
