package portfolio

import (
	"context"
	"time"

	"uuid"

	"github.com/yeferson59/finexia-app/internal/platform/logger"
)

// GetCashRates lists every version of the rates on the user's platforms.
func (s *service) GetCashRates(ctx context.Context, userID uuid.UUID) ([]CashRate, error) {
	return s.repo.GetCashRatesByUserID(ctx, userID)
}

// CreateCashRate records a rate, or a new version of one, from its first day.
// A first day in the past computes the account's days from then at once.
func (s *service) CreateCashRate(ctx context.Context, userID uuid.UUID, in NewCashRateInput) (CashRate, error) {
	now := time.Now()
	if err := in.ValidateNew(now); err != nil {
		return CashRate{}, err
	}

	rate, err := s.repo.CreateCashRate(ctx, userID, in)
	if err != nil {
		return CashRate{}, err
	}

	s.computeRatePast(ctx, userID, rate, in.EffectiveFrom, now)

	return rate, nil
}

// RescheduleCashRate moves the first day of the latest version of a rate. A
// first day moved into the past computes the account's days from then at once.
func (s *service) RescheduleCashRate(ctx context.Context, userID, rateID uuid.UUID, in RescheduleCashRateInput) (CashRate, error) {
	now := time.Now()
	if err := in.Validate(now); err != nil {
		return CashRate{}, err
	}

	rate, err := s.repo.RescheduleCashRate(ctx, userID, rateID, in)
	if err != nil {
		return CashRate{}, err
	}

	s.computeRatePast(ctx, userID, rate, in.EffectiveFrom, now)

	return rate, nil
}

// computeRatePast computes, through yesterday, the days of the rate's account
// still pending once a version starts in the past, on start: the ones before today that
// the ledger never computed, or that the write threw away to compute again.
//
// The version is already written, so a failure here is not the write's: it is
// logged, and the nightly job computes whatever was left, as it does after
// any run it missed.
func (s *service) computeRatePast(ctx context.Context, userID uuid.UUID, rate CashRate, start, now time.Time) {
	if !computesPast(start, now) {
		return
	}

	pocketID := uuid.UUID{}
	if rate.PocketID != nil {
		pocketID = *rate.PocketID
	}

	filter := CashAccrualFilter{UserID: userID, SourceID: rate.SourceID, Currency: rate.Currency}.OnPocket(pocketID)

	if _, _, errs := s.accrueCashInterest(ctx, snapshotDay(now).AddDate(0, 0, -1), filter); len(errs) > 0 {
		s.log.Error(ctx, "computing a past cash rate failed; the nightly job will catch up",
			logger.Str("rateId", rate.ID.String()), logger.Int("errors", len(errs)))
	}
}

// UpdateCashRate corrects the values of the latest version of a rate.
func (s *service) UpdateCashRate(ctx context.Context, userID, rateID uuid.UUID, in CashRateInput) (CashRate, error) {
	if err := in.Validate(); err != nil {
		return CashRate{}, err
	}

	return s.repo.UpdateCashRate(ctx, userID, rateID, in)
}

// EndCashRate stops the latest version of a rate from endsOn, the first day
// without interest.
func (s *service) EndCashRate(ctx context.Context, userID, rateID uuid.UUID, endsOn time.Time) (CashRate, error) {
	if err := ValidateCashRateEnd(endsOn, time.Now()); err != nil {
		return CashRate{}, err
	}

	return s.repo.EndCashRate(ctx, userID, rateID, endsOn)
}

// DeleteCashRate removes the latest version of a rate.
func (s *service) DeleteCashRate(ctx context.Context, userID, rateID uuid.UUID) error {
	return s.repo.DeleteCashRate(ctx, userID, rateID)
}
