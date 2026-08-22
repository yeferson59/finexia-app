-- Put the column back and refill it from the asset type, which is where its
-- values came from in the first place.
--
-- The CASE mirrors entryCategoryFor. It exists only here, so that a rollback
-- lands on populated data instead of a table of the 'others' default; the
-- forward path has no SQL copy of that mapping.
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'portfolio_entry_category') THEN
    CREATE TYPE portfolio_entry_category AS ENUM (
      'stocks',
      'etfs',
      'cryptos',
      'bonds',
      'cash',
      'real_estates',
      'commodities',
      'others'
    );
  END IF;
END$$;

ALTER TABLE portfolio_entries
  ADD COLUMN IF NOT EXISTS category portfolio_entry_category NOT NULL DEFAULT 'others';

UPDATE portfolio_entries pe
SET category = CASE a.asset_type
                 WHEN 'stock'       THEN 'stocks'
                 WHEN 'etf'         THEN 'etfs'
                 WHEN 'crypto'      THEN 'cryptos'
                 WHEN 'bond'        THEN 'bonds'
                 WHEN 'cash'        THEN 'cash'
                 WHEN 'real_estate' THEN 'real_estates'
                 WHEN 'commodity'   THEN 'commodities'
                 ELSE 'others'
               END::portfolio_entry_category
FROM assets a
WHERE a.id = pe.asset_id;
