-- Stop the summary from inventing a rate of 1.
--
-- Since 000018 the totals in portfolio_summary have converted each position
-- into the portfolio's base currency, and since 000020 they read the user's own
-- rate before the shared one. What neither addressed is the last arm of that
-- COALESCE: when no rate exists at all the position is multiplied by 1, so the
-- summary quietly asserts that one euro is one dollar. Nothing downstream can
-- tell that total apart from a converted one — the number has the right shape
-- and the wrong meaning, which is the worst way for a financial figure to be
-- wrong.
--
-- Two things change here.
--
-- First, the lookup stops being a pair of inline subqueries and becomes
-- fx_rate(), which resolves a rate the same way the application does
-- (portfolio.GetConversionRate): the user's own row before the shared one, the
-- inverse of either when only the opposite direction was stored, and a two-leg
-- hop through USD when no direct pair exists. Having one rule expressed twice
-- is bad enough; having two layers disagree about whether a portfolio can be
-- converted is what makes a warning in one screen contradict a total in
-- another. It returns NULL when nothing connects the pair, which is what makes
-- the second change possible.
--
-- Second, positions_unconverted counts the positions whose totals are NOT in
-- the base currency. The totals still include them at face value — dropping
-- them would understate the portfolio, and the holdings endpoint takes the same
-- decision with its fxConverted flag — but now a client can say so instead of
-- presenting a mixed-currency sum as a total. It is the same device 000020
-- introduced for price provenance: a count that qualifies a number rather than
-- a number that hides its own provenance.

-- fx_pair_rate resolves one direction of one pair, or NULL.
--
-- Precedence mirrors portfolio.pairRate: the rate this user's key fetched wins
-- over the shared, admin-or-public one, and either direction is usable because
-- a stored rate can be inverted. Non-positive rows are skipped rather than
-- returned: money.Convert rejects them, so a corrupt row must fall through to
-- the next candidate instead of poisoning the total.
CREATE OR REPLACE FUNCTION fx_pair_rate(p_user_id UUID, p_from CHAR(3), p_to CHAR(3))
RETURNS NUMERIC
LANGUAGE plpgsql
STABLE
AS $$
DECLARE
  v_rate NUMERIC;
BEGIN
  IF p_from IS NULL OR p_to IS NULL THEN
    RETURN NULL;
  END IF;

  IF p_from = p_to THEN
    RETURN 1;
  END IF;

  SELECT uer.rate INTO v_rate
  FROM user_exchange_rates uer
  WHERE uer.user_id = p_user_id AND uer.from_currency = p_from AND uer.to_currency = p_to;
  IF v_rate IS NOT NULL AND v_rate > 0 THEN
    RETURN v_rate;
  END IF;

  SELECT er.rate INTO v_rate
  FROM exchange_rates er
  WHERE er.from_currency = p_from AND er.to_currency = p_to
  ORDER BY er.rate_date DESC
  LIMIT 1;
  IF v_rate IS NOT NULL AND v_rate > 0 THEN
    RETURN v_rate;
  END IF;

  SELECT uer.rate INTO v_rate
  FROM user_exchange_rates uer
  WHERE uer.user_id = p_user_id AND uer.from_currency = p_to AND uer.to_currency = p_from;
  IF v_rate IS NOT NULL AND v_rate > 0 THEN
    RETURN 1 / v_rate;
  END IF;

  SELECT er.rate INTO v_rate
  FROM exchange_rates er
  WHERE er.from_currency = p_to AND er.to_currency = p_from
  ORDER BY er.rate_date DESC
  LIMIT 1;
  IF v_rate IS NOT NULL AND v_rate > 0 THEN
    RETURN 1 / v_rate;
  END IF;

  RETURN NULL;
END;
$$;

-- fx_rate is the whole rule: the direct pair, or a hop through USD.
--
-- USD is the hub because that is where the rates are: the public ECB feed
-- publishes USD↔major and dolarapi USD→COP, so a pair like EUR→COP exists only
-- as two legs. Returns NULL when even that fails, and a NULL is what the caller
-- must treat as "not converted" rather than as 1.
CREATE OR REPLACE FUNCTION fx_rate(p_user_id UUID, p_from CHAR(3), p_to CHAR(3))
RETURNS NUMERIC
LANGUAGE plpgsql
STABLE
AS $$
DECLARE
  v_rate     NUMERIC;
  v_from_usd NUMERIC;
  v_usd_to   NUMERIC;
BEGIN
  v_rate := fx_pair_rate(p_user_id, p_from, p_to);
  IF v_rate IS NOT NULL THEN
    RETURN v_rate;
  END IF;

  v_from_usd := fx_pair_rate(p_user_id, p_from, 'USD');
  v_usd_to   := fx_pair_rate(p_user_id, 'USD', p_to);
  IF v_from_usd IS NOT NULL AND v_usd_to IS NOT NULL THEN
    RETURN v_from_usd * v_usd_to;
  END IF;

  RETURN NULL;
END;
$$;

-- The view keeps every existing column with its name, type and position, and
-- appends one — the only shape change CREATE OR REPLACE VIEW allows, and the
-- reason the two queries that read it need no coordination.
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
