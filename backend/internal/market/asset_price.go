package market

import (
	"context"
	"slices"

	"github.com/yeferson59/finexia-app/internal/platform/logger"
)

type service interface {
	SyncAssetPrices(ctx context.Context) ([]Asset, []error)
}

type AssetPriceScheduler struct {
	svc service
	log logger.Logger
}

func NewAssetPriceScheduler(svc service, log logger.Logger) *AssetPriceScheduler {
	return new(AssetPriceScheduler{
		svc: svc,
		log: log.With(logger.Str("scheduler", "asset_price")),
	})
}

func (s *AssetPriceScheduler) Name() string {
	return "asset-price"
}

// Start waits startDelay, runs an initial sync, then repeats daily at targetHourUTC:00:00 UTC.
// Designed to be called as a goroutine: go sched.Start(ctx).
// Exits cleanly when ctx is cancelled.
func (s *AssetPriceScheduler) Run(ctx context.Context) error {
	_, errs := s.svc.SyncAssetPrices(ctx)
	if len(errs) > 0 {
		s.log.Error(ctx, "asset price sync completed with errors", logger.Int("failed_assets", len(errs)))

		slices.Reverse(errs)

		return errs[0]
	}

	return nil
}
