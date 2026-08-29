package portfolio

import (
	"context"
	"testing"

	"uuid"

	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/market"
)

// A position carries up to three currencies — the portfolio's base, the one the
// purchase settled in, and the one the asset is quoted in — and only the totals
// converted into the first can be added to anything. These tests cover the
// conversion GetPortfolio performs, including the case that has to degrade
// rather than fail: no rate at all.

// foreignEntry builds the shape GetEntriesByPortfolioID returns for a position
// bought in costCurrency and quoted in the asset's own currency.
func foreignEntry(qty, marketPrice, assetCurrency string, price money.Money) Entry {
	entry := Entry{
		ID:           uuid.New(),
		AssetID:      uuid.New(),
		Quantity:     decimal.MustFromString(qty),
		Price:        price,
		CostCurrency: price.GetCurrency(),
		PriceSource:  PriceSourceOwn,
		Asset: market.Asset{
			Ticker:   "MC.FR",
			Currency: assetCurrency,
		},
	}
	if marketPrice != "" {
		entry.Asset.CurrentPrice = new(moneyOf(marketPrice, assetCurrency))
	}

	return entry
}

func TestValueEntriesInBase(t *testing.T) {
	userID := uuid.New()

	// 2 shares bought at 100 EUR, now worth 110 EUR each, in a USD portfolio at
	// 1.10 USD per EUR: cost 220 USD, value 242 USD.
	t.Run("converts both legs at the stored rate", func(t *testing.T) {
		svc := newTestServices(new(fakeRepository{
			getUserExchangeRateByPair: eurToUSD(t, "1.10"),
		}), newMemStorage())

		got := svc.valueEntriesInBase(context.Background(), userID, "USD",
			[]Entry{foreignEntry("2", "100", "EUR", money.MustMoneyFromString("110", money.EUR))})

		if !got[0].FXConverted {
			t.Fatal("fxConverted = false, want true")
		}
		if cost := got[0].CostBasisBase.String(); cost != "220" {
			t.Errorf("costBasisBase = %s, want 220", cost)
		}
		if value := got[0].MarketValueBase.String(); value != "242" {
			t.Errorf("marketValueBase = %s, want 242", value)
		}
		// Per-unit figures are what the client shows next to their own currency
		// symbol; converting them would round a fractional share to cents.
		if price := got[0].Price.String(); price != "100" {
			t.Errorf("price = %s, want the native 100", price)
		}
	})

	// The MC.FR case: the purchase was recorded in USD while the asset trades in
	// EUR. Each leg has to be converted from its own currency, not from one
	// guessed for the whole position.
	t.Run("cost and market legs use their own currencies", func(t *testing.T) {
		svc := newTestServices(new(fakeRepository{
			getUserExchangeRateByPair: eurToUSD(t, "1.20"),
		}), newMemStorage())

		got := svc.valueEntriesInBase(context.Background(), userID, "USD",
			[]Entry{foreignEntry("3", "50", "USD", money.MustMoneyFromString("40", money.EUR))})

		if !got[0].FXConverted {
			t.Fatal("fxConverted = false, want true")
		}
		// Cost is already in USD: 3 × 50, untouched by the EUR rate.
		if cost := got[0].CostBasisBase.String(); cost != "150" {
			t.Errorf("costBasisBase = %s, want 150 (no conversion)", cost)
		}
		// Value is 3 × 40 EUR converted at 1.20.
		if value := got[0].MarketValueBase.String(); value != "144" {
			t.Errorf("marketValueBase = %s, want 144", value)
		}
	})

	// Rates are BYO-key: a user whose key has not fetched EUR→USD has none, and
	// that must leave the portfolio readable rather than erroring the page out.
	t.Run("missing rate degrades to native amounts", func(t *testing.T) {
		svc := newTestServices(new(fakeRepository{
			getUserExchangeRateByPair: func(context.Context, uuid.UUID, string, string) (decimal.Decimal, error) {
				return decimal.Decimal{}, ErrExchangeRateNotFound
			},
			getExchangeRateByPair: func(context.Context, string, string) (decimal.Decimal, error) {
				return decimal.Decimal{}, ErrExchangeRateNotFound
			},
		}), newMemStorage())

		got := svc.valueEntriesInBase(context.Background(), userID, "USD",
			[]Entry{foreignEntry("2", "100", "EUR", money.MustMoneyFromString("110", money.EUR))})

		if got[0].FXConverted {
			t.Error("fxConverted = true, want false when no rate connects the pair")
		}
		// The amounts are still reported, in EUR, so the client can show them
		// unconverted and say so.
		if cost := got[0].CostBasisBase.String(); cost != "200" {
			t.Errorf("costBasisBase = %s, want the native 200", cost)
		}
		if cur := got[0].MarketValueBase.GetCurrency(); cur != money.EUR {
			t.Errorf("marketValueBase currency = %s, want EUR", cur)
		}
	})

	// A position already in the base currency must not consult the rate store at
	// all: the fake panics on a nil hook, which is what asserts it here.
	t.Run("base currency needs no rate lookup", func(t *testing.T) {
		svc := newTestServices(new(fakeRepository{}), newMemStorage())

		got := svc.valueEntriesInBase(context.Background(), userID, "USD",
			[]Entry{foreignEntry("2", "100", "USD", money.MustMoneyFromString("110", money.USD))})

		if !got[0].FXConverted {
			t.Error("fxConverted = false, want true for a position already in base")
		}
		if cost, value := got[0].CostBasisBase.String(), got[0].MarketValueBase.String(); cost != "200" || value != "220" {
			t.Errorf("totals = (%s, %s), want (200, 220)", cost, value)
		}
	})

	// Without a market price the position is carried at cost — the same
	// convention priceSource "cost" reports — and the two totals agree.
	t.Run("unpriced position is valued at cost", func(t *testing.T) {
		svc := newTestServices(new(fakeRepository{
			getUserExchangeRateByPair: eurToUSD(t, "1.10"),
		}), newMemStorage())

		got := svc.valueEntriesInBase(context.Background(), userID, "USD",
			[]Entry{foreignEntry("2", "100", "EUR", money.MustMoneyFromString("", money.EUR))})

		if cost, value := got[0].CostBasisBase.String(), got[0].MarketValueBase.String(); cost != value || cost != "220" {
			t.Errorf("totals = (%s, %s), want both 220", cost, value)
		}
	})
}

// The conversion is what makes a multi-currency portfolio addable, so it has to
// survive the trip through GetPortfolio and into the response DTO.
func TestGetPortfolioConvertsHoldingsToBaseCurrency(t *testing.T) {
	userID, portfolioID := uuid.New(), uuid.New()

	repo := new(fakeRepository{
		getPortfolioByID: func(context.Context, uuid.UUID, uuid.UUID) (Portfolio, error) {
			return Portfolio{ID: portfolioID, BaseCurrency: "USD"}, nil
		},
		getEntriesByPortfolioID: func(context.Context, uuid.UUID) ([]Entry, error) {
			return []Entry{foreignEntry("2", "100", "EUR", money.MustMoneyFromString("110", money.EUR))}, nil
		},
		getUserExchangeRateByPair: eurToUSD(t, "1.10"),
	})

	got, err := newTestServices(repo, newMemStorage()).
		GetPortfolio(context.Background(), userID, portfolioID)
	if err != nil {
		t.Fatalf("GetPortfolio: %v", err)
	}

	holding := NewPortfolioDetailResponse(got).Holdings[0]
	if holding.CostBasisBase != "220" || holding.MarketValueBase != "242" {
		t.Errorf("holding totals = (%s, %s), want (220, 242)", holding.CostBasisBase, holding.MarketValueBase)
	}
	if !holding.FXConverted {
		t.Error("fxConverted = false, want true")
	}
	// The native price stays as it was: it is quoted in EUR and the client
	// formats it as such.
	if holding.MarketPrice != "110" || holding.Currency != "EUR" {
		t.Errorf("native price = %s %s, want 110 EUR", holding.MarketPrice, holding.Currency)
	}
}

// eurToUSD serves one rate and rejects every other pair, so a test that
// converts the wrong direction fails instead of silently passing.
func eurToUSD(t *testing.T, value string) func(context.Context, uuid.UUID, string, string) (decimal.Decimal, error) {
	t.Helper()

	return func(_ context.Context, _ uuid.UUID, from, to string) (decimal.Decimal, error) {
		if from == "EUR" && to == "USD" {
			return rate(t, value), nil
		}

		return decimal.Decimal{}, ErrExchangeRateNotFound
	}
}
