package portfolio

import (
	"context"
	"fmt"

	"uuid"

	"github.com/yeferson59/gofinance/v2/money"
)

// platformStatsSelect is the head shared by the two queries that report what a
// platform holds: the listing and the re-read that follows an update.
//
// The total is a cost basis — quantity × weighted-average cost — and until now
// it was summed straight out of the column, in whatever currency each entry
// was settled in. A platform holding one position bought in COP and three in
// USD returned their face values added together: a number in no currency at
// all, and wildly inflated whenever a minor-unit currency was in the mix. That
// is the same mistake portfolio_summary made before 000024, so the fix is the
// same one — resolve fx_rate() per entry, and count the entries it could not
// resolve instead of pretending a missing rate is 1.
//
// The target is the account's preferred currency unless the caller names one,
// and it travels back in display_currency: the amount is meaningless without
// knowing which currency it landed in, and the caller was previously left to
// assume dollars.
//
// %s is the placeholder carrying that requested currency — a query parameter
// in both callers, only at different positions.
const platformStatsSelect = `
	SELECT
		is_.id,
		is_.name,
		COALESCE(is_.description, ''),
		is_.source_type,
		is_.is_active,
		is_.created_at,
		is_.updated_at,
		COUNT(pe.id)::bigint AS investments,
		-- ROUND for the reason 000025 gives: an inverted or hopped rate carries
		-- twenty-odd decimals of its own, and the total leaves here as text for
		-- a decimal that errors past nineteen.
		ROUND(COALESCE(SUM(pe.quantity * pe.price * COALESCE(fx.cost_rate, 1)), 0), 8)::text AS total_value,
		-- What the platform is worth now, against what it cost above. The two
		-- are summed over the same rows and converted into the same currency, so
		-- their difference is a gain rather than an artefact of two scopes.
		ROUND(COALESCE(SUM(pe.quantity * v.price * COALESCE(fx.value_rate, 1)), 0), 8)::text AS market_value,
		target.currency AS display_currency,
		-- A position counts as unconverted when either leg failed: the cost and
		-- the market value are looked up separately, and a total that is right
		-- on one side and nominal on the other is still a mixed-currency total.
		(COUNT(pe.id) FILTER (
			WHERE fx.cost_rate IS NULL OR fx.value_rate IS NULL
		))::bigint AS positions_unconverted
	FROM investment_sources is_
	JOIN users u ON u.id = is_.user_id
	CROSS JOIN LATERAL (
		SELECT COALESCE(NULLIF(%s::text, ''), u.preferred_currency)::CHAR(3) AS currency
	) target
	LEFT JOIN portfolio_entries pe ON pe.source_id = is_.id
	LEFT JOIN assets a              ON a.id = pe.asset_id
	LEFT JOIN user_asset_prices uap ON uap.asset_id = pe.asset_id AND uap.user_id = is_.user_id
	-- The price and the currency it is quoted in are picked together, the way
	-- the holdings query picks them: a position valued at its own cost is an
	-- amount in the cost currency, not in the asset's, and converting it as if
	-- it were the asset's is how a position with no market price gets valued
	-- through the wrong rate.
	LEFT JOIN LATERAL (
		SELECT
			COALESCE(uap.price, a.current_price, pe.price) AS price,
			CASE
				WHEN COALESCE(uap.price, a.current_price) IS NOT NULL
					THEN COALESCE(a.currency, pe.cost_currency)
				ELSE pe.cost_currency
			END AS currency
	) v ON TRUE
	-- One lateral per entry so every sum and count reads the same two rates.
	LEFT JOIN LATERAL (
		SELECT
			fx_rate(is_.user_id, pe.cost_currency, target.currency) AS cost_rate,
			fx_rate(is_.user_id, v.currency, target.currency)       AS value_rate
	) fx ON TRUE`

func (r *PostgresRepository) GetPlatformsWithStats(ctx context.Context, userID uuid.UUID, displayCurrency money.Currency) ([]PlatformStats, error) {
	// Ordered by what the owner has put in, biggest first. Creation order said
	// nothing about the account — the platform opened last is not the one that
	// matters — and a list meant to answer "where is my money" has to open on
	// the answer. Ties break by name so the order is stable between reads
	// instead of drifting with whatever the planner returns.
	rows, err := r.db.Query(ctx, fmt.Sprintf(platformStatsSelect, "$2")+`
		WHERE is_.user_id = $1
		GROUP BY is_.id, target.currency
		ORDER BY COALESCE(SUM(pe.quantity * pe.price * COALESCE(fx.cost_rate, 1)), 0) DESC, is_.name ASC
	`, userID, currencyParam(displayCurrency))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]PlatformStats, 0)
	for rows.Next() {
		var p PlatformStats

		if err := rows.Scan(
			&p.ID, &p.Name, &p.Description, &p.SourceType,
			&p.IsActive, &p.CreatedAt, &p.UpdatedAt,
			&p.Investments, &p.TotalValue, &p.MarketValue,
			&p.DisplayCurrency, &p.PositionsUnconverted,
		); err != nil {
			return nil, err
		}

		result = append(result, p)
	}

	return result, rows.Err()
}

func (r *PostgresRepository) UpdatePlatform(ctx context.Context, userID, sourceID uuid.UUID, name, description string, sourceType SourceType, isActive bool) (PlatformStats, error) {
	tag, err := r.db.Exec(ctx, `
		UPDATE investment_sources
		SET name = $3, description = $4, source_type = $5, is_active = $6, updated_at = NOW()
		WHERE id = $1 AND user_id = $2
	`, sourceID, userID, name, description, sourceType, isActive)
	if err != nil {
		return PlatformStats{}, err
	}

	if tag.RowsAffected() == 0 {
		return PlatformStats{}, ErrPlatformNotFound
	}

	// An edit form has no display currency to ask for, so the empty string
	// leaves the totals in the account's preferred one.
	var p PlatformStats
	if err := r.db.QueryRow(ctx, fmt.Sprintf(platformStatsSelect, "$3")+`
		WHERE is_.id = $1 AND is_.user_id = $2
		GROUP BY is_.id, target.currency
	`, sourceID, userID, "").Scan(
		&p.ID, &p.Name, &p.Description, &p.SourceType,
		&p.IsActive, &p.CreatedAt, &p.UpdatedAt,
		&p.Investments, &p.TotalValue, &p.MarketValue,
		&p.DisplayCurrency, &p.PositionsUnconverted,
	); err != nil {
		return PlatformStats{}, err
	}

	return p, nil
}

func (r *PostgresRepository) DeletePlatform(ctx context.Context, userID, sourceID uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `
		DELETE FROM investment_sources WHERE id = $1 AND user_id = $2
	`, sourceID, userID)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return ErrPlatformNotFound
	}

	return nil
}

func (r *PostgresRepository) CreatePlatform(ctx context.Context, userID uuid.UUID, sourceType SourceType, name, desciption string) (InvestmentSource, error) {
	var platform InvestmentSource

	err := r.db.QueryRow(ctx, "INSERT INTO investment_sources(user_id, source_type, name, description) VALUES ($1, $2, $3, $4) RETURNING id, name, description, created_at, updated_at", userID, sourceType, name, desciption).Scan(&platform.ID, &platform.Name, &platform.Description, &platform.CreatedAt, &platform.UpdatedAt)
	if err != nil {
		return InvestmentSource{}, err
	}

	return platform, nil
}
