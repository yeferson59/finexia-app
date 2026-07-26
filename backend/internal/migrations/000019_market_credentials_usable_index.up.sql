-- The partial index 000018 created did not match the queries it was meant to
-- serve.
--
-- It was declared WHERE status = 'active', but both sync-facing queries filter
-- WHERE status <> 'invalid' — UsersWithCredentials, which picks the users the
-- daily job walks, and GetSealedCredentials, which loads the keys to build
-- their provider chain. That predicate also admits 'rate_limited', which is
-- deliberate: a spent quota is a reason to try again tomorrow, not a reason to
-- stop looking at the key. A partial index whose predicate is narrower than the
-- query's cannot be used to answer it, so the index sat unused.
DROP INDEX IF EXISTS idx_market_credentials_active;

CREATE INDEX idx_market_credentials_usable
  ON market_credentials(user_id)
  WHERE status <> 'invalid';
