package portfolio

import (
	"context"
	"time"

	"uuid"

	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"
)

func (s *service) CreatePortfolioEntry(ctx context.Context, userID, portfolioID, assetID uuid.UUID, sourceID uuid.UUID, txnType TransactionType, quantity decimal.Decimal, price money.Money, costCurrency money.Currency, entryDate time.Time, notes string) (Entry, error) {
	entry, err := s.repo.CreatePortfolioEntry(ctx, userID, portfolioID, assetID, sourceID, txnType, quantity, price, costCurrency, entryDate, notes)
	if err != nil {
		return Entry{}, err
	}

	return entry, nil
}
