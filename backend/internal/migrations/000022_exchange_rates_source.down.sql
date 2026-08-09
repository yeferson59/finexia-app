-- The rows the public feed wrote are left in place: they are ordinary shared
-- rates, and dropping the column only loses the record of where they came from.
ALTER TABLE exchange_rates
  DROP COLUMN IF EXISTS source;
