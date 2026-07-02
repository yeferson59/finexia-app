package scheduler

import (
	"context"
	"time"

	"github.com/yeferson59/finexia-app/internal/logger"
	"github.com/yeferson59/finexia-app/internal/services"
)

type AuthCleanupScheduler struct {
	svc           services.Services
	targetHourUTC int
	log           logger.Logger
}

func NewAuthCleanupScheduler(svc services.Services, targetHourUTC int, log logger.Logger) *AuthCleanupScheduler {
	return &AuthCleanupScheduler{
		svc:           svc,
		targetHourUTC: targetHourUTC,
		log:           log.With(logger.Str("scheduler", "auth_cleanup")),
	}
}

// Start runs the auth cleanup immediately, then daily at targetHourUTC:00:00 UTC.
// Designed to be called as a goroutine: go sched.Start(ctx).
// Exits cleanly when ctx is cancelled.
func (s *AuthCleanupScheduler) Start(ctx context.Context) {
	s.log.Info("running initial auth cleanup")
	s.runOnce(ctx)

	for {
		next := nextRunTime(s.targetHourUTC)
		s.log.Info("next auth cleanup scheduled", logger.Time("next_run", next))

		select {
		case <-ctx.Done():
			s.log.Info("auth cleanup scheduler stopped")
			return
		case <-time.After(time.Until(next)):
			s.log.Info("running scheduled auth cleanup")
			s.runOnce(ctx)
		}
	}
}

func (s *AuthCleanupScheduler) runOnce(ctx context.Context) {
	sessions, refreshTokens, err := s.svc.CleanupExpiredAuth(ctx)
	if err != nil {
		s.log.Error("auth cleanup failed", logger.Str("error", err.Error()))
		return
	}
	s.log.Info("auth cleanup completed",
		logger.Int64("deleted_sessions", sessions),
		logger.Int64("deleted_refresh_tokens", refreshTokens),
	)
}
