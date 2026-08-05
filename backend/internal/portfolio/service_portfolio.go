package portfolio

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/finance/returns"
	"github.com/yeferson59/gofinance/v2/money"
	"golang.org/x/sync/errgroup"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
)

// oneHundred turns the fractions gofinance's returns package works in into the
// percentages the API reports.
var oneHundred = decimal.MustFromString("100")

// growthUnit is the placeholder currency the growth series is read into. The
// series aggregates every portfolio a user owns, so it has no currency of its
// own; only the ratio between two of its points is reported, and that is the
// same whatever unit both ends share.
const growthUnit = money.USD

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
		converted, err := s.convertSummaryTotals(ctx, userID, summary, targetCurrency)
		if err != nil {
			return nil, err
		}
		summaries[i] = converted
	}

	return summaries, nil
}

func (s *Service) convertSummaryTotals(ctx context.Context, userID uuid.UUID, summary SummaryView, targetCurrency string) (SummaryView, error) {
	// Both ends go through gofinance's ISO 4217 table rather than being carried
	// as bare strings: money.Convert is what does the arithmetic below, and it
	// needs the target currency to know how many minor units to round to.
	from, err := money.CurrencyFromISOCode(summary.BaseCurrency)
	if err != nil {
		return SummaryView{}, httpx.AsBadRequest(fmt.Errorf("unknown portfolio currency %q: %w", summary.BaseCurrency, err))
	}
	to, err := money.CurrencyFromISOCode(targetCurrency)
	if err != nil {
		return SummaryView{}, httpx.AsBadRequest(fmt.Errorf("unknown display currency %q: %w", targetCurrency, err))
	}

	rate, err := s.GetConversionRate(ctx, userID, summary.BaseCurrency, targetCurrency)
	if err != nil {
		return SummaryView{}, err
	}

	// money.Convert re-tags the amount with the target currency and rounds
	// (half to even) to that currency's own precision, so a COP total no longer
	// carries the full width of a USD figure times a four-digit rate.
	convert := func(raw string) (string, error) {
		amount, err := money.NewMoneyFromString(raw, from)
		if err != nil {
			return raw, err
		}
		converted, err := amount.Convert(to, rate)
		if err != nil {
			return raw, err
		}
		return converted.String(), nil
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

// GetPortfolioValuesAsOf returns what each of the user's portfolios was worth
// at the last snapshot on or before asOf. Portfolios with no history that far
// back are absent from the result, and an empty result means the account has
// none at all — a caller comparing against a past value falls back to showing
// no comparison.
func (s *Service) GetPortfolioValuesAsOf(ctx context.Context, userID uuid.UUID, asOf time.Time) ([]PortfolioValuePoint, error) {
	return s.repo.GetPortfolioValuesAsOf(ctx, userID, asOf)
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

	// The growth series is a per-user aggregate with no single currency of its
	// own, and the percentage below is currency-invariant anyway. Both ends are
	// read into the same unit so returns.ROI — which is currency-checked — has
	// a matching pair to work on.
	initial := growthAmount(first.TotalValue)
	current := growthAmount(last.TotalValue)

	// ROI is exactly "profit relative to the amount invested". It rejects a
	// non-positive starting value, which is what keeps a series that begins at
	// zero from dividing by it; the summary then reports 0.00% as before.
	growthPct := decimal.Zero
	if pct, err := returns.ROI(initial, current); err == nil {
		growthPct = pct.Mul(oneHundred)
	}

	return GrowthSummary{
		FirstDate:      first.Date,
		InitialValue:   initial.RoundBank(2).StringFixed(2),
		CurrentValue:   current.RoundBank(2).StringFixed(2),
		TotalGrowthPct: growthPct.RoundBank(2).StringFixed(2),
	}
}

// growthAmount parses one point of the growth series, treating an unparsable
// value as zero the way the previous strconv.ParseFloat call did.
func growthAmount(raw string) money.Money {
	amount, err := money.NewMoneyFromString(raw, growthUnit)
	if err != nil {
		return money.FromDecimal(decimal.Zero, growthUnit)
	}

	return amount
}
