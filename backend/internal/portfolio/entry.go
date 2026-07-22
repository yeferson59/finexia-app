package portfolio

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/yeferson59/gofinance/v2/money"
)

func (s *Service) CreatePortfolioEntry(ctx context.Context, userID, portfolioID, assetID uuid.UUID, sourceID uuid.UUID, txnType TransactionType, quantity money.Decimal, price money.Money, costCurrency string, category EntryCategory, entryDate time.Time, notes string) (Entry, error) {
	entry, err := s.repo.CreatePortfolioEntry(ctx, userID, portfolioID, assetID, sourceID, txnType, quantity, price, costCurrency, category, entryDate, notes)
	if err != nil {
		return Entry{}, err
	}

	return entry, nil
}
