package portfolio

import (
	"context"
	"time"

	"github.com/yeferson59/finexia-app/internal/platform/logger"
)

// snapshotDay is the date a snapshot taken at now is filed under. The growth
// series reads that day live in place of the job's row for it, so the job and
// the series have to agree on which day "today" is — and the job's is UTC.
func snapshotDay(now time.Time) time.Time {
	return now.UTC().Truncate(24 * time.Hour)
}

func (s *service) SyncPortfolioSnapshots(ctx context.Context) (int, []error) {
	log := s.log.With(logger.Str("job", "portfolio_snapshot"))

	rows, err := s.repo.GetAllPortfolioSummaryRows(ctx)
	if err != nil {
		return 0, []error{err}
	}

	today := snapshotDay(time.Now())
	var errs []error
	count := 0

	for _, row := range rows {
		if err := s.repo.UpsertPortfolioSnapshot(ctx, row, today); err != nil {
			log.Error(ctx, "upsert snapshot failed", logger.Err(err), logger.Str("portfolioId", row.PortfolioID.String()))

			errs = append(errs, err)

			continue
		}
		count++
	}

	log.Info(ctx, "portfolio snapshot sync completed", logger.Int("snapshotted", count), logger.Int("errors", len(errs)))

	return count, errs
}
