package portfolio

import (
	"context"
	"time"

	"github.com/google/uuid"
)

func (r *PostgresRepository) GetAllPortfolioSummaryRows(ctx context.Context) ([]SnapshotRow, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			p.id,
			p.base_currency,
			COALESCE(ps.total_market_value, 0)::text,
			COALESCE(ps.total_cost_base,    0)::text,
			COALESCE(ps.total_gain_loss,    0)::text,
			COALESCE(ps.total_gain_loss_pct,0)::text
		FROM portfolios p
		LEFT JOIN portfolio_summary ps ON ps.portfolio_id = p.id
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

func (r *PostgresRepository) UpsertPortfolioSnapshot(
	ctx context.Context,
	portfolioID uuid.UUID,
	snapshotDate time.Time,
	totalValue, currency, totalGainLoss, totalGainLossPct string,
) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO portfolio_snapshots
			(portfolio_id, snapshot_date, total_value, currency, allocation, total_gain_loss, total_gain_loss_pct)
		VALUES ($1, $2::date, $3::numeric, $4, '{}', $5::numeric, $6::numeric)
		ON CONFLICT (portfolio_id, snapshot_date)
		DO UPDATE SET
			total_value         = EXCLUDED.total_value,
			total_gain_loss     = EXCLUDED.total_gain_loss,
			total_gain_loss_pct = EXCLUDED.total_gain_loss_pct
	`, portfolioID, snapshotDate, totalValue, currency, totalGainLoss, totalGainLossPct)
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
// owner deposited. Every transaction is attributed to the first snapshot on or
// after its date — that is the snapshot that first reflects it — and converted
// with the same rate as the snapshots it sits next to.
func (r *PostgresRepository) GetPortfolioGrowthByUserID(
	ctx context.Context,
	userID uuid.UUID,
	currency string,
	hasSince bool,
	since time.Time,
) ([]GrowthPoint, error) {
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
			SELECT
				(
					SELECT MIN(t.snapshot_date)
					FROM totals t
					WHERE t.snapshot_date >= tx.transaction_date
				) AS snapshot_date,
				transaction_cash_flow(tx.type, tx.quantity, tx.price, tx.fees)
					* COALESCE(fxt.rate, 1) AS amount
			FROM transactions tx
			JOIN portfolio_entries pe ON pe.id = tx.entry_id
			JOIN portfolios pf        ON pf.id = pe.portfolio_id
			JOIN users us             ON us.id = pf.user_id
			CROSS JOIN LATERAL (
				SELECT COALESCE(NULLIF($2::text, ''), us.preferred_currency, 'USD')::char(3) AS code
			) tgt
			CROSS JOIN LATERAL (
				SELECT fx_rate(pf.user_id, tx.currency, tgt.code) AS rate
			) fxt
			WHERE pf.user_id = $1
		), net_flows AS (
			-- A transaction later than the last snapshot lands on NULL and waits
			-- for the snapshot that will cover it; one earlier than the first
			-- lands on the first, whose subperiod nobody measures.
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
	`, userID, currency, hasSince, since)
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

func (r *PostgresRepository) GetPortfolioGrowthByPortfolioID(
	ctx context.Context,
	userID, portfolioID uuid.UUID,
	hasSince bool,
	since time.Time,
) ([]GrowthPoint, error) {
	// One portfolio, one base currency: nothing to convert and nothing that can
	// fail to, which is why the unconverted count is a literal zero here. The
	// flows are the exception — a transaction carries its own currency and can
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
			SELECT
				(
					SELECT MIN(pt.snapshot_date)
					FROM points pt
					WHERE pt.snapshot_date >= tx.transaction_date
				) AS snapshot_date,
				transaction_cash_flow(tx.type, tx.quantity, tx.price, tx.fees)
					* COALESCE(fxt.rate, 1) AS amount
			FROM transactions tx
			JOIN portfolio_entries pe ON pe.id = tx.entry_id
			JOIN portfolios pf        ON pf.id = pe.portfolio_id
			CROSS JOIN LATERAL (
				SELECT fx_rate(pf.user_id, tx.currency, pf.base_currency) AS rate
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
