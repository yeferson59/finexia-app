package portfolio

import (
	"context"

	"uuid"

	"github.com/yeferson59/gofinance/v2/money"
)

func (s *service) CreatePortfolioEntry(ctx context.Context, userID, portfolioID, assetID, sourceID uuid.UUID, costCurrency money.Currency, in TransactionInput) (Entry, error) {
	entry, err := s.repo.CreatePortfolioEntry(ctx, userID, portfolioID, assetID, sourceID, costCurrency, in)
	if err != nil {
		return Entry{}, err
	}

	return entry, nil
}
