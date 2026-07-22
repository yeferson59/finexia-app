package portfolio

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yeferson59/gofinance/v2/money"
	"golang.org/x/sync/errgroup"

	"github.com/yeferson59/finexia-app/internal/platform/mail"
)

func (s *Service) GetTransactionsByEntry(ctx context.Context, userID, entryID uuid.UUID) ([]Transaction, error) {
	return s.repo.GetTransactionsByEntryID(ctx, userID, entryID)
}

func (s *Service) GetAssetTransactionsPaginated(ctx context.Context, userID, portfolioID uuid.UUID, ticker string, page, limit int) ([]Transaction, int, error) {
	offset := (page - 1) * limit

	// The count and the page are independent reads; overlap them instead of
	// paying two sequential DB round-trips.
	var (
		total int
		txns  []Transaction
	)

	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		var err error
		total, err = s.repo.CountAssetTransactions(gctx, userID, portfolioID, ticker)
		return err
	})
	g.Go(func() error {
		var err error
		txns, err = s.repo.GetAssetTransactionsPaginated(gctx, userID, portfolioID, ticker, limit, offset)
		return err
	})
	if err := g.Wait(); err != nil {
		return nil, 0, err
	}

	return txns, total, nil
}

func (s *Service) GetRecentUserTransactions(ctx context.Context, userID uuid.UUID, limit int) ([]Transaction, error) {
	return s.repo.GetRecentTransactionsByUserID(ctx, userID, limit)
}

func (s *Service) GetAssetAllocation(ctx context.Context, userID uuid.UUID) ([]AllocationItem, error) {
	return s.repo.GetAssetAllocationByUserID(ctx, userID)
}

func (s *Service) UpdateTransaction(ctx context.Context, userID, txnID uuid.UUID, txnType TransactionType, quantity money.Decimal, price money.Money, currency string, fees money.Money, transactionDate time.Time, notes string) (Transaction, error) {
	return s.repo.UpdateTransaction(ctx, userID, txnID, txnType, quantity, price, currency, fees, transactionDate, notes)
}

func (s *Service) CreateTransaction(ctx context.Context, userID, entryID uuid.UUID, txnType TransactionType, quantity money.Decimal, price money.Money, currency string, fees money.Money, transactionDate time.Time, notes string) (Transaction, error) {
	txn, err := s.repo.CreateTransaction(ctx, userID, entryID, txnType, quantity, price, currency, fees, transactionDate, notes)
	if err != nil {
		return Transaction{}, err
	}

	go s.sendTransactionAlert(userID, entryID, txn)

	return txn, nil
}

func (s *Service) sendTransactionAlert(userID, entryID uuid.UUID, txn Transaction) {
	ctx := context.Background()

	prefs, err := s.user.GetUserPreferences(ctx, userID)
	if err != nil || !prefs.EmailAlerts {
		return
	}

	usr, err := s.user.GetUserByID(ctx, userID)
	if err != nil {
		return
	}

	entry, err := s.repo.GetEntryWithAsset(ctx, entryID)
	if err != nil {
		return
	}

	qty := txn.Quantity.String()
	priceStr := txn.Price.String()
	totalStr := fmt.Sprintf("%.2f", txn.Quantity.InexactFloat64()*txn.Price.InexactFloat64())

	data := mail.ActivityAlertData{
		UserName:        usr.Name,
		AssetTicker:     entry.Asset.Ticker,
		AssetName:       entry.Asset.Name,
		TransactionType: string(txn.Type),
		Quantity:        qty,
		Price:           priceStr,
		Total:           totalStr,
		Currency:        txn.Currency,
		TransactionDate: txn.TransactionDate.Format("02 Jan 2006"),
		DashboardURL:    s.cfg.FrontendURL + "/dashboard/portfolios",
	}

	_ = s.mail.SendActivityAlert(usr.Email, data)
}
