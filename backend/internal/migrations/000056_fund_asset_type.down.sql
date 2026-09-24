-- The enum value stays, and on purpose.
--
-- Removing a value from an enum means rebuilding the type and every column that
-- uses it, and an asset already recorded as a fund would have to be rewritten as
-- something it is not. An unused value costs nothing, and the up migration adds
-- it with IF NOT EXISTS, so rolling forward again finds it there. 000057's down
-- is what removes the data that depends on it.
SELECT 1;
