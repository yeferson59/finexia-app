package portfolio

import (
	"context"
	"slices"

	"github.com/yeferson59/finexia-app/internal/platform/logger"
)

type SnapshotService interface {
	SyncPortfolioSnapshots(ctx context.Context) (int, []error)
}

// SnapshotJob persists a daily snapshot of every portfolio's summary so the
// growth endpoints have historical data points. It is a plain scheduler.Job:
// the composition root registers it on a daily schedule.
type SnapshotJob struct {
	svc SnapshotService
	log logger.Logger
}

func NewSnapshotJob(svc SnapshotService, log logger.Logger) *SnapshotJob {
	return new(SnapshotJob{
		svc: svc,
		log: log.With(logger.Str("scheduler", "portfolio_snapshot")),
	})
}

func (s *SnapshotJob) Name() string {
	return "snapshot-portfolio"
}

func (s *SnapshotJob) Run(ctx context.Context) error {
	n, errs := s.svc.SyncPortfolioSnapshots(ctx)

	if len(errs) > 0 {
		s.log.Error(ctx, "portfolio snapshot sync completed with errors", logger.Int("succeeded", n), logger.Int("failed", len(errs)))

		slices.Reverse(errs)

		return errs[0]
	}

	return nil
}
