package portfolio

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/market"
)

// Under BYO-key a valuation can rest on this user's own price, on the
// operator's manual reference price, or on nothing at all — in which case the
// position is carried at cost. The three produce the same shape of number and
// mean different things, so the API has to say which one it served. These
// tests cover the two places that carry that answer out: the holdings of a
// portfolio detail, and the per-portfolio counts in the summary.

// pricedEntry builds an entry whose asset carries a resolved market price, as
// GetEntriesByPortfolioID returns it.
func pricedEntry(ticker string, source PriceSource, price string, fetchedAt *time.Time) Entry {
	entry := Entry{
		ID:           uuid.New(),
		AssetID:      uuid.New(),
		Quantity:     decimal.MustFromString("2"),
		Price:        money.MustMoneyFromString("100", money.USD),
		CostCurrency: "USD",
		PriceSource:  source,
		Asset: market.Asset{
			Ticker:         ticker,
			Currency:       "USD",
			PriceUpdatedAt: fetchedAt,
		},
	}
	if price != "" {
		entry.Asset.CurrentPrice = new(money.MustMoneyFromString(price, money.USD))
	}

	return entry
}

func TestPortfolioDetailReportsPriceProvenance(t *testing.T) {
	fetched := time.Date(2026, 7, 26, 9, 30, 0, 0, time.UTC)

	p := Portfolio{
		ID:           uuid.New(),
		Name:         "Main",
		BaseCurrency: "USD",
		Entries: []Entry{
			pricedEntry("AAPL", PriceSourceOwn, "190.55", new(fetched)),
			pricedEntry("MSFT", PriceSourceManual, "410.00", new(fetched)),
			pricedEntry("XYZ", PriceSourceCost, "", nil),
		},
	}

	got := NewPortfolioDetailResponse(p)
	if len(got.Holdings) != 3 {
		t.Fatalf("holdings = %d, want 3", len(got.Holdings))
	}

	byTicker := map[string]HoldingResponseDTO{}
	for _, h := range got.Holdings {
		byTicker[h.Ticker] = h
	}

	for ticker, want := range map[string]string{
		"AAPL": "own",
		"MSFT": "manual",
		"XYZ":  "cost",
	} {
		if got := byTicker[ticker].PriceSource; got != want {
			t.Errorf("%s priceSource = %q, want %q", ticker, got, want)
		}
	}

	// A priced holding carries the age of its price: a BYO-key sync is capped
	// per request and can trail by hours, so "190.55" alone is not enough to
	// act on.
	if h := byTicker["AAPL"]; h.MarketPrice != "190.55" || h.PriceUpdatedAt == nil || !h.PriceUpdatedAt.Equal(fetched) {
		t.Errorf("AAPL = {%q, %v}, want the price and its fetch time", h.MarketPrice, h.PriceUpdatedAt)
	}

	// The at-cost holding must not look like a market price of zero: the field
	// stays empty and the timestamp null, so a client cannot render it as one.
	if h := byTicker["XYZ"]; h.MarketPrice != "" || h.PriceUpdatedAt != nil {
		t.Errorf("XYZ = {%q, %v}, want no market price at all", h.MarketPrice, h.PriceUpdatedAt)
	}
}

func TestSummaryCountsSurviveCurrencyConversion(t *testing.T) {
	userID := uuid.New()

	repo := new(fakeRepository{
		getPortfoliosSummaryByUserID: func(context.Context, uuid.UUID) ([]SummaryView, error) {
			return []SummaryView{{
				ID:                    uuid.New(),
				BaseCurrency:          "USD",
				TotalPositions:        10,
				TotalCostBase:         "1000",
				TotalMarketValue:      "1200",
				TotalGainLoss:         "200",
				TotalGainLossPct:      "20",
				PositionsPricedOwn:    6,
				PositionsPricedManual: 3,
				PositionsAtCost:       1,
			}}, nil
		},
		getUserExchangeRateByPair: func(_ context.Context, _ uuid.UUID, from, to string) (money.Decimal, error) {
			if from == "USD" && to == "COP" {
				return rate(t, "4000"), nil
			}

			return money.Decimal{}, ErrExchangeRateNotFound
		},
	})

	got, err := newTestServices(repo, newMemStorage()).
		GetPortfoliosSummaryInCurrency(context.Background(), userID, "COP")
	if err != nil {
		t.Fatalf("GetPortfoliosSummaryInCurrency: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("summaries = %d, want 1", len(got))
	}

	summary := got[0]
	if summary.TotalMarketValue != "4800000" {
		t.Errorf("totalMarketValue = %s, want 4800000 (converted)", summary.TotalMarketValue)
	}

	// The counts are positions, not amounts. Multiplying them by the rate the
	// way the totals are is the mistake this guards against.
	if summary.PositionsPricedOwn != 6 || summary.PositionsPricedManual != 3 || summary.PositionsAtCost != 1 {
		t.Errorf("counts = (%d, %d, %d), want (6, 3, 1) unchanged by conversion",
			summary.PositionsPricedOwn, summary.PositionsPricedManual, summary.PositionsAtCost)
	}
	// They partition the positions, so a client can trust the arithmetic.
	if sum := summary.PositionsPricedOwn + summary.PositionsPricedManual + summary.PositionsAtCost; sum != summary.TotalPositions {
		t.Errorf("counts sum to %d, want totalPositions = %d", sum, summary.TotalPositions)
	}
}
