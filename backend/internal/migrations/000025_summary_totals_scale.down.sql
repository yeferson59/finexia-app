-- Restore the 000024 view: the same totals, unrounded. Their scale is then
-- whatever multiplying quantity, price and a rate produces, which is what the
-- up migration bounded.

CREATE OR REPLACE VIEW portfolio_summary AS
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
    -- A position is unconverted when either leg has no rate: the cost was
    -- settled in one currency and the asset is quoted in another, and the two
    -- are looked up separately. Counted on the same GROUP BY as the sums below
    -- so the count and the totals can never describe different rows.
    COUNT(DISTINCT pe.asset_id) FILTER (
      WHERE fx.cost_rate IS NULL OR fx.value_rate IS NULL
    )                              AS positions_unconverted,
    -- Cost basis: quantity × weighted-avg-cost, converted to base currency.
    COALESCE(SUM(
      pe.quantity * pe.price * COALESCE(fx.cost_rate, 1)
    ), 0)                          AS total_cost_base,
    -- Market value: quantity × the best price this user is entitled to.
    COALESCE(SUM(
      pe.quantity * COALESCE(uap.price, a.current_price, pe.price) * COALESCE(fx.value_rate, 1)
    ), 0)                          AS total_market_value,
    p.created_at
  FROM portfolios p
  LEFT JOIN portfolio_entries pe ON pe.portfolio_id = p.id
  LEFT JOIN assets a              ON a.id = pe.asset_id
  LEFT JOIN user_asset_prices uap ON uap.asset_id = pe.asset_id AND uap.user_id = p.user_id
  -- One lateral per entry, so each rate is resolved once and used by the sum,
  -- the count and nothing else.
  LEFT JOIN LATERAL (
    SELECT
      fx_rate(p.user_id, pe.cost_currency, p.base_currency)                        AS cost_rate,
      fx_rate(p.user_id, COALESCE(a.currency, pe.cost_currency), p.base_currency)  AS value_rate
  ) fx ON TRUE
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
  total_positions - positions_priced_own - positions_priced_manual  AS positions_at_cost,
  positions_unconverted
FROM base;
