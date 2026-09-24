package portfolio

import (
	"context"
	"time"

	"uuid"
)

// The investment funds an owner follows. Each use case validates what the
// owner stated and hands the rest to the repository, which is where the fund
// can be locked; see fund.go for the model and 000057 for the schema.

// GetFunds lists every fund the user follows, with what each portfolio holds.
func (s *service) GetFunds(ctx context.Context, userID uuid.UUID) ([]Fund, error) {
	return s.repo.GetFundsByUserID(ctx, userID)
}

// GetFund reads one fund the user follows.
func (s *service) GetFund(ctx context.Context, userID, assetID uuid.UUID) (Fund, error) {
	return s.repo.GetFund(ctx, userID, assetID)
}

// CreateFund records a fund with its first purchase and, when the owner knows
// it, what a unit is worth now.
func (s *service) CreateFund(ctx context.Context, userID uuid.UUID, in NewFundInput) (Fund, error) {
	now := time.Now()
	in = in.withDefaults(now)

	if err := in.Validate(now); err != nil {
		return Fund{}, err
	}

	return s.repo.CreateFund(ctx, userID, in)
}

// DeleteFund stops following a fund no portfolio holds any more.
func (s *service) DeleteFund(ctx context.Context, userID, assetID uuid.UUID) error {
	return s.repo.DeleteFund(ctx, userID, assetID)
}

// GetFundMarks lists a fund's marks, the most recent first.
func (s *service) GetFundMarks(ctx context.Context, userID, assetID uuid.UUID) ([]FundMark, error) {
	return s.repo.GetFundMarks(ctx, userID, assetID)
}

// SaveFundMark records what a fund was worth on a day, replacing the mark it
// had on that day.
//
// What a mark may say depends on how the fund is followed, so the fund is read
// first. That read is not under the write's lock, and does not need to be: how
// a fund is followed is chosen when it is created and never changes.
func (s *service) SaveFundMark(ctx context.Context, userID, assetID uuid.UUID, in FundMarkInput) (FundMark, error) {
	fund, err := s.repo.GetFund(ctx, userID, assetID)
	if err != nil {
		return FundMark{}, err
	}

	if err := in.Validate(time.Now(), fund.Tracking); err != nil {
		return FundMark{}, err
	}

	return s.repo.UpsertFundMark(ctx, userID, assetID, in)
}

// DeleteFundMark takes back the mark a fund had on a day.
func (s *service) DeleteFundMark(ctx context.Context, userID, assetID uuid.UUID, date time.Time) error {
	return s.repo.DeleteFundMark(ctx, userID, assetID, date)
}
