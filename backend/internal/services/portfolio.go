package services

import (
	"context"
	"strings"
	"time"

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

func (s *Services) GetPortfolio(ctx context.Context, userID, portfolioID uuid.UUID) (entities.Portfolio, error) {
	portfolio, err := s.repos.GetPortfolioByID(ctx, portfolioID, userID)
	if err != nil {
		return entities.Portfolio{}, err
	}

	entries, err := s.repos.GetEntriesByPortfolioID(ctx, portfolioID)
	if err != nil {
		return entities.Portfolio{}, err
	}

	portfolio.PortfolioEntries = entries

	return portfolio, nil
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

func (s *Services) GetPlatforms(ctx context.Context, userID uuid.UUID) ([]entities.InvestmentSource, error) {
	return s.repos.GetPlatforms(ctx, userID)
}

func (s *Services) GetAssets(ctx context.Context, offset, limit uint) ([]entities.Asset, error) {
	return s.repos.GetAssets(ctx, offset, limit)
}

func (s *Services) CreatePortfolioEntry(ctx context.Context, userID, portfolioID, assetID uuid.UUID, sourceID uuid.UUID, quantity money.Decimal, price money.Money, costCurrency string, category entities.PortfolioEntryCategory, entryDate time.Time, notes string) (entities.PortfolioEntry, error) {
	entry, err := s.repos.CreatePortfolioEntry(ctx, userID, portfolioID, assetID, sourceID, quantity, price, costCurrency, category, entryDate, notes)
	if err != nil {
		return entities.PortfolioEntry{}, err
	}

	return entry, nil
}
