package portfolio

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/entities"
)

func usd(t *testing.T, s string) money.Money {
	t.Helper()
	cur, err := money.CurrencyFromISOCode("USD")
	if err != nil {
		t.Fatalf("CurrencyFromISOCode: %v", err)
	}
	m, err := money.NewMoneyFromString(s, cur)
	if err != nil {
		t.Fatalf("NewMoneyFromString(%q): %v", s, err)
	}
	return m
}

func dec(t *testing.T, s string) money.Decimal {
	t.Helper()
	d, err := money.NewFromString(s)
	if err != nil {
		t.Fatalf("NewFromString(%q): %v", s, err)
	}
	return d
}

func TestNewTransactionResponse(t *testing.T) {
	txn := entities.Transaction{
		ID:              uuid.New(),
		EntryID:         uuid.New(),
		Type:            entities.Buy,
		Quantity:        dec(t, "2.5"),
		Price:           usd(t, "100.10"),
		Currency:        "USD",
		Fees:            usd(t, "1.25"),
		TransactionDate: time.Date(2026, 6, 20, 0, 0, 0, 0, time.UTC),
		Notes:           "first buy",
	}

	got := NewTransactionResponse(txn)
	if got.ID != txn.ID || got.EntryID != txn.EntryID {
		t.Error("IDs not mapped")
	}
	if got.Type != "buy" {
		t.Errorf("Type = %q, want buy", got.Type)
	}
	if got.Quantity != txn.Quantity.String() || got.Price != txn.Price.String() || got.Fees != txn.Fees.String() {
		t.Errorf("amounts = %q/%q/%q", got.Quantity, got.Price, got.Fees)
	}
	if got.Notes != "first buy" {
		t.Errorf("Notes = %q", got.Notes)
	}
}

func TestNewTransactionListResponse(t *testing.T) {
	got := NewTransactionListResponse(nil)
	if got == nil || len(got) != 0 {
		t.Errorf("nil input should map to an empty non-nil slice, got %#v", got)
	}

	txns := []entities.Transaction{{ID: uuid.New(), Type: entities.Sell}, {ID: uuid.New(), Type: entities.Buy}}
	got = NewTransactionListResponse(txns)
	if len(got) != 2 || got[0].Type != "sell" || got[1].Type != "buy" {
		t.Errorf("mapped = %+v", got)
	}
}

func TestNewUserTransactionResponse(t *testing.T) {
	txn := entities.Transaction{
		ID:   uuid.New(),
		Type: entities.Dividend,
		Entry: entities.PortfolioEntry{
			Asset: entities.Asset{Ticker: "AAPL", Name: "Apple Inc."},
		},
	}

	got := NewUserTransactionResponse(txn)
	if got.AssetTicker != "AAPL" || got.AssetName != "Apple Inc." {
		t.Errorf("asset fields = %q/%q", got.AssetTicker, got.AssetName)
	}
	if got.Type != "dividend" {
		t.Errorf("Type = %q", got.Type)
	}

	list := NewUserTransactionListResponse([]entities.Transaction{txn})
	if len(list) != 1 || list[0].AssetTicker != "AAPL" {
		t.Errorf("list = %+v", list)
	}
}

func TestNewAllocationResponse(t *testing.T) {
	t.Run("computes rounded percentages of the total", func(t *testing.T) {
		items := []entities.AllocationItem{
			{Category: entities.Stocks, MarketValue: "750.00"},
			{Category: entities.Cryptos, MarketValue: "250.00"},
		}

		got := NewAllocationResponse(items)
		if len(got) != 2 {
			t.Fatalf("len = %d, want 2", len(got))
		}
		if got[0].Percent != 75 || got[1].Percent != 25 {
			t.Errorf("percents = %v/%v, want 75/25", got[0].Percent, got[1].Percent)
		}
		if got[0].Category != "stocks" || got[0].MarketValue != "750.00" {
			t.Errorf("item = %+v", got[0])
		}
	})

	t.Run("rounds to two decimals", func(t *testing.T) {
		items := []entities.AllocationItem{
			{Category: entities.Stocks, MarketValue: "1.00"},
			{Category: entities.Bonds, MarketValue: "2.00"},
		}

		got := NewAllocationResponse(items)
		if got[0].Percent != 33.33 || got[1].Percent != 66.67 {
			t.Errorf("percents = %v/%v, want 33.33/66.67", got[0].Percent, got[1].Percent)
		}
	})

	t.Run("zero total yields zero percentages", func(t *testing.T) {
		items := []entities.AllocationItem{{Category: entities.Cashs, MarketValue: "0"}}

		got := NewAllocationResponse(items)
		if got[0].Percent != 0 {
			t.Errorf("percent = %v, want 0", got[0].Percent)
		}
	})

	t.Run("empty input yields empty output", func(t *testing.T) {
		got := NewAllocationResponse(nil)
		if got == nil || len(got) != 0 {
			t.Errorf("expected empty non-nil slice, got %#v", got)
		}
	})
}

func TestNewGrowthResponse(t *testing.T) {
	t.Run("formats dates and copies values", func(t *testing.T) {
		points := []entities.PortfolioGrowthPoint{
			{Date: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC), TotalValue: "1000.00", TotalCostBase: "900.00", GainLoss: "100.00", GainLossPct: "11.11"},
		}
		summary := entities.PortfolioGrowthSummary{
			FirstDate:      time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
			InitialValue:   "1000.00",
			CurrentValue:   "1250.00",
			TotalGrowthPct: "25.00",
		}

		got := NewGrowthResponse(points, summary)
		if len(got.Points) != 1 {
			t.Fatalf("points = %d, want 1", len(got.Points))
		}
		if got.Points[0].Date != "2026-05-01" {
			t.Errorf("date = %q, want 2026-05-01", got.Points[0].Date)
		}
		if got.Points[0].TotalValue != "1000.00" || got.Points[0].GainLossPct != "11.11" {
			t.Errorf("point = %+v", got.Points[0])
		}
		if got.Summary.FirstDate != "2026-05-01" || got.Summary.TotalGrowthPct != "25.00" {
			t.Errorf("summary = %+v", got.Summary)
		}
	})

	t.Run("zero first date maps to an empty string", func(t *testing.T) {
		got := NewGrowthResponse(nil, entities.PortfolioGrowthSummary{})
		if got.Summary.FirstDate != "" {
			t.Errorf("FirstDate = %q, want empty", got.Summary.FirstDate)
		}
		if got.Points == nil || len(got.Points) != 0 {
			t.Errorf("expected empty non-nil points, got %#v", got.Points)
		}
	})
}

func TestNewPortfolioDetailResponse(t *testing.T) {
	price := usd(t, "150.00")
	entry := entities.PortfolioEntry{
		ID:           uuid.New(),
		AssetID:      uuid.New(),
		Quantity:     dec(t, "3"),
		Price:        usd(t, "120.00"),
		CostCurrency: "USD",
		Category:     entities.Stocks,
		EntryDate:    time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		Notes:        "core position",
		Asset: entities.Asset{
			Ticker:       "AAPL",
			Name:         "Apple Inc.",
			AssetType:    entities.Stock,
			Exchange:     "NASDAQ",
			Currency:     "USD",
			CurrentPrice: &price,
		},
	}
	noPriceEntry := entities.PortfolioEntry{
		ID:    uuid.New(),
		Asset: entities.Asset{Ticker: "BND"},
	}

	p := entities.Portfolio{
		ID:               uuid.New(),
		UserID:           uuid.New(),
		Name:             "Main",
		Description:      "primary portfolio",
		Type:             entities.PortfolioTypeStocks,
		BaseCurrency:     "USD",
		IsDefault:        true,
		RiskID:           uuid.New(),
		Risk:             entities.Risk{Name: "moderate"},
		PortfolioEntries: []entities.PortfolioEntry{entry, noPriceEntry},
	}

	got := NewPortfolioDetailResponse(p)
	if got.ID != p.ID || got.UserID != p.UserID || got.RiskID != p.RiskID {
		t.Error("IDs not mapped")
	}
	if got.RiskName != "moderate" {
		t.Errorf("RiskName = %q", got.RiskName)
	}
	if len(got.Holdings) != 2 {
		t.Fatalf("holdings = %d, want 2", len(got.Holdings))
	}

	h := got.Holdings[0]
	if h.Ticker != "AAPL" || h.AssetType != "stock" || h.Exchange != "NASDAQ" {
		t.Errorf("holding = %+v", h)
	}
	if h.MarketPrice != price.String() {
		t.Errorf("MarketPrice = %q, want %q", h.MarketPrice, price.String())
	}
	if h.Quantity != entry.Quantity.String() || h.Price != entry.Price.String() {
		t.Errorf("quantity/price = %q/%q", h.Quantity, h.Price)
	}

	if got.Holdings[1].MarketPrice != "" {
		t.Errorf("MarketPrice = %q, want empty when the asset has no current price", got.Holdings[1].MarketPrice)
	}
}
