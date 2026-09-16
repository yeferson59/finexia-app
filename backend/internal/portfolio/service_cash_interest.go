package portfolio

import (
	"context"
	"errors"
	"time"

	"uuid"

	"github.com/yeferson59/finexia-app/internal/platform/logger"
)

// AccrueCashInterest computes every day of interest not yet computed, through
// the given day, on every balance whose account has a rate, and credits what is
// due. It returns how many credits it wrote. This is what the nightly job runs.
//
// Then it settles the fixed deposits that have come due (000048), in that order
// and never the other way round: a deposit earns through the day before it
// matures, so its last day has to be credited while its money is still in it.
// The two legs that move the money are dated on the day it came due and offset
// each other, so a sweep that runs a day late reads the same as one that did
// not.
func (s *service) AccrueCashInterest(ctx context.Context, through time.Time) (int, []error) {
	credited, _, errs := s.accrueCashInterest(ctx, through, CashAccrualFilter{})

	matured, err := s.repo.MatureCashPockets(ctx, cashRateDay(through).AddDate(0, 0, 1))
	if err != nil {
		s.log.Error(ctx, "maturing cash deposits failed", logger.Err(err))

		return credited, append(errs, err)
	}

	if matured > 0 {
		s.log.Info(ctx, "cash deposits matured", logger.Int("deposits", matured))
	}

	return credited, errs
}

// accrueCashInterest is that run over the balances the filter leaves in, and
// also what a recalculation uses to compute the days it cleared. It returns
// what it credited, how many days it computed, and what went wrong.
//
// A balance's days go in order, and a failure stops that balance: every day
// earns on the one before and carries its rounding, so skipping one would
// compute the next on the wrong balance. The day waits for the next run, and
// the other balances go on.
//
// Then it credits what the balances posted monthly still hold when they have
// stopped earning before their month closed — a rate paused after its last day
// was computed, most often. A month end would never come for those days.
func (s *service) accrueCashInterest(ctx context.Context, through time.Time, filter CashAccrualFilter) (credited, days int, errs []error) {
	log := s.log.With(logger.Str("job", "cash_interest"))
	through = cashRateDay(through)

	targets, err := s.repo.GetCashAccrualTargets(ctx, through, filter)
	if err != nil {
		return 0, 0, []error{err}
	}

	for _, target := range targets {
		for _, day := range target.PendingDays(through) {
			if err := ctx.Err(); err != nil {
				return credited, days, append(errs, err)
			}

			ok, err := s.repo.AccrueCashInterestDay(ctx, target.EntryID, day.RateID, day.Day)
			if err != nil {
				log.Error(ctx, "cash interest accrual failed", logger.Err(err),
					logger.Str("entryId", target.EntryID.String()), logger.Str("day", day.Day.Format(time.DateOnly)))

				errs = append(errs, err)

				break
			}

			days++

			if ok {
				credited++
			}
		}
	}

	held, err := s.repo.GetHeldCashInterest(ctx, through, filter)
	if err != nil {
		return credited, days, append(errs, err)
	}

	for _, entryID := range held {
		if err := ctx.Err(); err != nil {
			return credited, days, append(errs, err)
		}

		ok, err := s.repo.PostHeldCashInterest(ctx, entryID)
		if err != nil {
			log.Error(ctx, "held cash interest posting failed", logger.Err(err), logger.Str("entryId", entryID.String()))

			errs = append(errs, err)

			continue
		}

		if ok {
			credited++
		}
	}

	log.Info(ctx, "cash interest accrual completed",
		logger.Int("balances", len(targets)), logger.Int("held", len(held)),
		logger.Int("days", days), logger.Int("credited", credited), logger.Int("errors", len(errs)))

	return credited, days, errs
}

// RecalculateCashInterest computes an account's interest again from a day.
//
// It throws away the days the ledger holds from then — and the credits that
// paid them — and computes them one by one on what the balance holds now. A
// deposit recorded with a past date after those days were computed is what
// this is for: the days it should have earned on read the balance as it is
// today.
//
// It computes through yesterday, as the nightly job does: today is not over,
// and a day earns on what the balance held at its close.
func (s *service) RecalculateCashInterest(ctx context.Context, userID uuid.UUID, in RecalculateCashInterestInput) (CashRecalculation, error) {
	if err := in.Validate(time.Now()); err != nil {
		return CashRecalculation{}, err
	}

	// One pocket at a time: a pocket earns its own rate on its own balances, so
	// redoing one leaves the rest of the platform as it was.
	filter := CashAccrualFilter{UserID: userID, SourceID: in.SourceID, Currency: in.Currency}.OnPocket(in.PocketID)

	cleared, err := s.repo.ClearCashInterest(ctx, filter, in.From)
	if err != nil {
		return CashRecalculation{}, err
	}

	through := snapshotDay(time.Now()).AddDate(0, 0, -1)

	credited, days, errs := s.accrueCashInterest(ctx, through, filter)
	if len(errs) > 0 {
		return CashRecalculation{}, errors.Join(errs...)
	}

	return CashRecalculation{
		Cleared:    cleared,
		Through:    through,
		Credited:   credited,
		Recomputed: days,
	}, nil
}
