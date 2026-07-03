package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yeferson59/gofinance/money"

	"github.com/yeferson59/finexia-app/internal/entities"
	"github.com/yeferson59/finexia-app/internal/prices"
)

func TestWasExchangeRateSyncedRecently(t *testing.T) {
	storage := newMemStorage()
	svc := newTestServices(&fakeRepository{}, storage)

	if svc.WasExchangeRateSyncedRecently() {
		t.Error("expected false before any sync")
	}
	if err := storage.Set(rateSyncCacheKey, []byte("2026-07-03T00:00:00Z"), time.Hour); err != nil {
		t.Fatalf("storage.Set: %v", err)
	}
	if !svc.WasExchangeRateSyncedRecently() {
		t.Error("expected true after the sync marker is set")
	}
}

func TestSyncExchangeRates(t *testing.T) {
	fetchedAt := time.Date(2026, 7, 1, 6, 0, 0, 0, time.UTC)

	t.Run("fetch failures are collected per pair and the marker is still set", func(t *testing.T) {
		var pairsRequested []string
		provider := &fakePriceProvider{
			fetchExchangeRate: func(_ context.Context, from, to string) (prices.ExchangeRateResult, error) {
				pairsRequested = append(pairsRequested, from+"/"+to)
				return prices.ExchangeRateResult{}, errors.New("provider down")
			},
		}
		storage := newMemStorage()
		svc := newTestServicesFull(&fakeRepository{}, storage, nil, provider)

		results, errs := svc.SyncExchangeRates(context.Background())
		if len(results) != 0 {
			t.Errorf("results = %+v, want none", results)
		}
		if len(errs) != len(defaultPairs) {
			t.Errorf("errs = %d, want one per pair (%d)", len(errs), len(defaultPairs))
		}
		if len(pairsRequested) != len(defaultPairs) {
			t.Errorf("requested pairs = %v, want all %d defaults", pairsRequested, len(defaultPairs))
		}
		if !svc.WasExchangeRateSyncedRecently() {
			t.Error("the sync marker is set even when every pair fails")
		}
	})

	t.Run("unparseable rates are rejected", func(t *testing.T) {
		provider := &fakePriceProvider{
			fetchExchangeRate: func(context.Context, string, string) (prices.ExchangeRateResult, error) {
				return prices.ExchangeRateResult{Rate: "not-a-number", FetchedAt: fetchedAt}, nil
			},
		}
		repo := &fakeRepository{
			upsertExchangeRate: func(context.Context, string, string, money.Decimal, time.Time) (entities.ExchangeRate, error) {
				t.Error("upsert must not run for an unparseable rate")
				return entities.ExchangeRate{}, nil
			},
		}
		svc := newTestServicesFull(repo, newMemStorage(), nil, provider)

		results, errs := svc.SyncExchangeRates(context.Background())
		if len(results) != 0 || len(errs) != len(defaultPairs) {
			t.Errorf("results/errs = %d/%d, want 0/%d", len(results), len(errs), len(defaultPairs))
		}
	})

	t.Run("upsert failures are collected", func(t *testing.T) {
		provider := &fakePriceProvider{
			fetchExchangeRate: func(context.Context, string, string) (prices.ExchangeRateResult, error) {
				return prices.ExchangeRateResult{Rate: "1.0850", FetchedAt: fetchedAt}, nil
			},
		}
		repo := &fakeRepository{
			upsertExchangeRate: func(context.Context, string, string, money.Decimal, time.Time) (entities.ExchangeRate, error) {
				return entities.ExchangeRate{}, errors.New("db write failed")
			},
		}
		svc := newTestServicesFull(repo, newMemStorage(), nil, provider)

		results, errs := svc.SyncExchangeRates(context.Background())
		if len(results) != 0 || len(errs) != len(defaultPairs) {
			t.Errorf("results/errs = %d/%d, want 0/%d", len(results), len(errs), len(defaultPairs))
		}
	})

	t.Run("a successful pair is persisted and cancellation stops the run", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		provider := &fakePriceProvider{
			fetchExchangeRate: func(_ context.Context, from, to string) (prices.ExchangeRateResult, error) {
				if from != "EUR" || to != "USD" {
					t.Errorf("first pair = %s/%s, want EUR/USD", from, to)
				}
				return prices.ExchangeRateResult{Rate: "1.0850", FetchedAt: fetchedAt}, nil
			},
		}
		repo := &fakeRepository{
			upsertExchangeRate: func(_ context.Context, from, to string, rate money.Decimal, rateDate time.Time) (entities.ExchangeRate, error) {
				if rate.String() != "1.085" {
					t.Errorf("rate = %s, want 1.085", rate.String())
				}
				if !rateDate.Equal(fetchedAt) {
					t.Errorf("rateDate = %v, want %v", rateDate, fetchedAt)
				}
				// Cancel before the inter-pair rate-limit sleep so the test
				// exercises the early-return path without waiting 13s.
				cancel()
				return entities.ExchangeRate{FromCurrency: from, ToCurrency: to, Rate: rate, RateDate: rateDate}, nil
			},
		}
		storage := newMemStorage()
		svc := newTestServicesFull(repo, storage, nil, provider)

		start := time.Now()
		results, errs := svc.SyncExchangeRates(ctx)
		if elapsed := time.Since(start); elapsed > 5*time.Second {
			t.Fatalf("sync took %v; cancellation should skip the rate-limit sleep", elapsed)
		}
		if len(results) != 1 || len(errs) != 0 {
			t.Fatalf("results/errs = %d/%d, want 1/0", len(results), len(errs))
		}
		if results[0].FromCurrency != "EUR" || results[0].ToCurrency != "USD" {
			t.Errorf("result pair = %s/%s", results[0].FromCurrency, results[0].ToCurrency)
		}
		if svc.WasExchangeRateSyncedRecently() {
			t.Error("a cancelled run must not set the sync marker")
		}
	})
}
