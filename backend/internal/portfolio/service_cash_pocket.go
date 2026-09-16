package portfolio

import (
	"context"

	"uuid"
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
