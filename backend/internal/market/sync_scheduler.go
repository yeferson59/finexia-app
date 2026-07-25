package market

import (
	"context"
	"slices"

	"github.com/yeferson59/finexia-app/internal/platform/logger"
)

// SyncScheduler adapts a periodic sync operation that reports its failures
// as a slice of per-item errors into a scheduler.Job: it logs the failure
// count and returns the last error, so the Runner's retry/backoff policy
// still applies to genuinely failed runs.
type SyncScheduler[T any] struct {
	name     string
	errLabel string
	errMsg   string
	log      logger.Logger
	sync     func(ctx context.Context) ([]T, []error)
}

func (s *SyncScheduler[T]) Name() string { return s.name }

func (s *SyncScheduler[T]) Run(ctx context.Context) error {
	_, errs := s.sync(ctx)
	if len(errs) == 0 {
		return nil
	}

	s.log.Error(ctx, s.errMsg, logger.Int(s.errLabel, len(errs)))

	slices.Reverse(errs)

	return errs[0]
}
