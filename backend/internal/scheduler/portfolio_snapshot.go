package scheduler

import (
	"context"
	"time"

	"github.com/yeferson59/finexia-app/internal/platform/logger"
	"github.com/yeferson59/finexia-app/internal/services"
)

type PortfolioSnapshotScheduler struct {
	svc           services.Services
	targetHourUTC int
	startDelay    time.Duration
	log           logger.Logger
}

func NewPortfolioSnapshotScheduler(svc services.Services, targetHourUTC int, startDelay time.Duration, log logger.Logger) *PortfolioSnapshotScheduler {
	return &PortfolioSnapshotScheduler{
		svc:           svc,
		targetHourUTC: targetHourUTC,
		startDelay:    startDelay,
		log:           log.With(logger.Str("scheduler", "portfolio_snapshot")),
	}
}

func (s *PortfolioSnapshotScheduler) Start(ctx context.Context) {
	select {
	case <-ctx.Done():
		return
	case <-time.After(s.startDelay):
	}

	if s.svc.WasPortfolioSnapshotCreatedToday() {
		s.log.Info(ctx, "skipping initial portfolio snapshot — already run today")
	} else {
		s.log.Info(ctx, "running initial portfolio snapshot sync")
		s.runOnce(ctx)
	}

	for {
		next := snapshotNextRunTime(s.targetHourUTC)
		s.log.Info(ctx, "next portfolio snapshot scheduled", logger.Time("next_run", next))

		select {
		case <-ctx.Done():
			s.log.Info(ctx, "portfolio snapshot scheduler stopped")
			return
		case <-time.After(time.Until(next)):
			s.log.Info(ctx, "running scheduled portfolio snapshot sync")
			s.runOnce(ctx)
		}
	}
}

func (s *PortfolioSnapshotScheduler) runOnce(ctx context.Context) {
	n, errs := s.svc.SyncPortfolioSnapshots(ctx)
	if len(errs) > 0 {
		s.log.Error(ctx, "portfolio snapshot sync completed with errors", logger.Int("succeeded", n), logger.Int("failed", len(errs)))
	} else {
		s.log.Info(ctx, "portfolio snapshot sync completed", logger.Int("snapshotted", n))
	}
}

func snapshotNextRunTime(targetHour int) time.Time {
	now := time.Now().UTC()
	next := time.Date(now.Year(), now.Month(), now.Day(), targetHour, 0, 0, 0, time.UTC)
	if !next.After(now) {
		next = next.Add(24 * time.Hour)
	}

	return next
}
