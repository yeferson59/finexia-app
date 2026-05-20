package services

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/yeferson59/finexia-app/internal/entities"
	"github.com/yeferson59/gofinance/money"
)

func (s *Services) GetPortfoliosRisks(ctx context.Context) ([]entities.Risk, error) {
	risks, err := s.repos.GetPortfoliosRisks(ctx)
	if err != nil {
		return []entities.Risk{}, err
	}

	return risks, nil
}

func (s *Services) GetPortfolios(ctx context.Context, userID uuid.UUID) ([]entities.Portfolio, error) {
	portfolios, err := s.repos.GetPortfoliosByUserID(ctx, userID)
	if err != nil {
		return []entities.Portfolio{}, err
	}

	return portfolios, nil
}

func (s *Services) CreatePortfolio(ctx context.Context, userID uuid.UUID, name string, description string, baseCurrency string, riskID uuid.UUID, typePortfolio entities.PortfolioType, price_value money.Money, isDefault bool) (entities.Portfolio, error) {
	portfolio, err := s.repos.CreatePortfolio(ctx, userID, name, description, baseCurrency, riskID, typePortfolio, price_value, isDefault)
	if err != nil {
		return entities.Portfolio{}, err
	}

	return portfolio, nil
}

func (s *Services) CreatePlatform(ctx context.Context, userID uuid.UUID, sourceType entities.SourceType, name, description string) (entities.InvestmentSource, error) {
	platform, err := s.repos.CreatePlatform(ctx, userID, sourceType, strings.ToLower(name), description)
	if err != nil {
		return entities.InvestmentSource{}, err
	}

	return platform, nil
}
