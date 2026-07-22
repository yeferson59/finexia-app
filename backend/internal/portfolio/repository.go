package portfolio

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/yeferson59/gofinance/v2/money"
)

// Repository is the persistence surface the portfolio module needs, defined
// by the consumer (this module) and satisfied by *PostgresRepository. The
// asset methods migrate to the market module in Fase 7.
type Repository interface {
	// Portfolios
	GetPortfoliosRisks(ctx context.Context) ([]Risk, error)
	GetPortfoliosByUserID(ctx context.Context, userID uuid.UUID) ([]Portfolio, error)
	GetPortfoliosSummaryByUserID(ctx context.Context, userID uuid.UUID) ([]SummaryView, error)
	GetPortfolioByID(ctx context.Context, portfolioID, userID uuid.UUID) (Portfolio, error)
	CreatePortfolio(ctx context.Context, userID uuid.UUID, name, description, baseCurrency string, riskID uuid.UUID, typePortfolio Type, priceValue money.Money, isDefault bool) (Portfolio, error)
	UpdatePortfolio(ctx context.Context, userID, portfolioID uuid.UUID, name, description string, portfolioType Type, riskID uuid.UUID, isDefault bool) (Portfolio, error)
	GetEntriesByPortfolioID(ctx context.Context, portfolioID uuid.UUID) ([]Entry, error)
	GetTopTransactionByPortfolioID(ctx context.Context, userID, portfolioID uuid.UUID) (TopTransactionDTO, error)

	// Platforms (investment sources)
	CreatePlatform(ctx context.Context, userID uuid.UUID, sourceType SourceType, name, description string) (InvestmentSource, error)
	GetPlatformsWithStats(ctx context.Context, userID uuid.UUID) ([]PlatformStats, error)
	UpdatePlatform(ctx context.Context, userID, sourceID uuid.UUID, name, description string, sourceType SourceType, isActive bool) (PlatformStats, error)
	DeletePlatform(ctx context.Context, userID, sourceID uuid.UUID) error

	// Assets
	GetAssetByID(ctx context.Context, assetID uuid.UUID) (Asset, error)
	GetAssets(ctx context.Context, offset, limit uint) ([]Asset, error)
	SearchAssets(ctx context.Context, search string, offset, limit uint) ([]Asset, error)
	UpsertAsset(ctx context.Context, ticker, name string, assetType AssetType, exchange, currency string) (Asset, error)
	UpdateAssetPrice(ctx context.Context, assetID uuid.UUID, price money.Money) (Asset, error)

	// Entries & transactions
	CreatePortfolioEntry(ctx context.Context, userID, portfolioID, assetID uuid.UUID, sourceID uuid.UUID, txnType TransactionType, quantity money.Decimal, price money.Money, costCurrency string, category EntryCategory, entryDate time.Time, notes string) (Entry, error)
	GetEntryWithAsset(ctx context.Context, entryID uuid.UUID) (Entry, error)
	GetTransactionsByEntryID(ctx context.Context, userID, entryID uuid.UUID) ([]Transaction, error)
	CountAssetTransactions(ctx context.Context, userID, portfolioID uuid.UUID, ticker string) (int, error)
	GetAssetTransactionsPaginated(ctx context.Context, userID, portfolioID uuid.UUID, ticker string, limit, offset int) ([]Transaction, error)
	GetRecentTransactionsByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]Transaction, error)
	GetAssetAllocationByUserID(ctx context.Context, userID uuid.UUID) ([]AllocationItem, error)
	CreateTransaction(ctx context.Context, userID, entryID uuid.UUID, txnType TransactionType, quantity money.Decimal, price money.Money, currency string, fees money.Money, transactionDate time.Time, notes string) (Transaction, error)
	UpdateTransaction(ctx context.Context, userID, txnID uuid.UUID, txnType TransactionType, quantity money.Decimal, price money.Money, currency string, fees money.Money, transactionDate time.Time, notes string) (Transaction, error)
	ImportEntryTransactions(ctx context.Context, userID, portfolioID, sourceID uuid.UUID, rows []ImportTransactionRow) (int, error)

	// Snapshots & growth
	GetAllPortfolioSummaryRows(ctx context.Context) ([]SnapshotRow, error)
	UpsertPortfolioSnapshot(ctx context.Context, portfolioID uuid.UUID, snapshotDate time.Time, totalValue, currency, totalGainLoss, totalGainLossPct string) error
	GetPortfolioGrowthByUserID(ctx context.Context, userID uuid.UUID, hasSince bool, since time.Time) ([]GrowthPoint, error)
	GetPortfolioGrowthByPortfolioID(ctx context.Context, userID, portfolioID uuid.UUID, hasSince bool, since time.Time) ([]GrowthPoint, error)

	// Exchange rates (read-only lookup for display-currency conversion; the
	// exchange-rate domain itself stays in the legacy market area until Fase 7)
	GetExchangeRateByPair(ctx context.Context, from, to string) (money.Decimal, error)
}

// Ensure the concrete repository keeps satisfying the interface.
var _ Repository = (*PostgresRepository)(nil)
