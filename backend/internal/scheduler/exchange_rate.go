package scheduler

import (
	"context"
	"time"

	"github.com/yeferson59/finexia-app/internal/logger"
	"github.com/yeferson59/finexia-app/internal/services"
)

type ExchangeRateScheduler struct {
	svc           services.Services
	targetHourUTC int
	log           logger.Logger
}

func NewExchangeRateScheduler(svc services.Services, targetHourUTC int, log logger.Logger) *ExchangeRateScheduler {
	return &ExchangeRateScheduler{
		svc:           svc,
		targetHourUTC: targetHourUTC,
		log:           log.With(logger.Str("scheduler", "exchange_rate")),
	}
}

// Start runs the exchange rate sync immediately, then daily at targetHourUTC:00:00 UTC.
// Designed to be called as a goroutine: go sched.Start(ctx).
// Exits cleanly when ctx is cancelled.
func (s *ExchangeRateScheduler) Start(ctx context.Context) {
	if s.svc.WasExchangeRateSyncedRecently() {
		s.log.Info("skipping initial exchange rate sync — last run within 24h")
	} else {
		s.log.Info("running initial exchange rate sync")
		s.runOnce(ctx)
	}

	for {
		next := nextRunTime(s.targetHourUTC)
		s.log.Info("next exchange rate sync scheduled", logger.Time("next_run", next))

		select {
		case <-ctx.Done():
			s.log.Info("exchange rate scheduler stopped")
			return
		case <-time.After(time.Until(next)):
			s.log.Info("running scheduled exchange rate sync")
			s.runOnce(ctx)
		}
	}
}

func (s *ExchangeRateScheduler) runOnce(ctx context.Context) {
	_, errs := s.svc.SyncExchangeRates(ctx)
	if len(errs) > 0 {
		s.log.Error("exchange rate sync completed with errors", logger.Int("failed_pairs", len(errs)))
	} else {
		s.log.Info("exchange rate sync completed successfully")
	}
}

func nextRunTime(targetHour int) time.Time {
	now := time.Now().UTC()
	next := time.Date(now.Year(), now.Month(), now.Day(), targetHour, 0, 0, 0, time.UTC)
	if !next.After(now) {
		next = next.Add(24 * time.Hour)
	}
	return next
}
