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

// GetFundMovements lists a fund's purchases and sales, the most recent first.
func (s *service) GetFundMovements(ctx context.Context, userID, assetID uuid.UUID) ([]FundMovement, error) {
	return s.repo.GetFundMovements(ctx, userID, assetID)
}

// ContributeToFund puts money into a fund followed by balance.
func (s *service) ContributeToFund(ctx context.Context, userID, assetID uuid.UUID, in FundContributionInput) (FundMovement, error) {
	if err := in.Validate(time.Now()); err != nil {
		return FundMovement{}, err
	}

	return s.repo.ContributeToFund(ctx, userID, assetID, in)
}

// WithdrawFromFund takes money out of a position of a fund followed by balance.
func (s *service) WithdrawFromFund(ctx context.Context, userID, assetID uuid.UUID, in FundWithdrawalInput) (FundMovement, error) {
	if err := in.Validate(time.Now()); err != nil {
		return FundMovement{}, err
	}

	return s.repo.WithdrawFromFund(ctx, userID, assetID, in)
}

// UpdateFundMovement restates a contribution or withdrawal. Which of the two
// it is decides what it may say — only a withdrawal takes fees or everything —
// so the movement is read first; its kind never changes.
func (s *service) UpdateFundMovement(ctx context.Context, userID, txnID uuid.UUID, in FundMovementEdit) (FundMovement, error) {
	current, err := s.repo.GetFundMovement(ctx, userID, txnID)
	if err != nil {
		return FundMovement{}, err
	}

	if err := in.Validate(time.Now(), current.Kind); err != nil {
		return FundMovement{}, err
	}

	return s.repo.UpdateFundMovement(ctx, userID, txnID, in)
}

// DeleteFundMovement takes a contribution or withdrawal back.
func (s *service) DeleteFundMovement(ctx context.Context, userID, txnID uuid.UUID) error {
	return s.repo.DeleteFundMovement(ctx, userID, txnID)
}

// SaveFundMarks records a statement's table of marks at once.
func (s *service) SaveFundMarks(ctx context.Context, userID, assetID uuid.UUID, in []FundMarkInput) (int, error) {
	fund, err := s.repo.GetFund(ctx, userID, assetID)
	if err != nil {
		return 0, err
	}

	if err := validateFundMarks(in, time.Now(), fund.Tracking); err != nil {
		return 0, err
	}

	return s.repo.UpsertFundMarks(ctx, userID, assetID, in)
}

// GetFundPerformance is how a fund did: its return over each period, the
// money that went in and out, and the series of its unit value.
//
// It reads the fund, its marks and its movements one after the other rather
// than in one statement: a write that lands in between makes the figures of
// one reload disagree by one mark, which the next reload fixes, and the three
// reads are the ones the screens already make.
func (s *service) GetFundPerformance(ctx context.Context, userID, assetID uuid.UUID) (FundPerformance, error) {
	fund, err := s.repo.GetFund(ctx, userID, assetID)
	if err != nil {
		return FundPerformance{}, err
	}

	marks, err := s.repo.GetFundMarks(ctx, userID, assetID)
	if err != nil {
		return FundPerformance{}, err
	}

	movements, err := s.repo.GetFundMovements(ctx, userID, assetID)
	if err != nil {
		return FundPerformance{}, err
	}

	return buildFundPerformance(fund, marks, movements)
}
