-- Restore the 000020 view: inline rate lookups, direct pair only, and a rate of
-- 1 when none is found. The column has to go, and CREATE OR REPLACE VIEW cannot
-- drop one, so the view is dropped and rebuilt — nothing else depends on it.

DROP VIEW IF EXISTS portfolio_summary;

CREATE VIEW portfolio_summary AS
WITH base AS (
  SELECT
    p.id                           AS portfolio_id,
    p.user_id,
    p.name                         AS portfolio_name,
    p.base_currency,
    COUNT(DISTINCT pe.asset_id)    AS total_positions,
    COUNT(DISTINCT pe.asset_id) FILTER (
      WHERE uap.price IS NOT NULL
    )                              AS positions_priced_own,
    COUNT(DISTINCT pe.asset_id) FILTER (
      WHERE uap.price IS NULL AND a.current_price IS NOT NULL
    )                              AS positions_priced_manual,
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
  created_at,
  positions_priced_own,
  positions_priced_manual,
  total_positions - positions_priced_own - positions_priced_manual  AS positions_at_cost
FROM base;

DROP FUNCTION IF EXISTS fx_rate(UUID, CHAR(3), CHAR(3));
DROP FUNCTION IF EXISTS fx_pair_rate(UUID, CHAR(3), CHAR(3));
