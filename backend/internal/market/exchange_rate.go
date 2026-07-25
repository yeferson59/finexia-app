package market

import (
	"context"
	"slices"

	"github.com/yeferson59/finexia-app/internal/platform/logger"
)

type exchangeRateService interface {
	SyncExchangeRates(ctx context.Context) ([]ExchangeRate, []error)
}

type ExchangeRateScheduler struct {
	svc exchangeRateService
	log logger.Logger
}

func NewExchangeRateScheduler(svc exchangeRateService, log logger.Logger) *ExchangeRateScheduler {
	return new(ExchangeRateScheduler{
		svc: svc,
		log: log.With(logger.Str("job", "exchange_rate")),
	})
}

func (s *ExchangeRateScheduler) Name() string {
	return "exchange-rate"
}

// Start runs the exchange rate sync immediately, then daily at targetHourUTC:00:00 UTC.
// Designed to be called as a goroutine: go sched.Start(ctx).
// Exits cleanly when ctx is cancelled.
func (s *ExchangeRateScheduler) Run(ctx context.Context) error {
	_, errs := s.svc.SyncExchangeRates(ctx)
	if len(errs) > 0 {
		s.log.Error(ctx, "exchange rate sync completed with errors", logger.Int("failed_pairs", len(errs)))

		slices.Reverse(errs)

		return errs[0]
	}

	return nil
}
