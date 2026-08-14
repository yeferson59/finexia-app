-- Bound the scale of the summary totals.
--
-- The totals are a product of three numerics — quantity (scale 8), price
-- (scale 8) and, since 000024 always in practice, an exchange rate — and
-- Postgres adds the scales of what it multiplies. An inverted or hopped rate
-- carries twenty-odd decimals of its own, so the sums were coming out with
-- thirty and more.
--
-- That is not a rounding nicety. The totals leave the view as text and are
-- parsed by gofinance's decimal, which caps at 19 digits after the point and
-- errors beyond it, so the display-currency conversion (?currency=) turned an
-- ordinary portfolio into a 500 the moment a real rate existed for it. Nothing
-- surfaced it until now because the shared table only held pairs almost nobody
-- converted; the ECB feed made rates the normal case.
--
-- Eight decimals is what every money column in this schema stores, and six is
-- what the snapshot table keeps for a percentage — rounding to those here loses
-- nothing that was ever persisted.
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
    COUNT(DISTINCT pe.asset_id) FILTER (
      WHERE fx.cost_rate IS NULL OR fx.value_rate IS NULL
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
