-- Price provenance in the summary.
--
-- 000018 made the valuation prefer the user's own price, then the operator's
-- manual one, and fall back to the cost basis. It reported only the resulting
-- number, which makes two very different things numerically indistinguishable:
-- a portfolio priced by the market, and a portfolio valued at what it was
-- bought for. The second has a gain/loss of exactly zero by construction, not
-- because the market said so, and nothing downstream could tell.
--
-- These counts break total_positions down by where each position's price came
-- from. They are what lets a client label a holding, warn that a total is only
-- partly priced, and suppress a 0% return that means "no data" rather than
-- "no movement". They are counts, not prices: no provider data crosses between
-- users here.
--
-- Appended at the end of the column list because CREATE OR REPLACE VIEW can
-- only add columns there — the existing ones keep their names, order and types,
-- so the two queries that read this view are unaffected.
CREATE OR REPLACE VIEW portfolio_summary AS
WITH base AS (
  SELECT
    p.id                           AS portfolio_id,
    p.user_id,
    p.name                         AS portfolio_name,
    p.base_currency,
    COUNT(DISTINCT pe.asset_id)    AS total_positions,
    -- Same three-way preference the market value below applies, counted
    -- instead of summed. The manual arm repeats the "own price is absent"
    -- condition because COALESCE's precedence has to hold here too: a position
    -- with both prices is priced by its owner's key, not by the operator.
    COUNT(DISTINCT pe.asset_id) FILTER (
      WHERE uap.price IS NOT NULL
    )                              AS positions_priced_own,
    COUNT(DISTINCT pe.asset_id) FILTER (
      WHERE uap.price IS NULL AND a.current_price IS NOT NULL
    )                              AS positions_priced_manual,
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
  created_at,
  positions_priced_own,
  positions_priced_manual,
  -- Derived here rather than left to the client: the three counts partition
  -- total_positions exactly, and computing it from the same GROUP BY is what
  -- guarantees they keep adding up.
  total_positions - positions_priced_own - positions_priced_manual  AS positions_at_cost
FROM base;
