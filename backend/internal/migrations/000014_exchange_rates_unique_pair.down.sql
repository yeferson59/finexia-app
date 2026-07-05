DROP INDEX IF EXISTS idx_rates_pair;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_class WHERE relkind = 'i' AND relname = 'idx_rates_pair_date') THEN
    CREATE UNIQUE INDEX idx_rates_pair_date ON exchange_rates(from_currency, to_currency, rate_date);
  END IF;
END$$;
