-- Where a shared exchange rate came from.
--
-- Until now this table had a single kind of writer — an operator typing a rate
-- (POST /exchange-rates) or uploading a spreadsheet — so provenance was
-- implicit and a column for it would have said 'manual' on every row.
--
-- It stops being implicit with the public feed added alongside. 000018 emptied
-- this table and moved everything provider-fetched into user_exchange_rates,
-- because a rate paid for with one user's key may not be served to another.
-- That reasoning turns on the key: a source that needs no credential, publishes
-- public official data and places no redistribution terms on it can be fetched
-- once and served to everybody, which is exactly what this table is for.
-- dolarapi.com's TRM endpoint is such a source, and it is what fills the
-- USD→COP pair from now on.
--
-- 'manual' is the default so every existing row keeps its meaning and both
-- admin writers stay correct without a code change.
ALTER TABLE exchange_rates
  ADD COLUMN IF NOT EXISTS source VARCHAR(32) NOT NULL DEFAULT 'manual';
