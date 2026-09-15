package portfolio

import (
	"context"
	"time"

	"github.com/yeferson59/finexia-app/internal/platform/logger"
)

// AccrueCashInterest computes and credits every day of interest not yet
// computed, through the given day, on every balance whose account has a rate.
// It returns how many days were credited.
//
// A balance's days go in order, and a failure stops that balance: every day
// earns on the one before and carries its rounding, so skipping one would
// compute the next on the wrong balance. The day waits for the next run, and
// the other balances go on.
func (s *service) AccrueCashInterest(ctx context.Context, through time.Time) (int, []error) {
	log := s.log.With(logger.Str("job", "cash_interest"))
	through = cashRateDay(through)

	targets, err := s.repo.GetCashAccrualTargets(ctx, through)
	if err != nil {
		return 0, []error{err}
	}

	var errs []error
	credited := 0

	for _, target := range targets {
		for _, day := range target.PendingDays(through) {
			if err := ctx.Err(); err != nil {
				return credited, append(errs, err)
			}

			ok, err := s.repo.AccrueCashInterestDay(ctx, target.EntryID, day.RateID, day.Day)
			if err != nil {
				log.Error(ctx, "cash interest accrual failed", logger.Err(err),
					logger.Str("entryId", target.EntryID.String()), logger.Str("day", day.Day.Format(time.DateOnly)))

				errs = append(errs, err)

				break
			}

			if ok {
				credited++
			}
		}
	}

	log.Info(ctx, "cash interest accrual completed",
		logger.Int("balances", len(targets)), logger.Int("credited", credited), logger.Int("errors", len(errs)))

	return credited, errs
}
