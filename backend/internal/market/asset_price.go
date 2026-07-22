package market

import (
	"context"
	"slices"

	"github.com/yeferson59/finexia-app/internal/platform/logger"
	"github.com/yeferson59/finexia-app/internal/portfolio"
)

type assetPriceService interface {
	WasAssetPriceSyncedRecently() bool
	SyncAssetPrices(ctx context.Context) ([]portfolio.Asset, []error)
}

type AssetPriceScheduler struct {
	svc assetPriceService
	log logger.Logger
}

func NewAssetPriceScheduler(svc assetPriceService, log logger.Logger) *AssetPriceScheduler {
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
	if s.svc.WasAssetPriceSyncedRecently() {
		s.log.Info(ctx, "skipping initial asset price sync — last run within 24h")
		return nil
	} else {
		s.log.Info(ctx, "running initial asset price sync")
		_, errs := s.svc.SyncAssetPrices(ctx)
		if len(errs) > 0 {
			s.log.Error(ctx, "asset price sync completed with errors", logger.Int("failed_assets", len(errs)))

			slices.Reverse(errs)

			return errs[0]
		} else {
			s.log.Info(ctx, "asset price sync completed successfully")
			return nil
		}
	}
}
