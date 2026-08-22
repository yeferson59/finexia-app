-- Drop the entry's copy of the asset's type.
--
-- portfolio_entries.category was written once, when the entry was created, from
-- the type of the asset it holds: the create handler and the importer both
-- derive it through entryCategoryFor, and nothing ever wrote it again. It was a
-- second answer to a question the asset catalogue already answers, and the two
-- drifted the moment an asset was reclassified. Correcting an ETF of bonds
-- first filed as a plain ETF moved the position in the per-portfolio donut,
-- which groups holdings by assets.asset_type, and left it in the old slice on
-- the dashboard, which grouped by this column: two charts over the same
-- positions, disagreeing, with no way for the user to tell which was lying.
--
-- The dashboard allocation now derives the category from assets.asset_type like
-- everything else, so nothing reads this column any more. Keeping it would only
-- preserve the drift — a stale value still served in the entry payload, and an
-- inviting one for the next query to group by.
--
-- The enum goes with it: this was its only column.
ALTER TABLE portfolio_entries DROP COLUMN IF EXISTS category;

DROP TYPE IF EXISTS portfolio_entry_category;
