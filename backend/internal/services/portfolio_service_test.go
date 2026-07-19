package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/yeferson59/gofinance/v2/money"

	portfoliodto "github.com/yeferson59/finexia-app/internal/dtos/portfolio"
	"github.com/yeferson59/finexia-app/internal/entities"
)

func mustUSD(t *testing.T, s string) money.Money {
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

func mustDecimal(t *testing.T, s string) money.Decimal {
	t.Helper()
	d, err := money.NewFromString(s)
	if err != nil {
		t.Fatalf("NewFromString(%q): %v", s, err)
	}
	return d
}

// waitFor polls cond until it returns true or the deadline expires. Used to
// synchronise with the fire-and-forget alert goroutine.
func waitFor(t *testing.T, timeout time.Duration, cond func() bool) bool {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return true
		}
		time.Sleep(5 * time.Millisecond)
	}
	return cond()
}

func TestGetPortfoliosRisks(t *testing.T) {
	t.Run("returns risks from the repository", func(t *testing.T) {
		want := []entities.Risk{{ID: uuid.New(), Name: "conservative"}}
		repo := &fakeRepository{
			getPortfoliosRisks: func(context.Context) ([]entities.Risk, error) { return want, nil },
		}
		svc := newTestServices(repo, newMemStorage())

		got, err := svc.GetPortfoliosRisks(context.Background())
		if err != nil {
			t.Fatalf("GetPortfoliosRisks: %v", err)
		}
		if len(got) != 1 || got[0].Name != "conservative" {
			t.Errorf("risks = %+v, want %+v", got, want)
		}
	})

	t.Run("repository error yields empty slice", func(t *testing.T) {
		repo := &fakeRepository{
			getPortfoliosRisks: func(context.Context) ([]entities.Risk, error) {
				return nil, errors.New("db down")
			},
		}
		svc := newTestServices(repo, newMemStorage())

		got, err := svc.GetPortfoliosRisks(context.Background())
		if err == nil {
			t.Fatal("expected error")
		}
		if got == nil || len(got) != 0 {
			t.Errorf("expected empty non-nil slice, got %+v", got)
		}
	})

	t.Run("second call is served from the cache", func(t *testing.T) {
		calls := 0
		repo := &fakeRepository{
			getPortfoliosRisks: func(context.Context) ([]entities.Risk, error) {
				calls++
				return []entities.Risk{{ID: uuid.New(), Name: "moderate"}}, nil
			},
		}
		svc := newTestServices(repo, newMemStorage())

		for range 2 {
			got, err := svc.GetPortfoliosRisks(context.Background())
			if err != nil {
				t.Fatalf("GetPortfoliosRisks: %v", err)
			}
			if len(got) != 1 || got[0].Name != "moderate" {
				t.Errorf("risks = %+v", got)
			}
		}
		if calls != 1 {
			t.Errorf("repository calls = %d, want 1 (cached)", calls)
		}
	})

	t.Run("errors are not cached", func(t *testing.T) {
		calls := 0
		repo := &fakeRepository{
			getPortfoliosRisks: func(context.Context) ([]entities.Risk, error) {
				calls++
				if calls == 1 {
					return nil, errors.New("db down")
				}
				return []entities.Risk{{ID: uuid.New(), Name: "aggressive"}}, nil
			},
		}
		svc := newTestServices(repo, newMemStorage())

		if _, err := svc.GetPortfoliosRisks(context.Background()); err == nil {
			t.Fatal("expected error on first call")
		}
		got, err := svc.GetPortfoliosRisks(context.Background())
		if err != nil {
			t.Fatalf("GetPortfoliosRisks after error: %v", err)
		}
		if len(got) != 1 || got[0].Name != "aggressive" {
			t.Errorf("risks = %+v", got)
		}
	})
}

func TestGetPortfolios(t *testing.T) {
	userID := uuid.New()

	t.Run("passes the user ID through", func(t *testing.T) {
		repo := &fakeRepository{
			getPortfoliosByUserID: func(_ context.Context, uid uuid.UUID) ([]entities.Portfolio, error) {
				if uid != userID {
					t.Errorf("userID = %s, want %s", uid, userID)
				}
				return []entities.Portfolio{{ID: uuid.New(), Name: "Main"}}, nil
			},
		}
		svc := newTestServices(repo, newMemStorage())

		got, err := svc.GetPortfolios(context.Background(), userID)
		if err != nil {
			t.Fatalf("GetPortfolios: %v", err)
		}
		if len(got) != 1 || got[0].Name != "Main" {
			t.Errorf("portfolios = %+v", got)
		}
	})

	t.Run("repository error yields empty slice", func(t *testing.T) {
		repo := &fakeRepository{
			getPortfoliosByUserID: func(context.Context, uuid.UUID) ([]entities.Portfolio, error) {
				return nil, errors.New("boom")
			},
		}
		svc := newTestServices(repo, newMemStorage())

		got, err := svc.GetPortfolios(context.Background(), userID)
		if err == nil {
			t.Fatal("expected error")
		}
		if got == nil || len(got) != 0 {
			t.Errorf("expected empty non-nil slice, got %+v", got)
		}
	})
}

func TestGetPortfoliosSummary(t *testing.T) {
	userID := uuid.New()
	repo := &fakeRepository{
		getPortfoliosSummaryByUserID: func(_ context.Context, uid uuid.UUID) ([]entities.PortfolioSummaryView, error) {
			if uid != userID {
				t.Errorf("userID = %s, want %s", uid, userID)
			}
			return []entities.PortfolioSummaryView{{Name: "Growth", TotalMarketValue: "1000.00"}}, nil
		},
	}
	svc := newTestServices(repo, newMemStorage())

	got, err := svc.GetPortfoliosSummary(context.Background(), userID)
	if err != nil {
		t.Fatalf("GetPortfoliosSummary: %v", err)
	}
	if len(got) != 1 || got[0].TotalMarketValue != "1000.00" {
		t.Errorf("summaries = %+v", got)
	}
}

func TestGetPortfoliosSummaryInCurrency(t *testing.T) {
	userID := uuid.New()

	t.Run("same currency skips conversion but still sets DisplayCurrency", func(t *testing.T) {
		repo := &fakeRepository{
			getPortfoliosSummaryByUserID: func(context.Context, uuid.UUID) ([]entities.PortfolioSummaryView, error) {
				return []entities.PortfolioSummaryView{{
					BaseCurrency: "USD", TotalMarketValue: "1000", TotalCostBase: "900",
					TotalGainLoss: "100", TotalGainLossPct: "11.11",
				}}, nil
			},
			getExchangeRateByPair: func(context.Context, string, string) (entities.ExchangeRate, error) {
				t.Error("no rate lookup expected when currencies match")
				return entities.ExchangeRate{}, errors.New("unexpected call")
			},
		}
		svc := newTestServices(repo, newMemStorage())

		got, err := svc.GetPortfoliosSummaryInCurrency(context.Background(), userID, "USD")
		if err != nil {
			t.Fatalf("GetPortfoliosSummaryInCurrency: %v", err)
		}
		if len(got) != 1 || got[0].DisplayCurrency != "USD" || got[0].TotalMarketValue != "1000" {
			t.Errorf("summaries = %+v", got)
		}
	})

	t.Run("converts totals with the direct rate and leaves the percentage untouched", func(t *testing.T) {
		repo := &fakeRepository{
			getPortfoliosSummaryByUserID: func(context.Context, uuid.UUID) ([]entities.PortfolioSummaryView, error) {
				return []entities.PortfolioSummaryView{{
					BaseCurrency: "USD", TotalMarketValue: "1000", TotalCostBase: "900",
					TotalGainLoss: "100", TotalGainLossPct: "11.11",
				}}, nil
			},
			getExchangeRateByPair: func(_ context.Context, from, to string) (entities.ExchangeRate, error) {
				if from == "USD" && to == "COP" {
					return entities.ExchangeRate{FromCurrency: "USD", ToCurrency: "COP", Rate: mustDecimal(t, "4000")}, nil
				}
				return entities.ExchangeRate{}, errors.New("exchange rate not found")
			},
		}
		svc := newTestServices(repo, newMemStorage())

		got, err := svc.GetPortfoliosSummaryInCurrency(context.Background(), userID, "COP")
		if err != nil {
			t.Fatalf("GetPortfoliosSummaryInCurrency: %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("summaries = %+v, want one", got)
		}
		s := got[0]
		if s.DisplayCurrency != "COP" {
			t.Errorf("DisplayCurrency = %s, want COP", s.DisplayCurrency)
		}
		if s.TotalMarketValue != "4000000" || s.TotalCostBase != "3600000" || s.TotalGainLoss != "400000" {
			t.Errorf("converted totals = %+v", s)
		}
		if s.TotalGainLossPct != "11.11" {
			t.Errorf("TotalGainLossPct = %s, want unchanged 11.11", s.TotalGainLossPct)
		}
	})

	t.Run("falls back to the inverse rate when only the opposite pair is synced", func(t *testing.T) {
		repo := &fakeRepository{
			getPortfoliosSummaryByUserID: func(context.Context, uuid.UUID) ([]entities.PortfolioSummaryView, error) {
				return []entities.PortfolioSummaryView{{
					BaseCurrency: "COP", TotalMarketValue: "4000000", TotalCostBase: "0",
					TotalGainLoss: "0", TotalGainLossPct: "0",
				}}, nil
			},
			getExchangeRateByPair: func(_ context.Context, from, to string) (entities.ExchangeRate, error) {
				if from == "USD" && to == "COP" {
					return entities.ExchangeRate{FromCurrency: "USD", ToCurrency: "COP", Rate: mustDecimal(t, "4000")}, nil
				}
				return entities.ExchangeRate{}, errors.New("exchange rate not found")
			},
		}
		svc := newTestServices(repo, newMemStorage())

		got, err := svc.GetPortfoliosSummaryInCurrency(context.Background(), userID, "USD")
		if err != nil {
			t.Fatalf("GetPortfoliosSummaryInCurrency: %v", err)
		}
		if got[0].TotalMarketValue != "1000" {
			t.Errorf("TotalMarketValue = %s, want 1000", got[0].TotalMarketValue)
		}
	})

	t.Run("propagates an error when no rate connects the pair", func(t *testing.T) {
		repo := &fakeRepository{
			getPortfoliosSummaryByUserID: func(context.Context, uuid.UUID) ([]entities.PortfolioSummaryView, error) {
				return []entities.PortfolioSummaryView{{BaseCurrency: "EUR", TotalMarketValue: "100"}}, nil
			},
			getExchangeRateByPair: func(context.Context, string, string) (entities.ExchangeRate, error) {
				return entities.ExchangeRate{}, errors.New("exchange rate not found")
			},
		}
		svc := newTestServices(repo, newMemStorage())

		if _, err := svc.GetPortfoliosSummaryInCurrency(context.Background(), userID, "COP"); !errors.Is(err, ErrExchangeRateUnavailable) {
			t.Errorf("err = %v, want ErrExchangeRateUnavailable", err)
		}
	})
}

func TestGetPortfolio(t *testing.T) {
	userID, portfolioID := uuid.New(), uuid.New()

	t.Run("attaches entries to the portfolio", func(t *testing.T) {
		entryID := uuid.New()
		repo := &fakeRepository{
			getPortfolioByID: func(_ context.Context, pid, uid uuid.UUID) (entities.Portfolio, error) {
				if pid != portfolioID || uid != userID {
					t.Errorf("ids = %s/%s, want %s/%s", pid, uid, portfolioID, userID)
				}
				return entities.Portfolio{ID: portfolioID, Name: "Main"}, nil
			},
			getEntriesByPortfolioID: func(_ context.Context, pid uuid.UUID) ([]entities.PortfolioEntry, error) {
				return []entities.PortfolioEntry{{ID: entryID, PortfolioID: pid}}, nil
			},
		}
		svc := newTestServices(repo, newMemStorage())

		got, err := svc.GetPortfolio(context.Background(), userID, portfolioID)
		if err != nil {
			t.Fatalf("GetPortfolio: %v", err)
		}
		if len(got.PortfolioEntries) != 1 || got.PortfolioEntries[0].ID != entryID {
			t.Errorf("entries = %+v, want one entry %s", got.PortfolioEntries, entryID)
		}
	})

	// Both lookups run concurrently, so a failed portfolio lookup must return
	// its error and discard whatever the entries query produced.
	t.Run("portfolio lookup error discards fetched entries", func(t *testing.T) {
		repo := &fakeRepository{
			getPortfolioByID: func(context.Context, uuid.UUID, uuid.UUID) (entities.Portfolio, error) {
				return entities.Portfolio{}, errors.New("portfolio not found")
			},
			getEntriesByPortfolioID: func(context.Context, uuid.UUID) ([]entities.PortfolioEntry, error) {
				return []entities.PortfolioEntry{{ID: uuid.New()}}, nil
			},
		}
		svc := newTestServices(repo, newMemStorage())

		got, err := svc.GetPortfolio(context.Background(), userID, portfolioID)
		if err == nil {
			t.Fatal("expected error")
		}
		if len(got.PortfolioEntries) != 0 {
			t.Errorf("entries must not leak on error, got %+v", got.PortfolioEntries)
		}
	})

	t.Run("entries lookup error is returned", func(t *testing.T) {
		repo := &fakeRepository{
			getPortfolioByID: func(context.Context, uuid.UUID, uuid.UUID) (entities.Portfolio, error) {
				return entities.Portfolio{ID: portfolioID}, nil
			},
			getEntriesByPortfolioID: func(context.Context, uuid.UUID) ([]entities.PortfolioEntry, error) {
				return nil, errors.New("entries broke")
			},
		}
		svc := newTestServices(repo, newMemStorage())

		if _, err := svc.GetPortfolio(context.Background(), userID, portfolioID); err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestCreatePortfolio(t *testing.T) {
	userID, riskID := uuid.New(), uuid.New()
	price := mustUSD(t, "5000.00")

	t.Run("passes all fields to the repository", func(t *testing.T) {
		repo := &fakeRepository{
			createPortfolio: func(_ context.Context, uid uuid.UUID, name, description, baseCurrency string, rid uuid.UUID, typePortfolio entities.PortfolioType, priceValue money.Money, isDefault bool) (entities.Portfolio, error) {
				if uid != userID || rid != riskID {
					t.Errorf("ids = %s/%s, want %s/%s", uid, rid, userID, riskID)
				}
				if name != "Retirement" || description != "long term" || baseCurrency != "USD" {
					t.Errorf("fields = %q/%q/%q", name, description, baseCurrency)
				}
				if typePortfolio != entities.PortfolioTypeDiversified {
					t.Errorf("type = %q, want diversified", typePortfolio)
				}
				if !priceValue.Equal(price) {
					t.Errorf("price = %s, want %s", priceValue.String(), price.String())
				}
				if !isDefault {
					t.Error("isDefault should be true")
				}
				return entities.Portfolio{ID: uuid.New(), Name: name}, nil
			},
		}
		svc := newTestServices(repo, newMemStorage())

		got, err := svc.CreatePortfolio(context.Background(), userID, "Retirement", "long term", "USD", riskID, entities.PortfolioTypeDiversified, price, true)
		if err != nil {
			t.Fatalf("CreatePortfolio: %v", err)
		}
		if got.Name != "Retirement" {
			t.Errorf("name = %q, want Retirement", got.Name)
		}
	})

	t.Run("repository error returns zero portfolio", func(t *testing.T) {
		repo := &fakeRepository{
			createPortfolio: func(context.Context, uuid.UUID, string, string, string, uuid.UUID, entities.PortfolioType, money.Money, bool) (entities.Portfolio, error) {
				return entities.Portfolio{}, errors.New("duplicate name")
			},
		}
		svc := newTestServices(repo, newMemStorage())

		got, err := svc.CreatePortfolio(context.Background(), userID, "X", "", "USD", riskID, entities.PortfolioTypeStocks, price, false)
		if err == nil {
			t.Fatal("expected error")
		}
		if got.ID != uuid.Nil {
			t.Errorf("expected zero portfolio, got %+v", got)
		}
	})
}

func TestUpdatePortfolio(t *testing.T) {
	userID, portfolioID, riskID := uuid.New(), uuid.New(), uuid.New()
	repo := &fakeRepository{
		updatePortfolio: func(_ context.Context, uid, pid uuid.UUID, name, description string, portfolioType entities.PortfolioType, rid uuid.UUID, isDefault bool) (entities.Portfolio, error) {
			if uid != userID || pid != portfolioID || rid != riskID {
				t.Error("IDs not forwarded correctly")
			}
			return entities.Portfolio{ID: pid, Name: name, Type: portfolioType, IsDefault: isDefault}, nil
		},
	}
	svc := newTestServices(repo, newMemStorage())

	got, err := svc.UpdatePortfolio(context.Background(), userID, portfolioID, "Renamed", "desc", entities.PortfolioTypeCryptos, riskID, true)
	if err != nil {
		t.Fatalf("UpdatePortfolio: %v", err)
	}
	if got.Name != "Renamed" || got.Type != entities.PortfolioTypeCryptos || !got.IsDefault {
		t.Errorf("updated = %+v", got)
	}
}

func TestGetPortfolioTopTransaction(t *testing.T) {
	userID, portfolioID := uuid.New(), uuid.New()
	repo := &fakeRepository{
		getTopTransactionByPortfolio: func(_ context.Context, uid, pid uuid.UUID) (portfoliodto.PortfolioTopTransactionDTO, error) {
			if uid != userID || pid != portfolioID {
				t.Error("IDs not forwarded correctly")
			}
			return portfoliodto.PortfolioTopTransactionDTO{Value: "990.00", AssetTicker: "AAPL"}, nil
		},
	}
	svc := newTestServices(repo, newMemStorage())

	got, err := svc.GetPortfolioTopTransaction(context.Background(), userID, portfolioID)
	if err != nil {
		t.Fatalf("GetPortfolioTopTransaction: %v", err)
	}
	if got.Value != "990.00" || got.AssetTicker != "AAPL" {
		t.Errorf("top transaction = %+v", got)
	}
}

func TestCreatePlatform(t *testing.T) {
	userID := uuid.New()

	t.Run("lowercases the platform name", func(t *testing.T) {
		repo := &fakeRepository{
			createPlatform: func(_ context.Context, uid uuid.UUID, sourceType entities.SourceType, name, description string) (entities.InvestmentSource, error) {
				if name != "interactive brokers" {
					t.Errorf("name = %q, want lowercase %q", name, "interactive brokers")
				}
				if sourceType != entities.Broker {
					t.Errorf("sourceType = %q, want broker", sourceType)
				}
				return entities.InvestmentSource{ID: uuid.New(), Name: name, SourceType: sourceType}, nil
			},
		}
		svc := newTestServices(repo, newMemStorage())

		got, err := svc.CreatePlatform(context.Background(), userID, entities.Broker, "Interactive Brokers", "main broker")
		if err != nil {
			t.Fatalf("CreatePlatform: %v", err)
		}
		if got.Name != "interactive brokers" {
			t.Errorf("name = %q", got.Name)
		}
	})

	t.Run("repository error returns zero source", func(t *testing.T) {
		repo := &fakeRepository{
			createPlatform: func(context.Context, uuid.UUID, entities.SourceType, string, string) (entities.InvestmentSource, error) {
				return entities.InvestmentSource{}, errors.New("already exists")
			},
		}
		svc := newTestServices(repo, newMemStorage())

		got, err := svc.CreatePlatform(context.Background(), userID, entities.Broker, "X", "")
		if err == nil {
			t.Fatal("expected error")
		}
		if got.ID != uuid.Nil {
			t.Errorf("expected zero source, got %+v", got)
		}
	})
}

func TestPlatformLifecycle(t *testing.T) {
	userID, sourceID := uuid.New(), uuid.New()

	t.Run("GetPlatforms returns stats", func(t *testing.T) {
		repo := &fakeRepository{
			getPlatformsWithStats: func(_ context.Context, uid uuid.UUID) ([]entities.PlatformStats, error) {
				return []entities.PlatformStats{{ID: sourceID, Name: "binance", Investments: 4, TotalValue: "250.00"}}, nil
			},
		}
		svc := newTestServices(repo, newMemStorage())

		got, err := svc.GetPlatforms(context.Background(), userID)
		if err != nil {
			t.Fatalf("GetPlatforms: %v", err)
		}
		if len(got) != 1 || got[0].Investments != 4 {
			t.Errorf("platforms = %+v", got)
		}
	})

	t.Run("UpdatePlatform forwards every field", func(t *testing.T) {
		repo := &fakeRepository{
			updatePlatform: func(_ context.Context, uid, sid uuid.UUID, name, description string, sourceType entities.SourceType, isActive bool) (entities.PlatformStats, error) {
				if uid != userID || sid != sourceID {
					t.Error("IDs not forwarded correctly")
				}
				if sourceType != entities.CryptoWallet || isActive {
					t.Errorf("sourceType/isActive = %q/%v", sourceType, isActive)
				}
				return entities.PlatformStats{ID: sid, Name: name, IsActive: isActive}, nil
			},
		}
		svc := newTestServices(repo, newMemStorage())

		got, err := svc.UpdatePlatform(context.Background(), userID, sourceID, "ledger", "cold wallet", entities.CryptoWallet, false)
		if err != nil {
			t.Fatalf("UpdatePlatform: %v", err)
		}
		if got.Name != "ledger" {
			t.Errorf("name = %q", got.Name)
		}
	})

	t.Run("DeletePlatform propagates the error", func(t *testing.T) {
		want := errors.New("platform not found")
		repo := &fakeRepository{
			deletePlatform: func(_ context.Context, uid, sid uuid.UUID) error { return want },
		}
		svc := newTestServices(repo, newMemStorage())

		if err := svc.DeletePlatform(context.Background(), userID, sourceID); !errors.Is(err, want) {
			t.Errorf("err = %v, want %v", err, want)
		}
	})
}

func TestAssetOperations(t *testing.T) {
	t.Run("GetAssets forwards pagination", func(t *testing.T) {
		repo := &fakeRepository{
			getAssets: func(_ context.Context, offset, limit uint) ([]entities.Asset, error) {
				if offset != 20 || limit != 10 {
					t.Errorf("offset/limit = %d/%d, want 20/10", offset, limit)
				}
				return []entities.Asset{{Ticker: "AAPL"}}, nil
			},
		}
		svc := newTestServices(repo, newMemStorage())

		got, err := svc.GetAssets(context.Background(), 20, 10)
		if err != nil || len(got) != 1 {
			t.Fatalf("GetAssets = %+v, %v", got, err)
		}
	})

	t.Run("SearchAssets forwards the query", func(t *testing.T) {
		repo := &fakeRepository{
			searchAssets: func(_ context.Context, search string, offset, limit uint) ([]entities.Asset, error) {
				if search != "apple" {
					t.Errorf("search = %q, want apple", search)
				}
				return []entities.Asset{{Ticker: "AAPL"}}, nil
			},
		}
		svc := newTestServices(repo, newMemStorage())

		got, err := svc.SearchAssets(context.Background(), "apple", 0, 5)
		if err != nil || len(got) != 1 {
			t.Fatalf("SearchAssets = %+v, %v", got, err)
		}
	})

	t.Run("CreateAsset upserts", func(t *testing.T) {
		repo := &fakeRepository{
			upsertAsset: func(_ context.Context, ticker, name string, assetType entities.AssetType, exchange, currency string) (entities.Asset, error) {
				if ticker != "MSFT" || assetType != entities.Stock {
					t.Errorf("ticker/type = %q/%q", ticker, assetType)
				}
				return entities.Asset{ID: uuid.New(), Ticker: ticker}, nil
			},
		}
		svc := newTestServices(repo, newMemStorage())

		got, err := svc.CreateAsset(context.Background(), "MSFT", "Microsoft", entities.Stock, "NASDAQ", "USD")
		if err != nil || got.Ticker != "MSFT" {
			t.Fatalf("CreateAsset = %+v, %v", got, err)
		}
	})

	t.Run("UpdateAssetPrice forwards the price", func(t *testing.T) {
		assetID := uuid.New()
		price := mustUSD(t, "412.50")
		repo := &fakeRepository{
			updateAssetPrice: func(_ context.Context, aid uuid.UUID, p money.Money) (entities.Asset, error) {
				if aid != assetID || !p.Equal(price) {
					t.Errorf("asset/price = %s/%s", aid, p.String())
				}
				return entities.Asset{ID: aid, CurrentPrice: &p}, nil
			},
		}
		svc := newTestServices(repo, newMemStorage())

		got, err := svc.UpdateAssetPrice(context.Background(), assetID, price)
		if err != nil || got.CurrentPrice == nil {
			t.Fatalf("UpdateAssetPrice = %+v, %v", got, err)
		}
	})
}

func TestCreatePortfolioEntry(t *testing.T) {
	userID, portfolioID, assetID, sourceID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	qty := mustDecimal(t, "2.5")
	price := mustUSD(t, "100.00")
	entryDate := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)

	t.Run("forwards all fields", func(t *testing.T) {
		repo := &fakeRepository{
			createPortfolioEntry: func(_ context.Context, uid, pid, aid, sid uuid.UUID, txnType entities.TransactionType, quantity money.Decimal, p money.Money, costCurrency string, category entities.PortfolioEntryCategory, ed time.Time, notes string) (entities.PortfolioEntry, error) {
				if uid != userID || pid != portfolioID || aid != assetID || sid != sourceID {
					t.Error("IDs not forwarded correctly")
				}
				if txnType != entities.Buy || category != entities.Stocks {
					t.Errorf("type/category = %q/%q", txnType, category)
				}
				if !ed.Equal(entryDate) || notes != "first buy" {
					t.Errorf("date/notes = %v/%q", ed, notes)
				}
				return entities.PortfolioEntry{ID: uuid.New(), PortfolioID: pid}, nil
			},
		}
		svc := newTestServices(repo, newMemStorage())

		got, err := svc.CreatePortfolioEntry(context.Background(), userID, portfolioID, assetID, sourceID, entities.Buy, qty, price, "USD", entities.Stocks, entryDate, "first buy")
		if err != nil {
			t.Fatalf("CreatePortfolioEntry: %v", err)
		}
		if got.PortfolioID != portfolioID {
			t.Errorf("portfolioID = %s", got.PortfolioID)
		}
	})

	t.Run("repository error returns zero entry", func(t *testing.T) {
		repo := &fakeRepository{
			createPortfolioEntry: func(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID, entities.TransactionType, money.Decimal, money.Money, string, entities.PortfolioEntryCategory, time.Time, string) (entities.PortfolioEntry, error) {
				return entities.PortfolioEntry{}, errors.New("asset not found")
			},
		}
		svc := newTestServices(repo, newMemStorage())

		got, err := svc.CreatePortfolioEntry(context.Background(), userID, portfolioID, assetID, sourceID, entities.Buy, qty, price, "USD", entities.Stocks, entryDate, "")
		if err == nil {
			t.Fatal("expected error")
		}
		if got.ID != uuid.Nil {
			t.Errorf("expected zero entry, got %+v", got)
		}
	})
}

func TestTransactionQueries(t *testing.T) {
	userID := uuid.New()

	t.Run("GetTransactionsByEntry", func(t *testing.T) {
		entryID := uuid.New()
		repo := &fakeRepository{
			getTransactionsByEntryID: func(_ context.Context, uid, eid uuid.UUID) ([]entities.Transaction, error) {
				if uid != userID || eid != entryID {
					t.Error("IDs not forwarded correctly")
				}
				return []entities.Transaction{{ID: uuid.New(), EntryID: eid}}, nil
			},
		}
		svc := newTestServices(repo, newMemStorage())

		got, err := svc.GetTransactionsByEntry(context.Background(), userID, entryID)
		if err != nil || len(got) != 1 {
			t.Fatalf("GetTransactionsByEntry = %+v, %v", got, err)
		}
	})

	t.Run("GetRecentUserTransactions forwards the limit", func(t *testing.T) {
		repo := &fakeRepository{
			getRecentTransactionsByUserID: func(_ context.Context, uid uuid.UUID, limit int) ([]entities.Transaction, error) {
				if limit != 50 {
					t.Errorf("limit = %d, want 50", limit)
				}
				return []entities.Transaction{}, nil
			},
		}
		svc := newTestServices(repo, newMemStorage())

		if _, err := svc.GetRecentUserTransactions(context.Background(), userID, 50); err != nil {
			t.Fatalf("GetRecentUserTransactions: %v", err)
		}
	})

	t.Run("GetAssetAllocation", func(t *testing.T) {
		repo := &fakeRepository{
			getAssetAllocationByUserID: func(context.Context, uuid.UUID) ([]entities.AllocationItem, error) {
				return []entities.AllocationItem{{Category: entities.Stocks, MarketValue: "750.00"}}, nil
			},
		}
		svc := newTestServices(repo, newMemStorage())

		got, err := svc.GetAssetAllocation(context.Background(), userID)
		if err != nil || len(got) != 1 || got[0].Category != entities.Stocks {
			t.Fatalf("GetAssetAllocation = %+v, %v", got, err)
		}
	})
}

func TestUpdateTransaction(t *testing.T) {
	userID, txnID := uuid.New(), uuid.New()
	qty := mustDecimal(t, "1")
	price := mustUSD(t, "10.00")
	fees := mustUSD(t, "0.50")
	date := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)

	repo := &fakeRepository{
		updateTransaction: func(_ context.Context, uid, tid uuid.UUID, txnType entities.TransactionType, quantity money.Decimal, p money.Money, currency string, f money.Money, transactionDate time.Time, notes string) (entities.Transaction, error) {
			if uid != userID || tid != txnID {
				t.Error("IDs not forwarded correctly")
			}
			if txnType != entities.Sell || currency != "USD" || !transactionDate.Equal(date) {
				t.Errorf("type/currency/date = %q/%q/%v", txnType, currency, transactionDate)
			}
			return entities.Transaction{ID: tid, Type: txnType}, nil
		},
	}
	svc := newTestServices(repo, newMemStorage())

	got, err := svc.UpdateTransaction(context.Background(), userID, txnID, entities.Sell, qty, price, "USD", fees, date, "sold half")
	if err != nil {
		t.Fatalf("UpdateTransaction: %v", err)
	}
	if got.Type != entities.Sell {
		t.Errorf("type = %q, want sell", got.Type)
	}
}

func TestCreateTransactionSendsAlert(t *testing.T) {
	userID, entryID := uuid.New(), uuid.New()
	qty := mustDecimal(t, "2")
	price := mustUSD(t, "150.25")
	fees := mustUSD(t, "1.00")
	date := time.Date(2026, 6, 20, 0, 0, 0, 0, time.UTC)

	newTxnRepo := func() *fakeRepository {
		return &fakeRepository{
			createTransaction: func(_ context.Context, uid, eid uuid.UUID, txnType entities.TransactionType, quantity money.Decimal, p money.Money, currency string, f money.Money, transactionDate time.Time, notes string) (entities.Transaction, error) {
				return entities.Transaction{
					ID: uuid.New(), EntryID: eid, Type: txnType,
					Quantity: quantity, Price: p, Currency: currency,
					Fees: f, TransactionDate: transactionDate, Notes: notes,
				}, nil
			},
		}
	}

	t.Run("alert email sent when the user opted in", func(t *testing.T) {
		repo := newTxnRepo()

		repo.getEntryWithAsset = func(context.Context, uuid.UUID) (entities.PortfolioEntry, error) {
			return entities.PortfolioEntry{ID: entryID, Asset: entities.Asset{Ticker: "AAPL", Name: "Apple Inc."}}, nil
		}

		mailer := &fakeMailer{}
		svc := newTestServicesFull(repo, newMemStorage(), mailer, nil)
		svc.cfg.FrontendURL = "https://app.finexia.me"

		txn, err := svc.CreateTransaction(context.Background(), userID, entryID, entities.Buy, qty, price, "USD", fees, date, "note")
		if err != nil {
			t.Fatalf("CreateTransaction: %v", err)
		}
		if txn.Type != entities.Buy {
			t.Errorf("type = %q, want buy", txn.Type)
		}

		if !waitFor(t, 2*time.Second, func() bool {
			mailer.mu.Lock()
			defer mailer.mu.Unlock()
			return len(mailer.activity) == 1
		}) {
			t.Fatal("expected one activity alert email")
		}

		mailer.mu.Lock()
		defer mailer.mu.Unlock()
		sent := mailer.activity[0]
		if sent.To != "ada@example.com" {
			t.Errorf("recipient = %q", sent.To)
		}
		if sent.Data.AssetTicker != "AAPL" || sent.Data.UserName != "Ada" {
			t.Errorf("alert data = %+v", sent.Data)
		}
		if sent.Data.Total != "300.50" {
			t.Errorf("total = %q, want 300.50 (2 x 150.25)", sent.Data.Total)
		}
		if sent.Data.DashboardURL != "https://app.finexia.me/dashboard/portfolios" {
			t.Errorf("dashboard URL = %q", sent.Data.DashboardURL)
		}
	})

	t.Run("no alert when alerts are disabled", func(t *testing.T) {
		prefsChecked := make(chan struct{})
		repo := newTxnRepo()

		mailer := &fakeMailer{}
		svc := newTestServicesFull(repo, newMemStorage(), mailer, nil)

		if _, err := svc.CreateTransaction(context.Background(), userID, entryID, entities.Buy, qty, price, "USD", fees, date, ""); err != nil {
			t.Fatalf("CreateTransaction: %v", err)
		}

		select {
		case <-prefsChecked:
		case <-time.After(2 * time.Second):
			t.Fatal("preferences were never checked")
		}
		time.Sleep(20 * time.Millisecond)

		mailer.mu.Lock()
		defer mailer.mu.Unlock()
		if len(mailer.activity) != 0 {
			t.Errorf("expected no alert emails, got %d", len(mailer.activity))
		}
	})

	t.Run("preferences lookup failure suppresses the alert", func(t *testing.T) {
		prefsChecked := make(chan struct{})
		repo := newTxnRepo()

		mailer := &fakeMailer{}
		svc := newTestServicesFull(repo, newMemStorage(), mailer, nil)

		if _, err := svc.CreateTransaction(context.Background(), userID, entryID, entities.Buy, qty, price, "USD", fees, date, ""); err != nil {
			t.Fatalf("CreateTransaction: %v", err)
		}

		select {
		case <-prefsChecked:
		case <-time.After(2 * time.Second):
			t.Fatal("preferences were never checked")
		}
		time.Sleep(20 * time.Millisecond)

		mailer.mu.Lock()
		defer mailer.mu.Unlock()
		if len(mailer.activity) != 0 {
			t.Errorf("expected no alert emails, got %d", len(mailer.activity))
		}
	})

	t.Run("repository error is returned and no alert goes out", func(t *testing.T) {
		repo := &fakeRepository{
			createTransaction: func(context.Context, uuid.UUID, uuid.UUID, entities.TransactionType, money.Decimal, money.Money, string, money.Money, time.Time, string) (entities.Transaction, error) {
				return entities.Transaction{}, errors.New("entry not found")
			},
		}
		mailer := &fakeMailer{}
		svc := newTestServicesFull(repo, newMemStorage(), mailer, nil)

		got, err := svc.CreateTransaction(context.Background(), userID, entryID, entities.Buy, qty, price, "USD", fees, date, "")
		if err == nil {
			t.Fatal("expected error")
		}
		if got.ID != uuid.Nil {
			t.Errorf("expected zero transaction, got %+v", got)
		}

		time.Sleep(20 * time.Millisecond)
		mailer.mu.Lock()
		defer mailer.mu.Unlock()
		if len(mailer.activity) != 0 {
			t.Errorf("expected no alert emails, got %d", len(mailer.activity))
		}
	})
}
