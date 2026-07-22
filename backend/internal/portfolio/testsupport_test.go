package portfolio

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/identity"
	"github.com/yeferson59/finexia-app/internal/platform/config"
	"github.com/yeferson59/finexia-app/internal/platform/logger"
	"github.com/yeferson59/finexia-app/internal/platform/mail"
	"github.com/yeferson59/finexia-app/internal/user"
)

// mustUSD builds a USD money.Money from a decimal string, failing the test on
// a malformed input.
func mustUSD(t *testing.T, s string) money.Money {
	t.Helper()
	cur, err := money.CurrencyFromISOCode("USD")
	if err != nil {
		t.Fatalf("CurrencyFromISOCode: %v", err)
	}
	m, err := money.NewMoneyFromString(s, cur)
	if err != nil {
		t.Fatalf("NewMoneyFromString(%q): %v", s, err)
	}
	return m
}

func mustDecimal(t *testing.T, s string) money.Decimal {
	t.Helper()
	d, err := decimal.NewFromString(s)
	if err != nil {
		t.Fatalf("NewFromString(%q): %v", s, err)
	}
	return d
}

// waitFor polls cond until it returns true or the deadline expires. Used to
// synchronise with the fire-and-forget alert goroutine.
func waitFor(t *testing.T, timeout time.Duration, cond func() bool) bool {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return true
		}
		time.Sleep(5 * time.Millisecond)
	}
	return cond()
}

// fakeRepository implements portfolio.Repository through per-method hooks: a
// test sets only the hooks its scenario exercises and any other call panics
// with a nil func, flagging an unexpected dependency. The getUserPreferences
// and getUserByID hooks feed the fakeUserReader (the transaction activity
// alert reads them through the user module, not the repository).
type fakeRepository struct {
	getPortfoliosRisks              func(ctx context.Context) ([]Risk, error)
	getPortfoliosByUserID           func(ctx context.Context, userID uuid.UUID) ([]Portfolio, error)
	getPortfoliosSummaryByUserID    func(ctx context.Context, userID uuid.UUID) ([]SummaryView, error)
	getPortfolioByID                func(ctx context.Context, portfolioID, userID uuid.UUID) (Portfolio, error)
	createPortfolio                 func(ctx context.Context, userID uuid.UUID, name, description, baseCurrency string, riskID uuid.UUID, typePortfolio Type, priceValue money.Money, isDefault bool) (Portfolio, error)
	updatePortfolio                 func(ctx context.Context, userID, portfolioID uuid.UUID, name, description string, portfolioType Type, riskID uuid.UUID, isDefault bool) (Portfolio, error)
	getEntriesByPortfolioID         func(ctx context.Context, portfolioID uuid.UUID) ([]Entry, error)
	getTopTransactionByPortfolio    func(ctx context.Context, userID, portfolioID uuid.UUID) (TopTransactionDTO, error)
	createPlatform                  func(ctx context.Context, userID uuid.UUID, sourceType SourceType, name, description string) (InvestmentSource, error)
	getPlatformsWithStats           func(ctx context.Context, userID uuid.UUID) ([]PlatformStats, error)
	updatePlatform                  func(ctx context.Context, userID, sourceID uuid.UUID, name, description string, sourceType SourceType, isActive bool) (PlatformStats, error)
	deletePlatform                  func(ctx context.Context, userID, sourceID uuid.UUID) error
	createPortfolioEntry            func(ctx context.Context, userID, portfolioID, assetID, sourceID uuid.UUID, txnType TransactionType, quantity money.Decimal, price money.Money, costCurrency string, category EntryCategory, entryDate time.Time, notes string) (Entry, error)
	getEntryWithAsset               func(ctx context.Context, entryID uuid.UUID) (Entry, error)
	getTransactionsByEntryID        func(ctx context.Context, userID, entryID uuid.UUID) ([]Transaction, error)
	countAssetTransactions          func(ctx context.Context, userID, portfolioID uuid.UUID, ticker string) (int, error)
	getAssetTransactionsPaginated   func(ctx context.Context, userID, portfolioID uuid.UUID, ticker string, limit, offset int) ([]Transaction, error)
	getRecentTransactionsByUserID   func(ctx context.Context, userID uuid.UUID, limit int) ([]Transaction, error)
	getAssetAllocationByUserID      func(ctx context.Context, userID uuid.UUID) ([]AllocationItem, error)
	createTransaction               func(ctx context.Context, userID, entryID uuid.UUID, txnType TransactionType, quantity money.Decimal, price money.Money, currency string, fees money.Money, transactionDate time.Time, notes string) (Transaction, error)
	updateTransaction               func(ctx context.Context, userID, txnID uuid.UUID, txnType TransactionType, quantity money.Decimal, price money.Money, currency string, fees money.Money, transactionDate time.Time, notes string) (Transaction, error)
	importEntryTransactions         func(ctx context.Context, userID, portfolioID, sourceID uuid.UUID, rows []ImportTransactionRow) (int, error)
	getAllPortfolioSummaryRows      func(ctx context.Context) ([]SnapshotRow, error)
	upsertPortfolioSnapshot         func(ctx context.Context, portfolioID uuid.UUID, snapshotDate time.Time, totalValue, currency, totalGainLoss, totalGainLossPct string) error
	getPortfolioGrowthByUserID      func(ctx context.Context, userID uuid.UUID, hasSince bool, since time.Time) ([]GrowthPoint, error)
	getPortfolioGrowthByPortfolioID func(ctx context.Context, userID, portfolioID uuid.UUID, hasSince bool, since time.Time) ([]GrowthPoint, error)
	getExchangeRateByPair           func(ctx context.Context, from, to string) (money.Decimal, error)

	// Consumed by fakeUserReader, not part of portfolio.Repository.
	getUserPreferences func(ctx context.Context, userID uuid.UUID) (user.UserPreferences, error)
	getUserByID        func(ctx context.Context, id uuid.UUID) (identity.User, error)
}

var _ Repository = (*fakeRepository)(nil)

func (f *fakeRepository) GetPortfoliosRisks(ctx context.Context) ([]Risk, error) {
	return f.getPortfoliosRisks(ctx)
}

func (f *fakeRepository) GetPortfoliosByUserID(ctx context.Context, userID uuid.UUID) ([]Portfolio, error) {
	return f.getPortfoliosByUserID(ctx, userID)
}

func (f *fakeRepository) GetPortfoliosSummaryByUserID(ctx context.Context, userID uuid.UUID) ([]SummaryView, error) {
	return f.getPortfoliosSummaryByUserID(ctx, userID)
}

func (f *fakeRepository) GetPortfolioByID(ctx context.Context, portfolioID, userID uuid.UUID) (Portfolio, error) {
	return f.getPortfolioByID(ctx, portfolioID, userID)
}

func (f *fakeRepository) CreatePortfolio(ctx context.Context, userID uuid.UUID, name, description, baseCurrency string, riskID uuid.UUID, typePortfolio Type, priceValue money.Money, isDefault bool) (Portfolio, error) {
	return f.createPortfolio(ctx, userID, name, description, baseCurrency, riskID, typePortfolio, priceValue, isDefault)
}

func (f *fakeRepository) UpdatePortfolio(ctx context.Context, userID, portfolioID uuid.UUID, name, description string, portfolioType Type, riskID uuid.UUID, isDefault bool) (Portfolio, error) {
	return f.updatePortfolio(ctx, userID, portfolioID, name, description, portfolioType, riskID, isDefault)
}

func (f *fakeRepository) GetEntriesByPortfolioID(ctx context.Context, portfolioID uuid.UUID) ([]Entry, error) {
	return f.getEntriesByPortfolioID(ctx, portfolioID)
}

func (f *fakeRepository) GetTopTransactionByPortfolioID(ctx context.Context, userID, portfolioID uuid.UUID) (TopTransactionDTO, error) {
	return f.getTopTransactionByPortfolio(ctx, userID, portfolioID)
}

func (f *fakeRepository) CreatePlatform(ctx context.Context, userID uuid.UUID, sourceType SourceType, name, description string) (InvestmentSource, error) {
	return f.createPlatform(ctx, userID, sourceType, name, description)
}

func (f *fakeRepository) GetPlatformsWithStats(ctx context.Context, userID uuid.UUID) ([]PlatformStats, error) {
	return f.getPlatformsWithStats(ctx, userID)
}

func (f *fakeRepository) UpdatePlatform(ctx context.Context, userID, sourceID uuid.UUID, name, description string, sourceType SourceType, isActive bool) (PlatformStats, error) {
	return f.updatePlatform(ctx, userID, sourceID, name, description, sourceType, isActive)
}

func (f *fakeRepository) DeletePlatform(ctx context.Context, userID, sourceID uuid.UUID) error {
	return f.deletePlatform(ctx, userID, sourceID)
}

func (f *fakeRepository) CreatePortfolioEntry(ctx context.Context, userID, portfolioID, assetID, sourceID uuid.UUID, txnType TransactionType, quantity money.Decimal, price money.Money, costCurrency string, category EntryCategory, entryDate time.Time, notes string) (Entry, error) {
	return f.createPortfolioEntry(ctx, userID, portfolioID, assetID, sourceID, txnType, quantity, price, costCurrency, category, entryDate, notes)
}

func (f *fakeRepository) GetEntryWithAsset(ctx context.Context, entryID uuid.UUID) (Entry, error) {
	return f.getEntryWithAsset(ctx, entryID)
}

func (f *fakeRepository) GetTransactionsByEntryID(ctx context.Context, userID, entryID uuid.UUID) ([]Transaction, error) {
	return f.getTransactionsByEntryID(ctx, userID, entryID)
}

func (f *fakeRepository) CountAssetTransactions(ctx context.Context, userID, portfolioID uuid.UUID, ticker string) (int, error) {
	return f.countAssetTransactions(ctx, userID, portfolioID, ticker)
}

func (f *fakeRepository) GetAssetTransactionsPaginated(ctx context.Context, userID, portfolioID uuid.UUID, ticker string, limit, offset int) ([]Transaction, error) {
	return f.getAssetTransactionsPaginated(ctx, userID, portfolioID, ticker, limit, offset)
}

func (f *fakeRepository) GetRecentTransactionsByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]Transaction, error) {
	return f.getRecentTransactionsByUserID(ctx, userID, limit)
}

func (f *fakeRepository) GetAssetAllocationByUserID(ctx context.Context, userID uuid.UUID) ([]AllocationItem, error) {
	return f.getAssetAllocationByUserID(ctx, userID)
}

func (f *fakeRepository) CreateTransaction(ctx context.Context, userID, entryID uuid.UUID, txnType TransactionType, quantity money.Decimal, price money.Money, currency string, fees money.Money, transactionDate time.Time, notes string) (Transaction, error) {
	return f.createTransaction(ctx, userID, entryID, txnType, quantity, price, currency, fees, transactionDate, notes)
}

func (f *fakeRepository) UpdateTransaction(ctx context.Context, userID, txnID uuid.UUID, txnType TransactionType, quantity money.Decimal, price money.Money, currency string, fees money.Money, transactionDate time.Time, notes string) (Transaction, error) {
	return f.updateTransaction(ctx, userID, txnID, txnType, quantity, price, currency, fees, transactionDate, notes)
}

func (f *fakeRepository) ImportEntryTransactions(ctx context.Context, userID, portfolioID, sourceID uuid.UUID, rows []ImportTransactionRow) (int, error) {
	return f.importEntryTransactions(ctx, userID, portfolioID, sourceID, rows)
}

func (f *fakeRepository) GetAllPortfolioSummaryRows(ctx context.Context) ([]SnapshotRow, error) {
	return f.getAllPortfolioSummaryRows(ctx)
}

func (f *fakeRepository) UpsertPortfolioSnapshot(ctx context.Context, portfolioID uuid.UUID, snapshotDate time.Time, totalValue, currency, totalGainLoss, totalGainLossPct string) error {
	return f.upsertPortfolioSnapshot(ctx, portfolioID, snapshotDate, totalValue, currency, totalGainLoss, totalGainLossPct)
}

func (f *fakeRepository) GetPortfolioGrowthByUserID(ctx context.Context, userID uuid.UUID, hasSince bool, since time.Time) ([]GrowthPoint, error) {
	return f.getPortfolioGrowthByUserID(ctx, userID, hasSince, since)
}

func (f *fakeRepository) GetPortfolioGrowthByPortfolioID(ctx context.Context, userID, portfolioID uuid.UUID, hasSince bool, since time.Time) ([]GrowthPoint, error) {
	return f.getPortfolioGrowthByPortfolioID(ctx, userID, portfolioID, hasSince, since)
}

func (f *fakeRepository) GetExchangeRateByPair(ctx context.Context, from, to string) (money.Decimal, error) {
	return f.getExchangeRateByPair(ctx, from, to)
}

// fakeUserReader satisfies portfolio.UserReader by delegating to the
// repository's user hooks. Unset hooks default to a zero user with alerts
// off, so the fire-and-forget transaction alert exits early instead of
// dereferencing a nil func in its goroutine.
type fakeUserReader struct {
	repo *fakeRepository
}

func (u fakeUserReader) GetUserPreferences(ctx context.Context, userID uuid.UUID) (user.UserPreferences, error) {
	if u.repo == nil || u.repo.getUserPreferences == nil {
		return user.UserPreferences{}, nil
	}
	return u.repo.getUserPreferences(ctx, userID)
}

func (u fakeUserReader) GetUserByID(ctx context.Context, id uuid.UUID) (identity.User, error) {
	if u.repo == nil || u.repo.getUserByID == nil {
		return identity.User{}, nil
	}
	return u.repo.getUserByID(ctx, id)
}

// fakeMailer records the activity alerts the transaction flow emits so tests
// can assert on them without a Resend client.
type fakeMailer struct {
	mu sync.Mutex

	activityErr error

	activity []struct {
		To   string
		Data mail.ActivityAlertData
	}
}

func (m *fakeMailer) SendActivityAlert(email string, data mail.ActivityAlertData) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.activityErr != nil {
		return m.activityErr
	}
	m.activity = append(m.activity, struct {
		To   string
		Data mail.ActivityAlertData
	}{email, data})
	return nil
}

// memStorage is an in-memory fiber.Storage that honours TTLs, enough to
// exercise the snapshot-sync caching without Redis.
type memStorage struct {
	mu    sync.Mutex
	items map[string]memItem
}

type memItem struct {
	value     []byte
	expiresAt time.Time
}

func newMemStorage() *memStorage {
	return &memStorage{items: map[string]memItem{}}
}

func (s *memStorage) GetWithContext(_ context.Context, key string) ([]byte, error) {
	return s.Get(key)
}

func (s *memStorage) Get(key string) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.items[key]
	if !ok {
		return nil, nil
	}
	if !item.expiresAt.IsZero() && time.Now().After(item.expiresAt) {
		delete(s.items, key)
		return nil, nil
	}
	return item.value, nil
}

func (s *memStorage) SetWithContext(_ context.Context, key string, val []byte, exp time.Duration) error {
	return s.Set(key, val, exp)
}

func (s *memStorage) Set(key string, val []byte, exp time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	item := memItem{value: append([]byte(nil), val...)}
	if exp > 0 {
		item.expiresAt = time.Now().Add(exp)
	}
	s.items[key] = item
	return nil
}

func (s *memStorage) DeleteWithContext(_ context.Context, key string) error {
	return s.Delete(key)
}

func (s *memStorage) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.items, key)
	return nil
}

func (s *memStorage) ResetWithContext(_ context.Context) error {
	return s.Reset()
}

func (s *memStorage) Reset() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = map[string]memItem{}
	return nil
}

func (s *memStorage) Close() error { return nil }

func testConfig() *config.Env {
	return &config.Env{
		PublicURL:   "http://localhost:8080",
		FrontendURL: "http://localhost:5173",
	}
}

// newTestServices wires the portfolio service with a fake user reader and a
// discarded mailer, matching the common case where a test only cares about
// repository interactions.
func newTestServices(repo *fakeRepository, storage *memStorage) *Service {
	return NewService(repo, testConfig(), storage, &fakeMailer{}, fakeUserReader{repo}, logger.Noop())
}

// newTestServicesFull injects a caller-provided mailer for the flows that emit
// the transaction activity alert. The fourth argument is retained for parity
// with the legacy helper and ignored.
func newTestServicesFull(repo *fakeRepository, storage *memStorage, mailer Mailer, _ any) *Service {
	return NewService(repo, testConfig(), storage, mailer, fakeUserReader{repo}, logger.Noop())
}
