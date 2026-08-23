package market

import (
	"context"
	"testing"

	"uuid"

	"github.com/yeferson59/finexia-app/internal/platform/logger"
	"github.com/yeferson59/finexia-app/internal/platform/marketdata"
)

func TestSyncJobRun(t *testing.T) {
	userID := uuid.New()
	assetID := uuid.New()

	repoFor := func() *fakeRepository {
		return new(fakeRepository{
			getAssetByID: func(context.Context, uuid.UUID) (Asset, error) {
				return Asset{ID: assetID, Ticker: "AAPL", AssetType: Stock, Currency: "USD"}, nil
			},
			upsertAsset: func(context.Context, string, string, AssetType, string, string) (Asset, error) {
				return Asset{}, nil
			},
		})
	}

	holdingsFor := func() stubHoldings {
		return stubHoldings{
			assetIDs: []uuid.UUID{assetID},
			pairs:    []CurrencyPair{{From: "USD", To: "COP"}},
		}
	}

	// The regression: syncUser used to return as soon as any asset failed, so a
	// single ticker the provider does not cover left the user's currency pairs
	// unsynced — and with the shared table emptied, unconvertible.
	t.Run("a failing asset does not stop the rate sync", func(t *testing.T) {
		var ratesFetched int

		provider := new(fakePriceProvider{
			fetchQuote: func(context.Context, string) (marketdata.QuoteResult, error) {
				return marketdata.QuoteResult{}, providerErr(Finnhub, marketdata.ErrUnsupported, "finnhub: no data for AAPL")
			},
			fetchExchangeRate: func(context.Context, string, string) (marketdata.ExchangeRateResult, error) {
				ratesFetched++

				return marketdata.ExchangeRateResult{Rate: "4100.5", Source: Finnhub}, nil
			},
		})

		f := newBYOFixture(t, repoFor(), provider)
		f.creds.seed(t, f.ring, userID, Finnhub, "user-finnhub-key")

		job := NewSyncJob(f.svc, holdingsFor(), logger.Noop())
		if err := job.Run(context.Background()); err == nil {
			t.Error("Run() = nil, want the asset failure reported")
		}

		if ratesFetched != 1 {
			t.Fatalf("fetched %d rates, want 1 — the rate leg was skipped", ratesFetched)
		}
	})

	t.Run("a user with no key is not a failure", func(t *testing.T) {
		f := newBYOFixture(t, repoFor(), quoteOK())
		// No seeded credential, but the user is still walked: UsersWithCredentials
		// is answered by the same store, so this is the empty-store case.
		job := NewSyncJob(f.svc, holdingsFor(), logger.Noop())

		if err := job.Run(context.Background()); err != nil {
			t.Fatalf("Run() = %v, want nil when nobody has a key", err)
		}
	})

	t.Run("a healthy run reports no error", func(t *testing.T) {
		provider := new(fakePriceProvider{
			fetchQuote: func(context.Context, string) (marketdata.QuoteResult, error) {
				return marketdata.QuoteResult{Price: "190.55", Source: Finnhub}, nil
			},
			fetchExchangeRate: func(context.Context, string, string) (marketdata.ExchangeRateResult, error) {
				return marketdata.ExchangeRateResult{Rate: "4100.5", Source: Finnhub}, nil
			},
		})

		f := newBYOFixture(t, repoFor(), provider)
		f.creds.seed(t, f.ring, userID, Finnhub, "user-finnhub-key")

		job := NewSyncJob(f.svc, holdingsFor(), logger.Noop())
		if err := job.Run(context.Background()); err != nil {
			t.Fatalf("Run() = %v, want nil", err)
		}

		if _, ok := f.creds.priceOf(userID, assetID); !ok {
			t.Error("no price stored against the user")
		}
	})
}
