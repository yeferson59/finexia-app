package services

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/yeferson59/gofinance/money"

	"github.com/yeferson59/finexia-app/internal/config"
	portfoliodto "github.com/yeferson59/finexia-app/internal/dtos/portfolio"
	"github.com/yeferson59/finexia-app/internal/entities"
	"github.com/yeferson59/finexia-app/internal/logger"
	"github.com/yeferson59/finexia-app/internal/mail"
	"github.com/yeferson59/finexia-app/internal/prices"
)

// fakeRepository embeds the Repository interface so tests only override the
// methods a scenario needs; calling anything else panics loudly.
type fakeRepository struct {
	Repository

	getAccountByEmail                func(ctx context.Context, email string) (entities.User, error)
	getAccountByUserID               func(ctx context.Context, userID uuid.UUID) (entities.Account, error)
	createSession                    func(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) (uuid.UUID, error)
	updateSessionToken               func(ctx context.Context, sessionID uuid.UUID, newToken string, expiresAt time.Time) (string, error)
	createRefreshToken               func(ctx context.Context, userID uuid.UUID, tokenHash string, familyID, sessionID uuid.UUID, ip, ua *string, expiresAt time.Time) (uuid.UUID, error)
	getRefreshTokenByHash            func(ctx context.Context, tokenHash string) (entities.RefreshToken, error)
	markRefreshTokenUsed             func(ctx context.Context, id uuid.UUID) error
	revokeRefreshTokenFamily         func(ctx context.Context, familyID uuid.UUID) ([]string, error)
	getRefreshTokenFamiliesBySession func(ctx context.Context, userID uuid.UUID, sessionToken string) ([]string, []uuid.UUID, error)
	register                         func(ctx context.Context, name, email, password string) (entities.User, error)
	getSessionByToken                func(ctx context.Context, token string) (entities.User, error)
	deleteSessionByUserIDToken       func(ctx context.Context, userID uuid.UUID, token string) error
	deleteExpiredRefreshTokens       func(ctx context.Context) (int64, error)
	deleteExpiredSessions            func(ctx context.Context) (int64, error)
	getUserByEmail                   func(ctx context.Context, email string) (entities.User, error)
	getUserByID                      func(ctx context.Context, id uuid.UUID) (entities.User, error)
	updateUser                       func(ctx context.Context, id uuid.UUID, name, email, image string) (entities.User, error)
	updateUserProfile                func(ctx context.Context, id uuid.UUID, name, preferredCurrency, image string) (entities.User, error)
	updateUserPassword               func(ctx context.Context, userID uuid.UUID, hashedPassword string) error
	countAssetTransactions           func(ctx context.Context, userID, portfolioID uuid.UUID, ticker string) (int, error)
	getAssetTransactionsPaginated    func(ctx context.Context, userID, portfolioID uuid.UUID, ticker string, limit, offset int) ([]entities.Transaction, error)
	getPortfolioGrowthByUserID       func(ctx context.Context, userID uuid.UUID, hasSince bool, since time.Time) ([]entities.PortfolioGrowthPoint, error)

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
	getUserPreferences            func(ctx context.Context, userID uuid.UUID) (entities.UserPreferences, error)
	getUsersWithWeeklySummary     func(ctx context.Context) ([]entities.User, error)
	getAllPortfolioSummaryRows    func(ctx context.Context) ([]entities.PortfolioSnapshotRow, error)
	upsertPortfolioSnapshot       func(ctx context.Context, portfolioID uuid.UUID, snapshotDate time.Time, totalValue, currency, totalGainLoss, totalGainLossPct string) error
	upsertExchangeRate            func(ctx context.Context, from, to string, rate money.Decimal, rateDate time.Time) (entities.ExchangeRate, error)
	saveWaitlistEmail             func(ctx context.Context, email string) error
}

func (f *fakeRepository) GetAccountByEmail(ctx context.Context, email string) (entities.User, error) {
	return f.getAccountByEmail(ctx, email)
}

func (f *fakeRepository) GetAccountByUserID(ctx context.Context, userID uuid.UUID) (entities.Account, error) {
	return f.getAccountByUserID(ctx, userID)
}

func (f *fakeRepository) CreateSession(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) (uuid.UUID, error) {
	return f.createSession(ctx, userID, token, expiresAt)
}

func (f *fakeRepository) UpdateSessionToken(ctx context.Context, sessionID uuid.UUID, newToken string, expiresAt time.Time) (string, error) {
	return f.updateSessionToken(ctx, sessionID, newToken, expiresAt)
}

func (f *fakeRepository) CreateRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, familyID, sessionID uuid.UUID, ip, ua *string, expiresAt time.Time) (uuid.UUID, error) {
	return f.createRefreshToken(ctx, userID, tokenHash, familyID, sessionID, ip, ua, expiresAt)
}

func (f *fakeRepository) GetRefreshTokenByHash(ctx context.Context, tokenHash string) (entities.RefreshToken, error) {
	return f.getRefreshTokenByHash(ctx, tokenHash)
}

func (f *fakeRepository) MarkRefreshTokenUsed(ctx context.Context, id uuid.UUID) error {
	return f.markRefreshTokenUsed(ctx, id)
}

func (f *fakeRepository) RevokeRefreshTokenFamily(ctx context.Context, familyID uuid.UUID) ([]string, error) {
	return f.revokeRefreshTokenFamily(ctx, familyID)
}

func (f *fakeRepository) GetRefreshTokenFamiliesBySession(ctx context.Context, userID uuid.UUID, sessionToken string) ([]string, []uuid.UUID, error) {
	return f.getRefreshTokenFamiliesBySession(ctx, userID, sessionToken)
}

func (f *fakeRepository) Register(ctx context.Context, name, email, password string) (entities.User, error) {
	return f.register(ctx, name, email, password)
}

func (f *fakeRepository) GetSessionByToken(ctx context.Context, token string) (entities.User, error) {
	return f.getSessionByToken(ctx, token)
}

func (f *fakeRepository) DeleteSessionByUserIDToken(ctx context.Context, userID uuid.UUID, token string) error {
	return f.deleteSessionByUserIDToken(ctx, userID, token)
}

func (f *fakeRepository) DeleteExpiredRefreshTokens(ctx context.Context) (int64, error) {
	return f.deleteExpiredRefreshTokens(ctx)
}

func (f *fakeRepository) DeleteExpiredSessions(ctx context.Context) (int64, error) {
	return f.deleteExpiredSessions(ctx)
}

func (f *fakeRepository) GetUserByEmail(ctx context.Context, email string) (entities.User, error) {
	return f.getUserByEmail(ctx, email)
}

func (f *fakeRepository) GetUserByID(ctx context.Context, id uuid.UUID) (entities.User, error) {
	return f.getUserByID(ctx, id)
}

func (f *fakeRepository) UpdateUser(ctx context.Context, id uuid.UUID, name, email, image string) (entities.User, error) {
	return f.updateUser(ctx, id, name, email, image)
}

func (f *fakeRepository) UpdateUserProfile(ctx context.Context, id uuid.UUID, name, preferredCurrency, image string) (entities.User, error) {
	return f.updateUserProfile(ctx, id, name, preferredCurrency, image)
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

func (f *fakeRepository) GetUserPreferences(ctx context.Context, userID uuid.UUID) (entities.UserPreferences, error) {
	return f.getUserPreferences(ctx, userID)
}

func (f *fakeRepository) GetUsersWithWeeklySummary(ctx context.Context) ([]entities.User, error) {
	return f.getUsersWithWeeklySummary(ctx)
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

func (f *fakeRepository) SaveWaitlistEmail(ctx context.Context, email string) error {
	return f.saveWaitlistEmail(ctx, email)
}

// fakeMailer records outbound emails so tests can assert on the alert and
// summary flows without a Resend client.
type fakeMailer struct {
	mu sync.Mutex

	waitlistErr error
	activityErr error
	weeklyErr   error

	waitlistTo []string
	activity   []struct {
		To   string
		Data mail.ActivityAlertData
	}
	weekly []struct {
		To   string
		Data mail.WeeklySummaryData
	}
}

func (m *fakeMailer) SendWaitlistConfirmation(email string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.waitlistErr != nil {
		return m.waitlistErr
	}
	m.waitlistTo = append(m.waitlistTo, email)
	return nil
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
	fetchQuote        func(ctx context.Context, symbol string) (prices.QuoteResult, error)
	fetchExchangeRate func(ctx context.Context, from, to string) (prices.ExchangeRateResult, error)
}

func (p *fakePriceProvider) FetchQuote(ctx context.Context, symbol string) (prices.QuoteResult, error) {
	return p.fetchQuote(ctx, symbol)
}

func (p *fakePriceProvider) FetchExchangeRate(ctx context.Context, from, to string) (prices.ExchangeRateResult, error) {
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

func (s *memStorage) has(key string) bool {
	v, _ := s.Get(key)
	return v != nil
}

func testConfig() *config.Env {
	return &config.Env{
		JWTSecret:          "test-secret",
		JWTAccessDuration:  15 * time.Minute,
		JWTRefreshDuration: 30 * 24 * time.Hour,
		RefreshGracePeriod: 30 * time.Second,
		PublicURL:          "http://localhost:8080",
	}
}

func newTestServices(repo Repository, storage *memStorage) *Services {
	svc := New(repo, testConfig(), nil, storage, nil, logger.Noop(), nil)
	return &svc
}

// newTestServicesFull wires a fake mailer and price provider in addition to
// the repository, for flows that send email or hit market data.
func newTestServicesFull(repo Repository, storage *memStorage, mailer Mailer, provider prices.Provider) *Services {
	svc := New(repo, testConfig(), nil, storage, mailer, logger.Noop(), provider)
	return &svc
}
