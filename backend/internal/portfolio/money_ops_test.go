package portfolio

import (
	"context"
	"errors"
	"testing"

	"uuid"

	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/market"
)

func TestMoneyOfCarriesTheRowCurrency(t *testing.T) {
	m, _ := money.NewMoneyFromString("1234.50", money.COP)
	if m.GetCurrency() != money.COP {
		t.Errorf("currency = %v, want COP", m.GetCurrency())
	}
	if m.String() != "1234.5" {
		t.Errorf("value = %q, want 1234.5", m.String())
	}
}

func TestRetagCurrency(t *testing.T) {
	t.Run("replaces the currency Scan left behind and keeps the value", func(t *testing.T) {
		// money.Money.Scan only decodes the numeric value, leaving the zero
		// currency, which serializes as "XXX".
		var m money.Money
		if err := m.Scan("2500.75"); err != nil {
			t.Fatalf("Scan: %v", err)
		}
		if m.GetCurrency() == money.COP {
			t.Fatal("precondition: Scan should not have set COP")
		}

		m.SetCurrency(money.COP)

		if m.GetCurrency() != money.COP {
			t.Errorf("currency = %v, want COP", m.GetCurrency())
		}
		if m.String() != "2500.75" {
			t.Errorf("value = %q, want 2500.75 unchanged", m.String())
		}
	})
}

func TestConvertSummaryTotalsRoundsToTheTargetCurrency(t *testing.T) {
	userID := uuid.New()

	t.Run("converted totals are rounded to the display currency precision", func(t *testing.T) {
		summary := SummaryView{
			BaseCurrency: money.USD, TotalMarketValue: "1234.56", TotalCostBase: "1000.00",
			TotalGainLoss: "234.56", TotalGainLossPct: "23.46",
		}

		s := convertSummary(t, userID, summary, money.COP, summaryRates{"USD/COP": "4123.4567"}).first(t)

		// 1234.56 * 4123.4567 = 5090654.703552 — the raw product runs to six
		// decimals; money.Convert rounds it to the two COP minor units.
		if s.TotalMarketValue != "5090654.7" {
			t.Errorf("TotalMarketValue = %q, want 5090654.7", s.TotalMarketValue)
		}
		if s.TotalCostBase != "4123456.7" {
			t.Errorf("TotalCostBase = %q, want 4123456.7", s.TotalCostBase)
		}
	})

	t.Run("a loss keeps its sign through the conversion", func(t *testing.T) {
		summary := SummaryView{
			BaseCurrency: money.USD, TotalMarketValue: "900", TotalCostBase: "1000",
			TotalGainLoss: "-100", TotalGainLossPct: "-10",
		}

		s := convertSummary(t, userID, summary, money.COP, summaryRates{"USD/COP": "4000"}).first(t)

		if s.TotalGainLoss != "-400000" {
			t.Errorf("TotalGainLoss = %q, want -400000", s.TotalGainLoss)
		}
	})

	// A base currency gofinance cannot parse is stored data, not caller input:
	// there is no amount to convert, so the portfolio is reported as it stands
	// instead of failing the whole request over one bad row.
	t.Run("a base currency gofinance does not know is left unconverted", func(t *testing.T) {
		summary := SummaryView{BaseCurrency: money.Currency(255), TotalMarketValue: "100"}

		s := convertSummary(t, userID, summary, money.USD, summaryRates{}).first(t)

		if s.FXConverted || s.TotalMarketValue != "100" || s.DisplayCurrency != money.Currency(255) {
			t.Errorf("summary = %+v, want the stored totals flagged unconverted", s)
		}
	})

	// The target, by contrast, is what the caller asked for: an unknown one is
	// their mistake to hear about, and the handler's list already rejects it.
	t.Run("an unknown display currency is a bad request, not a 500", func(t *testing.T) {
		summary := SummaryView{BaseCurrency: money.USD, TotalMarketValue: "100"}

		got := convertSummary(t, userID, summary, money.Currency(255), summaryRates{})

		if got.err == nil {
			t.Fatal("want an error for an unknown display currency")
		}
	})
}

func TestGetConversionRateRejectsUnusableRates(t *testing.T) {
	userID := uuid.New()

	cases := []struct {
		name string
		rate string
	}{
		{"zero rate", "0"},
		{"negative rate", "-4000"},
	}

	for _, tc := range cases {
		t.Run(tc.name+" is reported as unavailable", func(t *testing.T) {
			repo := new(fakeRepository{
				getExchangeRateByPair: func(_ context.Context, from, to money.Currency) (decimal.Decimal, error) {
					if from == money.USD && to == money.COP {
						return mustDecimal(t, tc.rate), nil
					}
					return decimal.Decimal{}, errors.New("exchange rate not found")
				},
			})
			svc := newTestServices(repo, newMemStorage())

			if _, err := svc.GetConversionRate(context.Background(), userID, money.USD, money.COP); !errors.Is(err, ErrExchangeRateUnavailable) {
				t.Errorf("err = %v, want ErrExchangeRateUnavailable", err)
			}
		})
	}

	t.Run("a corrupt direct rate falls through to the inverse", func(t *testing.T) {
		repo := new(fakeRepository{
			getExchangeRateByPair: func(_ context.Context, from, to money.Currency) (decimal.Decimal, error) {
				switch {
				case from == money.USD && to == money.COP:
					return mustDecimal(t, "0"), nil
				case from == money.COP && to == money.USD:
					return mustDecimal(t, "0.00025"), nil
				}
				return decimal.Decimal{}, errors.New("exchange rate not found")
			},
		})
		svc := newTestServices(repo, newMemStorage())

		rate, err := svc.GetConversionRate(context.Background(), userID, money.USD, money.COP)
		if err != nil {
			t.Fatalf("GetConversionRate: %v", err)
		}
		if rate.String() != "4000" {
			t.Errorf("rate = %s, want 4000 (1/0.00025)", rate.String())
		}
	})
}

func TestNewAllocationResponsePercentages(t *testing.T) {
	t.Run("shares are computed on the decimal engine", func(t *testing.T) {
		items := []AllocationItem{
			{Category: market.Stock, MarketValue: "333.33"},
			{Category: market.Crypto, MarketValue: "333.33"},
			{Category: market.Bond, MarketValue: "333.34"},
		}

		got := NewAllocationResponse(items)
		if len(got) != 3 {
			t.Fatalf("items = %d, want 3", len(got))
		}
		// 33.333, 33.333 and 33.334 — all three round down to 33.33, so the
		// displayed shares add up to 99.99. Rounding each share independently
		// is what the endpoint has always done; the point here is that the
		// figures come off the decimal engine exactly.
		for i, want := range []float64{33.33, 33.33, 33.33} {
			if got[i].Percent != want {
				t.Errorf("item %d percent = %v, want %v", i, got[i].Percent, want)
			}
		}
	})

	t.Run("an exact half rounds away from zero, as math.Round did", func(t *testing.T) {
		// 333.35 / 1000 is 33.335%, a tie at the second decimal. Banker's
		// rounding would answer 33.34 here and 33.32 for 333.25; half-away
		// answers 33.34 and 33.33.
		got := NewAllocationResponse([]AllocationItem{
			{Category: market.Stock, MarketValue: "333.35"},
			{Category: market.Crypto, MarketValue: "333.25"},
			{Category: market.Bond, MarketValue: "333.40"},
		})
		if got[0].Percent != 33.34 {
			t.Errorf("percent = %v, want 33.34", got[0].Percent)
		}
		if got[1].Percent != 33.33 {
			t.Errorf("percent = %v, want 33.33", got[1].Percent)
		}
	})

	t.Run("an all-zero portfolio reports zero rather than dividing by it", func(t *testing.T) {
		got := NewAllocationResponse([]AllocationItem{{Category: market.Stock, MarketValue: "0"}})
		if got[0].Percent != 0 {
			t.Errorf("percent = %v, want 0", got[0].Percent)
		}
	})

	t.Run("an unparsable value counts as zero", func(t *testing.T) {
		got := NewAllocationResponse([]AllocationItem{
			{Category: market.Stock, MarketValue: "n/a"},
			{Category: market.Crypto, MarketValue: "100"},
		})
		if got[0].Percent != 0 || got[1].Percent != 100 {
			t.Errorf("percents = %v/%v, want 0/100", got[0].Percent, got[1].Percent)
		}
	})

	t.Run("market values are passed through untouched", func(t *testing.T) {
		got := NewAllocationResponse([]AllocationItem{{Category: market.Stock, MarketValue: "1234.5600"}})
		if got[0].MarketValue != "1234.5600" {
			t.Errorf("MarketValue = %q, want it unchanged", got[0].MarketValue)
		}
	})
}

func TestNormalizeCurrencyValidatesAgainstISO4217(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want string
		ok   bool
	}{
		{"plain code", "usd", "USD", true},
		{"code inside a label", "Precio (COP)", "COP", true},
		{"symbol", "$", "USD", true},
		{"euro symbol", "€", "EUR", true},
		{"three letters that are not a currency", "ABC", "", false},
		{"a word with no code in it", "pesos", "", false},
		{"empty", "  ", "", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := normalizeCurrency(tc.raw)
			if ok != tc.ok || got != tc.want {
				t.Errorf("normalizeCurrency(%q) = %q/%v, want %q/%v", tc.raw, got, ok, tc.want, tc.ok)
			}
		})
	}
}

func TestGrowthSummaryKeepsPrecisionThroughLargeValues(t *testing.T) {
	// A float64 round-trip of these two loses the trailing cents; the decimal
	// engine reports the exact 10% the values encode.
	points := []GrowthPoint{
		{TotalValue: "12345678901234.56"},
		{TotalValue: "13580246791358.016"},
	}

	got := buildGrowthSummary(points)
	if got.TotalGrowthPct != "10.00" {
		t.Errorf("TotalGrowthPct = %q, want 10.00", got.TotalGrowthPct)
	}
	if got.InitialValue != "12345678901234.56" {
		t.Errorf("InitialValue = %q, want 12345678901234.56", got.InitialValue)
	}
}

func TestGrowthSummaryTreatsUnparsableValuesAsZero(t *testing.T) {
	got := buildGrowthSummary([]GrowthPoint{{TotalValue: "n/a"}, {TotalValue: "500"}})
	if got.InitialValue != "0.00" || got.TotalGrowthPct != "0.00" {
		t.Errorf("summary = %+v, want a zero start and 0.00%%", got)
	}
}

func TestTransactionAlertTotalKeepsCents(t *testing.T) {
	// The exact product is 87943.90050551296. Routing it through float64, as
	// the alert used to, is what put cents at the mercy of binary rounding.
	price := money.MustMoneyFromString("0.07123456", money.USD)
	qty := decimal.MustFromString("1234567.891")

	got := price.MulDecimal(qty).RoundBank(2).StringFixed(2)
	if got != "87943.90" {
		t.Errorf("total = %q, want 87943.90", got)
	}
}

// The allocation groups by the asset's own type: 000026 dropped
// portfolio_entries.category, a copy taken at insert time that nothing updated,
// because correcting an asset afterwards moved it in the per-portfolio donut
// and left it in the old slice on the dashboard. The category is now the
// asset_type enum itself, so the only place a free-form label still has to be
// placed is the importer, and NormalizeAssetType is what places it.
func TestAllocationCategoriesComeFromTheAssetType(t *testing.T) {
	cases := []struct {
		assetType string
		want      market.AssetType
	}{
		{"stock", market.Stock},
		{"etf", market.ETF},
		{"crypto", market.Crypto},
		{"bond", market.Bond},
		{"cash", market.Cash},
		{"real_estate", market.RealEstate},
		{"commodity", market.Commodity},
		{"other", market.Other},
		// Anything the mapping does not know falls to Others rather than
		// leaking a raw enum value the clients have no label for.
		{"structured_note", market.Other},
	}

	for _, tc := range cases {
		// The importer's own fallback: a label the mapping cannot place is
		// filed as Other rather than written straight through.
		got, ok := market.NormalizeAssetType(tc.assetType)
		if !ok {
			got = market.Other
		}
		if got != tc.want {
			t.Errorf("NormalizeAssetType(%q) = %q, want %q", tc.assetType, got, tc.want)
		}
	}
}

func TestFoldAllocationByCategory(t *testing.T) {
	t.Run("adds the rows that share a category", func(t *testing.T) {
		// Two unrecognised asset types both land on Others. Keeping only one
		// would drop the other's money out of the chart entirely.
		got := foldAllocationByCategory([]AllocationItem{
			{Category: market.Other, MarketValue: "300.50", Currency: money.USD, PositionsUnconverted: 1},
			{Category: market.Stock, MarketValue: "250", Currency: money.USD},
			{Category: market.Other, MarketValue: "100.25", Currency: money.USD, PositionsUnconverted: 2},
		})

		if len(got) != 2 {
			t.Fatalf("items = %+v, want two categories", got)
		}
		if got[0].Category != market.Other || got[0].MarketValue != "400.75" {
			t.Errorf("first = %+v, want Others at 400.75", got[0])
		}
		if got[0].PositionsUnconverted != 3 {
			t.Errorf("positionsUnconverted = %d, want 3", got[0].PositionsUnconverted)
		}
	})

	t.Run("restores the by-value order a merge can disturb", func(t *testing.T) {
		// The query hands these back sorted, but folding the two Others rows
		// makes them outweigh the Stocks row that sat between them.
		got := foldAllocationByCategory([]AllocationItem{
			{Category: market.Other, MarketValue: "60"},
			{Category: market.Stock, MarketValue: "50"},
			{Category: market.Other, MarketValue: "40"},
		})

		if got[0].Category != market.Other || got[0].MarketValue != "100" {
			t.Errorf("first = %+v, want Others at 100", got[0])
		}
		if got[1].Category != market.Stock {
			t.Errorf("second = %+v, want Stocks", got[1])
		}
	})

	t.Run("an unparsable amount counts as zero instead of dropping the row", func(t *testing.T) {
		got := foldAllocationByCategory([]AllocationItem{
			{Category: market.Stock, MarketValue: "n/a"},
			{Category: market.Bond, MarketValue: "10"},
		})

		if len(got) != 2 || got[0].Category != market.Bond {
			t.Fatalf("items = %+v, want Bonds first and Stocks kept", got)
		}
	})
}
