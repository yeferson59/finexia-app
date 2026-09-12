-- The single-sector column in 000034 outlives this table, so a rollback loses
-- the breakdowns and nothing else: an asset that had one goes back to reading
-- as unclassified, which is what it read as before this migration existed.
DROP TABLE IF EXISTS asset_sector_weights;
