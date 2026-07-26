-- BYO-key: market data is fetched with each user's own provider API key.
--
-- The keys are sealed with envelope encryption (internal/platform/secretbox):
-- a random data key per credential, itself wrapped under a key encryption key
-- that lives in the environment and never in this database. A dump of these
-- rows is therefore not enough to recover anybody's key.
CREATE TABLE market_credentials (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id          UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  provider         VARCHAR(32) NOT NULL,
  -- Which KEK wrapped the data key, so a key can be rotated without
  -- rewriting every row at once.
  kek_version      INT NOT NULL,
  wrapped_dek      BYTEA NOT NULL,
  nonce            BYTEA NOT NULL,
  ciphertext       BYTEA NOT NULL,
  -- Last four characters of the key, the only plaintext fragment kept. It
  -- exists so the UI can show which key is configured without ever holding
  -- the secret.
  last4            VARCHAR(4) NOT NULL,
  -- active | invalid | rate_limited. Set by the verification call and by the
  -- sync job, so the user can be told their key stopped working instead of
  -- silently getting no prices.
  status           VARCHAR(16) NOT NULL DEFAULT 'active',
  last_verified_at TIMESTAMPTZ,
  last_error       TEXT,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT uq_market_credentials_user_provider UNIQUE (user_id, provider)
);

-- The sync job walks the users that actually have a usable credential.
CREATE INDEX idx_market_credentials_active
  ON market_credentials(user_id)
  WHERE status = 'active';

-- Prices and rates fetched with a personal API key belong to the user whose
-- key paid for them: provider terms do not allow redistributing them to other
-- users. These two tables are what keeps that data scoped to its owner,
-- replacing the shared assets.current_price for everything provider-sourced.
CREATE TABLE user_asset_prices (
  user_id    UUID NOT NULL REFERENCES users(id)  ON DELETE CASCADE,
  asset_id   UUID NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
  price      NUMERIC(20, 8) NOT NULL,
  currency   CHAR(3) NOT NULL,
  -- Provider that produced the value, for display and for debugging which key
  -- served a given price.
  source     VARCHAR(32) NOT NULL,
  fetched_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (user_id, asset_id)
);

CREATE TABLE user_exchange_rates (
  user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  from_currency CHAR(3) NOT NULL,
  to_currency   CHAR(3) NOT NULL,
  rate          NUMERIC(24, 10) NOT NULL,
  source        VARCHAR(32) NOT NULL,
  fetched_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (user_id, from_currency, to_currency)
);

-- Whatever sits in these columns today was fetched with the operator's key
-- under the old shared model, so it carries the same redistribution problem
-- this migration exists to solve. Clearing it means users briefly see their
-- holdings valued at cost until their first sync, which is the honest state.
-- From here on the column is only ever written by the admin manual-price
-- endpoint (PATCH /portfolios/assets/:id/price), which is operator-entered
-- data with no provider licence attached.
UPDATE assets SET current_price = NULL, price_updated_at = NULL;

-- Same reasoning for the shared exchange rate table: its rows came from the
-- operator's key. It stays in place for admin-entered rates.
DELETE FROM exchange_rates;

-- portfolio_summary is where the shared price actually reached the user: two
-- repository queries read this view for every portfolio screen. It now prefers
-- the price and the rate the user's own key fetched, and only then falls back.
--
-- The fallback order matters and is deliberate:
--   1. user_asset_prices  — this user's own data.
--   2. assets.current_price — admin-entered manual price, operator data with no
--      provider licence attached (PATCH /portfolios/assets/:id/price).
--   3. pe.price — the user's own cost basis, so a holding is valued at cost
--      rather than at somebody else's price.
-- At no point can one user's provider data reach another's valuation.
CREATE OR REPLACE VIEW portfolio_summary AS
WITH base AS (
  SELECT
    p.id                           AS portfolio_id,
    p.user_id,
    p.name                         AS portfolio_name,
    p.base_currency,
    COUNT(DISTINCT pe.asset_id)    AS total_positions,
    -- Cost basis: quantity × weighted-avg-cost, converted to base currency.
    COALESCE(SUM(
      pe.quantity * pe.price *
      COALESCE(
        (SELECT uer.rate FROM user_exchange_rates uer
         WHERE uer.user_id       = p.user_id
           AND uer.from_currency = pe.cost_currency
           AND uer.to_currency   = p.base_currency),
        (SELECT er.rate FROM exchange_rates er
         WHERE er.from_currency = pe.cost_currency
           AND er.to_currency   = p.base_currency
         ORDER BY er.rate_date DESC LIMIT 1),
        1
      )
    ), 0)                          AS total_cost_base,
    -- Market value: quantity × the best price this user is entitled to.
    COALESCE(SUM(
      pe.quantity *
      COALESCE(uap.price, a.current_price, pe.price) *
      COALESCE(
        (SELECT uer.rate FROM user_exchange_rates uer
         WHERE uer.user_id       = p.user_id
           AND uer.from_currency = COALESCE(a.currency, pe.cost_currency)
           AND uer.to_currency   = p.base_currency),
        (SELECT er.rate FROM exchange_rates er
         WHERE er.from_currency = COALESCE(a.currency, pe.cost_currency)
           AND er.to_currency   = p.base_currency
         ORDER BY er.rate_date DESC LIMIT 1),
        1
      )
    ), 0)                          AS total_market_value,
    p.created_at
  FROM portfolios p
  LEFT JOIN portfolio_entries pe ON pe.portfolio_id = p.id
  LEFT JOIN assets a              ON a.id = pe.asset_id
  LEFT JOIN user_asset_prices uap ON uap.asset_id = pe.asset_id AND uap.user_id = p.user_id
  GROUP BY p.id, p.user_id, p.name, p.base_currency, p.created_at
)
SELECT
  portfolio_id,
  user_id,
  portfolio_name,
  base_currency,
  total_positions,
  total_cost_base,
  total_market_value,
  total_market_value - total_cost_base                              AS total_gain_loss,
  CASE WHEN total_cost_base > 0
    THEN (total_market_value - total_cost_base) / total_cost_base * 100
    ELSE 0
  END                                                               AS total_gain_loss_pct,
  created_at
FROM base;
