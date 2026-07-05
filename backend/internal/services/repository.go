package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/yeferson59/gofinance/money"

	portfoliodto "github.com/yeferson59/finexia-app/internal/dtos/portfolio"
	"github.com/yeferson59/finexia-app/internal/entities"
	"github.com/yeferson59/finexia-app/internal/repositories"
)

// Repository abstracts the persistence layer used by the services. It is
// satisfied by *repositories.Repository and lets tests replace the database
// with in-memory fakes.
type Repository interface {
	// Auth
	GetAccountByUserID(ctx context.Context, userID uuid.UUID) (entities.Account, error)
	GetAccountByEmail(ctx context.Context, email string) (entities.User, error)
	CreateSession(ctx context.Context, userID uuid.UUID, token string, ip, ua *string, expiresAt time.Time) (uuid.UUID, error)
	UpdateSessionToken(ctx context.Context, sessionID uuid.UUID, newToken string, expiresAt time.Time) (string, error)
	ListSessionsByUserID(ctx context.Context, userID uuid.UUID) ([]entities.Session, error)
	GetRefreshTokensBySessionIDs(ctx context.Context, userID uuid.UUID, sessionIDs []uuid.UUID) ([]string, []uuid.UUID, error)
	DeleteSessionsByIDs(ctx context.Context, userID uuid.UUID, sessionIDs []uuid.UUID) (int64, error)
	HasSessionFromIP(ctx context.Context, userID uuid.UUID, ip string) (bool, error)
	CreateRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, familyID, sessionID uuid.UUID, ip, ua *string, expiresAt time.Time) (uuid.UUID, error)
	GetRefreshTokenByHash(ctx context.Context, tokenHash string) (entities.RefreshToken, error)
	MarkRefreshTokenUsed(ctx context.Context, id uuid.UUID) error
	RevokeRefreshTokenFamily(ctx context.Context, familyID uuid.UUID) ([]string, error)
	GetRefreshTokenFamiliesBySession(ctx context.Context, userID uuid.UUID, sessionToken string) ([]string, []uuid.UUID, error)
	Register(ctx context.Context, name, email, password string) (entities.User, error)
	GetSessionByUserIDToken(ctx context.Context, userID uuid.UUID, token string) (entities.User, error)
	GetSessionByToken(ctx context.Context, token string) (entities.User, error)
	DeleteSessionByUserIDToken(ctx context.Context, userID uuid.UUID, token string) error
	DeleteExpiredRefreshTokens(ctx context.Context) (int64, error)
	DeleteExpiredSessions(ctx context.Context) (int64, error)

	// Users
	ListUsers(ctx context.Context, offset, limit uint) ([]entities.User, uint, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (entities.User, error)
	GetUserByEmail(ctx context.Context, email string) (entities.User, error)
	CreateUser(ctx context.Context, name, email string) (entities.User, error)
	UpdateUser(ctx context.Context, id uuid.UUID, name, email, image string) (entities.User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
	BanUser(ctx context.Context, id uuid.UUID, ban bool) error
	UpdateUserProfile(ctx context.Context, id uuid.UUID, name, preferredCurrency, image string) (entities.User, error)
	UpdateUserImage(ctx context.Context, id uuid.UUID, image string) (entities.User, error)
	UpdateUserPassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error
	GetUserPreferences(ctx context.Context, userID uuid.UUID) (entities.UserPreferences, error)
	UpsertUserPreferences(ctx context.Context, userID uuid.UUID, emailAlerts, weeklySummary bool) (entities.UserPreferences, error)
	GetUsersWithWeeklySummary(ctx context.Context) ([]entities.User, error)

	// Portfolios
	GetPortfoliosRisks(ctx context.Context) ([]entities.Risk, error)
	GetPortfoliosByUserID(ctx context.Context, userID uuid.UUID) ([]entities.Portfolio, error)
	GetPortfoliosSummaryByUserID(ctx context.Context, userID uuid.UUID) ([]entities.PortfolioSummaryView, error)
	GetPortfolioByID(ctx context.Context, portfolioID, userID uuid.UUID) (entities.Portfolio, error)
	CreatePortfolio(ctx context.Context, userID uuid.UUID, name, description, baseCurrency string, riskID uuid.UUID, typePortfolio entities.PortfolioType, priceValue money.Money, isDefault bool) (entities.Portfolio, error)
	UpdatePortfolio(ctx context.Context, userID, portfolioID uuid.UUID, name, description string, portfolioType entities.PortfolioType, riskID uuid.UUID, isDefault bool) (entities.Portfolio, error)
	GetEntriesByPortfolioID(ctx context.Context, portfolioID uuid.UUID) ([]entities.PortfolioEntry, error)
	GetTopTransactionByPortfolioID(ctx context.Context, userID, portfolioID uuid.UUID) (portfoliodto.PortfolioTopTransactionDTO, error)
	CreatePlatform(ctx context.Context, userID uuid.UUID, sourceType entities.SourceType, name, description string) (entities.InvestmentSource, error)
	GetPlatformsWithStats(ctx context.Context, userID uuid.UUID) ([]entities.PlatformStats, error)
	UpdatePlatform(ctx context.Context, userID, sourceID uuid.UUID, name, description string, sourceType entities.SourceType, isActive bool) (entities.PlatformStats, error)
	DeletePlatform(ctx context.Context, userID, sourceID uuid.UUID) error

	// Assets
	GetAssetByID(ctx context.Context, assetID uuid.UUID) (entities.Asset, error)
	GetAssets(ctx context.Context, offset, limit uint) ([]entities.Asset, error)
	SearchAssets(ctx context.Context, search string, offset, limit uint) ([]entities.Asset, error)
	UpsertAsset(ctx context.Context, ticker, name string, assetType entities.AssetType, exchange, currency string) (entities.Asset, error)
	UpdateAssetPrice(ctx context.Context, assetID uuid.UUID, price money.Money) (entities.Asset, error)

	// Entries & transactions
	CreatePortfolioEntry(ctx context.Context, userID, portfolioID, assetID uuid.UUID, sourceID uuid.UUID, txnType entities.TransactionType, quantity money.Decimal, price money.Money, costCurrency string, category entities.PortfolioEntryCategory, entryDate time.Time, notes string) (entities.PortfolioEntry, error)
	GetEntryWithAsset(ctx context.Context, entryID uuid.UUID) (entities.PortfolioEntry, error)
	GetTransactionsByEntryID(ctx context.Context, userID, entryID uuid.UUID) ([]entities.Transaction, error)
	CountAssetTransactions(ctx context.Context, userID, portfolioID uuid.UUID, ticker string) (int, error)
	GetAssetTransactionsPaginated(ctx context.Context, userID, portfolioID uuid.UUID, ticker string, limit, offset int) ([]entities.Transaction, error)
	GetRecentTransactionsByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]entities.Transaction, error)
	GetAssetAllocationByUserID(ctx context.Context, userID uuid.UUID) ([]entities.AllocationItem, error)
	CreateTransaction(ctx context.Context, userID, entryID uuid.UUID, txnType entities.TransactionType, quantity money.Decimal, price money.Money, currency string, fees money.Money, transactionDate time.Time, notes string) (entities.Transaction, error)
	UpdateTransaction(ctx context.Context, userID, txnID uuid.UUID, txnType entities.TransactionType, quantity money.Decimal, price money.Money, currency string, fees money.Money, transactionDate time.Time, notes string) (entities.Transaction, error)
	ImportEntryTransactions(ctx context.Context, userID, portfolioID, sourceID uuid.UUID, rows []entities.ImportTransactionRow) (int, error)

	// Snapshots & growth
	GetAllPortfolioSummaryRows(ctx context.Context) ([]entities.PortfolioSnapshotRow, error)
	UpsertPortfolioSnapshot(ctx context.Context, portfolioID uuid.UUID, snapshotDate time.Time, totalValue, currency, totalGainLoss, totalGainLossPct string) error
	GetPortfolioGrowthByUserID(ctx context.Context, userID uuid.UUID, hasSince bool, since time.Time) ([]entities.PortfolioGrowthPoint, error)

	// Exchange rates & marketing
	UpsertExchangeRate(ctx context.Context, from, to string, rate money.Decimal, rateDate time.Time) (entities.ExchangeRate, error)
	SaveWaitlistEmail(ctx context.Context, email string) error
}

// Ensure the concrete repository keeps satisfying the interface.
var _ Repository = (*repositories.Repository)(nil)
