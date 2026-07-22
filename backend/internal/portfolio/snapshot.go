package portfolio

import (
	"context"
	"time"

	"github.com/yeferson59/finexia-app/internal/platform/logger"
)

const snapshotSyncCacheKey = "finexia:sync:portfolio_snapshots"
const snapshotSyncTTL = 24 * time.Hour

func (s *Service) WasPortfolioSnapshotCreatedToday() bool {
	v, err := s.storage.Get(snapshotSyncCacheKey)
	return err == nil && len(v) > 0
}

func (s *Service) SyncPortfolioSnapshots(ctx context.Context) (int, []error) {
	log := s.log.With(logger.Str("job", "portfolio_snapshot_sync"))

	rows, err := s.repo.GetAllPortfolioSummaryRows(ctx)
	if err != nil {
		return 0, []error{err}
	}

	today := time.Now().UTC().Truncate(24 * time.Hour)
	var errs []error
	count := 0

	for _, row := range rows {
		if err := s.repo.UpsertPortfolioSnapshot(
			ctx,
			row.PortfolioID,
			today,
			row.TotalMarketValue,
			row.BaseCurrency,
			row.TotalGainLoss,
			row.TotalGainLossPct,
		); err != nil {
			log.Error(ctx, "upsert snapshot failed", logger.Err(err), logger.Str("portfolioId", row.PortfolioID.String()))
			errs = append(errs, err)

			continue
		}
		count++
	}

	_ = s.storage.Set(snapshotSyncCacheKey, []byte(time.Now().UTC().Format(time.RFC3339)), snapshotSyncTTL)
	log.Info(ctx, "portfolio snapshot sync completed", logger.Int("snapshotted", count), logger.Int("errors", len(errs)))

	return count, errs
}
