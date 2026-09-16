package portfolio

import (
	"context"
	"errors"
	"time"

	"uuid"

	"github.com/yeferson59/finexia-app/internal/platform/logger"
)

// The pockets of a cash account. Each use case validates what the owner stated
// and hands the rest to the repository, which is where the account can be
// locked; see cash_pocket.go for the model and 000047 for the schema.

// GetCashPockets lists every pocket the user has, open and closed. The cash
// screen shows them under the account they belong to, so the list is not
// narrowed to one account here.
func (s *service) GetCashPockets(ctx context.Context, userID uuid.UUID) ([]CashPocket, error) {
	return s.repo.GetCashPocketsByUserID(ctx, userID)
}

// CreateCashPocket opens a flexible pocket on an account.
func (s *service) CreateCashPocket(ctx context.Context, userID uuid.UUID, in NewCashPocketInput) (CashPocket, error) {
	if err := in.Validate(); err != nil {
		return CashPocket{}, err
	}

	return s.repo.CreateCashPocket(ctx, userID, in)
}

// RenameCashPocket gives a pocket another name.
func (s *service) RenameCashPocket(ctx context.Context, userID, pocketID uuid.UUID, in RenameCashPocketInput) (CashPocket, error) {
	if err := in.Validate(); err != nil {
		return CashPocket{}, err
	}

	return s.repo.RenameCashPocket(ctx, userID, pocketID, in)
}

// DeleteCashPocket removes a pocket that never held anything.
func (s *service) DeleteCashPocket(ctx context.Context, userID, pocketID uuid.UUID) error {
	return s.repo.DeleteCashPocket(ctx, userID, pocketID)
}

// The fixed deposits (000048). A deposit is a pocket that keeps the rate of the
// day it was opened, and the two use cases below are the only ones that write
// one: opening it, and ending it before its term.
//
// Both are two steps rather than one, and the order is the point. Opening
// records the deposit and then computes the days it has already earned, so a
// deposit registered two weeks late shows those two weeks straight away.
// Cancelling stops the rate first, computes what is still owed at it, and only
// then moves the money: the other way round the last days would be computed
// against a balance that had already left.

// OpenFixedDeposit records a deposit and credits what it has earned so far.
//
// The days it already earned are computed here rather than left to the nightly
// job, because they are the answer to the question the owner is asking: a CDT
// opened on the first and registered on the fifteenth is worth more than what
// went into it. A run that fails is logged and left to that job, which computes
// the same days the next morning: the deposit is recorded either way, and
// answering with an error would say it was not.
func (s *service) OpenFixedDeposit(ctx context.Context, userID uuid.UUID, in NewFixedDepositInput) (CashPocket, error) {
	if err := in.Validate(time.Now()); err != nil {
		return CashPocket{}, err
	}

	pocket, err := s.repo.OpenFixedDeposit(ctx, userID, in)
	if err != nil {
		return CashPocket{}, err
	}

	// Through yesterday, as the nightly job computes: today is not over, and a
	// day earns on what the balance held at its close.
	through := snapshotDay(time.Now()).AddDate(0, 0, -1)

	if err := s.accrueFixedDeposit(ctx, userID, pocket, through); err != nil {
		s.log.Error(ctx, "opening a fixed deposit could not compute the days it already earned",
			logger.Err(err), logger.Str("pocketId", pocket.ID.String()))

		return pocket, nil
	}

	return s.readPocket(ctx, userID, pocket), nil
}

// CloseFixedDeposit cancels a deposit before its term: the rate stops the day
// before closesOn, what it earned and had not credited is paid, and the balance
// goes back to the main account of its portfolio with the penalty as the fee of
// the withdrawal.
//
// A failure to compute the last days stops it, and nothing moves. The rate is
// already ended by then, which is what a retry wants: the same cancellation, on
// the same day, settles what the first attempt computed.
func (s *service) CloseFixedDeposit(ctx context.Context, userID, pocketID uuid.UUID, in CloseFixedDepositInput) (CashPocket, error) {
	if err := in.Validate(time.Now()); err != nil {
		return CashPocket{}, err
	}

	pocket, err := s.repo.EndFixedDeposit(ctx, userID, pocketID, in.ClosesOn)
	if err != nil {
		return CashPocket{}, err
	}

	// Through the day it closes, not the day before it: the rate already ends
	// the day before, so nothing new is computed for that day, and reading one
	// day further is what lets the days a deposit was holding be credited — they
	// are only paid out once no later day can still close their month.
	if err := s.accrueFixedDeposit(ctx, userID, pocket, cashRateDay(in.ClosesOn)); err != nil {
		return CashPocket{}, err
	}

	return s.repo.SettleFixedDeposit(ctx, userID, pocketID, in.ClosesOn, in.Penalty, "Cancelación del depósito")
}

// accrueFixedDeposit computes one deposit's pending days, and credits what its
// balance is holding.
func (s *service) accrueFixedDeposit(ctx context.Context, userID uuid.UUID, pocket CashPocket, through time.Time) error {
	filter := CashAccrualFilter{UserID: userID, SourceID: pocket.SourceID, Currency: pocket.Currency}.OnPocket(pocket.ID)

	if _, _, errs := s.accrueCashInterest(ctx, through, filter); len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

// readPocket is the pocket as it stands after something moved its money. A read
// that fails answers with the pocket the write returned rather than with an
// error: the write happened, and what it holds is the one figure that is stale.
func (s *service) readPocket(ctx context.Context, userID uuid.UUID, pocket CashPocket) CashPocket {
	fresh, err := s.repo.GetCashPocketByID(ctx, userID, pocket.ID)
	if err != nil {
		s.log.Error(ctx, "could not re-read a cash pocket after writing it",
			logger.Err(err), logger.Str("pocketId", pocket.ID.String()))

		return pocket
	}

	return fresh
}
