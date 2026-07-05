-- Truncate any value longer than the old limit before narrowing the column,
-- otherwise the ALTER fails on rows written while 000011 was applied.
UPDATE sessions SET ip_address = LEFT(ip_address, 14) WHERE LENGTH(ip_address) > 14;
ALTER TABLE sessions ALTER COLUMN ip_address TYPE VARCHAR(14);
