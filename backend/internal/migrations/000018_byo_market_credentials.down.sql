-- The view must go back to the shared-price definition before the per-user
-- tables it references are dropped, or the drop fails on the dependency.
CREATE OR REPLACE VIEW portfolio_summary AS
WITH base AS (
  SELECT
    p.id                           AS portfolio_id,
    p.user_id,
    p.name                         AS portfolio_name,
    p.base_currency,
    COUNT(DISTINCT pe.asset_id)    AS total_positions,
    COALESCE(SUM(
      pe.quantity * pe.price *
      COALESCE(
        (SELECT er.rate FROM exchange_rates er
         WHERE er.from_currency = pe.cost_currency
           AND er.to_currency   = p.base_currency
         ORDER BY er.rate_date DESC LIMIT 1),
        1
      )
    ), 0)                          AS total_cost_base,
    COALESCE(SUM(
      pe.quantity *
      COALESCE(a.current_price, pe.price) *
      COALESCE(
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

-- The cleared assets.current_price and the deleted exchange_rates rows are not
-- restored: they were provider data the application should not have been
-- sharing, and a sync under the old model would repopulate them anyway.
DROP TABLE user_exchange_rates;

DROP TABLE user_asset_prices;

DROP INDEX IF EXISTS idx_market_credentials_active;

DROP TABLE market_credentials;
