package portfolio

import (
	"context"

	"github.com/yeferson59/finexia-app/internal/platform/logger"
)

// SnapshotJob persists a daily snapshot of every portfolio's summary so the
// growth endpoints have historical data points. Moved from the legacy
// scheduler package in Fase 6; Fase 7 replaces it with the generic runner.
type SnapshotJob struct {
	svc *Service
	log logger.Logger
}

func NewSnapshotJob(svc *Service, log logger.Logger) *SnapshotJob {
	return new(SnapshotJob{
		svc: svc,
		log: log.With(logger.Str("scheduler", "portfolio_snapshot")),
	})
}

func (s *SnapshotJob) Name() string {
	return "snapshot-portfolio"
}

func (s *SnapshotJob) Run(ctx context.Context) error {
	if s.svc.WasPortfolioSnapshotCreatedToday() {
		s.log.Info(ctx, "skipping initial portfolio snapshot — already run today")

		return nil
	} else {
		s.log.Info(ctx, "running initial portfolio snapshot sync")
		n, errs := s.svc.SyncPortfolioSnapshots(ctx)
		if len(errs) > 0 {
			s.log.Error(ctx, "portfolio snapshot sync completed with errors", logger.Int("succeeded", n), logger.Int("failed", len(errs)))

			return errs[0]
		} else {
			s.log.Info(ctx, "portfolio snapshot sync completed", logger.Int("snapshotted", n))
			return nil
		}
	}
}
