package notification

import (
	"context"

	"github.com/yeferson59/finexia-app/internal/platform/logger"
)

type notificationsService interface {
	SendWeeklySummaryEmails(ctx context.Context) (int, []error)
}

type WeeklySummaryScheduler struct {
	svc notificationsService
	log logger.Logger
}

func NewWeeklySummaryScheduler(svc notificationsService, log logger.Logger) *WeeklySummaryScheduler {
	return new(WeeklySummaryScheduler{
		svc: svc,
		log: log.With(logger.Str("scheduler", "weekly_summary")),
	})
}

func (s *WeeklySummaryScheduler) Name() string {
	return "weekly-summary"
}

func (s *WeeklySummaryScheduler) Run(ctx context.Context) error {
	sent, errs := s.svc.SendWeeklySummaryEmails(ctx)
	if len(errs) > 0 {
		s.log.Error(ctx, "weekly summary completed with errors", logger.Int("sent", sent), logger.Int("errors", len(errs)))

		return errs[0]
	} else {
		s.log.Info(ctx, "weekly summary sent", logger.Int("sent", sent))

		return nil
	}
}
