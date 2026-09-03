package portfolio

import (
	"context"
	"time"

	"uuid"

	"github.com/yeferson59/gofinance/v2/money"
)

// GetAllPortfolioSummaryRows reads what the snapshot job persists for each
// portfolio: the totals from the portfolio_summary view plus the day's
// composition by asset type.
//
// The composition is aggregated here, in the same statement, rather than from
// the allocation endpoint: that one answers across all of a user's portfolios
// and in a currency they ask for, while a snapshot is one portfolio in its own
// base currency. Above all, the value expression below is the view's
// total_market_value term spelled out per asset type — same price fallback
// (own price, then the catalog's, then cost), same rate, same 8-decimal
// rounding — so the slices add up to the total_value stored beside them
// instead of being a second opinion about the same positions.
func (r *PostgresRepository) GetAllPortfolioSummaryRows(ctx context.Context) ([]SnapshotRow, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			p.id,
			p.base_currency,
			COALESCE(ps.total_market_value, 0)::text,
			COALESCE(ps.total_cost_base,    0)::text,
			COALESCE(ps.total_gain_loss,    0)::text,
			COALESCE(ps.total_gain_loss_pct,0)::text,
			COALESCE(alloc.allocation, '{}'::jsonb)::text
		FROM portfolios p
		LEFT JOIN portfolio_summary ps ON ps.portfolio_id = p.id
		LEFT JOIN LATERAL (
			-- asset_type es un ENUM: jsonb_object_agg pide la clave en text y
			-- Postgres no lo convierte solo.
			SELECT jsonb_object_agg(byType.asset_type::text, byType.market_value::text) AS allocation
			FROM (
				SELECT
					a.asset_type,
					ROUND(COALESCE(SUM(
						pe.quantity * COALESCE(uap.price, a.current_price, pe.price)
						            * COALESCE(fx.value_rate, 1)
					), 0), 8) AS market_value
				FROM portfolio_entries pe
				JOIN assets a ON a.id = pe.asset_id
				LEFT JOIN user_asset_prices uap
					ON uap.asset_id = pe.asset_id AND uap.user_id = p.user_id
				LEFT JOIN LATERAL (
					SELECT fx_rate(
						p.user_id,
						COALESCE(a.currency, pe.cost_currency),
						p.base_currency
					) AS value_rate
				) fx ON TRUE
				WHERE pe.portfolio_id = p.id
				GROUP BY a.asset_type
			) byType
		) alloc ON TRUE
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]SnapshotRow, 0)
	for rows.Next() {
		var row SnapshotRow
		if err := rows.Scan(
			&row.PortfolioID,
			&row.BaseCurrency,
			&row.TotalMarketValue,
			&row.TotalCostBase,
			&row.TotalGainLoss,
			&row.TotalGainLossPct,
			&row.Allocation,
		); err != nil {
			return nil, err
		}

		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// UpsertPortfolioSnapshot writes one portfolio's row for a day, replacing it if
// the job already ran. It takes the row whole rather than six positional
// arguments: three of them were adjacent strings holding a value, a gain and a
// percentage, which a caller could reorder without the compiler noticing.
//
// allocation is refreshed along with the totals. A same-day re-run happens
// after positions moved, so leaving it would pair the new total with the old
// composition — the row would contradict itself.
func (r *PostgresRepository) UpsertPortfolioSnapshot(
	ctx context.Context,
	row SnapshotRow,
	snapshotDate time.Time,
) error {
	allocation := row.Allocation
	if allocation == "" {
		allocation = "{}"
	}

	_, err := r.db.Exec(ctx, `
		INSERT INTO portfolio_snapshots
			(portfolio_id, snapshot_date, total_value, currency, allocation, total_gain_loss, total_gain_loss_pct)
		VALUES ($1, $2::date, $3::numeric, $4, $5::jsonb, $6::numeric, $7::numeric)
		ON CONFLICT (portfolio_id, snapshot_date)
		DO UPDATE SET
			total_value         = EXCLUDED.total_value,
			allocation          = EXCLUDED.allocation,
			total_gain_loss     = EXCLUDED.total_gain_loss,
			total_gain_loss_pct = EXCLUDED.total_gain_loss_pct
	`,
		row.PortfolioID,
		snapshotDate,
		row.TotalMarketValue,
		row.BaseCurrency,
		allocation,
		row.TotalGainLoss,
		row.TotalGainLossPct,
	)

	return err
}

// GetPortfolioValuesAsOf returns what each of the user's portfolios was worth
// at its most recent snapshot on or before asOf, and which date that was.
//
// It looks backwards rather than at asOf exactly because the snapshot job can
// miss a day — an outage, a deploy — and a comparison against a day that was
// never snapshotted would report the whole portfolio as having appeared out of
// nowhere. Taking the latest snapshot before the cutoff instead widens the
// window rather than inventing a number, which is why the date comes back with
// the amount.
//
// DISTINCT ON gives one row per portfolio, so a caller can report each
// portfolio's own movement and sum the same figures for the account total —
// the parts then add up to the whole by construction.
//
// A portfolio with no snapshot that old is simply absent from the result: it
// did not exist a week ago, so there is nothing to compare it against. An
// empty result means the whole account has no history yet, the normal state
// in its first week.
func (r *PostgresRepository) GetPortfolioValuesAsOf(ctx context.Context, userID uuid.UUID, asOf time.Time) ([]PortfolioValuePoint, error) {
	rows, err := r.db.Query(ctx, `
		SELECT DISTINCT ON (ps.portfolio_id)
			ps.portfolio_id, ps.snapshot_date, ps.total_value::text
		FROM portfolio_snapshots ps
		JOIN portfolios p ON p.id = ps.portfolio_id
		WHERE p.user_id = $1
		  AND ps.snapshot_date <= $2::date
		ORDER BY ps.portfolio_id, ps.snapshot_date DESC
	`, userID, asOf)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	points := make([]PortfolioValuePoint, 0)
	for rows.Next() {
		var point PortfolioValuePoint
		if err := rows.Scan(&point.PortfolioID, &point.Date, &point.TotalValue); err != nil {
			return nil, err
		}
		points = append(points, point)
	}

	return points, rows.Err()
}

// GetPortfolioGrowthByUserID aggregates every portfolio's snapshots into one
// series, expressed in a single currency.
//
// Each snapshot is stored in its own portfolio's base currency, so the previous
// bare SUM added a EUR portfolio to a USD one and produced a total in no
// currency at all. Every row is converted first, to `currency` when the caller
// asks for one and to the account's preferred currency otherwise — the same
// rule the summary and allocation endpoints follow.
//
// The rate is today's, applied to every date in the series: the app keeps the
// latest rate per pair, not a history. Past points are therefore restated at
// the current rate, which is the usual constant-currency presentation and the
// honest limit of the data we keep — the alternative was adding the numbers up
// raw. A snapshot whose currency has no rate is left as-is and counted in
// PortfoliosUnconverted so the caller can say so rather than pass off a total
// that silently mixes units.
//
// Each point also carries the net external cash flow of the stretch that ends
// on it, so a caller can tell the value the market added from the value the
// owner deposited. A transaction counts on the first snapshot computed after it
// was recorded — the snapshot that first reflects it — and is converted with
// the same rate as the snapshots it sits next to.
//
// The flow of a position loaded with history behind it is its cost, while the
// value it brings in is what it is worth today, so the gain it accumulated
// before the app ever saw it lands on that one day. That is the honest limit of
// a series that only knows what it was told, when it was told: nobody watched
// that gain happen, and there is no earlier point to spread it over.
func (r *PostgresRepository) GetPortfolioGrowthByUserID(ctx context.Context, userID uuid.UUID, currency money.Currency, hasSince bool, since time.Time) ([]GrowthPoint, error) {
	rows, err := r.db.Query(ctx, `
		WITH converted AS (
			SELECT
				ps.snapshot_date,
				target.code                               AS currency,
				ps.total_value     * COALESCE(fx.rate, 1) AS total_value,
				ps.total_gain_loss * COALESCE(fx.rate, 1) AS total_gain_loss,
				(fx.rate IS NULL)::int                    AS unconverted
			FROM portfolio_snapshots ps
			JOIN portfolios p ON p.id = ps.portfolio_id
			JOIN users u      ON u.id = p.user_id
			CROSS JOIN LATERAL (
				SELECT COALESCE(NULLIF($2::text, ''), u.preferred_currency, 'USD')::char(3) AS code
			) target
			CROSS JOIN LATERAL (
				SELECT fx_rate(p.user_id, ps.currency, target.code) AS rate
			) fx
			WHERE p.user_id = $1
			  AND ($3::boolean = FALSE OR ps.snapshot_date >= $4::date)
		), totals AS (
			SELECT
				snapshot_date,
				currency,
				SUM(total_value)         AS total_value,
				SUM(total_gain_loss)     AS total_gain_loss,
				SUM(unconverted)::bigint AS unconverted
			FROM converted
			GROUP BY snapshot_date, currency
		), flows AS (
			-- Each transaction is attributed to the first snapshot computed
			-- after it was recorded, which is the point of the series where its
			-- money actually shows up. Not to its transaction_date: past
			-- snapshots are never recomputed, so a position registered today
			-- with a trade date of two months ago moves the series today and
			-- leaves the older date untouched.
			SELECT
				(
					SELECT MIN(ps.snapshot_date)
					FROM portfolio_snapshots ps
					WHERE ps.portfolio_id = pe.portfolio_id
					  AND ps.created_at >= tx.created_at
				) AS snapshot_date,
				-- Two rates, and the order of them is the point. tx.fx_rate is
				-- historical: it carries the trade from the currency it was
				-- quoted in to the one the account settled it in, at the rate of
				-- that day. fxt.rate is current: it carries that settled amount
				-- into whatever currency the series is being read in. Using the
				-- current rate for both legs — which is what this did before
				-- fx_rate existed — restates a December purchase at today's
				-- rate and books the difference as a flow the owner never made.
				--
				-- The price and the fee are converted separately and handed in
				-- already settled, so transaction_cash_flow is left as what its
				-- own comment says it is: the sign convention, nothing else. It
				-- cannot be given the raw amounts and one rate, because a
				-- commission billed to the account never rode the trade's
				-- conversion and multiplying it by that rate invents a cost.
				transaction_cash_flow(
					tx.type,
					tx.quantity,
					tx.price * tx.fx_rate,
					transaction_fees_in_cost(tx.fees, tx.fees_currency, tx.currency, tx.fx_rate)
				) * COALESCE(fxt.rate, 1) AS amount
			FROM transactions tx
			JOIN portfolio_entries pe ON pe.id = tx.entry_id
			JOIN portfolios pf        ON pf.id = pe.portfolio_id
			JOIN users us             ON us.id = pf.user_id
			CROSS JOIN LATERAL (
				SELECT COALESCE(NULLIF($2::text, ''), us.preferred_currency, 'USD')::char(3) AS code
			) tgt
			CROSS JOIN LATERAL (
				SELECT fx_rate(pf.user_id, pe.cost_currency, tgt.code) AS rate
			) fxt
			WHERE pf.user_id = $1
		), net_flows AS (
			-- A transaction the snapshot job has not caught up with yet lands on
			-- NULL and waits for the snapshot that will cover it. One that lands
			-- outside the asked-for window drops off the join below, which is
			-- what should happen: the opening point of a window has no previous
			-- point to be measured against.
			SELECT snapshot_date, SUM(amount) AS net_flow
			FROM flows
			WHERE snapshot_date IS NOT NULL
			GROUP BY snapshot_date
		)
		SELECT
			t.snapshot_date,
			t.currency,
			t.total_value::text,
			(t.total_value - t.total_gain_loss)::text,
			t.total_gain_loss::text,
			CASE
				WHEN (t.total_value - t.total_gain_loss) > 0
				THEN ((t.total_gain_loss / (t.total_value - t.total_gain_loss)) * 100)::text
				ELSE '0'
			END,
			t.unconverted,
			COALESCE(nf.net_flow, 0)::text
		FROM totals t
		LEFT JOIN net_flows nf ON nf.snapshot_date = t.snapshot_date
		ORDER BY t.snapshot_date ASC
	`, userID, currencyParam(currency), hasSince, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]GrowthPoint, 0)
	for rows.Next() {
		var point GrowthPoint
		if err := rows.Scan(
			&point.Date,
			&point.Currency,
			&point.TotalValue,
			&point.TotalCostBase,
			&point.GainLoss,
			&point.GainLossPct,
			&point.PortfoliosUnconverted,
			&point.NetFlow,
		); err != nil {
			return nil, err
		}
		result = append(result, point)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *PostgresRepository) GetPortfolioGrowthByPortfolioID(ctx context.Context, userID, portfolioID uuid.UUID, hasSince bool, since time.Time) ([]GrowthPoint, error) {
	// One portfolio, one base currency: nothing to convert and nothing that can
	// fail to, which is why the unconverted count is a literal zero here. The
	// flows are the exception — a position costs in its own currency and can
	// differ from the portfolio's base, so those do go through fx_rate.
	rows, err := r.db.Query(ctx, `
		WITH points AS (
			SELECT
				ps.snapshot_date,
				ps.currency,
				ps.total_value,
				ps.total_gain_loss,
				ps.total_gain_loss_pct
			FROM portfolio_snapshots ps
			JOIN portfolios p ON p.id = ps.portfolio_id
			WHERE ps.portfolio_id = $1 AND p.user_id = $2
			  AND ($3::boolean = FALSE OR ps.snapshot_date >= $4::date)
		), flows AS (
			-- By when the transaction was recorded, not when it was traded: see
			-- the account-wide query for why.
			SELECT
				(
					SELECT MIN(ps.snapshot_date)
					FROM portfolio_snapshots ps
					WHERE ps.portfolio_id = $1
					  AND ps.created_at >= tx.created_at
				) AS snapshot_date,
				-- Historical rate onto the position's own currency, then the
				-- current one onto the portfolio's base, with the fee converted
				-- on its own side: see the account-wide query for why neither
				-- pair can be collapsed into one.
				transaction_cash_flow(
					tx.type,
					tx.quantity,
					tx.price * tx.fx_rate,
					transaction_fees_in_cost(tx.fees, tx.fees_currency, tx.currency, tx.fx_rate)
				) * COALESCE(fxt.rate, 1) AS amount
			FROM transactions tx
			JOIN portfolio_entries pe ON pe.id = tx.entry_id
			JOIN portfolios pf        ON pf.id = pe.portfolio_id
			CROSS JOIN LATERAL (
				SELECT fx_rate(pf.user_id, pe.cost_currency, pf.base_currency) AS rate
			) fxt
			WHERE pe.portfolio_id = $1 AND pf.user_id = $2
		), net_flows AS (
			SELECT snapshot_date, SUM(amount) AS net_flow
			FROM flows
			WHERE snapshot_date IS NOT NULL
			GROUP BY snapshot_date
		)
		SELECT
			pt.snapshot_date,
			pt.currency,
			pt.total_value::text,
			(pt.total_value - pt.total_gain_loss)::text,
			pt.total_gain_loss::text,
			pt.total_gain_loss_pct::text,
			0::bigint,
			COALESCE(nf.net_flow, 0)::text
		FROM points pt
		LEFT JOIN net_flows nf ON nf.snapshot_date = pt.snapshot_date
		ORDER BY pt.snapshot_date ASC
	`, portfolioID, userID, hasSince, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]GrowthPoint, 0)
	for rows.Next() {
		var point GrowthPoint
		if err := rows.Scan(
			&point.Date,
			&point.Currency,
			&point.TotalValue,
			&point.TotalCostBase,
			&point.GainLoss,
			&point.GainLossPct,
			&point.PortfoliosUnconverted,
			&point.NetFlow,
		); err != nil {
			return nil, err
		}
		result = append(result, point)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
