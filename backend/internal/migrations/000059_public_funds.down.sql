-- The published values stay as marks: they are what the funds were worth, and
-- the prices and snapshots they set agree with them. Only who wrote them, and
-- the links that would keep adding more, go.
ALTER TABLE fund_marks DROP CONSTRAINT IF EXISTS fund_marks_source;
ALTER TABLE fund_marks DROP COLUMN IF EXISTS source;

DROP INDEX IF EXISTS idx_user_funds_public_fund;
ALTER TABLE user_funds DROP CONSTRAINT IF EXISTS user_funds_public_units;
ALTER TABLE user_funds DROP COLUMN IF EXISTS public_fund_id;

DROP TABLE IF EXISTS public_funds;
