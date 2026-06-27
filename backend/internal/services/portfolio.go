package services

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/yeferson59/gofinance/money"

	portfoliodto "github.com/yeferson59/finexia-app/internal/dtos/portfolio"
	"github.com/yeferson59/finexia-app/internal/entities"
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

func (s *Services) GetPortfoliosSummary(ctx context.Context, userID uuid.UUID) ([]entities.PortfolioSummaryView, error) {
	return s.repos.GetPortfoliosSummaryByUserID(ctx, userID)
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

func (s *Services) GetPortfolioTopTransaction(ctx context.Context, userID, portfolioID uuid.UUID) (portfoliodto.PortfolioTopTransactionDTO, error) {
	return s.repos.GetTopTransactionByPortfolioID(ctx, userID, portfolioID)
}

func (s *Services) UpdatePortfolio(ctx context.Context, userID, portfolioID uuid.UUID, name, description string, portfolioType entities.PortfolioType, riskID uuid.UUID, isDefault bool) (entities.Portfolio, error) {
	return s.repos.UpdatePortfolio(ctx, userID, portfolioID, name, description, portfolioType, riskID, isDefault)
}

func (s *Services) CreatePlatform(ctx context.Context, userID uuid.UUID, sourceType entities.SourceType, name, description string) (entities.InvestmentSource, error) {
	platform, err := s.repos.CreatePlatform(ctx, userID, sourceType, strings.ToLower(name), description)
	if err != nil {
		return entities.InvestmentSource{}, err
	}

	return platform, nil
}

func (s *Services) GetPlatforms(ctx context.Context, userID uuid.UUID) ([]entities.PlatformStats, error) {
	return s.repos.GetPlatformsWithStats(ctx, userID)
}

func (s *Services) UpdatePlatform(ctx context.Context, userID, sourceID uuid.UUID, name, description string, sourceType entities.SourceType, isActive bool) (entities.PlatformStats, error) {
	return s.repos.UpdatePlatform(ctx, userID, sourceID, name, description, sourceType, isActive)
}

func (s *Services) DeletePlatform(ctx context.Context, userID, sourceID uuid.UUID) error {
	return s.repos.DeletePlatform(ctx, userID, sourceID)
}

func (s *Services) GetAssets(ctx context.Context, offset, limit uint) ([]entities.Asset, error) {
	return s.repos.GetAssets(ctx, offset, limit)
}

func (s *Services) UpdateAssetPrice(ctx context.Context, assetID uuid.UUID, price money.Money) (entities.Asset, error) {
	return s.repos.UpdateAssetPrice(ctx, assetID, price)
}

func (s *Services) CreatePortfolioEntry(ctx context.Context, userID, portfolioID, assetID uuid.UUID, sourceID uuid.UUID, txnType entities.TransactionType, quantity money.Decimal, price money.Money, costCurrency string, category entities.PortfolioEntryCategory, entryDate time.Time, notes string) (entities.PortfolioEntry, error) {
	entry, err := s.repos.CreatePortfolioEntry(ctx, userID, portfolioID, assetID, sourceID, txnType, quantity, price, costCurrency, category, entryDate, notes)
	if err != nil {
		return entities.PortfolioEntry{}, err
	}

	return entry, nil
}

func (s *Services) GetTransactionsByEntry(ctx context.Context, userID, entryID uuid.UUID) ([]entities.Transaction, error) {
	return s.repos.GetTransactionsByEntryID(ctx, userID, entryID)
}

func (s *Services) GetRecentUserTransactions(ctx context.Context, userID uuid.UUID, limit int) ([]entities.Transaction, error) {
	return s.repos.GetRecentTransactionsByUserID(ctx, userID, limit)
}

func (s *Services) GetAssetAllocation(ctx context.Context, userID uuid.UUID) ([]entities.AllocationItem, error) {
	return s.repos.GetAssetAllocationByUserID(ctx, userID)
}

func (s *Services) CreateTransaction(ctx context.Context, userID, entryID uuid.UUID, txnType entities.TransactionType, quantity money.Decimal, price money.Money, currency string, fees money.Money, transactionDate time.Time, notes string) (entities.Transaction, error) {
	return s.repos.CreateTransaction(ctx, userID, entryID, txnType, quantity, price, currency, fees, transactionDate, notes)
}

func (s *Services) GetPortfolioGrowth(ctx context.Context, userID uuid.UUID, period string) ([]entities.PortfolioGrowthPoint, entities.PortfolioGrowthSummary, error) {
	hasSince, since := parsePeriod(period)
	points, err := s.repos.GetPortfolioGrowthByUserID(ctx, userID, hasSince, since)
	if err != nil {
		return nil, entities.PortfolioGrowthSummary{}, err
	}
	return points, buildGrowthSummary(points), nil
}

func parsePeriod(period string) (bool, time.Time) {
	now := time.Now().UTC()
	switch period {
	case "1M":
		return true, now.AddDate(0, -1, 0)
	case "3M":
		return true, now.AddDate(0, -3, 0)
	case "6M":
		return true, now.AddDate(0, -6, 0)
	case "1Y":
		return true, now.AddDate(-1, 0, 0)
	default:
		return false, time.Time{}
	}
}

func buildGrowthSummary(points []entities.PortfolioGrowthPoint) entities.PortfolioGrowthSummary {
	if len(points) == 0 {
		return entities.PortfolioGrowthSummary{}
	}
	first, last := points[0], points[len(points)-1]
	initialVal, _ := strconv.ParseFloat(first.TotalValue, 64)
	currentVal, _ := strconv.ParseFloat(last.TotalValue, 64)
	var growthPct float64
	if initialVal > 0 {
		growthPct = ((currentVal - initialVal) / initialVal) * 100
	}
	return entities.PortfolioGrowthSummary{
		FirstDate:      first.Date,
		InitialValue:   strconv.FormatFloat(initialVal, 'f', 2, 64),
		CurrentValue:   strconv.FormatFloat(currentVal, 'f', 2, 64),
		TotalGrowthPct: strconv.FormatFloat(growthPct, 'f', 2, 64),
	}
}
