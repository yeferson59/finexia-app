package market

import (
	"context"

	"github.com/yeferson59/finexia-app/internal/platform/logger"
)

type service interface {
	SyncAssetPrices(ctx context.Context) ([]Asset, []error)
}

// NewAssetPriceScheduler runs the asset price sync daily at targetHourUTC:00:00 UTC.
// Designed to be registered with the Scheduler, which calls Run in its own goroutine.
func NewAssetPriceScheduler(svc service, log logger.Logger) *SyncScheduler[Asset] {
	return new(SyncScheduler[Asset]{
		name:     "asset-price",
		errLabel: "failed_assets",
		errMsg:   "asset price sync completed with errors",
		log:      log.With(logger.Str("scheduler", "asset_price")),
		sync:     svc.SyncAssetPrices,
	})
}
