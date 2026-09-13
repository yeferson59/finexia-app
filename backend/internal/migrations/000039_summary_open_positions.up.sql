-- Count only the positions that are still held.
--
-- A position sold down to nothing keeps its row in portfolio_entries: its buys
-- and sells hang off it, and deleting it would take that history with it. The
-- summary counted it all the same, so a portfolio holding two assets with a
-- third sold out listed as "3 posiciones" while its own page — which sets closed
-- positions apart — said two. The platform stats and the sector allocation
-- already leave those rows out with quantity > 0; this is the same cut.
--
-- Every count takes it, not only total_positions. positions_at_cost is total
-- minus the two priced counts, so filtering one side alone could drive it
-- negative; and a sold-out position in a currency with no rate was reported as
-- unconverted, warning about a total that had nothing added at face value.
--
-- The totals need no filter: a quantity of zero adds zero to both sums.
CREATE OR REPLACE VIEW portfolio_summary AS
WITH base AS (
  SELECT
    p.id                           AS portfolio_id,
    p.user_id,
    p.name                         AS portfolio_name,
    p.base_currency,
    COUNT(DISTINCT pe.asset_id) FILTER (
      WHERE pe.quantity > 0
    )                              AS total_positions,
    COUNT(DISTINCT pe.asset_id) FILTER (
      WHERE pe.quantity > 0 AND uap.price IS NOT NULL
    )                              AS positions_priced_own,
    COUNT(DISTINCT pe.asset_id) FILTER (
      WHERE pe.quantity > 0 AND uap.price IS NULL AND a.current_price IS NOT NULL
    )                              AS positions_priced_manual,
    COUNT(DISTINCT pe.asset_id) FILTER (
      WHERE pe.quantity > 0 AND (fx.cost_rate IS NULL OR fx.value_rate IS NULL)
    )                              AS positions_unconverted,
    ROUND(COALESCE(SUM(
      pe.quantity * pe.price * COALESCE(fx.cost_rate, 1)
    ), 0), 8)                      AS total_cost_base,
    ROUND(COALESCE(SUM(
      pe.quantity * COALESCE(uap.price, a.current_price, pe.price) * COALESCE(fx.value_rate, 1)
    ), 0), 8)                      AS total_market_value,
    p.created_at
  FROM portfolios p
  LEFT JOIN portfolio_entries pe ON pe.portfolio_id = p.id
  LEFT JOIN assets a              ON a.id = pe.asset_id
  LEFT JOIN user_asset_prices uap ON uap.asset_id = pe.asset_id AND uap.user_id = p.user_id
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
    -- The division is the other unbounded operation here, and a percentage
    -- displayed to two decimals has no use for more than six.
    THEN ROUND((total_market_value - total_cost_base) / total_cost_base * 100, 6)
    ELSE 0
  END                                                               AS total_gain_loss_pct,
  created_at,
  positions_priced_own,
  positions_priced_manual,
  total_positions - positions_priced_own - positions_priced_manual  AS positions_at_cost,
  positions_unconverted
FROM base;
