package portfolio

import (
	"context"
	"fmt"
	"time"

	"uuid"

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

func (s *service) GetPortfoliosRisks(ctx context.Context) ([]Risk, error) {
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

func (s *service) GetPortfolios(ctx context.Context, userID uuid.UUID) ([]Portfolio, error) {
	portfolios, err := s.repo.GetPortfoliosByUserID(ctx, userID)
	if err != nil {
		return []Portfolio{}, err
	}

	return portfolios, nil
}

func (s *service) GetPortfoliosSummary(ctx context.Context, userID uuid.UUID) ([]SummaryView, error) {
	return s.repo.GetPortfoliosSummaryByUserID(ctx, userID)
}

// GetPortfoliosSummaryInCurrency behaves like GetPortfoliosSummary but
// converts each portfolio's totals from its own base currency into
// targetCurrency, so a user with portfolios in different currencies gets a
// single, comparable display currency.
func (s *service) GetPortfoliosSummaryInCurrency(ctx context.Context, userID uuid.UUID, targetCurrency money.Currency) ([]SummaryView, error) {
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

// convertSummaryTotals restates one portfolio's totals in targetCurrency.
//
// A portfolio it cannot convert is left in its own currency and returned with
// FXConverted false, rather than failing the request. The alternative — one
// unreachable pair turning the whole list into a 404 — is the wrong trade for a
// dashboard that now always asks for a display currency: a portfolio in a
// currency no source quotes would blank out every other portfolio with it. The
// same call is what holdings already make (see valueEntriesInBase).
func (s *service) convertSummaryTotals(ctx context.Context, userID uuid.UUID, summary SummaryView, targetCurrency money.Currency) (SummaryView, error) {
	// money.Currency is an integer type, so a code outside gofinance's ISO 4217
	// table travels this far without complaint and only fails inside
	// money.Convert, which needs the target's minor units to round to. Both ends
	// are screened here, because the two failures are not the same kind: a base
	// currency comes from a stored row and a target comes from the caller.
	if !summary.BaseCurrency.Valid() {
		return unconvertedSummary(summary), nil
	}
	if !targetCurrency.Valid() {
		// The target is the caller's own input, validated at the handler; an
		// unknown one here is a bad request, not a portfolio the app can't price.
		return SummaryView{}, httpx.AsBadRequest(fmt.Errorf("unknown display currency %q", targetCurrency))
	}

	rate, err := s.GetConversionRate(ctx, userID, summary.BaseCurrency, targetCurrency)
	if err != nil {
		return unconvertedSummary(summary), nil
	}

	// money.Convert re-tags the amount with the target currency and rounds
	// (half to even) to that currency's own precision, so a COP total no longer
	// carries the full width of a USD figure times a four-digit rate.
	convert := func(raw string) (string, error) {
		amount, err := money.NewMoneyFromString(raw, summary.BaseCurrency)
		if err != nil {
			return raw, err
		}
		converted, err := amount.Convert(targetCurrency, rate)
		if err != nil {
			return raw, err
		}

		return converted.String(), nil
	}

	// All three are converted before any is stored back, so a failure halfway
	// cannot leave a summary whose cost is in one currency and whose market
	// value is in another — the one shape no client could describe correctly.
	cost, costErr := convert(summary.TotalCostBase)
	value, valueErr := convert(summary.TotalMarketValue)
	gain, gainErr := convert(summary.TotalGainLoss)
	if costErr != nil || valueErr != nil || gainErr != nil {
		return unconvertedSummary(summary), nil
	}

	summary.TotalCostBase = cost
	summary.TotalMarketValue = value
	summary.TotalGainLoss = gain
	// TotalGainLossPct is a ratio, not a money amount — currency-invariant.

	summary.DisplayCurrency = targetCurrency
	summary.FXConverted = true
	return summary, nil
}

// unconvertedSummary hands back the totals as they were stored, in the
// portfolio's own currency, flagged so the client shows them for what they are
// instead of mixing them into a total in another currency.
func unconvertedSummary(summary SummaryView) SummaryView {
	summary.DisplayCurrency = summary.BaseCurrency
	summary.FXConverted = false
	return summary
}

func (s *service) GetPortfolio(ctx context.Context, userID, portfolioID uuid.UUID) (Portfolio, error) {
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

	// Positions arrive in their own currencies; the caller needs totals it can
	// add up, so they are valued in the portfolio's base currency before the
	// response is built.
	portfolio.Entries = s.valueEntriesInBase(ctx, userID, portfolio.BaseCurrency, entries)

	return portfolio, nil
}

func (s *service) CreatePortfolio(ctx context.Context, userID uuid.UUID, name, description string, baseCurrency money.Currency, riskID uuid.UUID, typePortfolio Type, priceValue money.Money, isDefault bool) (Portfolio, error) {
	portfolio, err := s.repo.CreatePortfolio(ctx, userID, name, description, baseCurrency, riskID, typePortfolio, priceValue, isDefault)
	if err != nil {
		return Portfolio{}, err
	}

	return portfolio, nil
}

func (s *service) GetPortfolioTopTransaction(ctx context.Context, userID, portfolioID uuid.UUID) (TopTransactionDTO, error) {
	return s.repo.GetTopTransactionByPortfolioID(ctx, userID, portfolioID)
}

func (s *service) UpdatePortfolio(ctx context.Context, userID, portfolioID uuid.UUID, name, description string, portfolioType Type, riskID uuid.UUID, isDefault bool) (Portfolio, error) {
	return s.repo.UpdatePortfolio(ctx, userID, portfolioID, name, description, portfolioType, riskID, isDefault)
}

// GetTrailingReturns is the account's return over the last day, week, month…
// in currency (money.XXX for the preferred one), as GetPortfolioGrowth reports
// it next to the series. It is for callers that want the figures and not the
// chart, such as the weekly digest.
func (s *service) GetTrailingReturns(ctx context.Context, userID uuid.UUID, currency money.Currency) ([]TrailingReturn, error) {
	_, summary, err := s.GetPortfolioGrowth(ctx, userID, currency, "ALL")
	if err != nil {
		return nil, err
	}

	return summary.Trailing, nil
}

// GetPortfolioTrailingReturns is GetTrailingReturns for one portfolio, in its
// base currency.
func (s *service) GetPortfolioTrailingReturns(ctx context.Context, userID, portfolioID uuid.UUID) ([]TrailingReturn, error) {
	_, summary, err := s.GetPortfolioGrowthByID(ctx, userID, portfolioID, "ALL")
	if err != nil {
		return nil, err
	}

	return summary.Trailing, nil
}

// GetPortfolioGrowth builds the account-wide series. An empty currency means
// "the account's preferred one", the same default the summary endpoints use.
//
// The series closes on today read live, not on the last snapshot, so its
// closing figures are the ones the summary endpoint reports next to it.
func (s *service) GetPortfolioGrowth(ctx context.Context, userID uuid.UUID, currency money.Currency, period string) ([]GrowthPoint, GrowthSummary, error) {
	history, err := s.repo.GetPortfolioGrowthByUserID(ctx, userID, currency, false, time.Time{}, snapshotDay(time.Now()))
	if err != nil {
		return nil, GrowthSummary{}, err
	}

	points, summary := growthWindow(history, period)

	return points, summary, nil
}

func (s *service) GetPortfolioGrowthByID(ctx context.Context, userID, portfolioID uuid.UUID, period string) ([]GrowthPoint, GrowthSummary, error) {
	history, err := s.repo.GetPortfolioGrowthByPortfolioID(ctx, userID, portfolioID, false, time.Time{}, snapshotDay(time.Now()))
	if err != nil {
		return nil, GrowthSummary{}, err
	}

	points, summary := growthWindow(history, period)

	return points, summary, nil
}

// growthWindow cuts the whole history down to the asked-for period and
// summarizes it, except for the trailing returns, which always come from the
// whole history: a one-month series has no anchor for the year, and the
// dashboard's "1 year" must not depend on how much of the chart was asked for.
//
// The cut is the one the query used to make (snapshot_date >= since), moved
// here so the history is read once. It gives the same points: a point's
// netFlow is attributed by landing date, not by window.
func growthWindow(history []GrowthPoint, period string) ([]GrowthPoint, GrowthSummary) {
	points := history
	if hasSince, since := parsePeriod(period); hasSince {
		points = pointsSince(history, since)
	}

	summary := buildGrowthSummary(points)
	summary.Trailing = BuildTrailingReturns(history)

	return points, summary
}

// pointsSince drops the points dated before since's UTC day. The series is in
// date order, so the kept points are a suffix of it.
func pointsSince(points []GrowthPoint, since time.Time) []GrowthPoint {
	cut := since.UTC().Truncate(24 * time.Hour)

	for i, p := range points {
		if !p.Date.Before(cut) {
			return points[i:]
		}
	}

	return []GrowthPoint{}
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
		// The profit of the latest point, which is a different quantity from the
		// growth above and has to travel next to it: TotalGrowthPct compares the
		// value of two dates, so opening a portfolio or adding a position counts
		// as growth. Only market - invested is a return, and it is what the chart
		// already shows for the point under the cursor.
		GainLoss:    last.GainLoss,
		GainLossPct: last.GainLossPct,
		Currency:    last.Currency,
	}
}

// growthAmount parses one point of the growth series, treating an unparsable
// value as zero the way the previous strconv.ParseFloat call did.
func growthAmount(raw string) money.Money {
	amount, err := money.NewMoneyFromString(raw, growthUnit)
	if err != nil {
		return money.NewFromDecimal(decimal.Zero, growthUnit)
	}

	return amount
}
