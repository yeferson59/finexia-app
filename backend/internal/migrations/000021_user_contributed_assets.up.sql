-- Assets contributed by users.
--
-- Under the provider model the catalog was the operator's: the operator paid
-- the quota, so the operator decided what got synced, and creating an asset was
-- admin-only. BYO-key removed that reason — each user now syncs their own
-- holdings with their own key — and left the admin gate half-closed anyway:
-- ImportEntryTransactions has always inserted rows into this same table for any
-- user who uploads a transaction file. The gate only ever blocked the front
-- door.
--
-- So the row itself stays shared (one AAPL, not one per user: it carries no
-- provider-licensed data — prices live in user_asset_prices), and what is added
-- here is provenance and reach.
--
-- created_by is the first contributor, kept for moderation. It is NOT the
-- visibility rule: two users tracking the same local ticker dedupe onto one row
-- and only one of them can be its creator.
ALTER TABLE assets
  ADD COLUMN IF NOT EXISTS created_by UUID REFERENCES users(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS is_curated BOOLEAN NOT NULL DEFAULT FALSE;

-- Everything that exists today is already served to everybody — the seed, the
-- admin's rows and the ones past transaction imports created — so it stays
-- visible to everybody. Only what is contributed from now on is scoped.
UPDATE assets SET is_curated = TRUE;

-- Reach: which users a non-curated asset is visible to. A membership table
-- rather than a column because it is many-to-many by construction — the second
-- user to contribute a ticker gets the existing row back and must still see it
-- in their own catalog.
CREATE TABLE IF NOT EXISTS user_catalog_assets (
  user_id    UUID NOT NULL REFERENCES users(id)  ON DELETE CASCADE,
  asset_id   UUID NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (user_id, asset_id)
);

-- Serves the catalog's visibility join, which looks up (user_id, asset_id) —
-- covered by the primary key — and the reverse direction, used when moderating
-- an asset.
CREATE INDEX IF NOT EXISTS idx_user_catalog_assets_asset ON user_catalog_assets(asset_id);

-- Serves the daily contribution quota: count this user's own rows since a
-- timestamp.
CREATE INDEX IF NOT EXISTS idx_assets_created_by
  ON assets(created_by, created_at)
  WHERE created_by IS NOT NULL;
