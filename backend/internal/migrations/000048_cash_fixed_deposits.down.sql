-- The index goes. The enum value stays, and on purpose.
--
-- Removing a value from an enum means rebuilding the type and every column that
-- uses it, and a rate already posted 'at_maturity' would have to be rewritten
-- as something it is not — a deposit that pays at the end is not a deposit that
-- pays monthly. An unused value costs nothing, and the up migration adds it
-- with IF NOT EXISTS, so rolling forward again finds it there.
DROP INDEX IF EXISTS idx_cash_pockets_maturing;
