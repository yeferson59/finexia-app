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

func (r *PostgresRepository) GetPortfolioGrowthByUserID(
	ctx context.Context,
	userID uuid.UUID,
	hasSince bool,
	since time.Time,
) ([]GrowthPoint, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			ps.snapshot_date,
			SUM(ps.total_value)::text,
			(SUM(ps.total_value) - SUM(ps.total_gain_loss))::text,
			SUM(ps.total_gain_loss)::text,
			CASE
				WHEN (SUM(ps.total_value) - SUM(ps.total_gain_loss)) > 0
				THEN ((SUM(ps.total_gain_loss) / (SUM(ps.total_value) - SUM(ps.total_gain_loss))) * 100)::text
				ELSE '0'
			END
		FROM portfolio_snapshots ps
		JOIN portfolios p ON p.id = ps.portfolio_id
		WHERE p.user_id = $1
		  AND ($2::boolean = FALSE OR ps.snapshot_date >= $3::date)
		GROUP BY ps.snapshot_date
		ORDER BY ps.snapshot_date ASC
	`, userID, hasSince, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]GrowthPoint, 0)
	for rows.Next() {
		var point GrowthPoint
		if err := rows.Scan(
			&point.Date,
			&point.TotalValue,
			&point.TotalCostBase,
			&point.GainLoss,
			&point.GainLossPct,
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
	rows, err := r.db.Query(ctx, `
		SELECT
			ps.snapshot_date,
			ps.total_value::text,
			(ps.total_value - ps.total_gain_loss)::text,
			ps.total_gain_loss::text,
			ps.total_gain_loss_pct::text
		FROM portfolio_snapshots ps
		JOIN portfolios p ON p.id = ps.portfolio_id
		WHERE ps.portfolio_id = $1 AND p.user_id = $2
		  AND ($3::boolean = FALSE OR ps.snapshot_date >= $4::date)
		ORDER BY ps.snapshot_date ASC
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
			&point.TotalValue,
			&point.TotalCostBase,
			&point.GainLoss,
			&point.GainLossPct,
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
