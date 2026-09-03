package portfolio

import (
	"context"
	"time"

	"uuid"

	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"
	"golang.org/x/sync/errgroup"

	"github.com/yeferson59/finexia-app/internal/platform/mail"
)

func (s *service) GetTransactionsByEntry(ctx context.Context, userID, entryID uuid.UUID) ([]Transaction, error) {
	return s.repo.GetTransactionsByEntryID(ctx, userID, entryID)
}

func (s *service) GetAssetTransactionsPaginated(ctx context.Context, userID, portfolioID uuid.UUID, ticker string, page, limit int) ([]Transaction, int, error) {
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

func (s *service) GetRecentUserTransactions(ctx context.Context, userID uuid.UUID, limit int) ([]Transaction, error) {
	return s.repo.GetRecentTransactionsByUserID(ctx, userID, limit)
}

// GetAssetAllocation totals the user's holdings per category in one currency.
// An empty targetCurrency reports in the user's stored preference.
func (s *service) GetAssetAllocation(ctx context.Context, userID uuid.UUID, targetCurrency money.Currency) ([]AllocationItem, error) {
	return s.repo.GetAssetAllocationByUserID(ctx, userID, targetCurrency)
}

func (s *service) UpdateTransaction(ctx context.Context, userID, txnID uuid.UUID, txnType TransactionType, quantity decimal.Decimal, price money.Money, currency string, fees money.Money, transactionDate time.Time, notes string) (Transaction, error) {
	return s.repo.UpdateTransaction(ctx, userID, txnID, txnType, quantity, price, currency, fees, transactionDate, notes)
}

// DeleteTransaction removes a transaction the user owns. No activity alert is
// sent: those announce trades, and undoing a mistyped one is not a trade.
func (s *service) DeleteTransaction(ctx context.Context, userID, txnID uuid.UUID) error {
	return s.repo.DeleteTransaction(ctx, userID, txnID)
}

func (s *service) CreateTransaction(ctx context.Context, userID, entryID uuid.UUID, txnType TransactionType, quantity decimal.Decimal, price money.Money, currency money.Currency, fees money.Money, transactionDate time.Time, notes string) (Transaction, error) {
	txn, err := s.repo.CreateTransaction(ctx, userID, entryID, txnType, quantity, price, currency, fees, transactionDate, notes)
	if err != nil {
		return Transaction{}, err
	}

	go s.sendTransactionAlert(userID, entryID, txn)

	return txn, nil
}

func (s *service) sendTransactionAlert(userID, entryID uuid.UUID, txn Transaction) {
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
	// The line total stays on the money type all the way to the string: going
	// through float64 lost cents on prices with more than a couple of decimals,
	// which is every crypto fill.
	totalStr := txn.Price.MulDecimal(txn.Quantity).RoundBank(2).StringFixed(2)

	data := mail.ActivityAlertData{
		UserName:        usr.Name,
		AssetTicker:     entry.Asset.Ticker,
		AssetName:       entry.Asset.Name,
		TransactionType: string(txn.Type),
		Quantity:        qty,
		Price:           priceStr,
		Total:           totalStr,
		Currency:        txn.Currency.String(),
		TransactionDate: txn.TransactionDate.Format("02 Jan 2006"),
		DashboardURL:    s.cfg.FrontendURL + "/dashboard/portfolios",
	}

	_ = s.mail.SendActivityAlert(usr.Email, data)
}
