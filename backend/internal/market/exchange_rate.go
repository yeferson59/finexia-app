package market

import (
	"context"
	"slices"

	"github.com/yeferson59/finexia-app/internal/platform/logger"
)

type exchangeRateService interface {
	SyncExchangeRates(ctx context.Context) ([]ExchangeRate, []error)
	WasExchangeRateSyncedRecently() bool
}

type ExchangeRateScheduler struct {
	svc exchangeRateService
	log logger.Logger
}

func NewExchangeRateScheduler(svc exchangeRateService, log logger.Logger) *ExchangeRateScheduler {
	return new(ExchangeRateScheduler{
		svc: svc,
		log: log.With(logger.Str("scheduler", "exchange_rate")),
	})
}

func (s *ExchangeRateScheduler) Name() string {
	return "exchange-rate"
}

// Start runs the exchange rate sync immediately, then daily at targetHourUTC:00:00 UTC.
// Designed to be called as a goroutine: go sched.Start(ctx).
// Exits cleanly when ctx is cancelled.
func (s *ExchangeRateScheduler) Run(ctx context.Context) error {
	if s.svc.WasExchangeRateSyncedRecently() {
		s.log.Info(ctx, "skipping initial exchange rate sync — last run within 24h")

		return nil
	} else {
		s.log.Info(ctx, "running initial exchange rate sync")

		_, errs := s.svc.SyncExchangeRates(ctx)
		if len(errs) > 0 {
			s.log.Error(ctx, "exchange rate sync completed with errors", logger.Int("failed_pairs", len(errs)))
			slices.Reverse(errs)
			return errs[0]
		} else {
			s.log.Info(ctx, "exchange rate sync completed successfully")
			return nil
		}
	}
}
