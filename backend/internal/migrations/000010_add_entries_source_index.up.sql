-- portfolio_entries.source_id is joined against investment_sources on every
-- platforms-with-stats query and walked by the ON DELETE cascade when a
-- platform is removed; without an index both degrade to sequential scans.
CREATE INDEX IF NOT EXISTS idx_entries_source ON portfolio_entries(source_id);
