package portfolio

import (
	"context"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"
	"golang.org/x/sync/errgroup"
)

func (s *Service) GetPortfoliosRisks(ctx context.Context) ([]Risk, error) {
	if c := s.risksCache; c != nil {
		c.mu.RLock()
		risks, fresh := c.risks, time.Now().Before(c.expiresAt)
		c.mu.RUnlock()
		if fresh {
			return risks, nil
		}
	}

	risks, err := s.repo.GetPortfoliosRisks(ctx)
	if err != nil {
		return []Risk{}, err
	}

	if c := s.risksCache; c != nil {
		c.mu.Lock()
		c.risks, c.expiresAt = risks, time.Now().Add(risksCacheTTL)
		c.mu.Unlock()
	}

	return risks, nil
}

func (s *Service) GetPortfolios(ctx context.Context, userID uuid.UUID) ([]Portfolio, error) {
	portfolios, err := s.repo.GetPortfoliosByUserID(ctx, userID)
	if err != nil {
		return []Portfolio{}, err
	}

	return portfolios, nil
}

func (s *Service) GetPortfoliosSummary(ctx context.Context, userID uuid.UUID) ([]SummaryView, error) {
	return s.repo.GetPortfoliosSummaryByUserID(ctx, userID)
}

// GetPortfoliosSummaryInCurrency behaves like GetPortfoliosSummary but
// converts each portfolio's totals from its own base currency into
// targetCurrency, so a user with portfolios in different currencies gets a
// single, comparable display currency.
func (s *Service) GetPortfoliosSummaryInCurrency(ctx context.Context, userID uuid.UUID, targetCurrency string) ([]SummaryView, error) {
	summaries, err := s.repo.GetPortfoliosSummaryByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	for i, summary := range summaries {
		converted, err := s.convertSummaryTotals(ctx, summary, targetCurrency)
		if err != nil {
			return nil, err
		}
		summaries[i] = converted
	}

	return summaries, nil
}

func (s *Service) convertSummaryTotals(ctx context.Context, summary SummaryView, targetCurrency string) (SummaryView, error) {
	rate, err := s.GetConversionRate(ctx, summary.BaseCurrency, targetCurrency)
	if err != nil {
		return SummaryView{}, err
	}

	convert := func(raw string) (string, error) {
		amount, err := decimal.NewFromString(raw)
		if err != nil {
			return raw, err
		}
		return amount.Mul(rate).String(), nil
	}

	var convErr error
	if summary.TotalCostBase, convErr = convert(summary.TotalCostBase); convErr != nil {
		return SummaryView{}, convErr
	}
	if summary.TotalMarketValue, convErr = convert(summary.TotalMarketValue); convErr != nil {
		return SummaryView{}, convErr
	}
	if summary.TotalGainLoss, convErr = convert(summary.TotalGainLoss); convErr != nil {
		return SummaryView{}, convErr
	}
	// TotalGainLossPct is a ratio, not a money amount — currency-invariant.

	summary.DisplayCurrency = targetCurrency
	return summary, nil
}

func (s *Service) GetPortfolio(ctx context.Context, userID, portfolioID uuid.UUID) (Portfolio, error) {
	// The portfolio header and its entries are independent queries; running
	// them concurrently halves the latency of the portfolio detail endpoint.
	// The ownership check in GetPortfolioByID still gates the response: if it
	// fails, the fetched entries are discarded with the error.
	var (
		portfolio Portfolio
		entries   []Entry
	)

	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		var err error
		portfolio, err = s.repo.GetPortfolioByID(gctx, portfolioID, userID)
		return err
	})
	g.Go(func() error {
		var err error
		entries, err = s.repo.GetEntriesByPortfolioID(gctx, portfolioID)
		return err
	})
	if err := g.Wait(); err != nil {
		return Portfolio{}, err
	}

	portfolio.Entries = entries

	return portfolio, nil
}

func (s *Service) CreatePortfolio(ctx context.Context, userID uuid.UUID, name string, description string, baseCurrency string, riskID uuid.UUID, typePortfolio Type, priceValue money.Money, isDefault bool) (Portfolio, error) {
	portfolio, err := s.repo.CreatePortfolio(ctx, userID, name, description, baseCurrency, riskID, typePortfolio, priceValue, isDefault)
	if err != nil {
		return Portfolio{}, err
	}

	return portfolio, nil
}

func (s *Service) GetPortfolioTopTransaction(ctx context.Context, userID, portfolioID uuid.UUID) (TopTransactionDTO, error) {
	return s.repo.GetTopTransactionByPortfolioID(ctx, userID, portfolioID)
}

func (s *Service) UpdatePortfolio(ctx context.Context, userID, portfolioID uuid.UUID, name, description string, portfolioType Type, riskID uuid.UUID, isDefault bool) (Portfolio, error) {
	return s.repo.UpdatePortfolio(ctx, userID, portfolioID, name, description, portfolioType, riskID, isDefault)
}

func (s *Service) GetPortfolioGrowth(ctx context.Context, userID uuid.UUID, period string) ([]GrowthPoint, GrowthSummary, error) {
	hasSince, since := parsePeriod(period)
	points, err := s.repo.GetPortfolioGrowthByUserID(ctx, userID, hasSince, since)
	if err != nil {
		return nil, GrowthSummary{}, err
	}
	return points, buildGrowthSummary(points), nil
}

func (s *Service) GetPortfolioGrowthByID(ctx context.Context, userID, portfolioID uuid.UUID, period string) ([]GrowthPoint, GrowthSummary, error) {
	hasSince, since := parsePeriod(period)
	points, err := s.repo.GetPortfolioGrowthByPortfolioID(ctx, userID, portfolioID, hasSince, since)
	if err != nil {
		return nil, GrowthSummary{}, err
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

func buildGrowthSummary(points []GrowthPoint) GrowthSummary {
	if len(points) == 0 {
		return GrowthSummary{}
	}
	first, last := points[0], points[len(points)-1]
	initialVal, _ := strconv.ParseFloat(first.TotalValue, 64)
	currentVal, _ := strconv.ParseFloat(last.TotalValue, 64)
	var growthPct float64
	if initialVal > 0 {
		growthPct = ((currentVal - initialVal) / initialVal) * 100
	}
	return GrowthSummary{
		FirstDate:      first.Date,
		InitialValue:   strconv.FormatFloat(initialVal, 'f', 2, 64),
		CurrentValue:   strconv.FormatFloat(currentVal, 'f', 2, 64),
		TotalGrowthPct: strconv.FormatFloat(growthPct, 'f', 2, 64),
	}
}
