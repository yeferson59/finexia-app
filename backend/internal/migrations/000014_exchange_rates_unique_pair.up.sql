-- exchange_rates previously kept one row per (from_currency, to_currency,
-- rate_date), so the daily sync cron inserted a new historical row every
-- day instead of replacing the current rate for a pair. Collapse existing
-- history down to the most recent row per pair, then re-key the unique
-- index on the pair alone so future syncs update in place.
DELETE FROM exchange_rates er
USING exchange_rates newer
WHERE er.from_currency = newer.from_currency
  AND er.to_currency = newer.to_currency
  AND (
    newer.rate_date > er.rate_date
    OR (newer.rate_date = er.rate_date AND newer.fetched_at > er.fetched_at)
    OR (newer.rate_date = er.rate_date AND newer.fetched_at = er.fetched_at AND newer.id > er.id)
  );

DROP INDEX IF EXISTS idx_rates_pair_date;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_class WHERE relkind = 'i' AND relname = 'idx_rates_pair') THEN
    CREATE UNIQUE INDEX idx_rates_pair ON exchange_rates(from_currency, to_currency);
  END IF;
END$$;
