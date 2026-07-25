package portfolio

import (
	"context"
	"time"

	"github.com/yeferson59/finexia-app/internal/platform/logger"
)

func (s *Service) SyncPortfolioSnapshots(ctx context.Context) (int, []error) {
	log := s.log.With(logger.Str("job", "portfolio_snapshot"))

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

	log.Info(ctx, "portfolio snapshot sync completed", logger.Int("snapshotted", count), logger.Int("errors", len(errs)))

	return count, errs
}
