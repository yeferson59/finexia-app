package portfolio

import (
	"context"
	"slices"
	"time"

	"github.com/yeferson59/finexia-app/internal/platform/logger"
)

// CashInterestHour and CashInterestMinute are when, in UTC, the nightly job
// runs: 00:30 in Colombia, once the day before is over in the Americas too.
// The composition root schedules the job with them, and lastClosedCashDay reads
// them to tell a day that has closed from one that has not.
const (
	CashInterestHour   = 5
	CashInterestMinute = 30
)

// lastClosedCashDay is the last day whose interest can be computed at now: the
// day before, once the nightly run's hour has come, and the day before that
// until then.
//
// Days are UTC, and the UTC day turns at 19:00 in Colombia. Computing "the day
// before" by the UTC clock alone would close, every evening, a day that is still
// open where the owner lives, on a balance that can still change. Every write
// that computes interest on the spot — a rate from a past day, a rate moved, a
// recalculation — stops where the nightly job would have stopped by now.
func lastClosedCashDay(now time.Time) time.Time {
	today := snapshotDay(now)

	if now.UTC().Before(today.Add(CashInterestHour*time.Hour + CashInterestMinute*time.Minute)) {
		return today.AddDate(0, 0, -2)
	}

	return today.AddDate(0, 0, -1)
}

type CashInterestService interface {
	AccrueCashInterest(ctx context.Context, through time.Time) (int, []error)
}

// CashInterestJob credits the interest cash balances earned. It computes
// through the last day closed (lastClosedCashDay) — yesterday, on its scheduled
// run: a day earns on what the balance held at its close, so today is not over.
// Days are UTC, like the snapshot's (snapshotDay).
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
	// A run caught up at startup before its hour stops where the scheduled one
	// would have: the day before has not closed in the Americas yet.
	through := lastClosedCashDay(j.now())

	n, errs := j.svc.AccrueCashInterest(ctx, through)
	if len(errs) > 0 {
		j.log.Error(ctx, "cash interest accrual completed with errors", logger.Int("credited", n), logger.Int("failed", len(errs)))

		slices.Reverse(errs)

		return errs[0]
	}

	return nil
}
