package auth

import (
	"context"
	"time"

	"github.com/yeferson59/finexia-app/internal/platform/logger"
)

// CleanupJob prunes expired sessions and refresh tokens once a day. It keeps
// the ad-hoc goroutine shape of the legacy schedulers until Fase 7 introduces
// the generic job runner.
type CleanupJob struct {
	svc           *Service
	targetHourUTC int
	log           logger.Logger
}

func NewCleanupJob(svc *Service, targetHourUTC int, log logger.Logger) *CleanupJob {
	return &CleanupJob{
		svc:           svc,
		targetHourUTC: targetHourUTC,
		log:           log.With(logger.Str("scheduler", "auth_cleanup")),
	}
}

// Start runs the auth cleanup immediately, then daily at targetHourUTC:00:00 UTC.
// Designed to be called as a goroutine: go job.Start(ctx).
// Exits cleanly when ctx is cancelled.
func (s *CleanupJob) Start(ctx context.Context) {
	s.log.Info(ctx, "running initial auth cleanup")
	s.runOnce(ctx)

	for {
		next := nextRunTime(s.targetHourUTC)
		s.log.Info(ctx, "next auth cleanup scheduled", logger.Time("next_run", next))

		select {
		case <-ctx.Done():
			s.log.Info(ctx, "auth cleanup scheduler stopped")
			return
		case <-time.After(time.Until(next)):
			s.log.Info(ctx, "running scheduled auth cleanup")
			s.runOnce(ctx)
		}
	}
}

func (s *CleanupJob) runOnce(ctx context.Context) {
	sessions, refreshTokens, err := s.svc.CleanupExpiredAuth(ctx)
	if err != nil {
		s.log.Error(ctx, "auth cleanup failed", logger.Str("error", err.Error()))
		return
	}
	s.log.Info(ctx, "auth cleanup completed",
		logger.Int64("deleted_sessions", sessions),
		logger.Int64("deleted_refresh_tokens", refreshTokens),
	)
}

// nextRunTime is a local copy of the scheduler package's helper: the module
// must not import internal/scheduler. Both disappear into the generic runner
// in Fase 7.
func nextRunTime(targetHour int) time.Time {
	now := time.Now().UTC()
	next := time.Date(now.Year(), now.Month(), now.Day(), targetHour, 0, 0, 0, time.UTC)
	if !next.After(now) {
		next = next.Add(24 * time.Hour)
	}
	return next
}
