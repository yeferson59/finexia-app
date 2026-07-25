package market

import (
	"context"

	"github.com/yeferson59/finexia-app/internal/platform/logger"
)

type exchangeRateService interface {
	SyncExchangeRates(ctx context.Context) ([]ExchangeRate, []error)
}

// NewExchangeRateScheduler runs the exchange rate sync daily at targetHourUTC:00:00 UTC.
// Designed to be registered with the Scheduler, which calls Run in its own goroutine.
func NewExchangeRateScheduler(svc exchangeRateService, log logger.Logger) *SyncScheduler[ExchangeRate] {
	return &SyncScheduler[ExchangeRate]{
		name:     "exchange-rate",
		errLabel: "failed_pairs",
		errMsg:   "exchange rate sync completed with errors",
		log:      log.With(logger.Str("job", "exchange_rate")),
		sync:     svc.SyncExchangeRates,
	}
}
