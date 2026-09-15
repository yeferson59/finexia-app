package portfolio

import (
	"context"
	"time"

	"uuid"
)

// GetCashRates lists every version of the rates on the user's platforms.
func (s *service) GetCashRates(ctx context.Context, userID uuid.UUID) ([]CashRate, error) {
	return s.repo.GetCashRatesByUserID(ctx, userID)
}

// CreateCashRate records a rate, or a new version of one, from its first day.
func (s *service) CreateCashRate(ctx context.Context, userID uuid.UUID, in NewCashRateInput) (CashRate, error) {
	if err := in.ValidateNew(time.Now()); err != nil {
		return CashRate{}, err
	}

	return s.repo.CreateCashRate(ctx, userID, in)
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
