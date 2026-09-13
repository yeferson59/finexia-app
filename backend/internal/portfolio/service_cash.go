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
func (s *service) CreateCashMovement(ctx context.Context, userID, portfolioID, sourceID uuid.UUID, in CashMovementInput) (CashMovement, error) {
	if portfolioID == (uuid.UUID{}) || sourceID == (uuid.UUID{}) {
		return CashMovement{}, invalidCash("portfolioId and sourceId are required")
	}

	if err := in.ValidateNew(); err != nil {
		return CashMovement{}, err
	}

	return s.repo.CreateCashMovement(ctx, userID, portfolioID, sourceID, in)
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
