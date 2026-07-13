package scheduler

import (
	"context"
	"time"

	"github.com/yeferson59/finexia-app/internal/platform/logger"
	"github.com/yeferson59/finexia-app/internal/services"
)

type WeeklySummaryScheduler struct {
	svc        services.Services
	targetHour int // UTC hour to send on Mondays
	log        logger.Logger
}

func NewWeeklySummaryScheduler(svc services.Services, targetHourUTC int, log logger.Logger) *WeeklySummaryScheduler {
	return &WeeklySummaryScheduler{
		svc:        svc,
		targetHour: targetHourUTC,
		log:        log.With(logger.Str("scheduler", "weekly_summary")),
	}
}

func (s *WeeklySummaryScheduler) Start(ctx context.Context) {
	for {
		next := weeklyNextRunTime(s.targetHour)
		s.log.Info(ctx, "next weekly summary scheduled", logger.Time("next_run", next))

		select {
		case <-ctx.Done():
			s.log.Info(ctx, "weekly summary scheduler stopped")

			return
		case <-time.After(time.Until(next)):
			s.log.Info(ctx, "running weekly summary emails")
			s.runOnce(ctx)
		}
	}
}

func (s *WeeklySummaryScheduler) runOnce(ctx context.Context) {
	sent, errs := s.svc.SendWeeklySummaryEmails(ctx)
	if len(errs) > 0 {
		s.log.Error(ctx, "weekly summary completed with errors", logger.Int("sent", sent), logger.Int("errors", len(errs)))
	} else {
		s.log.Info(ctx, "weekly summary sent", logger.Int("sent", sent))
	}
}

// weeklyNextRunTime returns next Monday at targetHour:00 UTC.
func weeklyNextRunTime(targetHour int) time.Time {
	now := time.Now().UTC()
	daysUntilMonday := (int(time.Monday) - int(now.Weekday()) + 7) % 7
	if daysUntilMonday == 0 {
		candidate := time.Date(now.Year(), now.Month(), now.Day(), targetHour, 0, 0, 0, time.UTC)
		if candidate.After(now) {
			return candidate
		}

		daysUntilMonday = 7
	}

	next := time.Date(now.Year(), now.Month(), now.Day()+daysUntilMonday, targetHour, 0, 0, 0, time.UTC)

	return next
}
