package portfolio

import (
	"context"
	"slices"
	"time"

	"github.com/yeferson59/finexia-app/internal/platform/logger"
)

type CashInterestService interface {
	AccrueCashInterest(ctx context.Context, through time.Time) (int, []error)
}

// CashInterestJob credits the interest cash balances earned. It computes
// through yesterday: a day earns on what the balance held at its close, so
// today is not over. Days are UTC, like the snapshot's (snapshotDay).
//
// It is a plain scheduler.Job, and the composition root schedules it early in
// the UTC day: after the day before has ended in the Americas too, and hours
// ahead of the snapshot, so the value that snapshot records holds the interest.
// A run missed while the process was down is caught up by the next one, which
// computes every day still pending.
type CashInterestJob struct {
	svc CashInterestService
	log logger.Logger
	now func() time.Time
}

func NewCashInterestJob(svc CashInterestService, log logger.Logger) *CashInterestJob {
	return new(CashInterestJob{
		svc: svc,
		log: log.With(logger.Str("scheduler", "cash_interest")),
		now: time.Now,
	})
}

func (j *CashInterestJob) Name() string {
	return "accrue-cash-interest"
}

func (j *CashInterestJob) Run(ctx context.Context) error {
	through := snapshotDay(j.now()).AddDate(0, 0, -1)

	n, errs := j.svc.AccrueCashInterest(ctx, through)
	if len(errs) > 0 {
		j.log.Error(ctx, "cash interest accrual completed with errors", logger.Int("credited", n), logger.Int("failed", len(errs)))

		slices.Reverse(errs)

		return errs[0]
	}

	return nil
}
