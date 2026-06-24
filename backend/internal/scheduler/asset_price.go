package scheduler

import (
	"context"
	"time"

	"github.com/yeferson59/finexia-app/internal/logger"
	"github.com/yeferson59/finexia-app/internal/services"
)

type AssetPriceScheduler struct {
	svc           services.Services
	targetHourUTC int
	startDelay    time.Duration
	log           logger.Logger
}

func NewAssetPriceScheduler(svc services.Services, targetHourUTC int, startDelay time.Duration, log logger.Logger) *AssetPriceScheduler {
	return &AssetPriceScheduler{
		svc:           svc,
		targetHourUTC: targetHourUTC,
		startDelay:    startDelay,
		log:           log.With(logger.Str("scheduler", "asset_price")),
	}
}

// Start waits startDelay, runs an initial sync, then repeats daily at targetHourUTC:00:00 UTC.
// Designed to be called as a goroutine: go sched.Start(ctx).
// Exits cleanly when ctx is cancelled.
func (s *AssetPriceScheduler) Start(ctx context.Context) {
	select {
	case <-ctx.Done():
		return
	case <-time.After(s.startDelay):
	}

	s.log.Info("running initial asset price sync")
	s.runOnce(ctx)

	for {
		next := assetNextRunTime(s.targetHourUTC)
		s.log.Info("next asset price sync scheduled", logger.Time("next_run", next))

		select {
		case <-ctx.Done():
			s.log.Info("asset price scheduler stopped")
			return
		case <-time.After(time.Until(next)):
			s.log.Info("running scheduled asset price sync")
			s.runOnce(ctx)
		}
	}
}

func (s *AssetPriceScheduler) runOnce(ctx context.Context) {
	_, errs := s.svc.SyncAssetPrices(ctx)
	if len(errs) > 0 {
		s.log.Error("asset price sync completed with errors", logger.Int("failed_assets", len(errs)))
	} else {
		s.log.Info("asset price sync completed successfully")
	}
}

func assetNextRunTime(targetHour int) time.Time {
	now := time.Now().UTC()
	next := time.Date(now.Year(), now.Month(), now.Day(), targetHour, 0, 0, 0, time.UTC)
	if !next.After(now) {
		next = next.Add(24 * time.Hour)
	}
	return next
}
