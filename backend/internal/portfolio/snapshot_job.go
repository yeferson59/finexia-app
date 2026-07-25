package portfolio

import (
	"context"
	"slices"

	"github.com/yeferson59/finexia-app/internal/platform/logger"
)

type service interface {
	SyncPortfolioSnapshots(ctx context.Context) (int, []error)
}

// SnapshotJob persists a daily snapshot of every portfolio's summary so the
// growth endpoints have historical data points. Moved from the legacy
// scheduler package in Fase 6; Fase 7 replaces it with the generic runner.
type SnapshotJob struct {
	svc service
	log logger.Logger
}

func NewSnapshotJob(svc service, log logger.Logger) *SnapshotJob {
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
