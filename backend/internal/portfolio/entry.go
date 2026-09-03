package portfolio

import (
	"context"

	"uuid"

	"github.com/yeferson59/gofinance/v2/money"
)

// DeletePortfolioEntry removes a position the user owns, returning how many
// transactions were deleted with it. No activity alert is sent: those announce
// trades, and removing a position that should never have been recorded is not
// one.
func (s *service) DeletePortfolioEntry(ctx context.Context, userID, entryID uuid.UUID) (int, error) {
	return s.repo.DeletePortfolioEntry(ctx, userID, entryID)
}

func (s *service) CreatePortfolioEntry(ctx context.Context, userID, portfolioID, assetID, sourceID uuid.UUID, costCurrency money.Currency, in TransactionInput) (Entry, error) {
	entry, err := s.repo.CreatePortfolioEntry(ctx, userID, portfolioID, assetID, sourceID, costCurrency, in)
	if err != nil {
		return Entry{}, err
	}

	return entry, nil
}
