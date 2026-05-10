package repositories

import (
	"context"

	"github.com/google/uuid"

	"github.com/yeferson59/finexia-app/internal/entities"
)

func (r *Repository) GetPortfoliosByUserID(ctx context.Context, userID uuid.UUID) ([]entities.Portfolio, error) {
	rows, err := r.db.Query(ctx, "SELECT * FROM portfolio WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	portfolios := make([]entities.Portfolio, 0)
	for rows.Next() {
		var portfolio entities.Portfolio
		if err := rows.Scan(&portfolio.ID, &portfolio.UserID, &portfolio.Name, &portfolio.Description, &portfolio.BaseCurrency, &portfolio.IsDefault, &portfolio.CreatedAt, &portfolio.UpdatedAt); err != nil {
			return nil, err
		}
		portfolios = append(portfolios, portfolio)
	}

	return portfolios, nil
}

func (r *Repository) CreatePortfolio(ctx context.Context, userID uuid.UUID, name, description, baseCurrency string, isDefault bool) (entities.Portfolio, error) {
	var portfolio entities.Portfolio
	if err := r.db.QueryRow(ctx, "INSERT INTO portfolio(user_id, name, description, base_currency, is_default) VALUES($1, $2, $3, $4, $5) RETURNING *", userID, name, description, baseCurrency, isDefault).Scan(&portfolio.ID, &portfolio.UserID, &portfolio.Name, &portfolio.Description, &portfolio.BaseCurrency, &portfolio.IsDefault, &portfolio.CreatedAt, &portfolio.UpdatedAt); err != nil {
		return entities.Portfolio{}, err
	}

	return portfolio, nil
}
