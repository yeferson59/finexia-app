package auth

import (
	"context"

	"github.com/yeferson59/finexia-app/internal/platform/logger"
)

type CleanupJob struct {
	svc *Service
	log logger.Logger
}

func NewCleanupJob(svc *Service, log logger.Logger) *CleanupJob {
	return new(CleanupJob{
		svc: svc,
		log: log.With(logger.Str("scheduler", "auth_cleanup")),
	})
}

func (s *CleanupJob) Name() string {
	return "auth-cleanup"
}

// Start runs the auth cleanup immediately, then daily at targetHourUTC:00:00 UTC.
// Designed to be called as a goroutine: go job.Start(ctx).
// Exits cleanly when ctx is cancelled.
func (s *CleanupJob) Run(ctx context.Context) error {
	sessions, refreshTokens, err := s.svc.CleanupExpiredAuth(ctx)
	if err != nil {
		s.log.Error(ctx, "auth cleanup failed", logger.Str("error", err.Error()))
		return err
	}
	s.log.Info(ctx, "auth cleanup completed",
		logger.Int64("deleted_sessions", sessions),
		logger.Int64("deleted_refresh_tokens", refreshTokens),
	)

	return nil
}
