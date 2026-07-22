package portfolio

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/yeferson59/gofinance/v2/money"
)

func (r *PostgresRepository) GetPortfoliosSummaryByUserID(ctx context.Context, userID uuid.UUID) ([]SummaryView, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			p.id,
			p.name,
			COALESCE(p.description, ''),
			p.type,
			p.base_currency,
			p.is_default,
			ri.id,
			ri.name,
			COALESCE(ps.total_positions, 0)::bigint,
			COALESCE(ps.total_cost_base,    0)::text,
			COALESCE(ps.total_market_value, 0)::text,
			COALESCE(ps.total_gain_loss,    0)::text,
			COALESCE(ps.total_gain_loss_pct,0)::text,
			p.created_at
		FROM portfolios p
		JOIN  risks ri          ON ri.id = p.risk_id
		LEFT JOIN portfolio_summary ps ON ps.portfolio_id = p.id
		WHERE p.user_id = $1
		ORDER BY p.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]SummaryView, 0)
	for rows.Next() {
		var item SummaryView
		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Description,
			&item.Type,
			&item.BaseCurrency,
			&item.IsDefault,
			&item.RiskID,
			&item.RiskName,
			&item.TotalPositions,
			&item.TotalCostBase,
			&item.TotalMarketValue,
			&item.TotalGainLoss,
			&item.TotalGainLossPct,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		item.DisplayCurrency = item.BaseCurrency
		result = append(result, item)
	}
	return result, nil
}

func (r *PostgresRepository) GetPortfoliosByUserID(ctx context.Context, userID uuid.UUID) ([]Portfolio, error) {
	rows, err := r.db.Query(ctx, "SELECT p.*, r.* FROM portfolios p JOIN risks r ON p.risk_id = r.id WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	portfolios := make([]Portfolio, 0)
	for rows.Next() {
		var portfolio Portfolio

		if err := rows.Scan(&portfolio.ID, &portfolio.UserID, &portfolio.Name, &portfolio.Description, &portfolio.Type, &portfolio.RiskID, &portfolio.BaseCurrency, &portfolio.IsDefault, &portfolio.PriceValue, &portfolio.CreatedAt, &portfolio.UpdatedAt, &portfolio.Risk.ID, &portfolio.Risk.Name, &portfolio.Risk.Description, &portfolio.Risk.CreatedAt, &portfolio.Risk.UpdatedAt); err != nil {
			return nil, err
		}

		portfolios = append(portfolios, portfolio)
	}

	return portfolios, nil
}

func (r *PostgresRepository) CreatePortfolio(ctx context.Context, userID uuid.UUID, name, description, baseCurrency string, riskID uuid.UUID, typePortfolio Type, priceValue money.Money, isDefault bool) (Portfolio, error) {
	var portfolio Portfolio
	if err := r.db.QueryRow(ctx, "INSERT INTO portfolios(user_id, name, description, base_currency, type, risk_id, price_value, is_default) VALUES($1, $2, $3, $4, $5, $6, $7, $8) RETURNING *", userID, name, description, baseCurrency, typePortfolio, riskID, priceValue, isDefault).Scan(
		&portfolio.ID,
		&portfolio.UserID,
		&portfolio.Name,
		&portfolio.Description,
		&portfolio.Type,
		&portfolio.RiskID,
		&portfolio.BaseCurrency,
		&portfolio.IsDefault,
		&portfolio.PriceValue,
		&portfolio.CreatedAt,
		&portfolio.UpdatedAt,
	); err != nil {
		return Portfolio{}, err
	}

	return portfolio, nil
}

func (r *PostgresRepository) GetPortfolioByID(ctx context.Context, portfolioID, userID uuid.UUID) (Portfolio, error) {
	var portfolio Portfolio
	err := r.db.QueryRow(ctx, `
		SELECT p.id, p.user_id, p.name, COALESCE(p.description, ''), p.type, p.risk_id, p.base_currency, p.is_default, p.price_value, p.created_at, p.updated_at,
		       r.id, r.name, COALESCE(r.description, ''), r.created_at, r.updated_at
		FROM portfolios p
		JOIN risks r ON p.risk_id = r.id
		WHERE p.id = $1 AND p.user_id = $2
	`, portfolioID, userID).Scan(
		&portfolio.ID,
		&portfolio.UserID,
		&portfolio.Name,
		&portfolio.Description,
		&portfolio.Type,
		&portfolio.RiskID,
		&portfolio.BaseCurrency,
		&portfolio.IsDefault,
		&portfolio.PriceValue,
		&portfolio.CreatedAt,
		&portfolio.UpdatedAt,
		&portfolio.Risk.ID,
		&portfolio.Risk.Name,
		&portfolio.Risk.Description,
		&portfolio.Risk.CreatedAt,
		&portfolio.Risk.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Portfolio{}, errors.New("portfolio not found")
		}
		return Portfolio{}, err
	}

	return portfolio, nil
}

func (r *PostgresRepository) UpdatePortfolio(ctx context.Context, userID, portfolioID uuid.UUID, name, description string, portfolioType Type, riskID uuid.UUID, isDefault bool) (Portfolio, error) {
	var portfolio Portfolio
	err := r.db.QueryRow(ctx, `
		UPDATE portfolios
		SET name = $1, description = $2, type = $3, risk_id = $4, is_default = $5, updated_at = NOW()
		WHERE id = $6 AND user_id = $7
		RETURNING id, user_id, name, COALESCE(description, ''), type, risk_id, base_currency, is_default, price_value, created_at, updated_at
	`, name, description, portfolioType, riskID, isDefault, portfolioID, userID).Scan(
		&portfolio.ID,
		&portfolio.UserID,
		&portfolio.Name,
		&portfolio.Description,
		&portfolio.Type,
		&portfolio.RiskID,
		&portfolio.BaseCurrency,
		&portfolio.IsDefault,
		&portfolio.PriceValue,
		&portfolio.CreatedAt,
		&portfolio.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Portfolio{}, errors.New("portfolio not found")
		}
		return Portfolio{}, err
	}

	if err := r.db.QueryRow(ctx, "SELECT id, name, COALESCE(description, ''), created_at, updated_at FROM risks WHERE id = $1", riskID).Scan(
		&portfolio.Risk.ID,
		&portfolio.Risk.Name,
		&portfolio.Risk.Description,
		&portfolio.Risk.CreatedAt,
		&portfolio.Risk.UpdatedAt,
	); err != nil {
		return Portfolio{}, err
	}

	return portfolio, nil
}

func (r *PostgresRepository) GetPortfoliosRisks(ctx context.Context) ([]Risk, error) {
	rows, err := r.db.Query(ctx, "SELECT * FROM risks")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	risks := make([]Risk, 0)
	for rows.Next() {
		var risk Risk

		if err := rows.Scan(&risk.ID, &risk.Name, &risk.Description, &risk.CreatedAt, &risk.UpdatedAt); err != nil {
			return nil, err
		}

		risks = append(risks, risk)
	}

	return risks, nil
}
