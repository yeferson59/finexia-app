package services

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/yeferson59/gofinance/v2/money"

	portfoliodto "github.com/yeferson59/finexia-app/internal/dtos/portfolio"
	"github.com/yeferson59/finexia-app/internal/entities"
	"github.com/yeferson59/finexia-app/internal/identity"
	"github.com/yeferson59/finexia-app/internal/platform/config"
	"github.com/yeferson59/finexia-app/internal/platform/logger"
	"github.com/yeferson59/finexia-app/internal/platform/mail"
	"github.com/yeferson59/finexia-app/internal/platform/marketdata"
	"github.com/yeferson59/finexia-app/internal/user"
)

// fakeRepository embeds the Repository interface so tests only override the
// methods a scenario needs; calling anything else panics loudly.
type fakeRepository struct {
	Repository

	updateUserPassword              func(ctx context.Context, userID uuid.UUID, hashedPassword string) error
	countAssetTransactions          func(ctx context.Context, userID, portfolioID uuid.UUID, ticker string) (int, error)
	getAssetTransactionsPaginated   func(ctx context.Context, userID, portfolioID uuid.UUID, ticker string, limit, offset int) ([]entities.Transaction, error)
	getPortfolioGrowthByUserID      func(ctx context.Context, userID uuid.UUID, hasSince bool, since time.Time) ([]entities.PortfolioGrowthPoint, error)
	getPortfolioGrowthByPortfolioID func(ctx context.Context, userID, portfolioID uuid.UUID, hasSince bool, since time.Time) ([]entities.PortfolioGrowthPoint, error)

	getPortfoliosRisks            func(ctx context.Context) ([]entities.Risk, error)
	getPortfoliosByUserID         func(ctx context.Context, userID uuid.UUID) ([]entities.Portfolio, error)
	getPortfoliosSummaryByUserID  func(ctx context.Context, userID uuid.UUID) ([]entities.PortfolioSummaryView, error)
	getPortfolioByID              func(ctx context.Context, portfolioID, userID uuid.UUID) (entities.Portfolio, error)
	createPortfolio               func(ctx context.Context, userID uuid.UUID, name, description, baseCurrency string, riskID uuid.UUID, typePortfolio entities.PortfolioType, priceValue money.Money, isDefault bool) (entities.Portfolio, error)
	updatePortfolio               func(ctx context.Context, userID, portfolioID uuid.UUID, name, description string, portfolioType entities.PortfolioType, riskID uuid.UUID, isDefault bool) (entities.Portfolio, error)
	getEntriesByPortfolioID       func(ctx context.Context, portfolioID uuid.UUID) ([]entities.PortfolioEntry, error)
	getTopTransactionByPortfolio  func(ctx context.Context, userID, portfolioID uuid.UUID) (portfoliodto.PortfolioTopTransactionDTO, error)
	createPlatform                func(ctx context.Context, userID uuid.UUID, sourceType entities.SourceType, name, description string) (entities.InvestmentSource, error)
	getPlatformsWithStats         func(ctx context.Context, userID uuid.UUID) ([]entities.PlatformStats, error)
	updatePlatform                func(ctx context.Context, userID, sourceID uuid.UUID, name, description string, sourceType entities.SourceType, isActive bool) (entities.PlatformStats, error)
	deletePlatform                func(ctx context.Context, userID, sourceID uuid.UUID) error
	getAssetByID                  func(ctx context.Context, assetID uuid.UUID) (entities.Asset, error)
	getAssets                     func(ctx context.Context, offset, limit uint) ([]entities.Asset, error)
	searchAssets                  func(ctx context.Context, search string, offset, limit uint) ([]entities.Asset, error)
	upsertAsset                   func(ctx context.Context, ticker, name string, assetType entities.AssetType, exchange, currency string) (entities.Asset, error)
	updateAssetPrice              func(ctx context.Context, assetID uuid.UUID, price money.Money) (entities.Asset, error)
	createPortfolioEntry          func(ctx context.Context, userID, portfolioID, assetID, sourceID uuid.UUID, txnType entities.TransactionType, quantity money.Decimal, price money.Money, costCurrency string, category entities.PortfolioEntryCategory, entryDate time.Time, notes string) (entities.PortfolioEntry, error)
	getEntryWithAsset             func(ctx context.Context, entryID uuid.UUID) (entities.PortfolioEntry, error)
	getTransactionsByEntryID      func(ctx context.Context, userID, entryID uuid.UUID) ([]entities.Transaction, error)
	getRecentTransactionsByUserID func(ctx context.Context, userID uuid.UUID, limit int) ([]entities.Transaction, error)
	getAssetAllocationByUserID    func(ctx context.Context, userID uuid.UUID) ([]entities.AllocationItem, error)
	createTransaction             func(ctx context.Context, userID, entryID uuid.UUID, txnType entities.TransactionType, quantity money.Decimal, price money.Money, currency string, fees money.Money, transactionDate time.Time, notes string) (entities.Transaction, error)
	updateTransaction             func(ctx context.Context, userID, txnID uuid.UUID, txnType entities.TransactionType, quantity money.Decimal, price money.Money, currency string, fees money.Money, transactionDate time.Time, notes string) (entities.Transaction, error)
	getAllPortfolioSummaryRows    func(ctx context.Context) ([]entities.PortfolioSnapshotRow, error)
	upsertPortfolioSnapshot       func(ctx context.Context, portfolioID uuid.UUID, snapshotDate time.Time, totalValue, currency, totalGainLoss, totalGainLossPct string) error
	upsertExchangeRate            func(ctx context.Context, from, to string, rate money.Decimal, rateDate time.Time) (entities.ExchangeRate, error)
	getExchangeRateByPair         func(ctx context.Context, from, to string) (entities.ExchangeRate, error)
	importEntryTransactions       func(ctx context.Context, userID, portfolioID, sourceID uuid.UUID, rows []entities.ImportTransactionRow) (int, error)
}

func (f *fakeRepository) ImportEntryTransactions(ctx context.Context, userID, portfolioID, sourceID uuid.UUID, rows []entities.ImportTransactionRow) (int, error) {
	return f.importEntryTransactions(ctx, userID, portfolioID, sourceID, rows)
}

func (f *fakeRepository) UpdateUserPassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error {
	return f.updateUserPassword(ctx, userID, hashedPassword)
}

func (f *fakeRepository) CountAssetTransactions(ctx context.Context, userID, portfolioID uuid.UUID, ticker string) (int, error) {
	return f.countAssetTransactions(ctx, userID, portfolioID, ticker)
}

func (f *fakeRepository) GetAssetTransactionsPaginated(ctx context.Context, userID, portfolioID uuid.UUID, ticker string, limit, offset int) ([]entities.Transaction, error) {
	return f.getAssetTransactionsPaginated(ctx, userID, portfolioID, ticker, limit, offset)
}

func (f *fakeRepository) GetPortfolioGrowthByUserID(ctx context.Context, userID uuid.UUID, hasSince bool, since time.Time) ([]entities.PortfolioGrowthPoint, error) {
	return f.getPortfolioGrowthByUserID(ctx, userID, hasSince, since)
}

func (f *fakeRepository) GetPortfolioGrowthByPortfolioID(ctx context.Context, userID, portfolioID uuid.UUID, hasSince bool, since time.Time) ([]entities.PortfolioGrowthPoint, error) {
	return f.getPortfolioGrowthByPortfolioID(ctx, userID, portfolioID, hasSince, since)
}

func (f *fakeRepository) GetPortfoliosRisks(ctx context.Context) ([]entities.Risk, error) {
	return f.getPortfoliosRisks(ctx)
}

func (f *fakeRepository) GetPortfoliosByUserID(ctx context.Context, userID uuid.UUID) ([]entities.Portfolio, error) {
	return f.getPortfoliosByUserID(ctx, userID)
}

func (f *fakeRepository) GetPortfoliosSummaryByUserID(ctx context.Context, userID uuid.UUID) ([]entities.PortfolioSummaryView, error) {
	return f.getPortfoliosSummaryByUserID(ctx, userID)
}

func (f *fakeRepository) GetPortfolioByID(ctx context.Context, portfolioID, userID uuid.UUID) (entities.Portfolio, error) {
	return f.getPortfolioByID(ctx, portfolioID, userID)
}

func (f *fakeRepository) CreatePortfolio(ctx context.Context, userID uuid.UUID, name, description, baseCurrency string, riskID uuid.UUID, typePortfolio entities.PortfolioType, priceValue money.Money, isDefault bool) (entities.Portfolio, error) {
	return f.createPortfolio(ctx, userID, name, description, baseCurrency, riskID, typePortfolio, priceValue, isDefault)
}

func (f *fakeRepository) UpdatePortfolio(ctx context.Context, userID, portfolioID uuid.UUID, name, description string, portfolioType entities.PortfolioType, riskID uuid.UUID, isDefault bool) (entities.Portfolio, error) {
	return f.updatePortfolio(ctx, userID, portfolioID, name, description, portfolioType, riskID, isDefault)
}

func (f *fakeRepository) GetEntriesByPortfolioID(ctx context.Context, portfolioID uuid.UUID) ([]entities.PortfolioEntry, error) {
	return f.getEntriesByPortfolioID(ctx, portfolioID)
}

func (f *fakeRepository) GetTopTransactionByPortfolioID(ctx context.Context, userID, portfolioID uuid.UUID) (portfoliodto.PortfolioTopTransactionDTO, error) {
	return f.getTopTransactionByPortfolio(ctx, userID, portfolioID)
}

func (f *fakeRepository) CreatePlatform(ctx context.Context, userID uuid.UUID, sourceType entities.SourceType, name, description string) (entities.InvestmentSource, error) {
	return f.createPlatform(ctx, userID, sourceType, name, description)
}

func (f *fakeRepository) GetPlatformsWithStats(ctx context.Context, userID uuid.UUID) ([]entities.PlatformStats, error) {
	return f.getPlatformsWithStats(ctx, userID)
}

func (f *fakeRepository) UpdatePlatform(ctx context.Context, userID, sourceID uuid.UUID, name, description string, sourceType entities.SourceType, isActive bool) (entities.PlatformStats, error) {
	return f.updatePlatform(ctx, userID, sourceID, name, description, sourceType, isActive)
}

func (f *fakeRepository) DeletePlatform(ctx context.Context, userID, sourceID uuid.UUID) error {
	return f.deletePlatform(ctx, userID, sourceID)
}

func (f *fakeRepository) GetAssetByID(ctx context.Context, assetID uuid.UUID) (entities.Asset, error) {
	return f.getAssetByID(ctx, assetID)
}

func (f *fakeRepository) GetAssets(ctx context.Context, offset, limit uint) ([]entities.Asset, error) {
	return f.getAssets(ctx, offset, limit)
}

func (f *fakeRepository) SearchAssets(ctx context.Context, search string, offset, limit uint) ([]entities.Asset, error) {
	return f.searchAssets(ctx, search, offset, limit)
}

func (f *fakeRepository) UpsertAsset(ctx context.Context, ticker, name string, assetType entities.AssetType, exchange, currency string) (entities.Asset, error) {
	return f.upsertAsset(ctx, ticker, name, assetType, exchange, currency)
}

func (f *fakeRepository) UpdateAssetPrice(ctx context.Context, assetID uuid.UUID, price money.Money) (entities.Asset, error) {
	return f.updateAssetPrice(ctx, assetID, price)
}

func (f *fakeRepository) CreatePortfolioEntry(ctx context.Context, userID, portfolioID, assetID uuid.UUID, sourceID uuid.UUID, txnType entities.TransactionType, quantity money.Decimal, price money.Money, costCurrency string, category entities.PortfolioEntryCategory, entryDate time.Time, notes string) (entities.PortfolioEntry, error) {
	return f.createPortfolioEntry(ctx, userID, portfolioID, assetID, sourceID, txnType, quantity, price, costCurrency, category, entryDate, notes)
}

func (f *fakeRepository) GetEntryWithAsset(ctx context.Context, entryID uuid.UUID) (entities.PortfolioEntry, error) {
	return f.getEntryWithAsset(ctx, entryID)
}

func (f *fakeRepository) GetTransactionsByEntryID(ctx context.Context, userID, entryID uuid.UUID) ([]entities.Transaction, error) {
	return f.getTransactionsByEntryID(ctx, userID, entryID)
}

func (f *fakeRepository) GetRecentTransactionsByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]entities.Transaction, error) {
	return f.getRecentTransactionsByUserID(ctx, userID, limit)
}

func (f *fakeRepository) GetAssetAllocationByUserID(ctx context.Context, userID uuid.UUID) ([]entities.AllocationItem, error) {
	return f.getAssetAllocationByUserID(ctx, userID)
}

func (f *fakeRepository) CreateTransaction(ctx context.Context, userID, entryID uuid.UUID, txnType entities.TransactionType, quantity money.Decimal, price money.Money, currency string, fees money.Money, transactionDate time.Time, notes string) (entities.Transaction, error) {
	return f.createTransaction(ctx, userID, entryID, txnType, quantity, price, currency, fees, transactionDate, notes)
}

func (f *fakeRepository) UpdateTransaction(ctx context.Context, userID, txnID uuid.UUID, txnType entities.TransactionType, quantity money.Decimal, price money.Money, currency string, fees money.Money, transactionDate time.Time, notes string) (entities.Transaction, error) {
	return f.updateTransaction(ctx, userID, txnID, txnType, quantity, price, currency, fees, transactionDate, notes)
}

func (f *fakeRepository) GetAllPortfolioSummaryRows(ctx context.Context) ([]entities.PortfolioSnapshotRow, error) {
	return f.getAllPortfolioSummaryRows(ctx)
}

func (f *fakeRepository) UpsertPortfolioSnapshot(ctx context.Context, portfolioID uuid.UUID, snapshotDate time.Time, totalValue, currency, totalGainLoss, totalGainLossPct string) error {
	return f.upsertPortfolioSnapshot(ctx, portfolioID, snapshotDate, totalValue, currency, totalGainLoss, totalGainLossPct)
}

func (f *fakeRepository) UpsertExchangeRate(ctx context.Context, from, to string, rate money.Decimal, rateDate time.Time) (entities.ExchangeRate, error) {
	return f.upsertExchangeRate(ctx, from, to, rate, rateDate)
}

func (f *fakeRepository) GetExchangeRateByPair(ctx context.Context, from, to string) (entities.ExchangeRate, error) {
	return f.getExchangeRateByPair(ctx, from, to)
}

// fakeMailer records outbound emails so tests can assert on the alert and
// summary flows without a Resend client.
type fakeMailer struct {
	mu sync.Mutex

	activityErr error
	securityErr error
	weeklyErr   error

	activity []struct {
		To   string
		Data mail.ActivityAlertData
	}
	security []struct {
		To   string
		Data mail.SecurityAlertData
	}
	weekly []struct {
		To   string
		Data mail.WeeklySummaryData
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

func (m *fakeMailer) SendSecurityAlert(email string, data mail.SecurityAlertData) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.securityErr != nil {
		return m.securityErr
	}
	m.security = append(m.security, struct {
		To   string
		Data mail.SecurityAlertData
	}{email, data})
	return nil
}

func (m *fakeMailer) SendWeeklySummary(email string, data mail.WeeklySummaryData) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.weeklyErr != nil {
		return m.weeklyErr
	}
	m.weekly = append(m.weekly, struct {
		To   string
		Data mail.WeeklySummaryData
	}{email, data})
	return nil
}

// fakePriceProvider stubs the market data provider used by the sync jobs.
type fakePriceProvider struct {
	fetchQuote        func(ctx context.Context, symbol string) (marketdata.QuoteResult, error)
	fetchExchangeRate func(ctx context.Context, from, to string) (marketdata.ExchangeRateResult, error)
}

func (p *fakePriceProvider) FetchQuote(ctx context.Context, symbol string) (marketdata.QuoteResult, error) {
	return p.fetchQuote(ctx, symbol)
}

func (p *fakePriceProvider) FetchExchangeRate(ctx context.Context, from, to string) (marketdata.ExchangeRateResult, error) {
	return p.fetchExchangeRate(ctx, from, to)
}

// memStorage is an in-memory fiber.Storage that honours TTLs, good enough to
// exercise the auth caching logic without Redis.
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
		JWTSecret:              "test-secret",
		JWTAccessDuration:      15 * time.Minute,
		JWTRefreshDuration:     30 * 24 * time.Hour,
		RefreshGracePeriod:     30 * time.Second,
		PublicURL:              "http://localhost:8080",
		TwoFactorPendingExpiry: 5 * time.Minute,
	}
}

// fakeAuthService stubs the auth module slice the legacy services delegate
// to. Nil hooks are no-ops so tests that never touch the password flows can
// ignore it.
type fakeAuthService struct {
	verifyPassword      func(ctx context.Context, userID uuid.UUID, currentPassword string) error
	revokeOtherSessions func(ctx context.Context, userID uuid.UUID, currentToken string) (int64, error)
}

func (f *fakeAuthService) VerifyPassword(ctx context.Context, userID uuid.UUID, currentPassword string) error {
	if f.verifyPassword == nil {
		return nil
	}
	return f.verifyPassword(ctx, userID, currentPassword)
}

func (f *fakeAuthService) RevokeOtherSessions(ctx context.Context, userID uuid.UUID, currentToken string) (int64, error) {
	if f.revokeOtherSessions == nil {
		return 0, nil
	}
	return f.revokeOtherSessions(ctx, userID, currentToken)
}

type fakeUserService struct {
	getUserPreferences        func(ctx context.Context, userID uuid.UUID) (user.UserPreferences, error)
	getUserByID               func(ctx context.Context, id uuid.UUID) (identity.User, error)
	getUsersWithWeeklySummary func(ctx context.Context) ([]identity.User, error)
}

func (f *fakeUserService) GetUserPreferences(ctx context.Context, userID uuid.UUID) (user.UserPreferences, error) {
	if f.getUserPreferences == nil {
		return user.UserPreferences{}, nil
	}

	return f.getUserPreferences(ctx, userID)
}

func (f *fakeUserService) GetUserByID(ctx context.Context, userID uuid.UUID) (identity.User, error) {
	if f.getUserByID == nil {
		return identity.User{}, nil
	}

	return f.getUserByID(ctx, userID)
}

func (f *fakeUserService) GetUsersWithWeeklySummary(ctx context.Context) ([]identity.User, error) {
	if f.getUsersWithWeeklySummary == nil {
		return []identity.User{}, nil
	}

	return f.getUsersWithWeeklySummary(ctx)
}

func newTestServices(repo Repository, storage *memStorage) *Services {
	svc := New(repo, testConfig(), nil, storage, nil, nil, logger.Noop(), nil, &fakeAuthService{}, &fakeUserService{})
	return &svc
}

// newTestServicesFull wires a fake mailer and price provider in addition to
// the repository, for flows that send email or hit market data.
func newTestServicesFull(repo Repository, storage *memStorage, mailer Mailer, provider marketdata.Provider) *Services {
	svc := New(repo, testConfig(), nil, storage, mailer, nil, logger.Noop(), provider, &fakeAuthService{}, &fakeUserService{})
	return &svc
}
