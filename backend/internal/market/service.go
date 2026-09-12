package market

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/platform/logger"
	"github.com/yeferson59/finexia-app/internal/platform/marketdata"
	"github.com/yeferson59/finexia-app/internal/platform/secretbox"
)

type defaultAsset struct {
	Ticker    string
	Name      string
	AssetType AssetType
	Exchange  string
	Currency  money.Currency
	Sector    Sector
}

// The two crypto rows carry SectorNone because a coin has no industry behind
// it, not because nobody got round to them — see AssetType.HasSector.
//
// SPY is the interesting one: a whole-market fund is spread across all eleven
// sectors, so filing it under any single one would be a lie the breakdown then
// repeats. What it needs is a SectorBreakdown, and the seed deliberately does
// not ship one. These weights change every quarter and this list is re-applied
// on every boot, so a breakdown here would overwrite the operator's own numbers
// with whatever was true the day the constant was typed. It stays unclassified
// until somebody loads the fund's published weights through the admin edit or
// the spreadsheet import, which is the one place they can also be kept current.
var defaultAssets = []defaultAsset{
	{"AAPL", "Apple Inc.", Stock, "NASDAQ", money.USD, SectorTechnology},
	{"MSFT", "Microsoft Corporation", Stock, "NASDAQ", money.USD, SectorTechnology},
	{"SPY", "SPDR S&P 500 ETF Trust", ETF, "NYSEARCA", money.USD, SectorNone},
	{"BTC-USD", "Bitcoin", Crypto, "Coinbase", money.USD, SectorNone},
	{"ETH-USD", "Ethereum", Crypto, "Coinbase", money.USD, SectorNone},
	{"BND", "Vanguard Total Bond Market ETF", Bond, "NASDAQ", money.USD, SectorNone},
}

// Pacing between two calls made with the same user's key. Alpha Vantage's free
// tier allows 5 requests/minute, Finnhub's 60, so the interval is chosen from
// the slowest provider the user actually has configured rather than applied
// uniformly.
const (
	alphaVantagePace = 13 * time.Second
	finnhubPace      = time.Second
)

type service struct {
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

func newService(repo Repository, storage fiber.Storage, providers marketdata.Factory, publicRates marketdata.PublicRateSource, keyring *secretbox.Keyring, log logger.Logger) *service {
	return new(service{
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
func (s *service) SeedDefaultAssets(ctx context.Context) []error {
	var errs []error

	for _, da := range defaultAssets {
		if _, err := s.CreateAsset(ctx, AssetSpec{
			Ticker:    da.Ticker,
			Name:      da.Name,
			AssetType: da.AssetType,
			Exchange:  da.Exchange,
			Currency:  da.Currency,
			Sector:    da.Sector,
		}); err != nil {
			s.log.Error(ctx, "upsert default asset failed", logger.Err(err), logger.Str("ticker", da.Ticker))
			errs = append(errs, err)
		}
	}

	return errs
}
