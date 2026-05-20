package repositories

import (
	"context"

	"github.com/google/uuid"

	"github.com/yeferson59/finexia-app/internal/entities"
	"github.com/yeferson59/gofinance/money"
)

func (r *Repository) GetPortfoliosByUserID(ctx context.Context, userID uuid.UUID) ([]entities.Portfolio, error) {
	rows, err := r.db.Query(ctx, "SELECT p.*, r.* FROM portfolios p JOIN risks r ON p.risk_id = r.id WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	portfolios := make([]entities.Portfolio, 0)
	for rows.Next() {
		var portfolio entities.Portfolio

		if err := rows.Scan(&portfolio.ID, &portfolio.UserID, &portfolio.Name, &portfolio.Description, &portfolio.Type, &portfolio.RiskID, &portfolio.BaseCurrency, &portfolio.IsDefault, &portfolio.PriceValue, &portfolio.CreatedAt, &portfolio.UpdatedAt, &portfolio.Risk.ID, &portfolio.Risk.Name, &portfolio.Risk.Description, &portfolio.Risk.CreatedAt, &portfolio.Risk.UpdatedAt); err != nil {
			return nil, err
		}

		portfolios = append(portfolios, portfolio)
	}

	return portfolios, nil
}

func (r *Repository) CreatePortfolio(ctx context.Context, userID uuid.UUID, name, description, baseCurrency string, riskID uuid.UUID, typePortfolio entities.PortfolioType, price_value money.Money, isDefault bool) (entities.Portfolio, error) {
	var portfolio entities.Portfolio
	if err := r.db.QueryRow(ctx, "INSERT INTO portfolios(user_id, name, description, base_currency, type, risk_id, price_value, is_default) VALUES($1, $2, $3, $4, $5, $6, $7, $8) RETURNING *", userID, name, description, baseCurrency, typePortfolio, riskID, price_value, isDefault).Scan(&portfolio.ID, &portfolio.UserID, &portfolio.Name, &portfolio.Description, &portfolio.Type, &portfolio.RiskID, &portfolio.BaseCurrency, &portfolio.IsDefault, &portfolio.PriceValue, &portfolio.CreatedAt, &portfolio.UpdatedAt); err != nil {
		return entities.Portfolio{}, err
	}

	return portfolio, nil
}

func (r *Repository) GetPortfoliosRisks(ctx context.Context) ([]entities.Risk, error) {
	rows, err := r.db.Query(ctx, "SELECT * FROM risks")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	risks := make([]entities.Risk, 0)
	for rows.Next() {
		var risk entities.Risk

		if err := rows.Scan(&risk.ID, &risk.Name, &risk.Description, &risk.CreatedAt, &risk.UpdatedAt); err != nil {
			return nil, err
		}

		risks = append(risks, risk)
	}

	return risks, nil
}
