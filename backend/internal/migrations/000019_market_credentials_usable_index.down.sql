DROP INDEX IF EXISTS idx_market_credentials_usable;

CREATE INDEX idx_market_credentials_active
  ON market_credentials(user_id)
  WHERE status = 'active';
