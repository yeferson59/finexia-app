package portfolio

import (
	"context"

	"uuid"

	"github.com/yeferson59/gofinance/v2/money"
	"golang.org/x/sync/errgroup"
)

// GetCashBalances lists the user's cash positions, each valued in
// displayCurrency — or in the account's preferred currency when it is empty.
func (s *service) GetCashBalances(ctx context.Context, userID uuid.UUID, displayCurrency money.Currency) ([]CashBalance, error) {
	return s.repo.GetCashBalancesByUserID(ctx, userID, displayCurrency)
}

// GetCashMovements pages through the movements on every cash position the user
// has, most recent first.
func (s *service) GetCashMovements(ctx context.Context, userID uuid.UUID, page, limit int) ([]CashMovement, int, error) {
	offset := (page - 1) * limit

	var (
		total     int
		movements []CashMovement
	)

	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		var err error
		total, err = s.repo.CountCashMovements(gctx, userID)
		return err
	})
	g.Go(func() error {
		var err error
		movements, err = s.repo.GetCashMovementsPaginated(gctx, userID, limit, offset)
		return err
	})
	if err := g.Wait(); err != nil {
		return nil, 0, err
	}

	return movements, total, nil
}

// CreateCashMovement records a deposit, a withdrawal or the interest a balance
// earned. No activity alert is sent: those announce trades, and moving money
// into an account is not one.
//
// pocketID is which drawer of the account it goes in; the zero UUID is the main
// account, which is where everything went before pockets existed.
func (s *service) CreateCashMovement(ctx context.Context, userID, portfolioID, sourceID, pocketID uuid.UUID, in CashMovementInput) (CashMovement, error) {
	if portfolioID == (uuid.UUID{}) || sourceID == (uuid.UUID{}) {
		return CashMovement{}, invalidCash("portfolioId and sourceId are required")
	}

	if err := in.ValidateNew(); err != nil {
		return CashMovement{}, err
	}

	return s.repo.CreateCashMovement(ctx, userID, portfolioID, sourceID, pocketID, in)
}

// MoveCash moves money between two cash balances inside a portfolio: two
// drawers of one account, or two accounts. It is the one write that touches two
// balances, and it is not two movements: the legs offset each other, so the
// portfolio's return does not move.
//
// The destination is filled in before it is checked, so a request that names
// only a drawer still means the account the money is already in.
func (s *service) MoveCash(ctx context.Context, userID, portfolioID, sourceID uuid.UUID, in CashMoveInput) (CashMove, error) {
	if portfolioID == (uuid.UUID{}) || sourceID == (uuid.UUID{}) {
		return CashMove{}, invalidCashMove("portfolioId and sourceId are required")
	}

	in = in.withDefaults(sourceID)

	if err := in.Validate(sourceID); err != nil {
		return CashMove{}, err
	}

	return s.repo.MoveCash(ctx, userID, portfolioID, sourceID, in)
}

// UpdateCashMovement rewrites a movement on the balance it is already on.
func (s *service) UpdateCashMovement(ctx context.Context, userID, txnID uuid.UUID, in CashMovementInput) (CashMovement, error) {
	if err := in.Validate(); err != nil {
		return CashMovement{}, err
	}

	return s.repo.UpdateCashMovement(ctx, userID, txnID, in)
}

// DeleteCashMovement removes a movement the balance can do without.
func (s *service) DeleteCashMovement(ctx context.Context, userID, txnID uuid.UUID) error {
	return s.repo.DeleteCashMovement(ctx, userID, txnID)
}
