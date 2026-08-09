package market

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/yeferson59/finexia-app/internal/platform/logger"
	"github.com/yeferson59/finexia-app/internal/platform/marketdata"
	"github.com/yeferson59/finexia-app/internal/platform/secretbox"
)

type defaultAsset struct {
	Ticker    string
	Name      string
	AssetType AssetType
	Exchange  string
	Currency  string
}

var defaultAssets = []defaultAsset{
	{"AAPL", "Apple Inc.", Stock, "NASDAQ", "USD"},
	{"MSFT", "Microsoft Corporation", Stock, "NASDAQ", "USD"},
	{"SPY", "SPDR S&P 500 ETF Trust", ETF, "NYSEARCA", "USD"},
	{"BTC-USD", "Bitcoin", Crypto, "Coinbase", "USD"},
	{"ETH-USD", "Ethereum", Crypto, "Coinbase", "USD"},
	{"BND", "Vanguard Total Bond Market ETF", Bond, "NASDAQ", "USD"},
}

// Pacing between two calls made with the same user's key. Alpha Vantage's free
// tier allows 5 requests/minute, Finnhub's 60, so the interval is chosen from
// the slowest provider the user actually has configured rather than applied
// uniformly.
const (
	alphaVantagePace = 13 * time.Second
	finnhubPace      = time.Second
)

type Service struct {
	repo    Repository
	storage fiber.Storage
	// providers builds a chain from a user's own keys, per sync run.
	providers marketdata.Factory
	// publicRates is the keyless feed behind the shared exchange rates. It is
	// not part of the chain above and cannot be: that one is assembled from
	// credentials, and this source takes none. Optional — a deployment without
	// it keeps the admin-entered rates and nothing else.
	publicRates marketdata.PublicRateSource
	// keyring seals and opens those keys.
	keyring *secretbox.Keyring
	log     logger.Logger
}

func newService(repo Repository, storage fiber.Storage, providers marketdata.Factory, publicRates marketdata.PublicRateSource, keyring *secretbox.Keyring, log logger.Logger) *Service {
	return new(Service{
		repo:        repo,
		storage:     storage,
		providers:   providers,
		publicRates: publicRates,
		keyring:     keyring,
		log:         log,
	})
}

// SeedDefaultAssets makes sure the shared catalog has the handful of well-known
// instruments the app ships with. It touches no provider and needs no key: it
// only creates catalog rows, whose prices each user then fills with their own
// key.
func (s *Service) SeedDefaultAssets(ctx context.Context) []error {
	var errs []error

	for _, da := range defaultAssets {
		if _, err := s.CreateAsset(ctx, da.Ticker, da.Name, da.AssetType, da.Exchange, da.Currency); err != nil {
			s.log.Error(ctx, "upsert default asset failed", logger.Err(err), logger.Str("ticker", da.Ticker))
			errs = append(errs, err)
		}
	}

	return errs
}
