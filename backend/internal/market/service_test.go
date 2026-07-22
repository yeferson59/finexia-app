package market

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/platform/marketdata"
	"github.com/yeferson59/finexia-app/internal/portfolio"
)

func TestWasAssetPriceSyncedRecently(t *testing.T) {
	storage := newMemStorage()
	svc := newTestServices(&fakeRepository{}, storage)

	if svc.WasAssetPriceSyncedRecently() {
		t.Error("expected false before any sync")
	}
	if err := storage.Set(assetSyncCacheKey, []byte("2026-07-03T00:00:00Z"), time.Hour); err != nil {
		t.Fatalf("storage.Set: %v", err)
	}
	if !svc.WasAssetPriceSyncedRecently() {
		t.Error("expected true after the sync marker is set")
	}
}

func TestSyncAssetByID(t *testing.T) {
	assetID := uuid.New()

	stockAsset := portfolio.Asset{ID: assetID, Ticker: "AAPL", AssetType: portfolio.Stock, Currency: "USD"}
	cryptoAsset := portfolio.Asset{ID: assetID, Ticker: "BTC-USD", AssetType: portfolio.Crypto, Currency: "USD"}

	storeFor := func(asset portfolio.Asset) *fakePortfolioService {
		return &fakePortfolioService{
			getAssetByID: func(_ context.Context, id uuid.UUID) (portfolio.Asset, error) {
				if id != assetID {
					t.Errorf("assetID = %s, want %s", id, assetID)
				}
				return asset, nil
			},
			updateAssetPrice: func(_ context.Context, id uuid.UUID, price money.Money) (portfolio.Asset, error) {
				updated := asset
				updated.CurrentPrice = &price
				return updated, nil
			},
		}
	}

	t.Run("stock price comes from a quote", func(t *testing.T) {
		provider := &fakePriceProvider{
			fetchQuote: func(_ context.Context, symbol string) (marketdata.QuoteResult, error) {
				if symbol != "AAPL" {
					t.Errorf("symbol = %q, want AAPL", symbol)
				}
				return marketdata.QuoteResult{Price: "190.55"}, nil
			},
		}
		svc := newTestServicesFull(&fakeRepository{}, newMemStorage(), nil, provider, storeFor(stockAsset))

		got, err := svc.SyncAssetByID(context.Background(), assetID)
		if err != nil {
			t.Fatalf("SyncAssetByID: %v", err)
		}
		if got.CurrentPrice == nil || got.CurrentPrice.String() != mustUSD(t, "190.55").String() {
			t.Errorf("price = %v, want 190.55 USD", got.CurrentPrice)
		}
	})

	t.Run("crypto price comes from an exchange rate on the split ticker", func(t *testing.T) {
		provider := &fakePriceProvider{
			fetchExchangeRate: func(_ context.Context, from, to string) (marketdata.ExchangeRateResult, error) {
				if from != "BTC" || to != "USD" {
					t.Errorf("pair = %s/%s, want BTC/USD", from, to)
				}
				return marketdata.ExchangeRateResult{Rate: "64000.10"}, nil
			},
		}
		svc := newTestServicesFull(&fakeRepository{}, newMemStorage(), nil, provider, storeFor(cryptoAsset))

		got, err := svc.SyncAssetByID(context.Background(), assetID)
		if err != nil {
			t.Fatalf("SyncAssetByID: %v", err)
		}
		if got.CurrentPrice == nil {
			t.Fatal("expected a price")
		}
	})

	t.Run("malformed crypto ticker is rejected", func(t *testing.T) {
		bad := portfolio.Asset{ID: assetID, Ticker: "BTCUSD", AssetType: portfolio.Crypto, Currency: "USD"}
		svc := newTestServicesFull(&fakeRepository{}, newMemStorage(), nil, &fakePriceProvider{}, storeFor(bad))

		if _, err := svc.SyncAssetByID(context.Background(), assetID); err == nil || !strings.Contains(err.Error(), "cannot parse crypto ticker") {
			t.Errorf("err = %v, want crypto ticker parse error", err)
		}
	})

	t.Run("unsupported asset type is reported", func(t *testing.T) {
		cash := portfolio.Asset{ID: assetID, Ticker: "CASH", AssetType: portfolio.Cash, Currency: "USD"}
		svc := newTestServicesFull(&fakeRepository{}, newMemStorage(), nil, &fakePriceProvider{}, storeFor(cash))

		if _, err := svc.SyncAssetByID(context.Background(), assetID); err == nil || !strings.Contains(err.Error(), "not supported") {
			t.Errorf("err = %v, want not-supported error", err)
		}
	})

	t.Run("provider failure is wrapped", func(t *testing.T) {
		provider := &fakePriceProvider{
			fetchQuote: func(context.Context, string) (marketdata.QuoteResult, error) {
				return marketdata.QuoteResult{}, errors.New("rate limited")
			},
		}
		svc := newTestServicesFull(&fakeRepository{}, newMemStorage(), nil, provider, storeFor(stockAsset))

		if _, err := svc.SyncAssetByID(context.Background(), assetID); err == nil || !strings.Contains(err.Error(), "rate limited") {
			t.Errorf("err = %v, want wrapped provider error", err)
		}
	})

	t.Run("unknown currency is rejected", func(t *testing.T) {
		weird := portfolio.Asset{ID: assetID, Ticker: "AAPL", AssetType: portfolio.Stock, Currency: "ZZZ"}
		provider := &fakePriceProvider{
			fetchQuote: func(context.Context, string) (marketdata.QuoteResult, error) {
				return marketdata.QuoteResult{Price: "10.00"}, nil
			},
		}
		svc := newTestServicesFull(&fakeRepository{}, newMemStorage(), nil, provider, storeFor(weird))

		if _, err := svc.SyncAssetByID(context.Background(), assetID); err == nil || !strings.Contains(err.Error(), "unknown currency") {
			t.Errorf("err = %v, want unknown currency error", err)
		}
	})

	t.Run("unparseable price is rejected", func(t *testing.T) {
		provider := &fakePriceProvider{
			fetchQuote: func(context.Context, string) (marketdata.QuoteResult, error) {
				return marketdata.QuoteResult{Price: "n/a"}, nil
			},
		}
		svc := newTestServicesFull(&fakeRepository{}, newMemStorage(), nil, provider, storeFor(stockAsset))

		if _, err := svc.SyncAssetByID(context.Background(), assetID); err == nil || !strings.Contains(err.Error(), "parse price") {
			t.Errorf("err = %v, want parse price error", err)
		}
	})

	t.Run("persistence failure is wrapped", func(t *testing.T) {
		store := storeFor(stockAsset)
		store.updateAssetPrice = func(context.Context, uuid.UUID, money.Money) (portfolio.Asset, error) {
			return portfolio.Asset{}, errors.New("db write failed")
		}
		provider := &fakePriceProvider{
			fetchQuote: func(context.Context, string) (marketdata.QuoteResult, error) {
				return marketdata.QuoteResult{Price: "10.00"}, nil
			},
		}
		svc := newTestServicesFull(&fakeRepository{}, newMemStorage(), nil, provider, store)

		if _, err := svc.SyncAssetByID(context.Background(), assetID); err == nil || !strings.Contains(err.Error(), "persist price") {
			t.Errorf("err = %v, want persist error", err)
		}
	})

	t.Run("asset lookup failure stops early", func(t *testing.T) {
		store := &fakePortfolioService{
			getAssetByID: func(context.Context, uuid.UUID) (portfolio.Asset, error) {
				return portfolio.Asset{}, errors.New("asset not found")
			},
		}
		svc := newTestServicesFull(&fakeRepository{}, newMemStorage(), nil, &fakePriceProvider{}, store)

		if _, err := svc.SyncAssetByID(context.Background(), assetID); err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestSyncAssetPrices(t *testing.T) {
	t.Run("seeds default assets and prices the catalog", func(t *testing.T) {
		var seeded []string
		stock := portfolio.Asset{ID: uuid.New(), Ticker: "AAPL", AssetType: portfolio.Stock, Currency: "USD"}
		cash := portfolio.Asset{ID: uuid.New(), Ticker: "CASH", AssetType: portfolio.Cash, Currency: "USD"}

		store := &fakePortfolioService{
			createAsset: func(_ context.Context, ticker, name string, assetType portfolio.AssetType, exchange, currency string) (portfolio.Asset, error) {
				seeded = append(seeded, ticker)
				return portfolio.Asset{Ticker: ticker}, nil
			},
			getAssets: func(_ context.Context, offset, limit uint) ([]portfolio.Asset, error) {
				// Cash first: it is skipped without an API call, so no
				// rate-limit sleep happens before the single priced asset.
				return []portfolio.Asset{cash, stock}, nil
			},
			updateAssetPrice: func(_ context.Context, id uuid.UUID, price money.Money) (portfolio.Asset, error) {
				updated := stock
				updated.CurrentPrice = &price
				return updated, nil
			},
		}
		provider := &fakePriceProvider{
			fetchQuote: func(context.Context, string) (marketdata.QuoteResult, error) {
				return marketdata.QuoteResult{Price: "190.55"}, nil
			},
		}
		storage := newMemStorage()
		svc := newTestServicesFull(&fakeRepository{}, storage, nil, provider, store)

		results, errs := svc.SyncAssetPrices(context.Background())
		if len(errs) != 0 {
			t.Fatalf("errs = %v", errs)
		}
		if len(seeded) != len(defaultAssets) {
			t.Errorf("seeded %d default assets, want %d", len(seeded), len(defaultAssets))
		}
		if len(results) != 1 || results[0].Ticker != "AAPL" {
			t.Errorf("results = %+v, want the priced AAPL asset only (cash skipped)", results)
		}
		if !svc.WasAssetPriceSyncedRecently() {
			t.Error("sync marker should be set after a run")
		}
	})

	t.Run("catalog fetch failure aborts", func(t *testing.T) {
		store := &fakePortfolioService{
			getAssets: func(context.Context, uint, uint) ([]portfolio.Asset, error) {
				return nil, errors.New("db down")
			},
		}
		storage := newMemStorage()
		svc := newTestServicesFull(&fakeRepository{}, storage, nil, &fakePriceProvider{}, store)

		results, errs := svc.SyncAssetPrices(context.Background())
		if results != nil {
			t.Errorf("results = %+v, want nil", results)
		}
		if len(errs) != 1 {
			t.Errorf("errs = %v, want the fetch error", errs)
		}
		if svc.WasAssetPriceSyncedRecently() {
			t.Error("sync marker must not be set when the catalog fetch fails")
		}
	})

	t.Run("seed failures are collected but do not stop pricing", func(t *testing.T) {
		store := &fakePortfolioService{
			createAsset: func(context.Context, string, string, portfolio.AssetType, string, string) (portfolio.Asset, error) {
				return portfolio.Asset{}, errors.New("seed failed")
			},
			getAssets: func(context.Context, uint, uint) ([]portfolio.Asset, error) {
				return []portfolio.Asset{}, nil
			},
		}
		svc := newTestServicesFull(&fakeRepository{}, newMemStorage(), nil, &fakePriceProvider{}, store)

		results, errs := svc.SyncAssetPrices(context.Background())
		if len(errs) != len(defaultAssets) {
			t.Errorf("errs = %d, want one per default asset (%d)", len(errs), len(defaultAssets))
		}
		if len(results) != 0 {
			t.Errorf("results = %+v, want none", results)
		}
	})

	t.Run("cancellation stops before the next rate-limited call", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		a1 := portfolio.Asset{ID: uuid.New(), Ticker: "AAPL", AssetType: portfolio.Stock, Currency: "USD"}
		a2 := portfolio.Asset{ID: uuid.New(), Ticker: "MSFT", AssetType: portfolio.Stock, Currency: "USD"}

		var quoteCalls int
		store := &fakePortfolioService{
			getAssets: func(context.Context, uint, uint) ([]portfolio.Asset, error) {
				return []portfolio.Asset{a1, a2}, nil
			},
		}
		provider := &fakePriceProvider{
			fetchQuote: func(context.Context, string) (marketdata.QuoteResult, error) {
				quoteCalls++
				cancel() // cancel after the first asset so the 13s sleep before the second is skipped
				return marketdata.QuoteResult{Price: "10.00"}, nil
			},
		}
		storage := newMemStorage()
		svc := newTestServicesFull(&fakeRepository{}, storage, nil, provider, store)

		start := time.Now()
		results, errs := svc.SyncAssetPrices(ctx)
		if elapsed := time.Since(start); elapsed > 5*time.Second {
			t.Fatalf("sync took %v; cancellation should skip the rate-limit sleep", elapsed)
		}
		if quoteCalls != 1 {
			t.Errorf("quote calls = %d, want 1", quoteCalls)
		}
		if len(results) != 1 || len(errs) != 0 {
			t.Errorf("results/errs = %d/%d, want 1/0", len(results), len(errs))
		}
		if svc.WasAssetPriceSyncedRecently() {
			t.Error("a cancelled run must not set the sync marker")
		}
	})
}
